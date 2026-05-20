package server

// MockFileMover for testing
type MockFileMover struct {
	MoveFilesFunc func(results []ClassificationResult) error
	CallCount     int
	LastResults   []ClassificationResult
}

func (m *MockFileMover) MoveFiles(results []ClassificationResult) error {
	m.CallCount++
	m.LastResults = results
	if m.MoveFilesFunc != nil {
		return m.MoveFilesFunc(results)
	}
	return nil
}