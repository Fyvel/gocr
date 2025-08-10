package ocr

import (
	"fmt"
	"ocr-tool/internal/ocr/engine"
)

func NewEngine(engineType string) (OCREngine, error) {
	var e OCREngine
	var err error

	switch engineType {
	case "ollama":
		e = engine.NewOllamaEngine("", "")
	case "gosseract", "":
		e, err = engine.NewGosseractEngine()
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unknown engine type: %s", engineType)
	}

	return e, nil
}
