package server

import (
	"bytes"
	"net/http/httptest"
	"testing"
)

func BenchmarkIngestHandler(b *testing.B) {
	mockMover := &MockFileMover{}
	body := `[{"filepath":"/test.txt","top_topic":"docs","confidence":0.9}]`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("POST", "/ingest", bytes.NewReader([]byte(body)))
		w := httptest.NewRecorder()
		IngestHandlerWithDeps(w, req, mockMover)
	}
}