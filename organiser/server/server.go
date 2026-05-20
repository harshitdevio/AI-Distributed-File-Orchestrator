// Package server implements HTTP routing and handling for processing and 
// routing classified file data.
package server

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// ClassificationResult represents the classification payload for a single file.
// It maps directly to the incoming JSON structure from the classification service.
type ClassificationResult struct {
	// Filepath is the absolute path to the processed file.
	Filepath   string             `json:"filepath"`
	
	// TopTopic is the primary category determined by the classifier.
	TopTopic   string             `json:"top_topic"`
	
	// Confidence represents the scoring probability for the top topic.
	Confidence float64            `json:"confidence"`
	
	// AllScores contains the confidence breakdowns across all evaluated categories.
	AllScores  map[string]float64 `json:"all_scores"`
	
	// Error contains the error message if classification failed for this specific file.
	// This is nil if the file was processed successfully.
	Error      *string            `json:"error"`
}

// FileMover defines the contract for executing file relocation based on 
// classification results. It is decoupled as an interface to support unit testing 
// and alternative filesystem implementations.
type FileMover interface {
	MoveFiles(results []ClassificationResult) error
}

// IngestHandlerWithDeps processes an incoming classification payload and routes 
// files using the provided FileMover implementation. 
//
// This function exposes the dependency directly to allow test fixtures to inject 
// mocked FileMover implementations.
//
// Protocol Expectations:
//   - Request Method: MUST be POST. Returns 405 Method Not Allowed otherwise.
//   - Request Body: MUST be a valid JSON array of ClassificationResult structures. 
//     Returns 400 Bad Request on parsing failure.
//   - Downstream Failures: Returns 500 Internal Server Error if the body cannot 
//     be read or if the mover implementation fails.
func IngestHandlerWithDeps(w http.ResponseWriter, r *http.Request, mover FileMover) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading body", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	var results []ClassificationResult
	if err := json.Unmarshal(body, &results); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if err := mover.MoveFiles(results); err != nil {
		http.Error(w, "Error processing files", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"status":"processed","count":%d}`, len(results))
}

// IngestHandler wraps IngestHandlerWithDeps into a standard http.HandlerFunc closure.
// It acts as the primary public entrypoint for wiring the ingest endpoint to HTTP 
// multiplexers.
func IngestHandler(mover FileMover) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		IngestHandlerWithDeps(w, r, mover)
	}
}