package registry

import (
	"fmt"
)

// Package represents a package with its metadata and dependencies
type Package struct {
	Name         string
	Version      string
	Dependencies []Dependency
}

// Dependency represents a package dependency
type Dependency struct {
	Package string
	Version VersionConstraint
}

// VersionConstraint represents a version constraint
type VersionConstraint struct {
	Min      string
	Max      string
	Specific string
}

// IsSpecific returns true if this constraint represents a specific version
func (vc VersionConstraint) IsSpecific() bool {
	return vc.Specific != ""
}

// String returns a string representation of the version constraint
func (vc VersionConstraint) String() string {
	if vc.IsSpecific() {
		return vc.Specific
	}
	
	if vc.Min != "" && vc.Max != "" {
		return fmt.Sprintf(">=%s <%s", vc.Min, vc.Max)
	} else if vc.Min != "" {
		return fmt.Sprintf(">=%s", vc.Min)
	} else if vc.Max != "" {
		return fmt.Sprintf("<%s", vc.Max)
	}
	return "any"
}

// Registry represents a package registry
type Registry interface {
	// GetPackage retrieves a package by name and version
	GetPackage(name, version string) (*Package, error)
	
	// GetVersions retrieves all available versions for a package
	GetVersions(name string) ([]string, error)
	
	// GetLatestVersion retrieves the latest version for a package
	GetLatestVersion(name string) (string, error)
	
	// Satisfies checks if a version satisfies a constraint
	Satisfies(version string, constraint VersionConstraint) bool
}

// InMemoryRegistry is a simple in-memory implementation of Registry
type InMemoryRegistry struct {
	packages map[string]map[string]*Package
}

// NewInMemoryRegistry creates a new in-memory registry
func NewInMemoryRegistry() *InMemoryRegistry {
	return &InMemoryRegistry{
		packages: make(map[string]map[string]*Package),
	}
}

// AddPackage adds a package to the registry
func (r *InMemoryRegistry) AddPackage(pkg *Package) {
	if r.packages[pkg.Name] == nil {
		r.packages[pkg.Name] = make(map[string]*Package)
	}
	r.packages[pkg.Name][pkg.Version] = pkg
}

// GetPackage retrieves a package by name and version
func (r *InMemoryRegistry) GetPackage(name, version string) (*Package, error) {
	if versions, exists := r.packages[name]; exists {
		if pkg, exists := versions[version]; exists {
			return pkg, nil
		}
	}
	return nil, fmt.Errorf("package %s %s not found", name, version)
}

// GetVersions retrieves all available versions for a package
func (r *InMemoryRegistry) GetVersions(name string) ([]string, error) {
	if versions, exists := r.packages[name]; exists {
		result := make([]string, 0, len(versions))
		for version := range versions {
			result = append(result, version)
		}
		return result, nil
	}
	return nil, fmt.Errorf("package %s not found", name)
}

// GetLatestVersion retrieves the latest version for a package
func (r *InMemoryRegistry) GetLatestVersion(name string) (string, error) {
	versions, err := r.GetVersions(name)
	if err != nil {
		return "", err
	}
	
	if len(versions) == 0 {
		return "", fmt.Errorf("no versions found for package %s", name)
	}
	
	// For simplicity, just return the first version
	// In a real implementation, this would compare versions properly
	return versions[0], nil
}

// Satisfies checks if a version satisfies a constraint
func (r *InMemoryRegistry) Satisfies(version string, constraint VersionConstraint) bool {
	// This is a simplified implementation
	// In a real implementation, this would properly compare semantic versions
	
	if constraint.IsSpecific() {
		return version == constraint.Specific
	}
	
	// For now, just return true for non-specific constraints
	// This is a placeholder implementation
	return true
} 