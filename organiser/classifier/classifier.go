// Package classifier handles directory scanning and identifies files by their 
// media types using specific file system and MIME detection abstractions.
package classifier

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/gabriel-vasile/mimetype"
)

// FileInfo holds the metadata discovered for an individual file during processing.
type FileInfo struct {
	// FilePath contains the path relative to the runtime or absolute path 
	// encountered during the directory walk.
	FilePath string `json:"filepath"`
	
	// MIMEType contains the detected media type string (e.g., "application/json").
	MIMEType string `json:"mime_type"`
}

// FileSystemOps abstracts basic filesystem calls to decouple package logic from 
// the underlying OS storage. This enables structural mock testing without 
// disk interaction.
type FileSystemOps interface {
	// Stat returns the FileInfo structure describing the named file.
	Stat(path string) (fs.FileInfo, error)
	
	// WalkDir walks the file tree rooted at root, calling fn for each file or 
	// directory in the tree, including root.
	WalkDir(root string, fn fs.WalkDirFunc) error
}

// MIMEDetector abstracts content sniffing and magic number analysis to 
// identify actual file headers.
type MIMEDetector interface {
	// DetectFile inspects the given file and returns its MIME type string.
	DetectFile(path string) (string, error)
}

// RealFileSystem implements FileSystemOps by invoking standard library 'os' 
// and 'path/filepath' functions against the physical local disk.
type RealFileSystem struct{}

// Stat proxies directly to os.Stat.
func (rfs *RealFileSystem) Stat(path string) (fs.FileInfo, error) {
	return os.Stat(path)
}

// WalkDir proxies directly to filepath.WalkDir.
func (rfs *RealFileSystem) WalkDir(root string, fn fs.WalkDirFunc) error {
	return filepath.WalkDir(root, fn)
}

// RealMIMEDetector implements MIMEDetector using the external 
// github.com/gabriel-vasile/mimetype package.
type RealMIMEDetector struct{}

// DetectFile parses the target file path via the gabriel-vasile/mimetype engine.
func (rmd *RealMIMEDetector) DetectFile(path string) (string, error) {
	mtype, err := mimetype.DetectFile(path)
	if err != nil {
		return "", err
	}
	return mtype.String(), nil
}

// ProcessDirectoryWithDeps walks a directory tree, resolves the MIME type of every 
// file found, and aggregates the data into a slice of FileInfo structures.
//
// If a file's MIME type cannot be determined, the failure is logged to the provided 
// errWriter, and the execution continues rather than aborting the entire process.
func ProcessDirectoryWithDeps(dirPath string, fsOps FileSystemOps, detector MIMEDetector, errWriter io.Writer) ([]FileInfo, error) {
	if valid, err := isValidDirectoryWithDeps(dirPath, fsOps); !valid {
		return nil, err
	}

	var files []FileInfo

	err := fsOps.WalkDir(dirPath, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		mimeType, err := detector.DetectFile(path)
		if err != nil {
			fmt.Fprintf(errWriter, "Error detecting MIME for %s: %v\n", path, err)
			return nil
		}

		files = append(files, FileInfo{
			FilePath: path,
			MIMEType: mimeType,
		})

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("error walking directory: %w", err)
	}

	return files, nil
}

// ProcessDirectory provides a standard production entrypoint for package consumers. 
// It initializes and injects the live, OS-backed structural dependencies 
// and routes internal errors to standard error.
func ProcessDirectory(dirPath string) ([]FileInfo, error) {
	return ProcessDirectoryWithDeps(dirPath, &RealFileSystem{}, &RealMIMEDetector{}, os.Stderr)
}

// isValidDirectoryWithDeps runs preconditions against the target folder path 
// using the passed FileSystemOps. It ensures the target exists and is a directory.
func isValidDirectoryWithDeps(path string, fsOps FileSystemOps) (bool, error) {
	info, err := fsOps.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, fmt.Errorf("directory does not exist: %s", path)
		}
		return false, fmt.Errorf("error accessing path: %w", err)
	}

	if !info.IsDir() {
		return false, fmt.Errorf("path is not a directory: %s", path)
	}

	return true, nil
}

// isValidDirectory evaluates a folder path's structural validity against the 
// live physical file system.
func isValidDirectory(path string) (bool, error) {
	return isValidDirectoryWithDeps(path, &RealFileSystem{})
}