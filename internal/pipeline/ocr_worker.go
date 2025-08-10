package pipeline

import (
	"context"
	"fmt"
	"ocr-tool/internal/logger"
	"ocr-tool/internal/ocr"
)

func performOcr(ctx context.Context, preprocessChan <-chan string, ocrChan chan<- ocr.OCRResult, errChan chan<- error) {
	ctxClients := ctx.Value(clientsKey)
	proc, ok := ctxClients.(*Clients)
	if !ok {
		logger.DebugLog("[performOcr]: missing clients in context")
		errChan <- fmt.Errorf("[performOcr]: missing clients in context")
		return
	}
	ocrEngine := proc.engine

	for imagePath := range preprocessChan {
		if ctx.Err() != nil {
			logger.DebugLog("[performOcr]: context cancelled")
			return
		}

		logger.DebugLog("[performOcr]: processing image %s", imagePath)
		data, err := ocrEngine.ProcessImage(imagePath)

		logger.DebugLog("[performOcr]: sending OCR result - %s (err=%v)", data, err)
		select {
		case ocrChan <- ocr.OCRResult{Json: data, Filename: imagePath, Error: err}:
		case <-ctx.Done():
			logger.DebugLog("[performOcr]: context done while sending OCR result for %s", imagePath)
			return
		}
	}
}
