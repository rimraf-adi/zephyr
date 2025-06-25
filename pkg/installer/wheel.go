package installer

import (
	"archive/zip"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"rimraf-adi.com/zephyr/pkg/pypi"
)

// WheelInstaller handles wheel file installation
type WheelInstaller struct {
	venvPath string
}

// NewWheelInstaller creates a new wheel installer
func NewWheelInstaller(venvPath string) *WheelInstaller {
	return &WheelInstaller{
		venvPath: venvPath,
	}
}

// InstallWheel installs a wheel file into the virtual environment
func (wi *WheelInstaller) InstallWheel(wheelPath, packageName string) error {
	reader, err := zip.OpenReader(wheelPath)
	if err != nil {
		return fmt.Errorf("failed to open wheel file '%s': %w. Ensure the file exists and is a valid .whl archive.", wheelPath, err)
	}
	defer reader.Close()
	metadata, err := wi.parseWheelMetadata(reader)
	if err != nil {
		return fmt.Errorf("failed to parse wheel metadata for '%s': %w. The wheel may be corrupted or missing METADATA.", wheelPath, err)
	}
	
	// Determine installation directory
	sitePackages := wi.getSitePackagesPath()
	
	// Extract wheel contents
	if err := wi.extractWheel(reader, sitePackages, metadata); err != nil {
		return fmt.Errorf("failed to extract wheel '%s' to site-packages: %w. Check permissions and disk space.", wheelPath, err)
	}
	
	// Install metadata
	if err := wi.installMetadata(sitePackages, metadata); err != nil {
		return fmt.Errorf("failed to install metadata for '%s': %w. The wheel may be malformed.", wheelPath, err)
	}
	
	return nil
}

// parseWheelMetadata parses metadata from wheel file
func (wi *WheelInstaller) parseWheelMetadata(reader *zip.ReadCloser) (*WheelMetadata, error) {
	metadata := &WheelMetadata{}
	
	// Look for METADATA file
	for _, file := range reader.File {
		if strings.HasSuffix(file.Name, ".dist-info/METADATA") {
			rc, err := file.Open()
			if err != nil {
				return nil, err
			}
			defer rc.Close()
			
			content, err := io.ReadAll(rc)
			if err != nil {
				return nil, err
			}
			
			metadata.RawMetadata = string(content)
			metadata.parseMetadata()
			break
		}
	}
	
	// Look for WHEEL file
	for _, file := range reader.File {
		if strings.HasSuffix(file.Name, ".dist-info/WHEEL") {
			rc, err := file.Open()
			if err != nil {
				return nil, err
			}
			defer rc.Close()
			
			content, err := io.ReadAll(rc)
			if err != nil {
				return nil, err
			}
			
			metadata.WheelInfo = string(content)
			break
		}
	}
	
	return metadata, nil
}

// extractWheel extracts wheel contents to site-packages
func (wi *WheelInstaller) extractWheel(reader *zip.ReadCloser, sitePackages string, metadata *WheelMetadata) error {
	for _, file := range reader.File {
		// Skip metadata files (they're handled separately)
		if strings.Contains(file.Name, ".dist-info/") {
			continue
		}
		
		// Determine target path
		targetPath := filepath.Join(sitePackages, file.Name)
		
		// Create directory if needed
		if file.FileInfo().IsDir() {
			if err := os.MkdirAll(targetPath, 0755); err != nil {
				return fmt.Errorf("failed to create directory '%s': %w. Check permissions.", targetPath, err)
			}
			continue
		}
		
		// Create parent directory
		parentDir := filepath.Dir(targetPath)
		if err := os.MkdirAll(parentDir, 0755); err != nil {
			return fmt.Errorf("failed to create parent directory '%s': %w. Check permissions.", parentDir, err)
		}
		
		// Extract file
		if err := wi.extractFile(file, targetPath); err != nil {
			return fmt.Errorf("failed to extract file '%s' to '%s': %w. Check disk space and permissions.", file.Name, targetPath, err)
		}
	}
	
	return nil
}

// extractFile extracts a single file from the wheel
func (wi *WheelInstaller) extractFile(file *zip.File, targetPath string) error {
	rc, err := file.Open()
	if err != nil {
		return fmt.Errorf("failed to open file in wheel: %w. The wheel may be corrupted.", err)
	}
	defer rc.Close()
	
	targetFile, err := os.Create(targetPath)
	if err != nil {
		return fmt.Errorf("failed to create file '%s': %w. Check permissions and disk space.", targetPath, err)
	}
	defer targetFile.Close()
	
	_, err = io.Copy(targetFile, rc)
	if err != nil {
		return fmt.Errorf("failed to copy data to '%s': %w. Check disk space.", targetPath, err)
	}
	
	return nil
}

// installMetadata installs wheel metadata
func (wi *WheelInstaller) installMetadata(sitePackages string, metadata *WheelMetadata) error {
	// Create dist-info directory
	distInfoDir := filepath.Join(sitePackages, metadata.DistInfoName)
	if err := os.MkdirAll(distInfoDir, 0755); err != nil {
		return fmt.Errorf("failed to create dist-info directory '%s': %w. Check permissions.", distInfoDir, err)
	}
	
	// Write METADATA file
	metadataPath := filepath.Join(distInfoDir, "METADATA")
	if err := os.WriteFile(metadataPath, []byte(metadata.RawMetadata), 0644); err != nil {
		return fmt.Errorf("failed to write METADATA file '%s': %w. Check permissions and disk space.", metadataPath, err)
	}
	
	// Write WHEEL file
	wheelPath := filepath.Join(distInfoDir, "WHEEL")
	if err := os.WriteFile(wheelPath, []byte(metadata.WheelInfo), 0644); err != nil {
		return fmt.Errorf("failed to write WHEEL file '%s': %w. Check permissions and disk space.", wheelPath, err)
	}
	
	// Write RECORD file (simplified)
	recordPath := filepath.Join(distInfoDir, "RECORD")
	recordContent := wi.generateRecordFile(sitePackages, metadata)
	if err := os.WriteFile(recordPath, []byte(recordContent), 0644); err != nil {
		return fmt.Errorf("failed to write RECORD file '%s': %w. Check permissions and disk space.", recordPath, err)
	}
	
	return nil
}

// generateRecordFile generates a RECORD file for the wheel
func (wi *WheelInstaller) generateRecordFile(sitePackages string, metadata *WheelMetadata) string {
	// This is a simplified implementation
	// A real implementation would calculate hashes and include all files
	var lines []string
	
	// Add metadata files
	lines = append(lines, fmt.Sprintf("%s/METADATA,sha256=...,%d", metadata.DistInfoName, len(metadata.RawMetadata)))
	lines = append(lines, fmt.Sprintf("%s/WHEEL,sha256=...,%d", metadata.DistInfoName, len(metadata.WheelInfo)))
	lines = append(lines, fmt.Sprintf("%s/RECORD,sha256=...,%d", metadata.DistInfoName, 0))
	
	return strings.Join(lines, "\n")
}

// getSitePackagesPath returns the site-packages path for the virtual environment
func (wi *WheelInstaller) getSitePackagesPath() string {
	// Determine Python version (simplified)
	pythonVersion := "3.11" // This should be detected from the venv
	
	// Construct site-packages path
	sitePackages := filepath.Join(wi.venvPath, "lib", "python"+pythonVersion, "site-packages")
	
	// Create directory if it doesn't exist
	if err := os.MkdirAll(sitePackages, 0755); err != nil {
		// Fallback to a simpler path
		sitePackages = filepath.Join(wi.venvPath, "site-packages")
		os.MkdirAll(sitePackages, 0755)
	}
	
	return sitePackages
}

// WheelMetadata represents wheel metadata
type WheelMetadata struct {
	Name         string
	Version      string
	Summary      string
	Description  string
	Author       string
	AuthorEmail  string
	License      string
	RequiresDist []string
	RawMetadata  string
	WheelInfo    string
	DistInfoName string
}

// parseMetadata parses the raw metadata string
func (wm *WheelMetadata) parseMetadata() {
	lines := strings.Split(wm.RawMetadata, "\n")
	
	for _, line := range lines {
		if strings.HasPrefix(line, "Name: ") {
			wm.Name = strings.TrimSpace(strings.TrimPrefix(line, "Name: "))
		} else if strings.HasPrefix(line, "Version: ") {
			wm.Version = strings.TrimSpace(strings.TrimPrefix(line, "Version: "))
		} else if strings.HasPrefix(line, "Summary: ") {
			wm.Summary = strings.TrimSpace(strings.TrimPrefix(line, "Summary: "))
		} else if strings.HasPrefix(line, "Author: ") {
			wm.Author = strings.TrimSpace(strings.TrimPrefix(line, "Author: "))
		} else if strings.HasPrefix(line, "Author-email: ") {
			wm.AuthorEmail = strings.TrimSpace(strings.TrimPrefix(line, "Author-email: "))
		} else if strings.HasPrefix(line, "License: ") {
			wm.License = strings.TrimSpace(strings.TrimPrefix(line, "License: "))
		} else if strings.HasPrefix(line, "Requires-Dist: ") {
			req := strings.TrimSpace(strings.TrimPrefix(line, "Requires-Dist: "))
			wm.RequiresDist = append(wm.RequiresDist, req)
		}
	}
	
	// Generate dist-info name
	if wm.Name != "" && wm.Version != "" {
		wm.DistInfoName = fmt.Sprintf("%s-%s.dist-info", wm.Name, wm.Version)
	}
}

// InstallWheelFromPyPI downloads and installs a wheel from PyPI with atomic rollback and hash verification
func (wi *WheelInstaller) InstallWheelFromPyPI(packageName, version string) error {
	client := pypi.NewPyPIClient()
	release, err := client.FindWheelForVersion(packageName, version, "any")
	if err != nil {
		return fmt.Errorf("failed to find wheel: %w", err)
	}

	// Download wheel
	reader, err := client.DownloadRelease(*release)
	if err != nil {
		return fmt.Errorf("failed to download wheel: %w", err)
	}
	defer reader.Close()

	// Create temporary file
	tempFile, err := os.CreateTemp("", "wheel-*.whl")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tempFile.Name())
	defer tempFile.Close()

	// Write downloaded content to temp file and compute SHA256
	hasher := sha256.New()
	multiWriter := io.MultiWriter(tempFile, hasher)
	if _, err := io.Copy(multiWriter, reader); err != nil {
		return fmt.Errorf("failed to write temp file: %w", err)
	}

	// Verify SHA256 hash if available
	if release.Digests.SHA256 != "" {
		actualHash := hex.EncodeToString(hasher.Sum(nil))
		if !strings.EqualFold(actualHash, release.Digests.SHA256) {
			return fmt.Errorf("SHA256 hash mismatch for %s: expected %s, got %s", packageName, release.Digests.SHA256, actualHash)
		}
	}

	// Atomic install: track created files/dirs
	createdPaths := []string{}
	rollback := func() {
		for i := len(createdPaths) - 1; i >= 0; i-- {
			os.RemoveAll(createdPaths[i])
		}
	}

	// Wrap InstallWheel to track created files
	origMkdirAll := os.MkdirAll
	os.MkdirAll = func(path string, perm os.FileMode) error {
		err := origMkdirAll(path, perm)
		if err == nil {
			createdPaths = append(createdPaths, path)
		}
		return err
	}
	origCreate := os.Create
	os.Create = func(path string) (*os.File, error) {
		f, err := origCreate(path)
		if err == nil {
			createdPaths = append(createdPaths, path)
		}
		return f, err
	}
	defer func() {
		os.MkdirAll = origMkdirAll
		os.Create = origCreate
	}()

	err = wi.InstallWheel(tempFile.Name(), packageName)
	if err != nil {
		rollback()
		return fmt.Errorf("atomic install failed, rolled back: %w", err)
	}

	return nil
} 