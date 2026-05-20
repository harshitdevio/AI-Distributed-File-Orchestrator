package classifier

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestProcessDirectory_RealFilesystem(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "classifier-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	testFiles := map[string]string{
		"test.txt":  "Hello",
		"data.json": `{"key":"value"}`,
	}

	for name, content := range testFiles {
		if err := os.WriteFile(filepath.Join(tempDir, name), []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create %s: %v", name, err)
		}
	}

	subDir := filepath.Join(tempDir, "subdir")
	if err := os.Mkdir(subDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(subDir, "nested.txt"), []byte("nested"), 0644); err != nil {
		t.Fatal(err)
	}

	files, err := ProcessDirectory(tempDir)
	if err != nil {
		t.Fatalf("ProcessDirectory error: %v", err)
	}

	if len(files) != 3 {
		t.Errorf("Found %d files, want 3", len(files))
	}

	for _, file := range files {
		if file.FilePath == "" || file.MIMEType == "" {
			t.Errorf("File has empty fields: %+v", file)
		}
		if _, err := os.Stat(file.FilePath); os.IsNotExist(err) {
			t.Errorf("File path doesn't exist: %s", file.FilePath)
		}
	}
}

func TestProcessDirectory_Errors(t *testing.T) {
	tests := []struct {
		name      string
		setup     func() string
		wantError string
	}{
		{
			name: "nonexistent directory",
			setup: func() string {
				return "/nonexistent/path/12345"
			},
			wantError: "directory does not exist",
		},
		{
			name: "file not directory",
			setup: func() string {
				f, _ := os.CreateTemp("", "test-*.txt")
				defer f.Close()
				return f.Name()
			},
			wantError: "path is not a directory",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := tt.setup()
			if !strings.Contains(path, "nonexistent") {
				defer os.Remove(path)
			}

			_, err := ProcessDirectory(path)
			if err == nil || !strings.Contains(err.Error(), tt.wantError) {
				t.Errorf("Error = %v, want containing %q", err, tt.wantError)
			}
		})
	}
}

func TestProcessDirectory_EmptyDirectory(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "empty-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	files, err := ProcessDirectory(tempDir)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if len(files) != 0 {
		t.Errorf("Got %d files in empty dir, want 0", len(files))
	}
}

// Benchmarks
func BenchmarkProcessDirectory(b *testing.B) {
	tempDir, _ := os.MkdirTemp("", "bench-*")
	defer os.RemoveAll(tempDir)

	for i := 0; i < 10; i++ {
		_ = os.WriteFile(filepath.Join(tempDir, fmt.Sprintf("file%d.txt", i)), []byte("test"), 0644)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ProcessDirectory(tempDir)
	}
}