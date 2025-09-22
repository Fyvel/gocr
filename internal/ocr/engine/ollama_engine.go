package engine

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

type OllamaEngine struct {
	baseURL string
	model   string
	client  *http.Client
}

type OllamaRequest struct {
	Model  string   `json:"model"`
	Prompt string   `json:"prompt"`
	Images []string `json:"images"`
	Stream bool     `json:"stream"`
}

type OllamaResponse struct {
	Response json.RawMessage `json:"response"`
	Done     bool            `json:"done"`
}

const (
	defaultBaseURL = "http://localhost:11434"
	defaultModel   = "llama3.2-vision"
)

func NewOllamaEngine(baseURL, model string) *OllamaEngine {
	if baseURL == "" {
		baseURL = defaultBaseURL
	}
	if model == "" {
		model = defaultModel
	}

	return &OllamaEngine{
		baseURL: baseURL,
		model:   model,
		client:  &http.Client{},
	}
}

func (o *OllamaEngine) ProcessImage(imagePath string) (json.RawMessage, error) {
	imageData, err := os.ReadFile(imagePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read image: %w", err)
	}

	encodedImage := base64.StdEncoding.EncodeToString(imageData)

	request := OllamaRequest{
		Model: o.model,
		Prompt: `
You are an OCR helper.  
The image contains the following fields:

• Name  
• Email  
• Phone  
• Tags

Your job:

1. Extract the text for each field.  
2. For *Tags*, capture every label.  
3. If the OCR can't see any tags at all set Tags to '["MISS"]''.  
4. Return **only** a JSON object with this exact schema:

{
  "Name": "<value or empty string>",
  "Email": "<value or empty string>",
  "Phone": "<value or empty string>",
  "Tags": ["<tag1>", "<tag2>", ...]   // defaults to ["MISS"] if none detected
}

* Do not add any other text, explanations, or formatting.  
* If a field is missing or unreadable, use an empty string (or default array for Tags).  
* Make sure the JSON is syntactically correct – double quotes, no trailing commas, no comments.
				`,
		Images: []string{encodedImage},
		Stream: false,
	}

	jsonData, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}
	// fmt.Printf("Sending request to Ollama: %s\n", string(jsonData))

	resp, err := o.client.Post(o.baseURL+"/api/generate", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ollama request failed with status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var ollamaResp OllamaResponse
	if err := json.Unmarshal(body, &ollamaResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	jsonObj, err := extractJSON(string(ollamaResp.Response))
	if err != nil {
		return nil, fmt.Errorf("failed to extract JSON from response: %w", err)
	}

	return jsonObj, nil
}

func (o *OllamaEngine) Close() error {
	return nil
}

func extractJSON(input string) (json.RawMessage, error) {
	log.Printf("Extracting JSON from input: %s\n", input)
	normalized := bytes.ReplaceAll([]byte(input), []byte("\\n"), []byte(""))
	normalized = bytes.ReplaceAll(normalized, []byte("\\r"), []byte(""))
	normalized = bytes.ReplaceAll(normalized, []byte("\\t"), []byte(""))
	normalized = bytes.ReplaceAll(normalized, []byte("\\"), []byte(""))
	normalized = bytes.ReplaceAll(normalized, []byte("  "), []byte(""))
	text := string(normalized)
	// Find opening brace
	start := -1
	for i, char := range text {
		if char == '{' {
			start = i
			break
		}
	}

	if start == -1 {
		return nil, fmt.Errorf("no JSON found in text")
	}

	// Track brace depth to find the matching closing brace
	braceCount := 0
	end := -1

matchingBrace:
	for i := start; i < len(text); i++ {
		switch text[i] {
		case '{':
			braceCount++
		case '}':
			braceCount--
			if braceCount == 0 {
				end = i + 1
				break matchingBrace
			}
		}
	}

	if end == -1 {
		return nil, fmt.Errorf("no matching closing brace found")
	}

	jsonStr := text[start:end]

	// Validate that it's actually valid JSON
	var temp any
	if err := json.Unmarshal([]byte(jsonStr), &temp); err != nil {
		return nil, fmt.Errorf("extracted text is not valid JSON: %w", err)
	}

	return json.RawMessage(jsonStr), nil
}
