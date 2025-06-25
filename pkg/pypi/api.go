package pypi

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"rimraf-adi.com/zephyr/pkg/netutil"
)

const (
	PyPIBaseURL     = "https://pypi.org"
	PyPIJSONEndpoint = "/pypi/%s/json"
	PyPISimpleEndpoint = "/simple/%s/"
)

// PyPIMetadata represents the JSON response from PyPI
type PyPIMetadata struct {
	Info     PackageInfo     `json:"info"`
	Releases map[string][]Release `json:"releases"`
	URLs     []Release       `json:"urls"`
}

// PackageInfo contains basic package information
type PackageInfo struct {
	Name         string   `json:"name"`
	Version      string   `json:"version"`
	Summary      string   `json:"summary"`
	Description  string   `json:"description"`
	Author       string   `json:"author"`
	AuthorEmail  string   `json:"author_email"`
	License      string   `json:"license"`
	HomePage     string   `json:"home_page"`
	ProjectURL   string   `json:"project_url"`
	RequiresPython string `json:"requires_python"`
	RequiresDist []string `json:"requires_dist"`
	Platform     []string `json:"platform"`
	Classifier   []string `json:"classifier"`
}

// Release represents a package release with download URLs
type Release struct {
	Filename    string    `json:"filename"`
	URL         string    `json:"url"`
	Size        int64     `json:"size"`
	UploadTime  time.Time `json:"upload_time"`
	Digests     Digests   `json:"digests"`
	PythonVersion string  `json:"python_version"`
	Packagetype string    `json:"packagetype"`
}

// Digests contains hash information
type Digests struct {
	MD5    string `json:"md5"`
	SHA256 string `json:"sha256"`
}

// PyPIClient handles communication with PyPI
type PyPIClient struct {
	httpClient *http.Client
	baseURL    string
}

// NewPyPIClient creates a new PyPI client
func NewPyPIClient() *PyPIClient {
	return &PyPIClient{
		httpClient: netutil.NewPyPIClient(),
		baseURL:    netutil.GetPyPIBaseURL(),
	}
}

// FetchPackageMetadata retrieves package metadata from PyPI
func (c *PyPIClient) FetchPackageMetadata(packageName string) (*PyPIMetadata, error) {
	endpoint := fmt.Sprintf(PyPIJSONEndpoint, packageName)
	url := c.baseURL + endpoint
	
	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch package metadata: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("PyPI API returned status %d", resp.StatusCode)
	}
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}
	
	var metadata PyPIMetadata
	if err := json.Unmarshal(body, &metadata); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}
	
	return &metadata, nil
}

// FetchSimpleIndex retrieves the simple HTML index for a package
func (c *PyPIClient) FetchSimpleIndex(packageName string) (string, error) {
	endpoint := fmt.Sprintf(PyPISimpleEndpoint, packageName)
	url := c.baseURL + endpoint
	
	resp, err := c.httpClient.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to fetch simple index: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("PyPI simple index returned status %d", resp.StatusCode)
	}
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}
	
	return string(body), nil
}

// GetLatestVersion gets the latest version of a package
func (c *PyPIClient) GetLatestVersion(packageName string) (string, error) {
	metadata, err := c.FetchPackageMetadata(packageName)
	if err != nil {
		return "", err
	}
	
	return metadata.Info.Version, nil
}

// GetVersions gets all available versions of a package
func (c *PyPIClient) GetVersions(packageName string) ([]string, error) {
	metadata, err := c.FetchPackageMetadata(packageName)
	if err != nil {
		return nil, err
	}
	
	versions := make([]string, 0, len(metadata.Releases))
	for version := range metadata.Releases {
		versions = append(versions, version)
	}
	
	return versions, nil
}

// GetReleasesForVersion gets all releases for a specific version
func (c *PyPIClient) GetReleasesForVersion(packageName, version string) ([]Release, error) {
	metadata, err := c.FetchPackageMetadata(packageName)
	if err != nil {
		return nil, err
	}
	
	releases, exists := metadata.Releases[version]
	if !exists {
		return nil, fmt.Errorf("version %s not found for package %s", version, packageName)
	}
	
	return releases, nil
}

// DownloadRelease downloads a specific release
func (c *PyPIClient) DownloadRelease(release Release) (io.ReadCloser, error) {
	resp, err := c.httpClient.Get(release.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to download release: %w", err)
	}
	
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("download failed with status %d", resp.StatusCode)
	}
	
	return resp.Body, nil
}

// FindWheelForVersion finds the best wheel for a given version and platform
func (c *PyPIClient) FindWheelForVersion(packageName, version, platform string) (*Release, error) {
	releases, err := c.GetReleasesForVersion(packageName, version)
	if err != nil {
		return nil, err
	}
	
	// Look for wheels first
	for _, release := range releases {
		if release.Packagetype == "bdist_wheel" {
			// TODO: Implement platform matching logic
			return &release, nil
		}
	}
	
	// Fall back to source distribution
	for _, release := range releases {
		if release.Packagetype == "sdist" {
			return &release, nil
		}
	}
	
	return nil, fmt.Errorf("no suitable distribution found for %s %s", packageName, version)
} 