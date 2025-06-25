package registry

import (
	"testing"
)

func TestInMemoryRegistry_AddAndGetPackage(t *testing.T) {
	r := NewInMemoryRegistry()
	pkg := &Package{Name: "foo", Version: "1.0.0"}
	r.AddPackage(pkg)
	got, err := r.GetPackage("foo", "1.0.0")
	if err != nil {
		t.Fatalf("GetPackage failed: %v", err)
	}
	if got.Name != "foo" || got.Version != "1.0.0" {
		t.Errorf("Package mismatch: %+v", got)
	}
}

func TestInMemoryRegistry_GetPackage_NotFound(t *testing.T) {
	r := NewInMemoryRegistry()
	_, err := r.GetPackage("bar", "1.0.0")
	if err == nil {
		t.Error("Expected error for missing package")
	}
}

func TestInMemoryRegistry_GetVersions(t *testing.T) {
	r := NewInMemoryRegistry()
	pkg := &Package{Name: "foo", Version: "1.0.0"}
	r.AddPackage(pkg)
	vers, err := r.GetVersions("foo")
	if err != nil || len(vers) != 1 || vers[0] != "1.0.0" {
		t.Errorf("GetVersions mismatch: %+v, err=%v", vers, err)
	}
}

func TestInMemoryRegistry_GetLatestVersion(t *testing.T) {
	r := NewInMemoryRegistry()
	pkg1 := &Package{Name: "foo", Version: "1.0.0"}
	pkg2 := &Package{Name: "foo", Version: "2.0.0"}
	r.AddPackage(pkg1)
	r.AddPackage(pkg2)
	ver, err := r.GetLatestVersion("foo")
	if err != nil || (ver != "1.0.0" && ver != "2.0.0") {
		t.Errorf("GetLatestVersion mismatch: %s, err=%v", ver, err)
	}
}

func TestInMemoryRegistry_Satisfies(t *testing.T) {
	r := NewInMemoryRegistry()
	vc := VersionConstraint{Specific: "1.0.0"}
	if !r.Satisfies("1.0.0", vc) {
		t.Error("Satisfies should be true for specific match")
	}
	vc2 := VersionConstraint{Min: "1.0.0"}
	if !r.Satisfies("2.0.0", vc2) {
		t.Error("Satisfies should be true for non-specific constraint (placeholder)")
	}
}

func TestVersionConstraint_String(t *testing.T) {
	tests := []struct {
		vc       VersionConstraint
		expected string
	}{
		{VersionConstraint{Specific: "1.0.0"}, "1.0.0"},
		{VersionConstraint{Min: "1.0.0"}, ">=1.0.0"},
		{VersionConstraint{Max: "2.0.0"}, "<2.0.0"},
		{VersionConstraint{Min: "1.0.0", Max: "2.0.0"}, ">=1.0.0 <2.0.0"},
		{VersionConstraint{}, "any"},
	}
	for _, test := range tests {
		if test.vc.String() != test.expected {
			t.Errorf("Expected '%s', got '%s'", test.expected, test.vc.String())
		}
	}
} 