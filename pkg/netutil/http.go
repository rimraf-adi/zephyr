package netutil

import (
	"net/http"
	"time"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

const (
	DefaultTimeout = 30 * time.Second
	DefaultUserAgent = "Zephyr/1.0.0 (Python Package Manager)"
	DefaultPyPIBaseURL = "https://pypi.org"
)

// Config represents Zephyr configuration
// Supports global (~/.zephyr/config.toml or config.yaml) and project-level (.zephyrrc or pyproject.toml)
type Config struct {
	IndexURL string `yaml:"index_url"`
}

var globalConfig *Config
var projectConfig *Config

// LoadConfig loads global and project config
func LoadConfig() (*Config, error) {
	if globalConfig != nil && projectConfig != nil {
		return mergeConfig(globalConfig, projectConfig), nil
	}
	// Load global config
	home, err := os.UserHomeDir()
	if err == nil {
		globalPath := filepath.Join(home, ".zephyr", "config.yaml")
		if _, err := os.Stat(globalPath); err == nil {
			cfg, err := parseConfigFile(globalPath)
			if err == nil {
				globalConfig = cfg
			}
		}
	}
	// Load project config
	projectPath := ".zephyrrc"
	if _, err := os.Stat(projectPath); err == nil {
		cfg, err := parseConfigFile(projectPath)
		if err == nil {
			projectConfig = cfg
		}
	}
	return mergeConfig(globalConfig, projectConfig), nil
}

func parseConfigFile(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func mergeConfig(global, project *Config) *Config {
	cfg := &Config{}
	if global != nil {
		*cfg = *global
	}
	if project != nil {
		if project.IndexURL != "" {
			cfg.IndexURL = project.IndexURL
		}
	}
	// Environment variable override
	if env := os.Getenv("ZEPHYR_INDEX_URL"); env != "" {
		cfg.IndexURL = env
	}
	return cfg
}

// NewPyPIClient creates a new HTTP client configured for PyPI or custom index
func NewPyPIClient() *http.Client {
	return &http.Client{
		Timeout: DefaultTimeout,
		Transport: &http.Transport{
			MaxIdleConns:        10,
			IdleConnTimeout:     30 * time.Second,
			DisableCompression:  false,
		},
	}
}

// GetPyPIBaseURL returns the configured index URL or the default PyPI URL
func GetPyPIBaseURL() string {
	cfg, _ := LoadConfig()
	if cfg != nil && cfg.IndexURL != "" {
		return strings.TrimRight(cfg.IndexURL, "/")
	}
	return DefaultPyPIBaseURL
}

// NewHTTPClient creates a new HTTP client with custom configuration
func NewHTTPClient(timeout time.Duration) *http.Client {
	if timeout == 0 {
		timeout = DefaultTimeout
	}
	
	return &http.Client{
		Timeout: timeout,
		Transport: &http.Transport{
			MaxIdleConns:        10,
			IdleConnTimeout:     30 * time.Second,
			DisableCompression:  false,
		},
	}
}

// AddPyPIHeaders adds PyPI-compatible headers to an HTTP request
func AddPyPIHeaders(req *http.Request) {
	req.Header.Set("User-Agent", DefaultUserAgent)
	req.Header.Set("Accept", "application/json, text/html, */*")
	req.Header.Set("Accept-Encoding", "gzip, deflate")
	req.Header.Set("Connection", "keep-alive")
}

// CreatePyPIRequest creates a new HTTP request with PyPI headers
func CreatePyPIRequest(method, url string) (*http.Request, error) {
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, err
	}
	
	AddPyPIHeaders(req)
	return req, nil
}

// SetCustomUserAgent sets a custom user agent for requests
func SetCustomUserAgent(userAgent string) {
	// This would be used to override the default user agent
	// Implementation depends on how you want to manage global state
}

// RetryableHTTPClient creates an HTTP client with retry logic
type RetryableHTTPClient struct {
	client  *http.Client
	maxRetries int
}

// NewRetryableHTTPClient creates a new retryable HTTP client
func NewRetryableHTTPClient(maxRetries int) *RetryableHTTPClient {
	return &RetryableHTTPClient{
		client:     NewPyPIClient(),
		maxRetries: maxRetries,
	}
}

// Do performs an HTTP request with retry logic
func (c *RetryableHTTPClient) Do(req *http.Request) (*http.Response, error) {
	var lastErr error
	
	for i := 0; i <= c.maxRetries; i++ {
		resp, err := c.client.Do(req)
		if err == nil {
			return resp, nil
		}
		
		lastErr = err
		
		// Don't retry on the last attempt
		if i == c.maxRetries {
			break
		}
		
		// Wait before retrying (exponential backoff)
		time.Sleep(time.Duration(1<<uint(i)) * time.Second)
	}
	
	return nil, lastErr
}

// HTTPError represents an HTTP error
type HTTPError struct {
	StatusCode int
	Status     string
	Body       string
}

// Error implements the error interface
func (e *HTTPError) Error() string {
	return fmt.Sprintf("HTTP %d: %s", e.StatusCode, e.Status)
}

// IsRetryableError checks if an error is retryable
func IsRetryableError(err error) bool {
	if httpErr, ok := err.(*HTTPError); ok {
		// Retry on 5xx errors and some 4xx errors
		return httpErr.StatusCode >= 500 || httpErr.StatusCode == 429
	}
	
	// Retry on network errors
	return true
}

// DownloadFile downloads a file from a URL to a local path
func DownloadFile(client *http.Client, url, filepath string) error {
	req, err := CreatePyPIRequest("GET", url)
	if err != nil {
		return err
	}
	
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return &HTTPError{
			StatusCode: resp.StatusCode,
			Status:     resp.Status,
		}
	}
	
	// TODO: Implement file writing logic
	// This would use os.Create and io.Copy to write the response body to the file
	
	return nil
} 