package pypi

import (
	"testing"
)

func TestNewPEP517BuildBackend(t *testing.T) {
	b := NewPEP517BuildBackend("/path/to/backend", "backend")
	if b.BackendPath != "/path/to/backend" || b.BackendName != "backend" {
		t.Errorf("NewPEP517BuildBackend fields mismatch: %+v", b)
	}
}

// Integration tests for BuildWheel/BuildSdist would require a real Python environment and are skipped here.
func TestPEP517BuildBackend_Methods(t *testing.T) {
	b := NewPEP517BuildBackend("/path", "backend")
	// These should return errors if run in a test environment without Python/pep517
	_, err := b.BuildWheel(BuildRequest{})
	if err == nil {
		t.Error("Expected error for BuildWheel in test env")
	}
	_, err = b.BuildSdist(BuildRequest{})
	if err == nil {
		t.Error("Expected error for BuildSdist in test env")
	}
	_, err = b.GetRequiresForBuildWheel("/path")
	if err == nil {
		t.Error("Expected error for GetRequiresForBuildWheel in test env")
	}
	_, err = b.GetRequiresForBuildSdist("/path")
	if err == nil {
		t.Error("Expected error for GetRequiresForBuildSdist in test env")
	}
	_, err = b.PrepareMetadataForBuildWheel("/path", "/meta")
	if err == nil {
		t.Error("Expected error for PrepareMetadataForBuildWheel in test env")
	}
} 