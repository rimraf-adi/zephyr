package netutil

import (
	"net/http"
	"os"
	"path/filepath"
	"testing"
)

func TestMergeConfig(t *testing.T) {
	global := &Config{IndexURL: "https://global.example.com"}
	project := &Config{IndexURL: "https://project.example.com"}
	cfg := mergeConfig(global, project)
	if cfg.IndexURL != "https://project.example.com" {
		t.Errorf("Expected project IndexURL to override global, got %s", cfg.IndexURL)
	}
	os.Setenv("ZEPHYR_INDEX_URL", "https://env.example.com")
	cfg = mergeConfig(global, project)
	if cfg.IndexURL != "https://env.example.com" {
		t.Errorf("Expected env var to override config, got %s", cfg.IndexURL)
	}
	os.Unsetenv("ZEPHYR_INDEX_URL")
}

func TestAddPyPIHeaders(t *testing.T) {
	req, _ := http.NewRequest("GET", "https://pypi.org", nil)
	AddPyPIHeaders(req)
	if req.Header.Get("User-Agent") == "" {
		t.Error("User-Agent header not set")
	}
	if req.Header.Get("Accept") == "" {
		t.Error("Accept header not set")
	}
}

func TestCreatePyPIRequest(t *testing.T) {
	req, err := CreatePyPIRequest("GET", "https://pypi.org")
	if err != nil {
		t.Fatalf("CreatePyPIRequest failed: %v", err)
	}
	if req.Method != "GET" {
		t.Error("Request method mismatch")
	}
}

func TestDownloadFile_NotFound(t *testing.T) {
	client := NewPyPIClient()
	dir := t.TempDir()
	file := filepath.Join(dir, "out.txt")
	err := DownloadFile(client, "http://localhost:9999/notfound", file)
	if err == nil {
		t.Error("Expected error for download from invalid URL")
	}
} 