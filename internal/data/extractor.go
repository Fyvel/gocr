package data

import (
	"encoding/json"
	"regexp"
	"strings"
)

type ExtractedData struct {
	Filename string   `json:"Filename,omitempty"`
	Name     string   `json:"Name,omitempty"`
	Email    string   `json:"Email,omitempty"`
	Phone    string   `json:"Phone,omitempty"`
	Tags     []string `json:"Tags,omitempty"`
	Text     string   `json:"Text,omitempty"`
}

type DataExtractor struct{}

var (
	emailRegex = regexp.MustCompile(`(?i)[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}`)
	phoneRegex = regexp.MustCompile(`\+?[0-9]{10,15}`)
)

func NewDataExtractor() *DataExtractor {
	return &DataExtractor{}
}

func (de *DataExtractor) ExtractFromJson(data json.RawMessage, filename string) *ExtractedData {
	dataStr := string(data)

	// log.Printf("Extracting data from JSON: %s\n", dataStr)
	extractedData := ExtractedData{
		Filename: filename,
		Name:     "",
		Email:    "",
		Phone:    "",
		Tags:     []string{},
		Text:     "",
	}
	if err := json.Unmarshal([]byte(dataStr), &extractedData); err != nil {
		return &ExtractedData{
			Filename: filename,
		}
	}

	// log.Printf("Extracted data from %s: %+v\n", filename, extractedData)

	if extractedData.Text != "" {
		return &ExtractedData{
			Filename: filename,
			Name:     extractedData.Name,
			Email:    de.extractEmail(extractedData.Text),
			Phone:    de.extractPhone(extractedData.Text),
			Tags:     de.extractTags(extractedData.Text),
			Text:     extractedData.Text,
		}
	}

	return &ExtractedData{
		Filename: filename,
		Name:     extractedData.Name,
		Email:    de.extractEmail(extractedData.Email),
		Phone:    de.extractPhone(extractedData.Phone),
		Tags:     extractedData.Tags,
	}
}

func (de *DataExtractor) extractEmail(text string) string {
	emails := emailRegex.FindAllString(text, -1)
	if len(emails) == 0 {
		// log.Printf("No emails found in text: %s\n", text)
		return ""
	}
	if len(emails) > 1 {
		return strings.Join(emails, "; ")
	}
	return emails[0]
}

func (de *DataExtractor) extractPhone(text string) string {
	// log.Printf("Extracting phone from text: %s\n", text)
	// First try with original text
	phones := phoneRegex.FindAllString(text, -1)
	if len(phones) == 0 {
		// log.Printf("No phones found in original text: %s\n", text)
		return ""
	}

	// Then try with cleaned text for phone-like patterns
	cleaned := de.cleanOCRText(text)
	cleanedPhones := phoneRegex.FindAllString(cleaned, -1)

	// Combine and deduplicate
	allPhones := append(phones, cleanedPhones...)
	unique := make(map[string]bool)
	var result []string

	for _, phone := range allPhones {
		cleanPhone := de.cleanOCRText(phone)
		// Fix common OCR mistake: '4' at start should be '+'
		if len(cleanPhone) >= 11 && cleanPhone[0] == '4' &&
			(strings.HasPrefix(cleanPhone, "441") || strings.HasPrefix(cleanPhone, "447") || strings.HasPrefix(cleanPhone, "449")) {
			cleanPhone = "+" + cleanPhone[1:]
		}

		if !unique[cleanPhone] && len(cleanPhone) >= 10 {
			unique[cleanPhone] = true
			result = append(result, cleanPhone)
		}
	}

	if len(result) > 1 {
		return strings.Join(result, "; ")
	}
	return result[0]
}

func (de *DataExtractor) extractTags(text string) []string {
	// TODO
	return []string{}
}

func (de *DataExtractor) cleanOCRText(text string) string {
	replacements := map[string]string{
		"S": "5",
		"O": "0",
		"I": "1",
		"l": "1",
		"B": "8",
		"G": "6",
		"Z": "2",
	}

	cleaned := text
	for wrong, right := range replacements {
		cleaned = strings.ReplaceAll(cleaned, wrong, right)
	}
	return cleaned
}

func MapCSVRecord(item ExtractedData) []string {
	return []string{
		item.Filename,
		item.Name,
		item.Email,
		item.Phone,
		strings.Join(item.Tags, "; "),
		item.Text,
	}
}

func GetCSVHeader() []string {
	return []string{"Filename", "Name", "Email", "Phone", "Tags", "Text"}
}
