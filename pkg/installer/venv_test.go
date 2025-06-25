package installer

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestVirtualEnvironmentCreateAndExists(t *testing.T) {
	dir := t.TempDir()
	venvPath := filepath.Join(dir, "venvtest")
	venv := NewVirtualEnvironment(venvPath)
	if venv.Exists() {
		t.Error("Venv should not exist before creation")
	}
	if err := venv.Create(); err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if !venv.Exists() {
		t.Error("Venv should exist after creation")
	}
}

func TestVirtualEnvironmentGetPaths(t *testing.T) {
	venv := NewVirtualEnvironment("/tmp/venvtest")
	py := venv.GetPythonPath()
	pip := venv.GetPipPath()
	bin := venv.GetBinPath()
	if runtime.GOOS == "windows" {
		if filepath.Base(py) != "python.exe" || filepath.Base(pip) != "pip.exe" {
			t.Error("Windows paths should end with .exe")
		}
	} else {
		if filepath.Base(py) != "python" || filepath.Base(pip) != "pip" {
			t.Error("Unix paths should end with python/pip")
		}
	}
	if bin == "" {
		t.Error("GetBinPath should not be empty")
	}
}

func TestVirtualEnvironmentRemove(t *testing.T) {
	dir := t.TempDir()
	venvPath := filepath.Join(dir, "venvtest")
	venv := NewVirtualEnvironment(venvPath)
	_ = venv.Create()
	if err := venv.Remove(); err != nil {
		t.Errorf("Remove failed: %v", err)
	}
	if venv.Exists() {
		t.Error("Venv should not exist after remove")
	}
}

func TestVirtualEnvironmentFindPython(t *testing.T) {
	venv := NewVirtualEnvironment("/tmp/venvtest")
	py, err := venv.findPython()
	if err != nil {
		t.Errorf("findPython failed: %v", err)
	}
	if py == "" {
		t.Error("findPython returned empty string")
	}
} 