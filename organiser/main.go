package main

import (
	"fmt"
	"log"
	"net/http"
	"organiser/classifier"
	"organiser/mover"
	"organiser/network"
	"organiser/scanner"
	"organiser/server"
)

// FileMoverAdapter implements server.FileMover interface by delegating to mover package.
// It acts as a bridge between the HTTP server and the file organization logic.
type FileMoverAdapter struct{}

// MoveFiles processes classification results and moves files to appropriate directories.
// It iterates through each result and calls mover.ProcessMove to organize the file
// based on its topic classification and confidence score.
//
// Parameters:
//   - results: Slice of ClassificationResult containing filepath, topic, and confidence
//
// Returns:
//   - error: Always returns nil in current implementation, but interface allows for error handling
func (fma *FileMoverAdapter) MoveFiles(results []server.ClassificationResult) error {
	for _, result := range results {
		mover.ProcessMove(result.Filepath, result.TopTopic, result.Confidence)
	}
	return nil
}

// main is the entry point of the file organization application.
// It coordinates three main operations:
//  1. Starts an HTTP server to receive classification results from the Python LLM service
//  2. Scans a directory for files and classifies them by MIME type
//  3. Sends files to the Python LLM service for topic detection
//
// The application blocks indefinitely waiting for classification results via HTTP callbacks.
func main() {
	// Start HTTP server in background goroutine to receive classification results
	// from Python LLM service. The server listens on port 8080 and accepts POST
	// requests at /api/ingest endpoint with classification results.
	go func() {
		fileMover := &FileMoverAdapter{}
		http.HandleFunc("/api/ingest", server.IngestHandler(fileMover))
		fmt.Println("[Server] Listening for results on :8080...")
		if err := http.ListenAndServe(":8080", nil); err != nil {
			log.Fatalf("Receiver server failed: %v", err)
		}
	}()

	// Scan for directory path from user input or command-line flag.
	// Prompts user if no path is provided via -path flag.
	dir, err := scanner.Scan()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// Process the directory to discover all files and detect their MIME types.
	// Recursively walks through subdirectories and catalogs each file.
	files, err := classifier.ProcessDirectory(dir)
	if err != nil {
		log.Fatalf("Something went wrong: %v", err)
	}

	// Display summary of discovered files
	fmt.Printf("Found %d files:\n", len(files))
	for _, file := range files {
		fmt.Printf("File: %s | Type: %s\n", file.FilePath, file.MIMEType)
	}

	// Send file list to Python LLM service for topic classification.
	// The Python service analyzes file content and determines appropriate topics.
	apiURL := "http://127.0.0.1:8000/detect-topics-batch"
	fmt.Printf("Sending %d files to Python LLM...\n", len(files))

	if err := network.SendToService(apiURL, files); err != nil {
		log.Fatalf("Dispatch failed: %v", err)
	}

	fmt.Println("Transfer complete. Waiting for Python to finish and send results back...")

	// Block indefinitely to keep the HTTP server running.
	// Classification results will arrive asynchronously via HTTP callbacks
	// to the /api/ingest endpoint, where FileMoverAdapter will organize files.
	select {}
}