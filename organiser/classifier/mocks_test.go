package classifier

import (
	"errors"
	"io/fs"
	"time"
)

// MockFileSystem mocks the filesystem for unit tests.
type MockFileSystem struct {
	StatFunc    func(path string) (fs.FileInfo, error)
	WalkDirFunc func(root string, fn fs.WalkDirFunc) error
}

func (m *MockFileSystem) Stat(path string) (fs.FileInfo, error) {
	if m.StatFunc != nil {
		return m.StatFunc(path)
	}
	return nil, errors.New("StatFunc not implemented")
}

func (m *MockFileSystem) WalkDir(root string, fn fs.WalkDirFunc) error {
	if m.WalkDirFunc != nil {
		return m.WalkDirFunc(root, fn)
	}
	return errors.New("WalkDirFunc not implemented")
}

// MockFileInfo satisfies the fs.FileInfo interface.
type MockFileInfo struct {
	NameValue  string
	IsDirValue bool
}

func (m MockFileInfo) Name() string         { return m.NameValue }
func (m MockFileInfo) Size() int64          { return 0 }
func (m MockFileInfo) Mode() fs.FileMode    { return 0 }
func (m MockFileInfo) ModTime() time.Time   { return time.Time{} }
func (m MockFileInfo) IsDir() bool          { return m.IsDirValue }
func (m MockFileInfo) Sys() interface{}     { return nil }

// MockDirEntry satisfies the fs.DirEntry interface.
type MockDirEntry struct {
	NameValue  string
	IsDirValue bool
}

func (m MockDirEntry) Name() string               { return m.NameValue }
func (m MockDirEntry) IsDir() bool                { return m.IsDirValue }
func (m MockDirEntry) Type() fs.FileMode          { return 0 }
func (m MockDirEntry) Info() (fs.FileInfo, error) { return nil, nil }

// MockMIMEDetector mocks the MIME detection behavior.
type MockMIMEDetector struct {
	DetectFunc func(path string) (string, error)
}

func (m *MockMIMEDetector) DetectFile(path string) (string, error) {
	if m.DetectFunc != nil {
		return m.DetectFunc(path)
	}
	return "application/octet-stream", nil
}