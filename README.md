# OCR Tool - Extract Data from Images

This tool extracts text from images using OCR.

## Pipeline Overview

1. Discover files
   - Goroutine: [walkFiles]
   - Channel: `files` (unbuffered)
2. Preprocess images (enhance)
   - Goroutine: [enhanceImage]
   - In: `files`
   - Out: `enhancedChan` (buffered, size 2 — bounded throttle)
3. Perform OCR (parallel workers)
   - Goroutines: [performOcr] (N=2)
   - In: `enhancedChan`
   - Out: `ocrChan` (unbuffered)
4. Fan-out to:
   - Extraction
     - Channel path: `ocrChan` -> `extractInput` (unbuffered) -> `extractChan` (buffered, size 10)
     - Goroutine: [extractData]
   - Cleanup
     - Channel path: `ocrChan` -> `cleanupInput` (buffered, size 10)
     - Goroutine: [cleanupImage]
5. Write output
   - Goroutine: [writeOutput]
   - In: `extractChan`
   - Shared result map guarded by mutex (`writeResult`)

# Features

✅ Concurrent processing of images  
✅ Configurable OCR engine (Tesseract or Ollama **AI Powered btw**)  
✅ Text postprocessing (cleanup, formatting, etc.)  
✅ Output options (JSON or CSV)

# How to

## 0. Create a [Go](https://golang.org/downloads/) Project

```bash
go mod init ocr-tool
```

## 1. Install Dependencies

### Tesseract OCR Installation

**Windows:**
Download from [UB-Mannheim Tesseract](https://github.com/UB-Mannheim/tesseract/wiki)

**macOS:**

```bash
brew install tesseract
```

**Ubuntu/Debian:**

```bash
sudo apt-get update
sudo apt-get install tesseract-ocr
```

**Test it works**

```bash
tesseract ./images/sample.png stdout
```

---

### Ollama setup for OCR

Install [Ollama](https://ollama.com/)  
Pull llama3.2-vision latest model  
And that _should_ be it!

### Go Dependencies

**Gosseract** (Go client for Tesseract OCR)

```bash
go get -t github.com/otiai10/gosseract/v2
export LIBRARY_PATH="/opt/homebrew/lib"
export CPATH="/opt/homebrew/include"
```

**Imaging** (Image processing library)

```bash
go get -t github.com/disintegration/imaging
```

## 2. Run the OCR Tool

```bash
# See options
go run main.go --help

# Example
go run main.go --input ./images --output ./output --engine gosseract
```

## 3. Run Tests

```bash
# Run all tests in the project
go test -v ./...

# Run tests with coverage
go test -v ./... -coverprofile=coverage/coverage.out && go tool cover -html=coverage/coverage.out -o coverage/coverage.html
```
