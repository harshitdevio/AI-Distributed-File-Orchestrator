package server

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestIngestHandlerWithDeps_MethodNotAllowed tests non-POST requests
func TestIngestHandlerWithDeps_MethodNotAllowed(t *testing.T) {
	methods := []string{"GET", "PUT", "DELETE", "PATCH"}

	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			req := httptest.NewRequest(method, "/ingest", nil)
			w := httptest.NewRecorder()
			mockMover := &MockFileMover{}

			IngestHandlerWithDeps(w, req, mockMover)

			if w.Code != http.StatusMethodNotAllowed {
				t.Errorf("Status = %d, want %d", w.Code, http.StatusMethodNotAllowed)
			}
			if mockMover.CallCount != 0 {
				t.Error("MoveFiles should not be called for non-POST")
			}
		})
	}
}

// TestIngestHandlerWithDeps_InvalidJSON tests malformed JSON
func TestIngestHandlerWithDeps_InvalidJSON(t *testing.T) {
	tests := []struct {
		name string
		body string
	}{
		{"empty body", ""},
		{"invalid json", "{invalid}"},
		{"plain text", "not json at all"},
		{"incomplete json", `[{"filepath":"test"`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/ingest", strings.NewReader(tt.body))
			w := httptest.NewRecorder()
			mockMover := &MockFileMover{}

			IngestHandlerWithDeps(w, req, mockMover)

			if w.Code != http.StatusBadRequest {
				t.Errorf("Status = %d, want %d", w.Code, http.StatusBadRequest)
			}
		})
	}
}

// TestIngestHandlerWithDeps_ValidRequest tests successful processing
func TestIngestHandlerWithDeps_ValidRequest(t *testing.T) {
	tests := []struct {
		name          string
		body          string
		expectedCount int
		expectedCalls int
	}{
		{
			name:          "single result",
			body:          `[{"filepath":"/test.txt","top_topic":"docs","confidence":0.9}]`,
			expectedCount: 1,
			expectedCalls: 1,
		},
		{
			name: "multiple results",
			body: `[
				{"filepath":"/a.txt","top_topic":"docs","confidence":0.8},
				{"filepath":"/b.jpg","top_topic":"images","confidence":0.95}
			]`,
			expectedCount: 2,
			expectedCalls: 1,
		},
		{
			name:          "empty array",
			body:          `[]`,
			expectedCount: 0,
			expectedCalls: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/ingest", strings.NewReader(tt.body))
			w := httptest.NewRecorder()
			mockMover := &MockFileMover{}

			IngestHandlerWithDeps(w, req, mockMover)

			if w.Code != http.StatusOK {
				t.Errorf("Status = %d, want %d", w.Code, http.StatusOK)
			}

			if mockMover.CallCount != tt.expectedCalls {
				t.Errorf("MoveFiles called %d times, want %d", mockMover.CallCount, tt.expectedCalls)
			}

			if len(mockMover.LastResults) != tt.expectedCount {
				t.Errorf("Results count = %d, want %d", len(mockMover.LastResults), tt.expectedCount)
			}

			expectedBody := fmt.Sprintf(`{"status":"processed","count":%d}`, tt.expectedCount)
			if strings.TrimSpace(w.Body.String()) != expectedBody {
				t.Errorf("Body = %q, want %q", w.Body.String(), expectedBody)
			}
		})
	}
}

// TestIngestHandlerWithDeps_MoverError tests error from mover
func TestIngestHandlerWithDeps_MoverError(t *testing.T) {
	req := httptest.NewRequest("POST", "/ingest", strings.NewReader(`[{"filepath":"/test.txt"}]`))
	w := httptest.NewRecorder()
	mockMover := &MockFileMover{
		MoveFilesFunc: func(results []ClassificationResult) error {
			return errors.New("disk full")
		},
	}

	IngestHandlerWithDeps(w, req, mockMover)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusInternalServerError)
	}
	if !strings.Contains(w.Body.String(), "Error processing files") {
		t.Errorf("Body should contain error message, got: %q", w.Body.String())
	}
}

// TestIngestHandlerWithDeps_ParsesAllFields tests complete struct parsing
func TestIngestHandlerWithDeps_ParsesAllFields(t *testing.T) {
	body := `[{
		"filepath": "/data/report.pdf",
		"top_topic": "documents",
		"confidence": 0.87,
		"all_scores": {"documents": 0.87, "images": 0.13},
		"error": null
	}]`

	req := httptest.NewRequest("POST", "/ingest", strings.NewReader(body))
	w := httptest.NewRecorder()
	mockMover := &MockFileMover{}

	IngestHandlerWithDeps(w, req, mockMover)

	if w.Code != http.StatusOK {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusOK)
	}

	result := mockMover.LastResults[0]
	if result.Filepath != "/data/report.pdf" {
		t.Errorf("Filepath = %q, want %q", result.Filepath, "/data/report.pdf")
	}
	if result.TopTopic != "documents" {
		t.Errorf("TopTopic = %q, want %q", result.TopTopic, "documents")
	}
	if result.Confidence != 0.87 {
		t.Errorf("Confidence = %f, want %f", result.Confidence, 0.87)
	}
	if len(result.AllScores) != 2 {
		t.Errorf("AllScores length = %d, want 2", len(result.AllScores))
	}
}

// TestIngestHandler_Integration tests the wrapped handler
func TestIngestHandler_Integration(t *testing.T) {
	mockMover := &MockFileMover{}
	handler := IngestHandler(mockMover)

	body := `[{"filepath":"/test.txt","top_topic":"docs","confidence":0.9}]`
	req := httptest.NewRequest("POST", "/ingest", strings.NewReader(body))
	w := httptest.NewRecorder()

	handler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusOK)
	}
	if mockMover.CallCount != 1 {
		t.Errorf("MoveFiles called %d times, want 1", mockMover.CallCount)
	}
}