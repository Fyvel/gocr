package ocr

import "encoding/json"

type OCRResult struct {
	Json     json.RawMessage
	Filename string
	Error    error
}

type OCREngine interface {
	ProcessImage(imagePath string) (json.RawMessage, error)
	Close() error
}
