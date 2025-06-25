package pypi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

// PEP517BuildBackend represents a PEP 517 build backend
type PEP517BuildBackend struct {
	BackendPath string
	BackendName string
}

// BuildRequest represents a PEP 517 build request
type BuildRequest struct {
	SourceDir string
	BuildDir  string
	TargetDir string
	ConfigSettings map[string]interface{}
}

// BuildResponse represents a PEP 517 build response
type BuildResponse struct {
	Artifacts []BuildArtifact
}

// BuildArtifact represents a build artifact
type BuildArtifact struct {
	Path string
	Type string
}

// NewPEP517BuildBackend creates a new PEP 517 build backend
func NewPEP517BuildBackend(backendPath, backendName string) *PEP517BuildBackend {
	return &PEP517BuildBackend{
		BackendPath: backendPath,
		BackendName: backendName,
	}
}

// BuildWheel builds a wheel using the PEP 517 backend
func (b *PEP517BuildBackend) BuildWheel(req BuildRequest) (*BuildResponse, error) {
	// Create the build request JSON
	buildReq := map[string]interface{}{
		"source_dir": req.SourceDir,
		"build_dir":  req.BuildDir,
		"target_dir": req.TargetDir,
		"config_settings": req.ConfigSettings,
	}
	
	reqJSON, err := json.Marshal(buildReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal build request: %w", err)
	}
	
	// Execute the build backend
	cmd := exec.Command("python", "-m", "pep517.build", "wheel")
	cmd.Dir = req.SourceDir
	cmd.Stdin = bytes.NewReader(reqJSON)
	
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("build failed: %w, output: %s", err, string(output))
	}
	
	// Parse the response
	var response BuildResponse
	if err := json.Unmarshal(output, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal build response: %w", err)
	}
	
	return &response, nil
}

// BuildSdist builds a source distribution using the PEP 517 backend
func (b *PEP517BuildBackend) BuildSdist(req BuildRequest) (*BuildResponse, error) {
	// Create the build request JSON
	buildReq := map[string]interface{}{
		"source_dir": req.SourceDir,
		"build_dir":  req.BuildDir,
		"target_dir": req.TargetDir,
		"config_settings": req.ConfigSettings,
	}
	
	reqJSON, err := json.Marshal(buildReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal build request: %w", err)
	}
	
	// Execute the build backend
	cmd := exec.Command("python", "-m", "pep517.build", "sdist")
	cmd.Dir = req.SourceDir
	cmd.Stdin = bytes.NewReader(reqJSON)
	
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("build failed: %w, output: %s", err, string(output))
	}
	
	// Parse the response
	var response BuildResponse
	if err := json.Unmarshal(output, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal build response: %w", err)
	}
	
	return &response, nil
}

// GetRequiresForBuildWheel gets the requirements for building a wheel
func (b *PEP517BuildBackend) GetRequiresForBuildWheel(sourceDir string) ([]string, error) {
	cmd := exec.Command("python", "-m", "pep517.meta", "get_requires_for_build_wheel")
	cmd.Dir = sourceDir
	
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get wheel build requirements: %w", err)
	}
	
	requirements := strings.Split(strings.TrimSpace(string(output)), "\n")
	var result []string
	for _, req := range requirements {
		if req != "" {
			result = append(result, req)
		}
	}
	
	return result, nil
}

// GetRequiresForBuildSdist gets the requirements for building a source distribution
func (b *PEP517BuildBackend) GetRequiresForBuildSdist(sourceDir string) ([]string, error) {
	cmd := exec.Command("python", "-m", "pep517.meta", "get_requires_for_build_sdist")
	cmd.Dir = sourceDir
	
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get sdist build requirements: %w", err)
	}
	
	requirements := strings.Split(strings.TrimSpace(string(output)), "\n")
	var result []string
	for _, req := range requirements {
		if req != "" {
			result = append(result, req)
		}
	}
	
	return result, nil
}

// PrepareMetadataForBuildWheel prepares metadata for building a wheel
func (b *PEP517BuildBackend) PrepareMetadataForBuildWheel(sourceDir, metadataDir string) (string, error) {
	cmd := exec.Command("python", "-m", "pep517.meta", "prepare_metadata_for_build_wheel")
	cmd.Dir = sourceDir
	cmd.Env = append(cmd.Env, fmt.Sprintf("PEP517_METADATA_DIR=%s", metadataDir))
	
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to prepare metadata: %w", err)
	}
	
	return strings.TrimSpace(string(output)), nil
} 