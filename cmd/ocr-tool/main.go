package main

import (
	"log"
	"os"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	cli := NewCLI()
	if err := cli.Run(os.Args[1:]); err != nil {
		log.Fatal("Error:", err)
	}
}
