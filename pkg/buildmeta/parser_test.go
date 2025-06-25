package buildmeta

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseAndWriteBuildMeta(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "buildmeta.yaml")
	bm := NewBuildMeta("foo", "1.0.0")
	bm.Description = "desc"
	if err := WriteToDirectory(dir, bm); err != nil {
		t.Fatalf("WriteToDirectory failed: %v", err)
	}
	bm2, err := ParseFromDirectory(dir)
	if err != nil {
		t.Fatalf("ParseFromDirectory failed: %v", err)
	}
	if bm2.Name != "foo" || bm2.Version != "1.0.0" || bm2.Description != "desc" {
		t.Errorf("Parsed buildmeta mismatch: %+v", bm2)
	}
}

func TestBuildMetaValidation(t *testing.T) {
	bm := &BuildMeta{}
	if err := bm.Validate(); err == nil {
		t.Error("Validate should fail for missing name/version")
	}
	bm.Name = "foo"
	if err := bm.Validate(); err == nil {
		t.Error("Validate should fail for missing version")
	}
	bm.Version = "1.0.0"
	if err := bm.Validate(); err != nil {
		t.Errorf("Validate should succeed: %v", err)
	}
}

func TestRequirementsImportExport(t *testing.T) {
	dir := t.TempDir()
	reqPath := filepath.Join(dir, "requirements.txt")
	os.WriteFile(reqPath, []byte("foo==1.2.3\nbar>=2.0.0"), 0644)
	reqs, err := ParseRequirementsFile(reqPath)
	if err != nil {
		t.Fatalf("ParseRequirementsFile failed: %v", err)
	}
	if reqs["foo"] != "==1.2.3" || reqs["bar"] != ">=2.0.0" {
		t.Errorf("Parsed requirements mismatch: %+v", reqs)
	}
	exportPath := filepath.Join(dir, "out.txt")
	if err := ExportRequirementsFile(exportPath, reqs); err != nil {
		t.Fatalf("ExportRequirementsFile failed: %v", err)
	}
	data, _ := os.ReadFile(exportPath)
	if string(data) == "" {
		t.Error("Exported requirements.txt is empty")
	}
}

func TestPyProjectImportExport(t *testing.T) {
	dir := t.TempDir()
	pyPath := filepath.Join(dir, "pyproject.toml")
	os.WriteFile(pyPath, []byte(`[project]\nname = "foo"\nversion = "1.0.0"\n[project.dependencies]\nbar = ">=2.0.0"\n`), 0644)
	meta, err := ParsePyProjectToml(pyPath)
	if err != nil {
		t.Fatalf("ParsePyProjectToml failed: %v", err)
	}
	if meta.Name != "foo" || meta.Version != "1.0.0" || meta.Dependencies["bar"] != ">=2.0.0" {
		t.Errorf("Parsed pyproject.toml mismatch: %+v", meta)
	}
	bm := NewBuildMeta(meta.Name, meta.Version)
	for k, v := range meta.Dependencies {
		bm.AddDependency(k, v)
	}
	exportPath := filepath.Join(dir, "out.toml")
	if err := ExportPyProjectToml(exportPath, bm); err != nil {
		t.Fatalf("ExportPyProjectToml failed: %v", err)
	}
	data, _ := os.ReadFile(exportPath)
	if string(data) == "" {
		t.Error("Exported pyproject.toml is empty")
	}
} 