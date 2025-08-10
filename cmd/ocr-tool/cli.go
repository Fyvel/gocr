package main

import (
	"flag"
	"fmt"
	"ocr-tool/internal/pipeline"
)

type CLI struct {
	imagesDir  string
	outputDir  string
	engineType string
	outputFile string
}

func NewCLI() *CLI {
	return &CLI{
		imagesDir:  "images",
		outputDir:  "output",
		engineType: "gosseract",
	}
}

func (c *CLI) Run(args []string) error {
	fs := flag.NewFlagSet("ocr-tool", flag.ExitOnError)

	fs.StringVar(&c.imagesDir, "images", c.imagesDir, "Directory containing images to process")
	fs.StringVar(&c.outputDir, "output", c.outputDir, "Output directory for results")
	fs.StringVar(&c.engineType, "engine", c.engineType, "OCR engine type (ollama, gosseract)")

	if err := fs.Parse(args); err != nil {
		return fmt.Errorf("parsing flags: %w", err)
	}

	// Set output file based on engine type
	c.outputFile = fmt.Sprintf("%s/%s_extracted_data.csv", c.outputDir, c.engineType)

	return c.process_new()
}

func (c *CLI) process_new() error {
	results, errors := pipeline.Run(c.engineType, c.imagesDir, c.outputFile)
	for path, err := range errors {
		fmt.Printf("Error processing %s: %v\n", path, err)
	}
	for path, data := range results {
		fmt.Printf("Processed %s: %v\n", path, data)
	}
	fmt.Printf("\nProcessing complete! Results saved to: %s\n", c.outputFile)
	fmt.Printf("Processed %d records\n", len(results)+len(errors))
	return nil
}
