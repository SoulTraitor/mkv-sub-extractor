package mkvinfo

import (
	"os"
	"strings"
	"testing"
)

func TestGetMKVInfo_NonExistentFile(t *testing.T) {
	_, err := GetMKVInfo("/nonexistent/path/to/file.mkv")
	if err == nil {
		t.Fatal("expected error for non-existent file, got nil")
	}
	if !strings.Contains(err.Error(), "cannot open file") {
		t.Errorf("expected 'cannot open file' error, got: %v", err)
	}
}

func TestGetMKVInfo_NonMKVFile(t *testing.T) {
	// Create a temporary file with random bytes (not a valid MKV)
	tmpFile, err := os.CreateTemp("", "test-nonmkv-*.bin")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	// Write some random data that is not a valid EBML header
	_, err = tmpFile.Write([]byte("this is not a valid MKV file content at all"))
	if err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}
	tmpFile.Close()

	_, err = GetMKVInfo(tmpFile.Name())
	if err == nil {
		t.Fatal("expected error for non-MKV file, got nil")
	}
	if !strings.Contains(err.Error(), "not a valid MKV file") {
		t.Errorf("expected 'not a valid MKV file' error, got: %v", err)
	}
}

// Integration testing with real MKV files should be done manually:
//   go run ./cmd/mkv-sub-extractor/main.go path/to/test.mkv
