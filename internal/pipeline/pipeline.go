package pipeline

import (
	"context"
	"ocr-tool/internal/data"
	"ocr-tool/internal/image"
	"ocr-tool/internal/logger"
	"ocr-tool/internal/ocr"
	"ocr-tool/internal/writer"
	"sync"
)

type result[T any] struct {
	path string
	data T
	err  error
}

type writeResult[T any] struct {
	mu       sync.Mutex
	writes   map[string]T
	failures map[string]error
}

type Clients struct {
	engine ocr.OCREngine
	image  image.ImageProcessor
	data   data.DataExtractor
	writer *writer.CSVWriter[data.ExtractedData]
}

type contextKey string

const clientsKey contextKey = "all_my_clients"

const outputFileKey contextKey = "output_file"

func Run(engineType string, directory string, outputFile string) (writes map[string]data.ExtractedData, failures map[string]error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	logger.DebugLog("Pipeline started with engineType=%s, directory=%s, output=%s", engineType, directory, outputFile)

	ocrEngine, err := ocr.NewEngine(engineType)
	if err != nil {
		logger.DebugLog("Failed to create OCR engine: %v", err)
		return nil, map[string]error{"engine": err}
	}
	defer func() {
		logger.DebugLog("Closing OCR engine")
		ocrEngine.Close()
	}()

	clients := &Clients{
		engine: ocrEngine,
		image:  *image.NewImageProcessor(),
		data:   *data.NewDataExtractor(),
		writer: writer.NewCSVWriter(data.MapCSVRecord, data.GetCSVHeader),
	}

	// Embed clients in context
	ctx = context.WithValue(ctx, clientsKey, clients)
	ctx = context.WithValue(ctx, outputFileKey, outputFile)

	errChan := make(chan error, 10)                          // Buffered channel to collect errors
	files := make(chan string)                               // Unbuffered channel for file paths
	enhancedChan := make(chan string, 2)                     // Bounded buffer to throttle image preprocessing
	ocrChan := make(chan ocr.OCRResult)                      // Buffered channel for OCR results from enhanced images
	extractChan := make(chan result[data.ExtractedData], 10) // Buffered channel for extraction tasks (OCR results to extracted data)
	results := &writeResult[data.ExtractedData]{
		writes:   make(map[string]data.ExtractedData), // Map to store extracted data
		failures: make(map[string]error),              // Map to store failures
	}

	go func() {
		defer close(files)
		logger.DebugLog("Starting [walkFiles] goroutine")
		walkFiles(ctx, directory, files, errChan)
		defer logger.DebugLog("[walkFiles] goroutine finished")
	}()

	go func() {
		defer close(enhancedChan)
		logger.DebugLog("Starting [enhanceImage] goroutine")
		enhanceImage(ctx, files, enhancedChan, errChan)
		defer logger.DebugLog("[enhanceImage] goroutine finished")
	}()

	processCount := 2 // hard limit on OCR workers for now
	var wg sync.WaitGroup

	for i := 0; i < processCount; i++ {
		wg.Add(1)
		go func(worker int) {
			defer wg.Done()
			logger.DebugLog("Starting [performOcr] worker #%d", worker+1)
			performOcr(ctx, enhancedChan, ocrChan, errChan)
			defer logger.DebugLog("[performOcr] worker #%d finished", worker+1)
		}(i)
	}
	go func() {
		wg.Wait()
		logger.DebugLog("All [performOcr] workers finished, closing ocrChan")
		close(ocrChan)
	}()

	// fan-out - forward ocr results to extraction + cleanup
	extractInput := make(chan ocr.OCRResult)
	cleanupInput := make(chan ocr.OCRResult, 10) // Buffered channel for cleanup files created during enhancement

	go func() {
		logger.DebugLog("Starting [forwardChan] for ocrChan -> extractInput, cleanupInput")
		forwardChan(ctx, ocrChan, extractInput, cleanupInput)
		defer logger.DebugLog("[forwardChan] for ocrChan finished")
	}()

	go func() {
		defer close(extractChan)
		logger.DebugLog("Starting [extractData] goroutine")
		extractData(ctx, extractInput, extractChan, errChan)
		defer logger.DebugLog("[extractData] goroutine finished")
	}()

	go func() {
		logger.DebugLog("Starting [cleanupImage] goroutine")
		cleanupImage(ctx, cleanupInput, errChan)
		defer logger.DebugLog("[cleanupImage] goroutine finished")
	}()

	var writeWg sync.WaitGroup
	writeWg.Add(1)
	go func() {
		defer writeWg.Done()
		logger.DebugLog("Starting [writeOutput] goroutine")
		writeOutput(ctx, extractChan, results, errChan)
		defer logger.DebugLog("[writeOutput] goroutine finished")
	}()

	writeWg.Wait()
	logger.DebugLog("All [writeOutput] finished, closing errChan")
	close(errChan)
	for err := range errChan {
		if err != nil {
			logger.DebugLog("Error received in errChan: %v", err)
			results.addFailure("pipeline_error", err)
		}
	}

	logger.DebugLog("Pipeline finished")
	return results.writes, results.failures
}

func forwardChan[T any](ctx context.Context, in <-chan T, outs ...chan<- T) {
	defer func() {
		for _, out := range outs {
			close(out)
		}
	}()

	for res := range in {
		if ctx.Err() != nil {
			logger.DebugLog("forwardChan: context cancelled")
			return
		}

		logger.DebugLog("forwardChan: forwarding result to %d outputs", len(outs))
		for _, out := range outs {
			select {
			case out <- res:
			case <-ctx.Done():
				logger.DebugLog("forwardChan: context done while forwarding")
				return
			}
		}
	}
}
