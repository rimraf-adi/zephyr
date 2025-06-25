package pypi

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParsePEP518Config_Success(t *testing.T) {
	dir := t.TempDir()
	pyproject := `[build-system]
requires = ["setuptools>=61.0", "wheel"]
build-backend = "setuptools.build_meta"
`
	path := filepath.Join(dir, "pyproject.toml")
	os.WriteFile(path, []byte(pyproject), 0644)
	cfg, err := ParsePEP518Config(dir)
	if err != nil {
		t.Fatalf("ParsePEP518Config failed: %v", err)
	}
	if cfg.BuildSystem.Backend != "setuptools.build_meta" || len(cfg.BuildSystem.Requires) != 2 {
		t.Errorf("Parsed config mismatch: %+v", cfg.BuildSystem)
	}
}

func TestParsePEP518Config_FileNotFound(t *testing.T) {
	_, err := ParsePEP518Config("/nonexistent")
	if err == nil {
		t.Error("Expected error for missing file")
	}
}

func TestGetBuildDependenciesAndBackend(t *testing.T) {
	dir := t.TempDir()
	pyproject := `[build-system]
requires = ["setuptools>=61.0", "wheel"]
build-backend = "setuptools.build_meta"
`
	path := filepath.Join(dir, "pyproject.toml")
	os.WriteFile(path, []byte(pyproject), 0644)
	deps, err := GetBuildDependencies(dir)
	if err != nil || len(deps) != 2 {
		t.Errorf("GetBuildDependencies failed: %v, deps=%v", err, deps)
	}
	backend, err := GetBuildBackend(dir)
	if err != nil || backend != "setuptools.build_meta" {
		t.Errorf("GetBuildBackend failed: %v, backend=%s", err, backend)
	}
}

func TestValidateBuildSystem(t *testing.T) {
	cfg := DefaultBuildSystem()
	if err := ValidateBuildSystem(cfg); err != nil {
		t.Errorf("ValidateBuildSystem failed: %v", err)
	}
	cfg.BuildSystem.Backend = ""
	if err := ValidateBuildSystem(cfg); err == nil {
		t.Error("Expected error for missing backend")
	}
	cfg.BuildSystem.Backend = "setuptools.build_meta"
	cfg.BuildSystem.Requires = nil
	if err := ValidateBuildSystem(cfg); err == nil {
		t.Error("Expected error for empty requires")
	}
}

func TestDefaultBuildSystem(t *testing.T) {
	cfg := DefaultBuildSystem()
	if cfg.BuildSystem.Backend == "" || len(cfg.BuildSystem.Requires) == 0 {
		t.Error("DefaultBuildSystem should set backend and requires")
	}
} 