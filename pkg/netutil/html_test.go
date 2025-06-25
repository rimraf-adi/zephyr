package netutil

import (
	"strings"
	"testing"
)

func TestNewHTMLParser_ValidHTML(t *testing.T) {
	html := `<html><body><a href="foo/">foo</a><a href="bar/">bar</a></body></html>`
	parser, err := NewHTMLParser(html)
	if err != nil {
		t.Fatalf("NewHTMLParser failed: %v", err)
	}
	links, err := parser.ExtractPackageLinks()
	if err != nil {
		t.Fatalf("ExtractPackageLinks failed: %v", err)
	}
	if len(links) != 2 || links[0] != "foo" || links[1] != "bar" {
		t.Errorf("Extracted links mismatch: %+v", links)
	}
}

func TestNewHTMLParser_InvalidHTML(t *testing.T) {
	_, err := NewHTMLParser("<html><body><a href='foo'>foo")
	if err == nil {
		t.Error("Expected error for invalid HTML, got nil")
	}
}

func TestExtractDownloadLinks(t *testing.T) {
	html := `<html><body><a href="file.whl">file.whl</a><a href="file.tar.gz">file.tar.gz</a></body></html>`
	parser, _ := NewHTMLParser(html)
	links, err := parser.ExtractDownloadLinks()
	if err != nil {
		t.Fatalf("ExtractDownloadLinks failed: %v", err)
	}
	if len(links) != 2 || links[0].URL != "file.whl" || links[1].URL != "file.tar.gz" {
		t.Errorf("Download links mismatch: %+v", links)
	}
}

func TestParsePyPIPackagePage(t *testing.T) {
	html := `<html><head><title>foo · PyPI</title></head><body><div class='package-description'>desc</div><a href="foo.whl">foo.whl</a></body></html>`
	info, err := ParsePyPIPackagePage(html)
	if err != nil {
		t.Fatalf("ParsePyPIPackagePage failed: %v", err)
	}
	if info.Name != "foo" || info.Description != "desc" || len(info.DownloadLinks) != 1 {
		t.Errorf("PyPIPackageInfo mismatch: %+v", info)
	}
} 