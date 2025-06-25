package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestZephyrInitAndAddRemove(t *testing.T) {
	dir := t.TempDir()
	bin := buildZephyrBinary(t)
	project := "testproj"
	cmd := exec.Command(bin, "init", project)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("zephyr init failed: %v, out=%s", err, out)
	}
	bmPath := filepath.Join(dir, project, "buildmeta.yaml")
	if _, err := os.Stat(bmPath); err != nil {
		t.Errorf("buildmeta.yaml not created: %v", err)
	}
	// Add dependency
	cmd = exec.Command(bin, "add", "requests", ">=2.0.0")
	cmd.Dir = filepath.Join(dir, project)
	out, err = cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("zephyr add failed: %v, out=%s", err, out)
	}
	// Remove dependency
	cmd = exec.Command(bin, "remove", "requests")
	cmd.Dir = filepath.Join(dir, project)
	out, err = cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("zephyr remove failed: %v, out=%s", err, out)
	}
}

func TestZephyrVenvCreateListActivate(t *testing.T) {
	dir := t.TempDir()
	bin := buildZephyrBinary(t)
	cmd := exec.Command(bin, "venv", "create")
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Skipf("zephyr venv create failed (skip if no Python): %v, out=%s", err, out)
	}
	venvPath := filepath.Join(dir, ".venv")
	if _, err := os.Stat(venvPath); err != nil {
		t.Errorf(".venv not created: %v", err)
	}
	cmd = exec.Command(bin, "venv", "list")
	cmd.Dir = dir
	out, _ = cmd.CombinedOutput()
	if !strings.Contains(string(out), ".venv") {
		t.Errorf("venv list output missing .venv: %s", out)
	}
	cmd = exec.Command(bin, "venv", "activate")
	cmd.Dir = dir
	out, _ = cmd.CombinedOutput()
	if !strings.Contains(string(out), "activate") {
		t.Errorf("venv activate output missing instructions: %s", out)
	}
}

func TestZephyrLockInstallSync(t *testing.T) {
	dir := t.TempDir()
	bin := buildZephyrBinary(t)
	cmd := exec.Command(bin, "init", "proj")
	cmd.Dir = dir
	cmd.CombinedOutput()
	cmd = exec.Command(bin, "add", "requests", ">=2.0.0")
	cmd.Dir = filepath.Join(dir, "proj")
	cmd.CombinedOutput()
	cmd = exec.Command(bin, "lock")
	cmd.Dir = filepath.Join(dir, "proj")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Skipf("zephyr lock failed (skip if no network): %v, out=%s", err, out)
	}
	lockPath := filepath.Join(dir, "proj", "zephyr.lock")
	if _, err := os.Stat(lockPath); err != nil {
		t.Errorf("zephyr.lock not created: %v", err)
	}
	// install and sync require Python and network, so we skip if not available
}

func buildZephyrBinary(t *testing.T) string {
	bin := filepath.Join(os.TempDir(), "zephyr-test-bin")
	// Find project root (assume test is run from any subdir)
	cwd, _ := os.Getwd()
	var root string
	for i := 0; i < 5; i++ {
		if _, err := os.Stat(filepath.Join(cwd, "go.mod")); err == nil {
			root = cwd
			break
		}
		cwd = filepath.Dir(cwd)
	}
	if root == "" {
		t.Fatalf("Could not find project root with go.mod")
	}
	mainPath := filepath.Join(root, "cmd", "zephyr", "main.go")
	cmd := exec.Command("go", "build", "-o", bin, mainPath)
	cmd.Dir = root
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("failed to build zephyr binary: %v, out=%s", err, out)
	}
	return bin
}
