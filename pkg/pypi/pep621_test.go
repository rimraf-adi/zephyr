package pypi

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCreateAndParsePEP621Config(t *testing.T) {
	dir := t.TempDir()
	cfg := CreateDefaultProject("foo", "1.0.0")
	cfg.Project.Description = "desc"
	if err := WritePEP621Config(dir, cfg); err != nil {
		t.Fatalf("WritePEP621Config failed: %v", err)
	}
	parsed, err := ParsePEP621Config(dir)
	if err != nil {
		t.Fatalf("ParsePEP621Config failed: %v", err)
	}
	if parsed.Project.Name != "foo" || parsed.Project.Version != "1.0.0" || parsed.Project.Description != "desc" {
		t.Errorf("Parsed config mismatch: %+v", parsed.Project)
	}
}

func TestGetProjectNameVersionDependencies(t *testing.T) {
	dir := t.TempDir()
	cfg := CreateDefaultProject("bar", "2.0.0")
	cfg.Project.Dependencies["baz"] = ">=1.0.0"
	WritePEP621Config(dir, cfg)
	name, err := GetProjectName(dir)
	if err != nil || name != "bar" {
		t.Errorf("GetProjectName failed: %v, name=%s", err, name)
	}
	ver, err := GetProjectVersion(dir)
	if err != nil || ver != "2.0.0" {
		t.Errorf("GetProjectVersion failed: %v, ver=%s", err, ver)
	}
	deps, err := GetProjectDependencies(dir)
	if err != nil || deps["baz"] != ">=1.0.0" {
		t.Errorf("GetProjectDependencies failed: %v, deps=%v", err, deps)
	}
}

func TestValidateProject(t *testing.T) {
	cfg := CreateDefaultProject("foo", "1.0.0")
	if err := ValidateProject(cfg); err != nil {
		t.Errorf("ValidateProject failed: %v", err)
	}
	cfg.Project.Name = ""
	if err := ValidateProject(cfg); err == nil {
		t.Error("Expected error for missing name")
	}
	cfg.Project.Name = "foo"
	cfg.Project.Version = ""
	if err := ValidateProject(cfg); err == nil {
		t.Error("Expected error for missing version")
	}
}

func TestAddAndRemoveDependency(t *testing.T) {
	dir := t.TempDir()
	cfg := CreateDefaultProject("foo", "1.0.0")
	WritePEP621Config(dir, cfg)
	if err := AddDependency(dir, "bar", ">=2.0.0"); err != nil {
		t.Fatalf("AddDependency failed: %v", err)
	}
	parsed, _ := ParsePEP621Config(dir)
	if parsed.Project.Dependencies["bar"] != ">=2.0.0" {
		t.Error("Dependency not added")
	}
	if err := RemoveDependency(dir, "bar"); err != nil {
		t.Fatalf("RemoveDependency failed: %v", err)
	}
	parsed, _ = ParsePEP621Config(dir)
	if _, ok := parsed.Project.Dependencies["bar"]; ok {
		t.Error("Dependency not removed")
	}
} 