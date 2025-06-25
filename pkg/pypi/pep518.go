package pypi

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// PEP518BuildSystem represents the build-system section in pyproject.toml
type PEP518BuildSystem struct {
	Requires []string `yaml:"requires"`
	Backend  string   `yaml:"build-backend"`
}

// PEP518Config represents the pyproject.toml configuration
type PEP518Config struct {
	BuildSystem PEP518BuildSystem `yaml:"build-system"`
}

// ParsePEP518Config parses pyproject.toml for PEP 518 build dependencies
func ParsePEP518Config(projectDir string) (*PEP518Config, error) {
	pyprojectPath := filepath.Join(projectDir, "pyproject.toml")
	
	data, err := os.ReadFile(pyprojectPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read pyproject.toml: %w", err)
	}
	
	var config PEP518Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse pyproject.toml: %w", err)
	}
	
	return &config, nil
}

// GetBuildDependencies gets the build dependencies for a project
func GetBuildDependencies(projectDir string) ([]string, error) {
	config, err := ParsePEP518Config(projectDir)
	if err != nil {
		return nil, err
	}
	
	return config.BuildSystem.Requires, nil
}

// GetBuildBackend gets the build backend for a project
func GetBuildBackend(projectDir string) (string, error) {
	config, err := ParsePEP518Config(projectDir)
	if err != nil {
		return "", err
	}
	
	return config.BuildSystem.Backend, nil
}

// ValidateBuildSystem validates the build system configuration
func ValidateBuildSystem(config *PEP518Config) error {
	if config.BuildSystem.Backend == "" {
		return fmt.Errorf("build-backend is required")
	}
	
	if len(config.BuildSystem.Requires) == 0 {
		return fmt.Errorf("build-system.requires cannot be empty")
	}
	
	return nil
}

// DefaultBuildSystem returns the default build system configuration
func DefaultBuildSystem() *PEP518Config {
	return &PEP518Config{
		BuildSystem: PEP518BuildSystem{
			Requires: []string{
				"setuptools>=61.0",
				"wheel",
			},
			Backend: "setuptools.build_meta",
		},
	}
}

// CreateDefaultPyProject creates a default pyproject.toml file
func CreateDefaultPyProject(projectDir string) error {
	config := DefaultBuildSystem()
	
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

// InstallBuildDependencies installs the build dependencies in a virtual environment
func InstallBuildDependencies(projectDir, venvPath string) error {
	deps, err := GetBuildDependencies(projectDir)
	if err != nil {
		return err
	}
	
	// Create a temporary requirements file
	requirementsContent := strings.Join(deps, "\n")
	requirementsPath := filepath.Join(projectDir, "build-requirements.txt")
	
	if err := os.WriteFile(requirementsPath, []byte(requirementsContent), 0644); err != nil {
		return fmt.Errorf("failed to write build requirements: %w", err)
	}
	defer os.Remove(requirementsPath)
	
	// Install dependencies using pip
	pipCmd := filepath.Join(venvPath, "bin", "pip")
	if _, err := os.Stat(pipCmd); os.IsNotExist(err) {
		// Windows path
		pipCmd = filepath.Join(venvPath, "Scripts", "pip.exe")
	}
	
	// TODO: Implement pip install command execution
	// This would use os/exec to run: pip install -r build-requirements.txt
	
	return nil
} 