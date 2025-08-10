package pipeline

import (
	"context"
	"fmt"
	"ocr-tool/internal/logger"
	"ocr-tool/internal/ocr"
)

func enhanceImage(ctx context.Context, files <-chan string, results chan<- string, errChan chan<- error) {
	ctxClients := ctx.Value(clientsKey)
	proc, ok := ctxClients.(*Clients)
	if !ok {
		logger.DebugLog("[enhanceImage]: missing clients in context")
		errChan <- fmt.Errorf("[enhanceImage]: missing clients in context")
		return
	}
	imageProcessor := proc.image

	for file := range files {
		if ctx.Err() != nil {
			logger.DebugLog("[enhanceImage]: context cancelled")
			return
		}

		logger.DebugLog("[enhanceImage]: enhancing file %s", file)
		processed, err := imageProcessor.EnhanceQuality(file)
		if err != nil {
			logger.DebugLog("[enhanceImage]: error processing %s: %v", file, err)
			errChan <- fmt.Errorf("preprocessing image %s: %w", file, err)
			continue
		}

		logger.DebugLog("[enhanceImage]: sending processed file %s", processed)
		select {
		case results <- processed:
		case <-ctx.Done():
			logger.DebugLog("[enhanceImage]: context done while sending %s", processed)
			return
		}
	}
}

func cleanupImage(ctx context.Context, ocrChan <-chan ocr.OCRResult, errChan chan<- error) {
	ctxClients := ctx.Value(clientsKey)
	proc, ok := ctxClients.(*Clients)
	if !ok {
		logger.DebugLog("[cleanupImage]: missing clients in context")
		errChan <- fmt.Errorf("[cleanupImage]: missing clients in context")
		return
	}
	imageProcessor := proc.image

	for ocrOutput := range ocrChan {
		if ctx.Err() != nil {
			logger.DebugLog("[cleanupImage]: context cancelled")
			return
		}

		if ocrOutput.Error != nil {
			logger.DebugLog("[cleanupImage]: skipping file %s due to OCR error", ocrOutput.Filename)
			continue
		}

		logger.DebugLog("[cleanupImage]: cleaning up %s", ocrOutput.Filename)
		if err := imageProcessor.Cleanup(ocrOutput.Filename); err != nil {
			logger.DebugLog("[cleanupImage]: error cleaning up %s: %v", ocrOutput.Filename, err)
			errChan <- fmt.Errorf("cleanup image %s: %w", ocrOutput.Filename, err)
		}
	}
}
