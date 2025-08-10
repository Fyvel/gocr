package pipeline

import (
	"context"
	"fmt"
	"ocr-tool/internal/data"
	"ocr-tool/internal/logger"
	"ocr-tool/internal/ocr"
)

func extractData(ctx context.Context, ocrChan <-chan ocr.OCRResult, results chan<- result[data.ExtractedData], errChan chan<- error) {
	ctxClients := ctx.Value(clientsKey)
	proc, ok := ctxClients.(*Clients)
	if !ok {
		logger.DebugLog("extractData: missing clients in context")
		errChan <- fmt.Errorf("extractData: missing clients in context")
		return
	}
	dataExtractor := proc.data

	for ocrOutput := range ocrChan {
		if ctx.Err() != nil {
			logger.DebugLog("extractData: context cancelled")
			return
		}

		if ocrOutput.Error != nil {
			logger.DebugLog("extractData: OCR error for %s: %v", ocrOutput.Filename, ocrOutput.Error)
			results <- result[data.ExtractedData]{path: ocrOutput.Filename, err: ocrOutput.Error}
			continue
		}

		logger.DebugLog("extractData: extracting data from %s", ocrOutput.Filename)
		res := dataExtractor.ExtractFromJson(ocrOutput.Json, ocrOutput.Filename)
		if res == nil {
			logger.DebugLog("extractData: extraction returned nil for %s", ocrOutput.Filename)
			results <- result[data.ExtractedData]{path: ocrOutput.Filename, err: fmt.Errorf("extraction returned nil for %s", ocrOutput.Filename)}
			continue
		}
		logger.DebugLog("extractData: sending extracted data for %s", ocrOutput.Filename)
		results <- result[data.ExtractedData]{path: ocrOutput.Filename, data: *res}
	}
}
