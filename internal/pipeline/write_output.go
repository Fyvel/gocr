package pipeline

import (
	"context"
	"fmt"
	"ocr-tool/internal/data"
	"ocr-tool/internal/logger"
)

func writeOutput(ctx context.Context,
	extractedChan <-chan result[data.ExtractedData],
	results *writeResult[data.ExtractedData],
	errChan chan<- error) {
	ctxClients := ctx.Value(clientsKey)
	proc, ok := ctxClients.(*Clients)
	if !ok {
		logger.DebugLog("[writeOutput]: missing clients in context")
		errChan <- fmt.Errorf("[writeOutput]: missing clients in context")
		return
	}
	writer := proc.writer

	output := ctx.Value(outputFileKey).(string)

	for res := range extractedChan {
		if ctx.Err() != nil {
			logger.DebugLog("[writeOutput]: context cancelled")
			return
		}

		if res.err != nil {
			logger.DebugLog("[writeOutput]: failure for %s: %v", res.path, res.err)
			results.addFailure(res.path, res.err)
			continue
		}

		logger.DebugLog("[writeOutput]: writing data for %s", res.path)
		if err := writer.WriteToFile([]data.ExtractedData{res.data}, output); err != nil {
			logger.DebugLog("[writeOutput]: error writing to file %s: %v", output, err)
			results.addFailure(res.path, fmt.Errorf("writing to file %s: %w", output, err))
			continue
		}

		logger.DebugLog("[writeOutput]: successfully wrote data for %s", res.path)
		results.addWrite(res.path, res.data)
	}

	logger.DebugLog("[writeOutput]: closing CSV writer")
	writer.Close()
	logger.DebugLog("[writeOutput]: CSV writer closed")
}

func (r *writeResult[T]) addWrite(path string, data T) {
	r.mu.Lock()
	r.writes[path] = data
	r.mu.Unlock()
}

func (r *writeResult[T]) addFailure(path string, err error) {
	r.mu.Lock()
	r.failures[path] = err
	r.mu.Unlock()
}
