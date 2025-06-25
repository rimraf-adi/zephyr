package installer

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// VirtualEnvironment represents a Python virtual environment
type VirtualEnvironment struct {
	Path string
}

// NewVirtualEnvironment creates a new virtual environment
func NewVirtualEnvironment(path string) *VirtualEnvironment {
	return &VirtualEnvironment{
		Path: path,
	}
}

// Create creates a new virtual environment
func (venv *VirtualEnvironment) Create() error {
	pythonCmd, err := venv.findPython()
	if err != nil {
		return fmt.Errorf("Python not found: %w. Please install Python 3.7+ and ensure it is in your PATH.", err)
	}
	cmd := exec.Command(pythonCmd, "-m", "venv", venv.Path)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create virtual environment at '%s': %w. Ensure you have write permissions and sufficient disk space.", venv.Path, err)
	}
	return nil
}

// Activate activates the virtual environment
func (venv *VirtualEnvironment) Activate() error {
	// This would set environment variables
	// In a real implementation, this would modify the current process environment
	
	// Set VIRTUAL_ENV
	os.Setenv("VIRTUAL_ENV", venv.Path)
	
	// Modify PATH to include virtual environment's bin directory
	binDir := venv.GetBinPath()
	currentPath := os.Getenv("PATH")
	
	if runtime.GOOS == "windows" {
		os.Setenv("PATH", binDir+";"+currentPath)
	} else {
		os.Setenv("PATH", binDir+":"+currentPath)
	}
	
	return nil
}

// Deactivate deactivates the virtual environment
func (venv *VirtualEnvironment) Deactivate() error {
	// Restore original environment variables
	os.Unsetenv("VIRTUAL_ENV")
	
	// Restore original PATH (this is simplified)
	// In a real implementation, you'd need to track the original PATH
	
	return nil
}

// GetPythonPath returns the path to the Python executable in the virtual environment
func (venv *VirtualEnvironment) GetPythonPath() string {
	if runtime.GOOS == "windows" {
		return filepath.Join(venv.Path, "Scripts", "python.exe")
	}
	return filepath.Join(venv.Path, "bin", "python")
}

// GetPipPath returns the path to the pip executable in the virtual environment
func (venv *VirtualEnvironment) GetPipPath() string {
	if runtime.GOOS == "windows" {
		return filepath.Join(venv.Path, "Scripts", "pip.exe")
	}
	return filepath.Join(venv.Path, "bin", "pip")
}

// GetBinPath returns the bin directory path
func (venv *VirtualEnvironment) GetBinPath() string {
	if runtime.GOOS == "windows" {
		return filepath.Join(venv.Path, "Scripts")
	}
	return filepath.Join(venv.Path, "bin")
}

// GetSitePackagesPath returns the site-packages directory path
func (venv *VirtualEnvironment) GetSitePackagesPath() string {
	// Try to determine Python version
	pythonPath := venv.GetPythonPath()
	if _, err := os.Stat(pythonPath); err == nil {
		// Get Python version
		cmd := exec.Command(pythonPath, "--version")
		output, err := cmd.Output()
		if err == nil {
			version := strings.TrimSpace(string(output))
			// Extract version number (e.g., "Python 3.11.0" -> "3.11")
			if strings.HasPrefix(version, "Python ") {
				parts := strings.Split(version, " ")
				if len(parts) >= 2 {
					versionParts := strings.Split(parts[1], ".")
					if len(versionParts) >= 2 {
						pythonVersion := versionParts[0] + "." + versionParts[1]
						return filepath.Join(venv.Path, "lib", "python"+pythonVersion, "site-packages")
					}
				}
			}
		}
	}
	
	// Fallback to a default path
	return filepath.Join(venv.Path, "lib", "python3.11", "site-packages")
}

// InstallPackage installs a package using pip
func (venv *VirtualEnvironment) InstallPackage(packageSpec string) error {
	pipPath := venv.GetPipPath()
	cmd := exec.Command(pipPath, "install", packageSpec)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to install package '%s': %w. Check your internet connection and package name.", packageSpec, err)
	}
	return nil
}

// InstallRequirements installs packages from a requirements file
func (venv *VirtualEnvironment) InstallRequirements(requirementsPath string) error {
	pipPath := venv.GetPipPath()
	cmd := exec.Command(pipPath, "install", "-r", requirementsPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to install requirements from '%s': %w. Check the file exists and is valid.", requirementsPath, err)
	}
	return nil
}

// ListInstalledPackages lists installed packages
func (venv *VirtualEnvironment) ListInstalledPackages() ([]string, error) {
	pipPath := venv.GetPipPath()
	cmd := exec.Command(pipPath, "list", "--format=freeze")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list packages: %w. Ensure the virtual environment is valid.", err)
	}
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	var packages []string
	for _, line := range lines {
		if line != "" {
			packages = append(packages, line)
		}
	}
	return packages, nil
}

// UninstallPackage uninstalls a package
func (venv *VirtualEnvironment) UninstallPackage(packageName string) error {
	pipPath := venv.GetPipPath()
	cmd := exec.Command(pipPath, "uninstall", "-y", packageName)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to uninstall package '%s': %w. The package may not be installed.", packageName, err)
	}
	return nil
}

// Exists checks if the virtual environment exists
func (venv *VirtualEnvironment) Exists() bool {
	pythonPath := venv.GetPythonPath()
	_, err := os.Stat(pythonPath)
	return err == nil
}

// Remove removes the virtual environment
func (venv *VirtualEnvironment) Remove() error {
	return os.RemoveAll(venv.Path)
}

// findPython finds the Python executable
func (venv *VirtualEnvironment) findPython() (string, error) {
	// Try common Python commands
	commands := []string{"python3", "python", "py"}
	
	for _, cmd := range commands {
		if path, err := exec.LookPath(cmd); err == nil {
			return path, nil
		}
	}
	
	return "", fmt.Errorf("Python not found in PATH")
}

// GetPythonVersion gets the Python version
func (venv *VirtualEnvironment) GetPythonVersion() (string, error) {
	pythonPath := venv.GetPythonPath()
	
	cmd := exec.Command(pythonPath, "--version")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get Python version: %w", err)
	}
	
	return strings.TrimSpace(string(output)), nil
}

// CreateFromRequirements creates a virtual environment and installs requirements
func (venv *VirtualEnvironment) CreateFromRequirements(requirementsPath string) error {
	// Create virtual environment
	if err := venv.Create(); err != nil {
		return err
	}
	
	// Install requirements
	if err := venv.InstallRequirements(requirementsPath); err != nil {
		return err
	}
	
	return nil
}

// UpgradePip upgrades pip in the virtual environment
func (venv *VirtualEnvironment) UpgradePip() error {
	pipPath := venv.GetPipPath()
	cmd := exec.Command(pipPath, "install", "--upgrade", "pip")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to upgrade pip: %w. Check your internet connection.", err)
	}
	return nil
} 