package buildmeta

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Parser handles parsing and writing of buildmeta.yaml files
type Parser struct {
	filePath string
}

// NewParser creates a new parser for buildmeta.yaml
func NewParser(filePath string) *Parser {
	return &Parser{
		filePath: filePath,
	}
}

// Parse parses a buildmeta.yaml file
func (p *Parser) Parse() (*BuildMeta, error) {
	data, err := os.ReadFile(p.filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read buildmeta.yaml: %w", err)
	}
	
	var buildMeta BuildMeta
	if err := yaml.Unmarshal(data, &buildMeta); err != nil {
		return nil, fmt.Errorf("failed to parse buildmeta.yaml: %w", err)
	}
	
	// Validate the parsed data
	if err := buildMeta.Validate(); err != nil {
		return nil, fmt.Errorf("invalid buildmeta.yaml: %w", err)
	}
	
	return &buildMeta, nil
}

// Write writes a BuildMeta to buildmeta.yaml
func (p *Parser) Write(buildMeta *BuildMeta) error {
	// Validate before writing
	if err := buildMeta.Validate(); err != nil {
		return fmt.Errorf("invalid buildmeta configuration: %w", err)
	}
	
	data, err := yaml.Marshal(buildMeta)
	if err != nil {
		return fmt.Errorf("failed to marshal buildmeta: %w", err)
	}
	
	// Create directory if it doesn't exist
	dir := filepath.Dir(p.filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}
	
	if err := os.WriteFile(p.filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write buildmeta.yaml: %w", err)
	}
	
	return nil
}

// Exists checks if the buildmeta.yaml file exists
func (p *Parser) Exists() bool {
	_, err := os.Stat(p.filePath)
	return err == nil
}

// Remove removes the buildmeta.yaml file
func (p *Parser) Remove() error {
	return os.Remove(p.filePath)
}

// ParseFromDirectory parses buildmeta.yaml from a directory
func ParseFromDirectory(dir string) (*BuildMeta, error) {
	filePath := filepath.Join(dir, "buildmeta.yaml")
	parser := NewParser(filePath)
	return parser.Parse()
}

// WriteToDirectory writes buildmeta.yaml to a directory
func WriteToDirectory(dir string, buildMeta *BuildMeta) error {
	filePath := filepath.Join(dir, "buildmeta.yaml")
	parser := NewParser(filePath)
	return parser.Write(buildMeta)
}

// ConvertFromPyProject converts pyproject.toml to buildmeta.yaml
func ConvertFromPyProject(pyprojectPath string) (*BuildMeta, error) {
	// This is a simplified conversion
	// In a real implementation, you'd parse pyproject.toml and convert it
	
	// For now, create a default buildmeta
	buildMeta := NewBuildMeta("converted-package", "0.1.0")
	buildMeta.Description = "Converted from pyproject.toml"
	
	return buildMeta, nil
}

// ConvertToPyProject converts buildmeta.yaml to pyproject.toml
func ConvertToPyProject(buildMeta *BuildMeta) (string, error) {
	// This is a simplified conversion
	// In a real implementation, you'd generate pyproject.toml content
	
	content := fmt.Sprintf(`[build-system]
requires = ["setuptools>=61.0", "wheel"]
build-backend = "setuptools.build_meta"

[project]
name = "%s"
version = "%s"
description = "%s"
authors = [{name = "%s", email = "%s"}]
license = {text = "%s"}
requires-python = "%s"
`, 
		buildMeta.Name,
		buildMeta.Version,
		buildMeta.Description,
		buildMeta.Author,
		buildMeta.Email,
		buildMeta.License,
		buildMeta.Python.Requires,
	)
	
	// Add dependencies
	if len(buildMeta.Dependencies.Direct) > 0 {
		content += "\ndependencies = [\n"
		for name, constraint := range buildMeta.Dependencies.Direct {
			content += fmt.Sprintf(`    "%s%s",`+"\n", name, constraint)
		}
		content += "]\n"
	}
	
	// Add optional dependencies
	if len(buildMeta.OptionalDependencies) > 0 {
		content += "\n[project.optional-dependencies]\n"
		for group, deps := range buildMeta.OptionalDependencies {
			content += fmt.Sprintf("%s = [\n", group)
			for name, constraint := range deps.Direct {
				content += fmt.Sprintf(`    "%s%s",`+"\n", name, constraint)
			}
			content += "]\n"
		}
	}
	
	// Add entry points
	if len(buildMeta.EntryPoints) > 0 {
		content += "\n[project.entry-points]\n"
		for group, entries := range buildMeta.EntryPoints {
			content += fmt.Sprintf("[project.entry-points.%s]\n", group)
			for name, target := range entries {
				content += fmt.Sprintf(`%s = "%s"`+"\n", name, target)
			}
		}
	}
	
	return content, nil
}

// ValidateFile validates a buildmeta.yaml file
func ValidateFile(filePath string) error {
	parser := NewParser(filePath)
	_, err := parser.Parse()
	return err
}

// CreateDefault creates a default buildmeta.yaml file
func CreateDefault(filePath, name, version string) error {
	buildMeta := NewBuildMeta(name, version)
	parser := NewParser(filePath)
	return parser.Write(buildMeta)
}

// UpdateFromRequirements updates buildmeta.yaml from a requirements.txt file
func UpdateFromRequirements(buildmetaPath, requirementsPath string) error {
	// Parse existing buildmeta
	parser := NewParser(buildmetaPath)
	buildMeta, err := parser.Parse()
	if err != nil {
		return err
	}
	
	// Parse requirements.txt
	requirements, err := parseRequirementsFile(requirementsPath)
	if err != nil {
		return err
	}
	
	// Update dependencies
	for name, constraint := range requirements {
		buildMeta.AddDependency(name, constraint)
	}
	
	// Write updated buildmeta
	return parser.Write(buildMeta)
}

// parseRequirementsFile parses a requirements.txt file
func parseRequirementsFile(filePath string) (map[string]string, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read requirements.txt: %w", err)
	}
	
	requirements := make(map[string]string)
	lines := strings.Split(string(data), "\n")
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		
		// Parse package specification
		parts := strings.SplitN(line, "==", 2)
		if len(parts) == 2 {
			requirements[parts[0]] = "==" + parts[1]
		} else {
			parts = strings.SplitN(line, ">=", 2)
			if len(parts) == 2 {
				requirements[parts[0]] = ">=" + parts[1]
			} else {
				parts = strings.SplitN(line, "<=", 2)
				if len(parts) == 2 {
					requirements[parts[0]] = "<=" + parts[1]
				} else {
					// No version constraint
					requirements[line] = ""
				}
			}
		}
	}
	
	return requirements, nil
}

// ParseRequirementsFile parses a requirements.txt file
func ParseRequirementsFile(filePath string) (map[string]string, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read requirements.txt: %w", err)
	}
	requirements := make(map[string]string)
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "==", 2)
		if len(parts) == 2 {
			requirements[parts[0]] = "==" + parts[1]
		} else {
			parts = strings.SplitN(line, ">=", 2)
			if len(parts) == 2 {
				requirements[parts[0]] = ">=" + parts[1]
			} else {
				parts = strings.SplitN(line, "<=", 2)
				if len(parts) == 2 {
					requirements[parts[0]] = "<=" + parts[1]
				} else {
					requirements[line] = ""
				}
			}
		}
	}
	return requirements, nil
}

// ExportRequirementsFile writes dependencies to requirements.txt
func ExportRequirementsFile(filePath string, deps map[string]string) error {
	var lines []string
	for name, constraint := range deps {
		if constraint != "" {
			lines = append(lines, fmt.Sprintf("%s%s", name, constraint))
		} else {
			lines = append(lines, name)
		}
	}
	content := strings.Join(lines, "\n")
	return os.WriteFile(filePath, []byte(content), 0644)
}

// PyProjectMeta is a minimal struct for pyproject.toml import/export
type PyProjectMeta struct {
	Name         string
	Version      string
	Dependencies map[string]string
}

// ParsePyProjectToml parses pyproject.toml for dependencies (very basic)
func ParsePyProjectToml(filePath string) (*PyProjectMeta, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read pyproject.toml: %w", err)
	}
	meta := &PyProjectMeta{Dependencies: make(map[string]string)}
	lines := strings.Split(string(data), "\n")
	inDeps := false
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "name = ") {
			meta.Name = strings.Trim(line[7:], `"`)
		} else if strings.HasPrefix(line, "version = ") {
			meta.Version = strings.Trim(line[10:], `"`)
		} else if strings.HasPrefix(line, "[project.dependencies]") || strings.HasPrefix(line, "[tool.poetry.dependencies]") {
			inDeps = true
			continue
		} else if strings.HasPrefix(line, "[") && inDeps {
			inDeps = false
		}
		if inDeps && strings.Contains(line, "=") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				name := strings.TrimSpace(parts[0])
				constraint := strings.TrimSpace(parts[1])
				meta.Dependencies[name] = strings.Trim(constraint, `"`)
			}
		}
	}
	return meta, nil
}

// ExportPyProjectToml writes dependencies to pyproject.toml (basic)
func ExportPyProjectToml(filePath string, buildMeta *BuildMeta) error {
	content := fmt.Sprintf(`[project]
name = "%s"
version = "%s"

[project.dependencies]
`, buildMeta.Name, buildMeta.Version)
	for name, constraint := range buildMeta.GetDependencies() {
		if constraint != "" {
			content += fmt.Sprintf("%s = \"%s\"\n", name, constraint)
		} else {
			content += fmt.Sprintf("%s = \"*\"\n", name)
		}
	}
	return os.WriteFile(filePath, []byte(content), 0644)
} 