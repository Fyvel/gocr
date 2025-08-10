package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
)

func main() {
	// This is a deprecated wrapper - use cmd/ocr-tool/main.go instead
	fmt.Println("Note: Please use the new CLI tool at cmd/ocr-tool/")
	fmt.Println("Running: go run cmd/ocr-tool/main.go cmd/ocr-tool/cli.go")

	cmd := exec.Command("go", append([]string{"run", "cmd/ocr-tool/main.go", "cmd/ocr-tool/cli.go"}, os.Args[1:]...)...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Run(); err != nil {
		log.Fatal("Failed to run new CLI:", err)
	}
}
