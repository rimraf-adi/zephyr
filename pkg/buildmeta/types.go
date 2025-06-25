package buildmeta

import (
	"fmt"
	"time"
)

// BuildMeta represents the buildmeta.yaml structure
type BuildMeta struct {
	Version     string            `yaml:"version"`
	Name        string            `yaml:"name"`
	Description string            `yaml:"description,omitempty"`
	Author      string            `yaml:"author,omitempty"`
	Email       string            `yaml:"email,omitempty"`
	License     string            `yaml:"license,omitempty"`
	Homepage    string            `yaml:"homepage,omitempty"`
	Repository  string            `yaml:"repository,omitempty"`
	Keywords    []string          `yaml:"keywords,omitempty"`
	Classifiers []string          `yaml:"classifiers,omitempty"`
	
	// Python-specific fields
	Python      PythonConfig      `yaml:"python"`
	Build       BuildConfig       `yaml:"build"`
	Dependencies DependenciesConfig `yaml:"dependencies"`
	DevDependencies DependenciesConfig `yaml:"dev-dependencies,omitempty"`
	OptionalDependencies map[string]DependenciesConfig `yaml:"optional-dependencies,omitempty"`
	
	// Scripts and entry points
	Scripts     map[string]string `yaml:"scripts,omitempty"`
	EntryPoints map[string]map[string]string `yaml:"entry-points,omitempty"`
	
	// Metadata
	Created     time.Time         `yaml:"created,omitempty"`
	Updated     time.Time         `yaml:"updated,omitempty"`
	Maintainers []Maintainer      `yaml:"maintainers,omitempty"`
}

// PythonConfig represents Python-specific configuration
type PythonConfig struct {
	Requires     string   `yaml:"requires,omitempty"`
	Exclude      []string `yaml:"exclude,omitempty"`
	Include      []string `yaml:"include,omitempty"`
	Packages     []string `yaml:"packages,omitempty"`
	PyModules    []string `yaml:"py-modules,omitempty"`
	DataFiles    []DataFile `yaml:"data-files,omitempty"`
}

// BuildConfig represents build configuration
type BuildConfig struct {
	Backend     string            `yaml:"backend,omitempty"`
	BackendPath string            `yaml:"backend-path,omitempty"`
	Scripts     map[string]string `yaml:"scripts,omitempty"`
	Config      map[string]interface{} `yaml:"config,omitempty"`
}

// DependenciesConfig represents dependencies configuration
type DependenciesConfig struct {
	Direct      map[string]string `yaml:"direct,omitempty"`
	Transitive  map[string]string `yaml:"transitive,omitempty"`
	Groups      map[string][]string `yaml:"groups,omitempty"`
	Platform    map[string]map[string]string `yaml:"platform,omitempty"`
}

// DataFile represents a data file entry
type DataFile struct {
	Source      string   `yaml:"source"`
	Destination string   `yaml:"destination"`
	Pattern     string   `yaml:"pattern,omitempty"`
}

// Maintainer represents a maintainer
type Maintainer struct {
	Name  string `yaml:"name"`
	Email string `yaml:"email,omitempty"`
}

// NewBuildMeta creates a new BuildMeta with default values
func NewBuildMeta(name, version string) *BuildMeta {
	return &BuildMeta{
		Version:     version,
		Name:        name,
		Description: "A Python package",
		Python: PythonConfig{
			Requires: ">=3.8",
		},
		Build: BuildConfig{
			Backend: "setuptools.build_meta",
		},
		Dependencies: DependenciesConfig{
			Direct: make(map[string]string),
		},
		DevDependencies: DependenciesConfig{
			Direct: make(map[string]string),
		},
		OptionalDependencies: make(map[string]DependenciesConfig),
		Scripts:             make(map[string]string),
		EntryPoints:         make(map[string]map[string]string),
		Maintainers:         []Maintainer{},
		Created:             time.Now(),
		Updated:             time.Now(),
	}
}

// AddDependency adds a dependency to the main dependencies
func (bm *BuildMeta) AddDependency(name, constraint string) {
	if bm.Dependencies.Direct == nil {
		bm.Dependencies.Direct = make(map[string]string)
	}
	bm.Dependencies.Direct[name] = constraint
	bm.Updated = time.Now()
}

// AddDevDependency adds a development dependency
func (bm *BuildMeta) AddDevDependency(name, constraint string) {
	if bm.DevDependencies.Direct == nil {
		bm.DevDependencies.Direct = make(map[string]string)
	}
	bm.DevDependencies.Direct[name] = constraint
	bm.Updated = time.Now()
}

// AddOptionalDependency adds an optional dependency group
func (bm *BuildMeta) AddOptionalDependency(group, name, constraint string) {
	if bm.OptionalDependencies == nil {
		bm.OptionalDependencies = make(map[string]DependenciesConfig)
	}
	
	if bm.OptionalDependencies[group].Direct == nil {
		bm.OptionalDependencies[group] = DependenciesConfig{
			Direct: make(map[string]string),
		}
	}
	
	bm.OptionalDependencies[group].Direct[name] = constraint
	bm.Updated = time.Now()
}

// RemoveDependency removes a dependency
func (bm *BuildMeta) RemoveDependency(name string) {
	if bm.Dependencies.Direct != nil {
		delete(bm.Dependencies.Direct, name)
		bm.Updated = time.Now()
	}
}

// RemoveDevDependency removes a development dependency
func (bm *BuildMeta) RemoveDevDependency(name string) {
	if bm.DevDependencies.Direct != nil {
		delete(bm.DevDependencies.Direct, name)
		bm.Updated = time.Now()
	}
}

// GetDependencies returns all direct dependencies
func (bm *BuildMeta) GetDependencies() map[string]string {
	if bm.Dependencies.Direct == nil {
		return make(map[string]string)
	}
	return bm.Dependencies.Direct
}

// GetDevDependencies returns all development dependencies
func (bm *BuildMeta) GetDevDependencies() map[string]string {
	if bm.DevDependencies.Direct == nil {
		return make(map[string]string)
	}
	return bm.DevDependencies.Direct
}

// GetOptionalDependencies returns optional dependencies for a group
func (bm *BuildMeta) GetOptionalDependencies(group string) map[string]string {
	if bm.OptionalDependencies == nil {
		return make(map[string]string)
	}
	
	if deps, exists := bm.OptionalDependencies[group]; exists && deps.Direct != nil {
		return deps.Direct
	}
	
	return make(map[string]string)
}

// AddScript adds a script entry
func (bm *BuildMeta) AddScript(name, command string) {
	if bm.Scripts == nil {
		bm.Scripts = make(map[string]string)
	}
	bm.Scripts[name] = command
	bm.Updated = time.Now()
}

// AddEntryPoint adds an entry point
func (bm *BuildMeta) AddEntryPoint(group, name, target string) {
	if bm.EntryPoints == nil {
		bm.EntryPoints = make(map[string]map[string]string)
	}
	
	if bm.EntryPoints[group] == nil {
		bm.EntryPoints[group] = make(map[string]string)
	}
	
	bm.EntryPoints[group][name] = target
	bm.Updated = time.Now()
}

// AddMaintainer adds a maintainer
func (bm *BuildMeta) AddMaintainer(name, email string) {
	maintainer := Maintainer{
		Name:  name,
		Email: email,
	}
	bm.Maintainers = append(bm.Maintainers, maintainer)
	bm.Updated = time.Now()
}

// SetPythonRequirement sets the Python version requirement
func (bm *BuildMeta) SetPythonRequirement(requirement string) {
	bm.Python.Requires = requirement
	bm.Updated = time.Now()
}

// AddPackage adds a package to include
func (bm *BuildMeta) AddPackage(pkg string) {
	bm.Python.Packages = append(bm.Python.Packages, pkg)
	bm.Updated = time.Now()
}

// AddPyModule adds a Python module
func (bm *BuildMeta) AddPyModule(module string) {
	bm.Python.PyModules = append(bm.Python.PyModules, module)
	bm.Updated = time.Now()
}

// AddDataFile adds a data file
func (bm *BuildMeta) AddDataFile(source, destination string) {
	dataFile := DataFile{
		Source:      source,
		Destination: destination,
	}
	bm.Python.DataFiles = append(bm.Python.DataFiles, dataFile)
	bm.Updated = time.Now()
}

// Validate validates the BuildMeta configuration
func (bm *BuildMeta) Validate() error {
	if bm.Name == "" {
		return fmt.Errorf("name is required")
	}
	
	if bm.Version == "" {
		return fmt.Errorf("version is required")
	}
	
	// Validate package name format
	if !isValidPackageName(bm.Name) {
		return fmt.Errorf("invalid package name: %s", bm.Name)
	}
	
	return nil
}

// isValidPackageName checks if a package name is valid
func isValidPackageName(name string) bool {
	if name == "" {
		return false
	}
	
	// Basic validation - package names should be lowercase, alphanumeric, with hyphens/underscores
	for _, char := range name {
		if !((char >= 'a' && char <= 'z') || (char >= '0' && char <= '9') || char == '-' || char == '_') {
			return false
		}
	}
	
	return true
} 