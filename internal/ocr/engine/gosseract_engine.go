package engine

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/otiai10/gosseract/v2"
)

type GosseractEngine struct {
	client *gosseract.Client
}

func NewGosseractEngine() (*GosseractEngine, error) {
	return &GosseractEngine{}, nil
}

func (g *GosseractEngine) ProcessImage(imagePath string) (json.RawMessage, error) {
	client := gosseract.NewClient()
	defer client.Close()
	client.SetConfigFile("digits")
	client.SetPageSegMode(gosseract.PSM_AUTO)
	client.SetVariable("tessedit_char_whitelist", "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz@+-.() ")

	client.SetImage(imagePath)
	text, err := client.Text()
	if err != nil {
		return nil, fmt.Errorf("failed to extract text from image %s: %w", imagePath, err)
	}
	jsonBytes, err := textToJSON(text)
	if err != nil {
		log.Printf("Failed to convert text to JSON: %v\n", err)
		return json.RawMessage{}, nil
	}
	return jsonBytes, nil
}

func (g *GosseractEngine) Close() error {
	if g.client != nil {
		return g.client.Close()
	}
	return nil
}

func textToJSON(text string) (json.RawMessage, error) {
	cleanText := strings.TrimSpace(text)
	cleanText = strings.ReplaceAll(cleanText, "\n", " ")
	cleanText = strings.ReplaceAll(cleanText, "\r", " ")
	cleanText = strings.ReplaceAll(cleanText, "\t", " ")
	cleanText = strings.ReplaceAll(cleanText, "  ", " ")
	data := map[string]string{"text": cleanText}
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal text to JSON: %w", err)
	}
	return json.RawMessage(jsonBytes), nil
}
