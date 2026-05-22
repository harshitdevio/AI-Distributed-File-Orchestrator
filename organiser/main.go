package main

import (
	"fmt"
	"log"
	"organiser/classifier"
	"organiser/mover"
	"organiser/network"
	"organiser/scanner"
)

func main() {
	dir, err := scanner.Scan()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	files, err := classifier.ProcessDirectory(dir)
	if err != nil {
		log.Fatalf("Something went wrong: %v", err)
	}

	fmt.Printf("Found %d files:\n", len(files))
	for _, file := range files {
		fmt.Printf("File: %s | Type: %s\n", file.FilePath, file.MIMEType)
	}

	fmt.Printf("Sending %d files to Python LLM...\n", len(files))

	results, err := network.DispatchClassificationRequest(files)
	if err != nil {
		log.Fatalf("Dispatch failed: %v", err)
	}

	fmt.Println("Transfer complete. Waiting for Python to finish and send results back...")

	for _, result := range results {
		mover.ProcessMove(result.Filepath, result.TopTopic, result.Confidence)
	}

	fmt.Println("Pipeline processing complete. Shutting down system execution cleanly.")
}