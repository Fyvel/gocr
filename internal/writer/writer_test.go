package writer

import (
	"encoding/csv"
	"fmt"
	"ocr-tool/internal/data"
	"os"
	"path/filepath"
	"sync"
	"testing"
)

func TestCSVWriter_AppendMode(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	outputPath := filepath.Join(tempDir, "append_test.csv")
	writer := NewCSVWriter(data.MapCSVRecord, data.GetCSVHeader)
	defer writer.Close()

	data1 := []data.ExtractedData{
		{
			Filename: "test1.jpg",
			Name:     "John Doe",
			Email:    "john@example.com",
			Phone:    "123-456-7890",
			Tags:     []string{"tag1", "tag2"},
			Text:     "Test text 1",
		},
	}

	data2 := []data.ExtractedData{
		{
			Filename: "test2.jpg",
			Name:     "Jane Smith",
			Email:    "jane@example.com",
			Phone:    "098-765-4321",
			Tags:     []string{"tag3"},
			Text:     "Test text 2",
		},
	}

	expectedHeader := []string{"Filename", "Name", "Email", "Phone", "Tags", "Text"}
	expectedRecords := 3 // header + 2 data rows

	// Act
	err1 := writer.WriteToFile(data1, outputPath)
	err2 := writer.WriteToFile(data2, outputPath)

	// Assert
	if err1 != nil {
		t.Fatalf("First write failed: %v", err1)
	}
	if err2 != nil {
		t.Fatalf("Second write failed: %v", err2)
	}

	records := readCSVFile(t, outputPath)
	if len(records) != expectedRecords {
		t.Errorf("expected %d records (header + data), got %d", expectedRecords, len(records))
	}

	if !stringSlicesEqual(records[0], expectedHeader) {
		t.Errorf("expected header %v, got %v", expectedHeader, records[0])
	}

	if records[1][0] != "test1.jpg" || records[2][0] != "test2.jpg" {
		t.Errorf("data integrity check failed")
	}
}

func TestCSVWriter_ReplaceMode(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	outputPath := filepath.Join(tempDir, "replace_test.csv")
	writer := NewCSVWriter(data.MapCSVRecord, data.GetCSVHeader)
	defer writer.Close()

	data1 := []data.ExtractedData{
		{Filename: "original.jpg", Name: "Original User"},
	}
	data2 := []data.ExtractedData{
		{Filename: "replaced.jpg", Name: "Replaced User"},
	}

	// Act
	err1 := writer.WriteToFile(data1, outputPath)
	err2 := writer.WriteToFile(data2, outputPath, true) // overwrite = true

	// Assert
	if err1 != nil {
		t.Fatalf("First write failed: %v", err1)
	}
	if err2 != nil {
		t.Fatalf("Replace write failed: %v", err2)
	}

	records := readCSVFile(t, outputPath)
	expectedRecords := 2 // header + 1 data row
	if len(records) != expectedRecords {
		t.Errorf("expected %d records after replace, got %d", expectedRecords, len(records))
	}

	if records[1][0] != "replaced.jpg" {
		t.Errorf("expected replaced content, got %s", records[1][0])
	}
}

func TestCSVWriter_ConcurrentWrites(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	outputPath := filepath.Join(tempDir, "concurrent_test.csv")
	writer := NewCSVWriter(data.MapCSVRecord, data.GetCSVHeader)
	defer writer.Close()

	numGoroutines := 5
	expectedRecords := 1 + numGoroutines // header + data
	var wg sync.WaitGroup

	// Act
	for i := range numGoroutines {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			data := []data.ExtractedData{
				{
					Filename: fmt.Sprintf("concurrent_%d.jpg", id),
					Name:     fmt.Sprintf("User %d", id),
					Email:    fmt.Sprintf("user%d@example.com", id),
					Phone:    fmt.Sprintf("123-456-%04d", id),
					Tags:     []string{fmt.Sprintf("tag-%d", id)},
					Text:     fmt.Sprintf("Text from goroutine %d", id),
				},
			}
			if err := writer.WriteToFile(data, outputPath); err != nil {
				t.Errorf("Goroutine %d failed: %v", id, err)
			}
		}(i)
	}

	wg.Wait()

	// Assert
	records := readCSVFile(t, outputPath)
	if len(records) != expectedRecords {
		t.Errorf("expected %d records, got %d", expectedRecords, len(records))
	}
}

func TestCSVWriter_EmptyData(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	outputPath := filepath.Join(tempDir, "empty_test.csv")
	writer := NewCSVWriter(data.MapCSVRecord, data.GetCSVHeader)
	defer writer.Close()

	// Act
	err := writer.WriteToFile([]data.ExtractedData{}, outputPath)

	// Assert
	if err != nil {
		t.Fatalf("Writing empty data failed: %v", err)
	}

	if _, err := os.Stat(outputPath); err == nil {
		records := readCSVFile(t, outputPath)
		if len(records) > 0 {
			t.Errorf("expected no records for empty data, got %d", len(records))
		}
	}
}

func TestCSVWriter_InvalidPath(t *testing.T) {
	// Arrange
	writer := NewCSVWriter(data.MapCSVRecord, data.GetCSVHeader)
	defer writer.Close()

	invalidPath := "/root/invalid/path/test.csv"
	data := []data.ExtractedData{{Filename: "test.jpg", Name: "Test"}}

	// Act
	err := writer.WriteToFile(data, invalidPath)

	// Assert
	if err == nil {
		t.Errorf("expected error for invalid path, got none")
	}
}

// Helper functions
func readCSVFile(t *testing.T, path string) [][]string {
	file, err := os.Open(path)
	if err != nil {
		t.Fatalf("failed to open CSV file: %v", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		t.Fatalf("failed to read CSV: %v", err)
	}
	return records
}

func stringSlicesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}
