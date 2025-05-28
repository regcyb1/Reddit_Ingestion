package fixtures

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
)

func LoadFixture(filename string) (json.RawMessage, error) {
	// Get the directory of the current file
	_, currentFile, _, _ := runtime.Caller(0)
	dir := filepath.Dir(currentFile)
	
	// Build path to fixtures directory
	path := filepath.Join(dir, filename)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return json.RawMessage(data), nil
}
