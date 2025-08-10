package pipeline

import (
	"context"
	"fmt"
	"ocr-tool/internal/logger"
	"os"
	"path/filepath"
	"strings"
)

func walkFiles(ctx context.Context, directory string, results chan<- string, errChan chan<- error) {
	files, err := os.ReadDir(directory)
	if err != nil {
		logger.DebugLog("[walkFiles]: failed to read directory %s: %v", directory, err)
		errChan <- fmt.Errorf("[walkFiles]: reading directory %s: %w", directory, err)
		return
	}

	for _, file := range files {
		if ctx.Err() != nil {
			logger.DebugLog("[walkFiles]: context cancelled")
			return
		}

		fileName := file.Name()
		if file.IsDir() || isProcessedFile(fileName) || !isImageFile(fileName) {
			continue
		}
		fullPath := filepath.Join(directory, fileName)
		logger.DebugLog("[walkFiles]: sending file %s", fullPath)
		select {
		case results <- fullPath:
		case <-ctx.Done():
			logger.DebugLog("[walkFiles]: context done while sending file %s", fullPath)
			return
		}
	}
}

func isProcessedFile(filename string) bool {
	return strings.Contains(filename, "_processed")
}

func isImageFile(filename string) bool {
	ext := strings.ToLower(filename[strings.LastIndex(filename, ".")+1:])
	return ext == "jpg" || ext == "jpeg" || ext == "png" || ext == "tiff" || ext == "bmp"
}
