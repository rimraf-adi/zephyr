package pypi

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// PEP621Project represents the project metadata section in pyproject.toml
type PEP621Project struct {
	Name         string            `yaml:"name"`
	Version      string            `yaml:"version"`
	Description  string            `yaml:"description,omitempty"`
	Readme       string            `yaml:"readme,omitempty"`
	RequiresPython string          `yaml:"requires-python,omitempty"`
	License      PEP621License     `yaml:"license,omitempty"`
	Authors      []PEP621Author    `yaml:"authors,omitempty"`
	Maintainers  []PEP621Author    `yaml:"maintainers,omitempty"`
	Keywords     []string          `yaml:"keywords,omitempty"`
	Classifiers  []string          `yaml:"classifiers,omitempty"`
	Dependencies map[string]string `yaml:"dependencies,omitempty"`
	OptionalDependencies map[string]map[string]string `yaml:"optional-dependencies,omitempty"`
	URLs         map[string]string `yaml:"urls,omitempty"`
	EntryPoints  map[string]map[string]string `yaml:"entry-points,omitempty"`
}

// PEP621Author represents an author or maintainer
type PEP621Author struct {
	Name  string `yaml:"name"`
	Email string `yaml:"email,omitempty"`
}

// PEP621License represents license information
type PEP621License struct {
	Text string `yaml:"text,omitempty"`
	File string `yaml:"file,omitempty"`
}

// PEP621Config represents the complete pyproject.toml configuration
type PEP621Config struct {
	Project PEP621Project `yaml:"project"`
}

// ParsePEP621Config parses pyproject.toml for PEP 621 project metadata
func ParsePEP621Config(projectDir string) (*PEP621Config, error) {
	pyprojectPath := filepath.Join(projectDir, "pyproject.toml")
	
	data, err := os.ReadFile(pyprojectPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read pyproject.toml: %w", err)
	}
	
	var config PEP621Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse pyproject.toml: %w", err)
	}
	
	return &config, nil
}

// GetProjectName gets the project name from pyproject.toml
func GetProjectName(projectDir string) (string, error) {
	config, err := ParsePEP621Config(projectDir)
	if err != nil {
		return "", err
	}
	
	return config.Project.Name, nil
}

// GetProjectVersion gets the project version from pyproject.toml
func GetProjectVersion(projectDir string) (string, error) {
	config, err := ParsePEP621Config(projectDir)
	if err != nil {
		return "", err
	}
	
	return config.Project.Version, nil
}

// GetProjectDependencies gets the project dependencies from pyproject.toml
func GetProjectDependencies(projectDir string) (map[string]string, error) {
	config, err := ParsePEP621Config(projectDir)
	if err != nil {
		return nil, err
	}
	
	return config.Project.Dependencies, nil
}

// GetOptionalDependencies gets the optional dependencies from pyproject.toml
func GetOptionalDependencies(projectDir string) (map[string]map[string]string, error) {
	config, err := ParsePEP621Config(projectDir)
	if err != nil {
		return nil, err
	}
	
	return config.Project.OptionalDependencies, nil
}

// ValidateProject validates the project metadata
func ValidateProject(config *PEP621Config) error {
	if config.Project.Name == "" {
		return fmt.Errorf("project name is required")
	}
	
	if config.Project.Version == "" {
		return fmt.Errorf("project version is required")
	}
	
	// Validate name format (PEP 508)
	if !isValidPackageName(config.Project.Name) {
		return fmt.Errorf("invalid package name: %s", config.Project.Name)
	}
	
	return nil
}

// isValidPackageName checks if a package name is valid according to PEP 508
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

// CreateDefaultProject creates a default project configuration
func CreateDefaultProject(name, version string) *PEP621Config {
	return &PEP621Config{
		Project: PEP621Project{
			Name:    name,
			Version: version,
			Description: "A Python project",
			RequiresPython: ">=3.8",
			Authors: []PEP621Author{
				{Name: "Your Name", Email: "your.email@example.com"},
			},
			Dependencies: make(map[string]string),
			OptionalDependencies: make(map[string]map[string]string),
			URLs: make(map[string]string),
		},
	}
}

// WritePEP621Config writes a PEP 621 configuration to pyproject.toml
func WritePEP621Config(projectDir string, config *PEP621Config) error {
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}
	
	pyprojectPath := filepath.Join(projectDir, "pyproject.toml")
	if err := os.WriteFile(pyprojectPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write pyproject.toml: %w", err)
	}
	
	return nil
}

// AddDependency adds a dependency to the project
func AddDependency(projectDir, packageName, versionConstraint string) error {
	config, err := ParsePEP621Config(projectDir)
	if err != nil {
		return err
	}
	
	if config.Project.Dependencies == nil {
		config.Project.Dependencies = make(map[string]string)
	}
	
	config.Project.Dependencies[packageName] = versionConstraint
	
	return WritePEP621Config(projectDir, config)
}

// RemoveDependency removes a dependency from the project
func RemoveDependency(projectDir, packageName string) error {
	config, err := ParsePEP621Config(projectDir)
	if err != nil {
		return err
	}
	
	delete(config.Project.Dependencies, packageName)
	
	return WritePEP621Config(projectDir, config)
} 