package installer

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLockfileLifecycle(t *testing.T) {
	dir := t.TempDir()
	lockPath := filepath.Join(dir, "zephyr.lock")
	lf := NewLockfile("3.11")
	lf.Packages["foo"] = LockPackage{Version: "1.2.3", Source: "pypi"}
	if err := lf.Save(lockPath); err != nil {
		t.Fatalf("Save failed: %v", err)
	}
	lf2, err := LoadLockfile(lockPath)
	if err != nil {
		t.Fatalf("LoadLockfile failed: %v", err)
	}
	if lf2.Python != "3.11" || lf2.Packages["foo"].Version != "1.2.3" {
		t.Errorf("Loaded lockfile mismatch: got %+v", lf2)
	}
	if err := lf2.Validate(); err != nil {
		t.Errorf("Validate failed: %v", err)
	}
}

func TestLockfileManager(t *testing.T) {
	dir := t.TempDir()
	mgr := NewLockfileManager(dir)
	lf := mgr.Create("3.10")
	lf.Packages["bar"] = LockPackage{Version: "2.0.0", Source: "pypi"}
	if err := mgr.Save(lf); err != nil {
		t.Fatalf("Save failed: %v", err)
	}
	lf2, err := mgr.Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if !mgr.Exists() {
		t.Error("Exists() should be true after save")
	}
	if err := mgr.Remove(); err != nil {
		t.Errorf("Remove failed: %v", err)
	}
	if mgr.Exists() {
		t.Error("Exists() should be false after remove")
	}
}

func TestLockfileHashAndStale(t *testing.T) {
	dir := t.TempDir()
	reqPath := filepath.Join(dir, "requirements.txt")
	os.WriteFile(reqPath, []byte("foo==1.2.3\nbar>=2.0.0"), 0644)
	lf := NewLockfile("3.9")
	if err := lf.UpdateHash(reqPath); err != nil {
		t.Fatalf("UpdateHash failed: %v", err)
	}
	stale, err := lf.IsStale(reqPath)
	if err != nil {
		t.Fatalf("IsStale failed: %v", err)
	}
	if stale {
		t.Error("Lockfile should not be stale after hash update")
	}
	os.WriteFile(reqPath, []byte("foo==1.2.4"), 0644)
	stale, _ = lf.IsStale(reqPath)
	if !stale {
		t.Error("Lockfile should be stale after requirements change")
	}
}

func TestLockfileValidationErrors(t *testing.T) {
	lf := &Lockfile{}
	err := lf.Validate()
	if err == nil {
		t.Error("Validate should fail for empty lockfile")
	}
	lf.Version = "1.0"
	err = lf.Validate()
	if err == nil {
		t.Error("Validate should fail for missing Python version")
	}
	lf.Python = "3.11"
	err = lf.Validate()
	if err == nil {
		t.Error("Validate should fail for nil Packages")
	}
	lf.Packages = make(map[string]LockPackage)
	if err := lf.Validate(); err != nil {
		t.Errorf("Validate should succeed for valid lockfile: %v", err)
	}
} 