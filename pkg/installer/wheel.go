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
	createdPaths := []string{}
	sitePackages := wi.getSitePackagesPath()
	if err := wi.extractWheel(reader, sitePackages, metadata, &createdPaths); err != nil {
		wi.rollbackCreatedPaths(createdPaths)
		return fmt.Errorf("failed to extract wheel '%s' to site-packages: %w. Check permissions and disk space.", wheelPath, err)
	}
	if err := wi.installMetadata(sitePackages, metadata, &createdPaths); err != nil {
		wi.rollbackCreatedPaths(createdPaths)
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

// Helper for atomic install: track created dirs
func trackMkdirAll(path string, perm os.FileMode, createdPaths *[]string) error {
	err := os.MkdirAll(path, perm)
	if err == nil {
		*createdPaths = append(*createdPaths, path)
	}
	return err
}

// Helper for atomic install: track created files
func trackCreateFile(path string, createdPaths *[]string) (*os.File, error) {
	f, err := os.Create(path)
	if err == nil {
		*createdPaths = append(*createdPaths, path)
	}
	return f, err
}

// extractWheel extracts wheel contents to site-packages
func (wi *WheelInstaller) extractWheel(reader *zip.ReadCloser, sitePackages string, metadata *WheelMetadata, createdPaths *[]string) error {
	for _, file := range reader.File {
		if strings.Contains(file.Name, ".dist-info/") {
			continue
		}
		targetPath := filepath.Join(sitePackages, file.Name)
		if file.FileInfo().IsDir() {
			if err := trackMkdirAll(targetPath, 0755, createdPaths); err != nil {
				return fmt.Errorf("failed to create directory '%s': %w. Check permissions.", targetPath, err)
			}
			continue
		}
		parentDir := filepath.Dir(targetPath)
		if err := trackMkdirAll(parentDir, 0755, createdPaths); err != nil {
			return fmt.Errorf("failed to create parent directory '%s': %w. Check permissions.", parentDir, err)
		}
		if err := wi.extractFileTracked(file, targetPath, createdPaths); err != nil {
			return fmt.Errorf("failed to extract file '%s' to '%s': %w. Check disk space and permissions.", file.Name, targetPath, err)
		}
	}
	return nil
}

// extractFile extracts a single file from the wheel
func (wi *WheelInstaller) extractFileTracked(file *zip.File, targetPath string, createdPaths *[]string) error {
	rc, err := file.Open()
	if err != nil {
		return fmt.Errorf("failed to open file in wheel: %w. The wheel may be corrupted.", err)
	}
	defer rc.Close()
	targetFile, err := trackCreateFile(targetPath, createdPaths)
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
func (wi *WheelInstaller) installMetadata(sitePackages string, metadata *WheelMetadata, createdPaths *[]string) error {
	distInfoDir := filepath.Join(sitePackages, metadata.DistInfoName)
	if err := trackMkdirAll(distInfoDir, 0755, createdPaths); err != nil {
		return fmt.Errorf("failed to create dist-info directory '%s': %w. Check permissions.", distInfoDir, err)
	}
	metadataPath := filepath.Join(distInfoDir, "METADATA")
	f, err := trackCreateFile(metadataPath, createdPaths)
	if err != nil {
		return fmt.Errorf("failed to write METADATA file '%s': %w. Check permissions and disk space.", metadataPath, err)
	}
	f.Write([]byte(metadata.RawMetadata))
	f.Close()
	wheelPath := filepath.Join(distInfoDir, "WHEEL")
	f, err = trackCreateFile(wheelPath, createdPaths)
	if err != nil {
		return fmt.Errorf("failed to write WHEEL file '%s': %w. Check permissions and disk space.", wheelPath, err)
	}
	f.Write([]byte(metadata.WheelInfo))
	f.Close()
	recordPath := filepath.Join(distInfoDir, "RECORD")
	recordContent := wi.generateRecordFile(sitePackages, metadata)
	f, err = trackCreateFile(recordPath, createdPaths)
	if err != nil {
		return fmt.Errorf("failed to write RECORD file '%s': %w. Check permissions and disk space.", recordPath, err)
	}
	f.Write([]byte(recordContent))
	f.Close()
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

// Helper to rollback created files/dirs
func (wi *WheelInstaller) rollbackCreatedPaths(createdPaths []string) {
	for i := len(createdPaths) - 1; i >= 0; i-- {
		os.RemoveAll(createdPaths[i])
	}
}

// InstallWheelFromPyPI downloads and installs a wheel from PyPI with atomic rollback and hash verification
func (wi *WheelInstaller) InstallWheelFromPyPI(packageName, version string) error {
	client := pypi.NewPyPIClient()
	release, err := client.FindWheelForVersion(packageName, version, "any")
	if err != nil {
		return fmt.Errorf("failed to find wheel: %w", err)
	}
	reader, err := client.DownloadRelease(*release)
	if err != nil {
		return fmt.Errorf("failed to download wheel: %w", err)
	}
	defer reader.Close()
	tempFile, err := os.CreateTemp("", "wheel-*.whl")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tempFile.Name())
	defer tempFile.Close()
	hasher := sha256.New()
	multiWriter := io.MultiWriter(tempFile, hasher)
	if _, err := io.Copy(multiWriter, reader); err != nil {
		return fmt.Errorf("failed to write temp file: %w", err)
	}
	if release.Digests.SHA256 != "" {
		actualHash := hex.EncodeToString(hasher.Sum(nil))
		if !strings.EqualFold(actualHash, release.Digests.SHA256) {
			return fmt.Errorf("SHA256 hash mismatch for %s: expected %s, got %s", packageName, release.Digests.SHA256, actualHash)
		}
	}
	createdPaths := []string{}
	err = wi.InstallWheelTracked(tempFile.Name(), packageName, &createdPaths)
	if err != nil {
		wi.rollbackCreatedPaths(createdPaths)
		return fmt.Errorf("atomic install failed, rolled back: %w", err)
	}
	return nil
}

// InstallWheelTracked is like InstallWheel but takes createdPaths for rollback
func (wi *WheelInstaller) InstallWheelTracked(wheelPath, packageName string, createdPaths *[]string) error {
	reader, err := zip.OpenReader(wheelPath)
	if err != nil {
		return fmt.Errorf("failed to open wheel file '%s': %w. Ensure the file exists and is a valid .whl archive.", wheelPath, err)
	}
	defer reader.Close()
	metadata, err := wi.parseWheelMetadata(reader)
	if err != nil {
		return fmt.Errorf("failed to parse wheel metadata for '%s': %w. The wheel may be corrupted or missing METADATA.", wheelPath, err)
	}
	sitePackages := wi.getSitePackagesPath()
	if err := wi.extractWheel(reader, sitePackages, metadata, createdPaths); err != nil {
		return err
	}
	if err := wi.installMetadata(sitePackages, metadata, createdPaths); err != nil {
		return err
	}
	return nil
} 