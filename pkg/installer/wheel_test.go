package installer

import (
	"archive/zip"
	"os"
	"path/filepath"
	"testing"
)

func createTestWheel(t *testing.T, dir, name string) string {
	wheelPath := filepath.Join(dir, name)
	f, err := os.Create(wheelPath)
	if err != nil {
		t.Fatalf("Failed to create wheel: %v", err)
	}
	w := zip.NewWriter(f)
	// Add a dummy .dist-info/METADATA file
	meta, _ := w.Create("foo-1.0.0.dist-info/METADATA")
	meta.Write([]byte("Name: foo\nVersion: 1.0.0\n"))
	// Add a dummy .dist-info/WHEEL file
	wheel, _ := w.Create("foo-1.0.0.dist-info/WHEEL")
	wheel.Write([]byte("Wheel-Version: 1.0\n"))
	// Add a dummy package file
	pkgfile, _ := w.Create("foo/__init__.py")
	pkgfile.Write([]byte("# test package"))
	w.Close()
	f.Close()
	return wheelPath
}

func TestInstallWheel_Success(t *testing.T) {
	dir := t.TempDir()
	venvPath := filepath.Join(dir, "venv")
	os.MkdirAll(venvPath, 0755)
	wi := NewWheelInstaller(venvPath)
	wheelPath := createTestWheel(t, dir, "foo-1.0.0-py3-none-any.whl")
	err := wi.InstallWheel(wheelPath, "foo")
	if err != nil {
		t.Fatalf("InstallWheel failed: %v", err)
	}
	// Check that .dist-info directory exists
	distInfo := filepath.Join(venvPath, "lib", "python3.11", "site-packages", "foo-1.0.0.dist-info")
	if _, err := os.Stat(distInfo); err != nil {
		t.Errorf("dist-info directory not created: %v", err)
	}
}

func TestInstallWheel_InvalidWheel(t *testing.T) {
	dir := t.TempDir()
	venvPath := filepath.Join(dir, "venv")
	os.MkdirAll(venvPath, 0755)
	wi := NewWheelInstaller(venvPath)
	// Create an invalid wheel file
	badWheel := filepath.Join(dir, "bad.whl")
	os.WriteFile(badWheel, []byte("not a zip"), 0644)
	err := wi.InstallWheel(badWheel, "foo")
	if err == nil {
		t.Error("Expected error for invalid wheel, got nil")
	}
} 