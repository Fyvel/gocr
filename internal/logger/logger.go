package logger

import (
	"log"
	"os"
)

func DebugLog(format string, args ...any) {
	if os.Getenv("DEBUG") == "1" {
		log.Printf("[DEBUG] "+format, args...)
	}
}
