package classifier

import (
	"bytes"
	"errors"
	"io/fs"
	"os"
	"strings"
	"testing"
)

func TestIsValidDirectoryWithDeps(t *testing.T) {
	tests := []struct {
		name      string
		path      string
		mockStat  func(path string) (fs.FileInfo, error)
		want      bool
		wantError string
	}{
		{
			name: "valid directory",
			path: "/valid/dir",
			mockStat: func(path string) (fs.FileInfo, error) {
				return MockFileInfo{NameValue: "dir", IsDirValue: true}, nil
			},
			want: true,
		},
		{
			name: "directory does not exist",
			path: "/nonexistent",
			mockStat: func(path string) (fs.FileInfo, error) {
				return nil, os.ErrNotExist
			},
			want:      false,
			wantError: "directory does not exist",
		},
		{
			name: "path is file",
			path: "/file.txt",
			mockStat: func(path string) (fs.FileInfo, error) {
				return MockFileInfo{NameValue: "file.txt", IsDirValue: false}, nil
			},
			want:      false,
			wantError: "path is not a directory",
		},
		{
			name: "permission error",
			path: "/forbidden",
			mockStat: func(path string) (fs.FileInfo, error) {
				return nil, os.ErrPermission
			},
			want:      false,
			wantError: "error accessing path",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockFS := &MockFileSystem{StatFunc: tt.mockStat}
			got, err := isValidDirectoryWithDeps(tt.path, mockFS)

			if tt.wantError != "" {
				if err == nil {
					t.Errorf("Expected error containing %q, got nil", tt.wantError)
					return
				}
				if !strings.Contains(err.Error(), tt.wantError) {
					t.Errorf("Error = %q, want containing %q", err.Error(), tt.wantError)
				}
			} else if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if got != tt.want {
				t.Errorf("Got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProcessDirectoryWithDeps_MockedFilesystem(t *testing.T) {
	tests := []struct {
		name         string
		dirPath      string
		mockStat     func(path string) (fs.FileInfo, error)
		mockWalkDir  func(root string, fn fs.WalkDirFunc) error
		mockDetector func(path string) (string, error)
		want         []FileInfo
		wantError    string
	}{
		{
			name:    "empty directory",
			dirPath: "/empty",
			mockStat: func(path string) (fs.FileInfo, error) {
				return MockFileInfo{IsDirValue: true}, nil
			},
			mockWalkDir: func(root string, fn fs.WalkDirFunc) error {
				return nil
			},
			want: []FileInfo{},
		},
		{
			name:    "single file",
			dirPath: "/single",
			mockStat: func(path string) (fs.FileInfo, error) {
				return MockFileInfo{IsDirValue: true}, nil
			},
			mockWalkDir: func(root string, fn fs.WalkDirFunc) error {
				return fn("/single/file.txt", MockDirEntry{NameValue: "file.txt"}, nil)
			},
			mockDetector: func(path string) (string, error) {
				return "text/plain", nil
			},
			want: []FileInfo{{FilePath: "/single/file.txt", MIMEType: "text/plain"}},
		},
		{
			name:    "multiple files",
			dirPath: "/multi",
			mockStat: func(path string) (fs.FileInfo, error) {
				return MockFileInfo{IsDirValue: true}, nil
			},
			mockWalkDir: func(root string, fn fs.WalkDirFunc) error {
				files := []struct {
					path  string
					name  string
					isDir bool
				}{
					{"/multi/doc.pdf", "doc.pdf", false},
					{"/multi/img.png", "img.png", false},
					{"/multi/subdir", "subdir", true},
				}
				for _, f := range files {
					if err := fn(f.path, MockDirEntry{NameValue: f.name, IsDirValue: f.isDir}, nil); err != nil {
						return err
					}
				}
				return nil
			},
			mockDetector: func(path string) (string, error) {
				if strings.HasSuffix(path, ".pdf") {
					return "application/pdf", nil
				}
				if strings.HasSuffix(path, ".png") {
					return "image/png", nil
				}
				return "application/octet-stream", nil
			},
			want: []FileInfo{
				{FilePath: "/multi/doc.pdf", MIMEType: "application/pdf"},
				{FilePath: "/multi/img.png", MIMEType: "image/png"},
			},
		},
		{
			name:    "invalid directory",
			dirPath: "/invalid",
			mockStat: func(path string) (fs.FileInfo, error) {
				return nil, os.ErrNotExist
			},
			wantError: "directory does not exist",
		},
		{
			name:    "walk error",
			dirPath: "/error",
			mockStat: func(path string) (fs.FileInfo, error) {
				return MockFileInfo{IsDirValue: true}, nil
			},
			mockWalkDir: func(root string, fn fs.WalkDirFunc) error {
				return errors.New("permission denied")
			},
			wantError: "error walking directory",
		},
		{
			name:    "mime error logged not returned",
			dirPath: "/mime-err",
			mockStat: func(path string) (fs.FileInfo, error) {
				return MockFileInfo{IsDirValue: true}, nil
			},
			mockWalkDir: func(root string, fn fs.WalkDirFunc) error {
				return fn("/mime-err/file", MockDirEntry{NameValue: "file"}, nil)
			},
			mockDetector: func(path string) (string, error) {
				return "", errors.New("mime error")
			},
			want: []FileInfo{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockFS := &MockFileSystem{
				StatFunc:    tt.mockStat,
				WalkDirFunc: tt.mockWalkDir,
			}
			mockDetector := &MockMIMEDetector{DetectFunc: tt.mockDetector}
			errWriter := &bytes.Buffer{}

			got, err := ProcessDirectoryWithDeps(tt.dirPath, mockFS, mockDetector, errWriter)

			if tt.wantError != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantError) {
					t.Errorf("Error = %v, want containing %q", err, tt.wantError)
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if len(got) != len(tt.want) {
				t.Errorf("Got %d files, want %d", len(got), len(tt.want))
				return
			}

			for i, file := range got {
				if file.FilePath != tt.want[i].FilePath || file.MIMEType != tt.want[i].MIMEType {
					t.Errorf("File[%d] = %+v, want %+v", i, file, tt.want[i])
				}
			}
		})
	}
}