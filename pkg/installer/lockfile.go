package installer

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"rimraf-adi.com/zephyr/pkg/solver"
)

// Lockfile represents a dependency lockfile
type Lockfile struct {
	Version     string                 `json:"version"`
	GeneratedAt time.Time              `json:"generated_at"`
	Python      string                 `json:"python"`
	Packages    map[string]LockPackage `json:"packages"`
	Groups      map[string]LockGroup   `json:"groups,omitempty"`
	Metadata    LockMetadata           `json:"metadata"`
}

// LockPackage represents a locked package
type LockPackage struct {
	Version     string            `json:"version"`
	Source      string            `json:"source"`
	URL         string            `json:"url,omitempty"`
	Hash        string            `json:"hash,omitempty"`
	Dependencies map[string]string `json:"dependencies,omitempty"`
	Extras      []string          `json:"extras,omitempty"`
	Markers     string            `json:"markers,omitempty"`
}

// LockGroup represents a group of packages
type LockGroup struct {
	Packages []string `json:"packages"`
}

// LockMetadata contains lockfile metadata
type LockMetadata struct {
	Hash         string            `json:"hash"`
	Timestamp    time.Time         `json:"timestamp"`
	PyPIVersion  string            `json:"pypi_version"`
	ResolvedBy   string            `json:"resolved_by"`
	ResolvedAt   time.Time         `json:"resolved_at"`
	Constraints  map[string]string `json:"constraints"`
	Conflicts    []string          `json:"conflicts,omitempty"`
}

// NewLockfile creates a new lockfile
func NewLockfile(pythonVersion string) *Lockfile {
	return &Lockfile{
		Version:     "1.0",
		GeneratedAt: time.Now(),
		Python:      pythonVersion,
		Packages:    make(map[string]LockPackage),
		Groups:      make(map[string]LockGroup),
		Metadata: LockMetadata{
			Timestamp:   time.Now(),
			ResolvedBy:  "zephyr",
			ResolvedAt:  time.Now(),
			Constraints: make(map[string]string),
		},
	}
}

// LoadLockfile loads a lockfile from disk
func LoadLockfile(path string) (*Lockfile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read lockfile '%s': %w. Ensure the file exists and is readable.", path, err)
	}
	var lockfile Lockfile
	if err := json.Unmarshal(data, &lockfile); err != nil {
		return nil, fmt.Errorf("failed to parse lockfile '%s': %w. The file may be corrupted or not a valid lockfile.", path, err)
	}
	return &lockfile, nil
}

// Save saves the lockfile to disk
func (lf *Lockfile) Save(path string) error {
	data, err := json.MarshalIndent(lf, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal lockfile: %w. This is likely a bug in Zephyr.", err)
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write lockfile '%s': %w. Check permissions and disk space.", path, err)
	}
	return nil
}

// AddPackage adds a package to the lockfile
func (lf *Lockfile) AddPackage(name string, pkg LockPackage) {
	lf.Packages[name] = pkg
}

// RemovePackage removes a package from the lockfile
func (lf *Lockfile) RemovePackage(name string) {
	delete(lf.Packages, name)
}

// GetPackage gets a package from the lockfile
func (lf *Lockfile) GetPackage(name string) (LockPackage, bool) {
	pkg, exists := lf.Packages[name]
	return pkg, exists
}

// HasPackage checks if a package exists in the lockfile
func (lf *Lockfile) HasPackage(name string) bool {
	_, exists := lf.Packages[name]
	return exists
}

// UpdateFromSolution updates the lockfile from a solver solution
func (lf *Lockfile) UpdateFromSolution(solution *solver.PartialSolution) error {
	// Clear existing packages
	lf.Packages = make(map[string]LockPackage)
	
	// Add packages from solution
	for _, assignment := range solution.Assignments {
		if assignment.IsDecision {
			packageName := assignment.Term.Package
			version := assignment.Term.Version.String()
			
			// Create lock package
			lockPkg := LockPackage{
				Version: version,
				Source:  "pypi",
				URL:     fmt.Sprintf("https://pypi.org/pypi/%s/%s/json", packageName, version),
			}
			
			lf.AddPackage(packageName, lockPkg)
		}
	}
	
	// Update metadata
	lf.GeneratedAt = time.Now()
	lf.Metadata.ResolvedAt = time.Now()
	
	return nil
}

// Validate validates the lockfile
func (lf *Lockfile) Validate() error {
	if lf.Version == "" {
		return fmt.Errorf("lockfile version is required. The lockfile may be corrupted.")
	}
	if lf.Python == "" {
		return fmt.Errorf("Python version is required in lockfile. The lockfile may be corrupted.")
	}
	if lf.Packages == nil {
		return fmt.Errorf("packages cannot be nil in lockfile. The lockfile may be corrupted.")
	}
	return nil
}

// IsStale checks if the lockfile is stale compared to requirements
func (lf *Lockfile) IsStale(requirementsPath string) (bool, error) {
	data, err := os.ReadFile(requirementsPath)
	if err != nil {
		return false, fmt.Errorf("failed to read requirements file '%s': %w. Ensure the file exists and is readable.", requirementsPath, err)
	}
	requirementsHash := calculateHash(string(data))
	return requirementsHash != lf.Metadata.Hash, nil
}

// UpdateHash updates the lockfile hash
func (lf *Lockfile) UpdateHash(requirementsPath string) error {
	data, err := os.ReadFile(requirementsPath)
	if err != nil {
		return fmt.Errorf("failed to read requirements file '%s': %w. Ensure the file exists and is readable.", requirementsPath, err)
	}
	lf.Metadata.Hash = calculateHash(string(data))
	return nil
}

// calculateHash calculates a simple hash of a string
func calculateHash(s string) string {
	// This is a simplified hash function
	// In a real implementation, you'd use a proper hash like SHA256
	hash := 0
	for _, char := range s {
		hash = (hash*31 + int(char)) % 1000000007
	}
	return fmt.Sprintf("%d", hash)
}

// GetDependencyTree returns the dependency tree from the lockfile
func (lf *Lockfile) GetDependencyTree() map[string][]string {
	tree := make(map[string][]string)
	
	for name, pkg := range lf.Packages {
		if len(pkg.Dependencies) > 0 {
			deps := make([]string, 0, len(pkg.Dependencies))
			for dep := range pkg.Dependencies {
				deps = append(deps, dep)
			}
			tree[name] = deps
		}
	}
	
	return tree
}

// GetDirectDependencies returns direct dependencies (no transitive deps)
func (lf *Lockfile) GetDirectDependencies() []string {
	// This is a simplified implementation
	// In a real implementation, you'd need to parse the original requirements
	var direct []string
	
	// For now, return all packages (this should be improved)
	for name := range lf.Packages {
		direct = append(direct, name)
	}
	
	return direct
}

// LockfileManager manages lockfile operations
type LockfileManager struct {
	ProjectDir string
	LockPath   string
}

// NewLockfileManager creates a new lockfile manager
func NewLockfileManager(projectDir string) *LockfileManager {
	return &LockfileManager{
		ProjectDir: projectDir,
		LockPath:   filepath.Join(projectDir, "zephyr.lock"),
	}
}

// Load loads the lockfile
func (lm *LockfileManager) Load() (*Lockfile, error) {
	if _, err := os.Stat(lm.LockPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("lockfile does not exist")
	}
	
	return LoadLockfile(lm.LockPath)
}

// Save saves the lockfile
func (lm *LockfileManager) Save(lockfile *Lockfile) error {
	return lockfile.Save(lm.LockPath)
}

// Create creates a new lockfile
func (lm *LockfileManager) Create(pythonVersion string) *Lockfile {
	return NewLockfile(pythonVersion)
}

// Exists checks if the lockfile exists
func (lm *LockfileManager) Exists() bool {
	_, err := os.Stat(lm.LockPath)
	return err == nil
}

// Remove removes the lockfile
func (lm *LockfileManager) Remove() error {
	return os.Remove(lm.LockPath)
}

// Update updates the lockfile from requirements and solution
func (lm *LockfileManager) Update(requirementsPath string, solution *solver.PartialSolution, pythonVersion string) error {
	lockfile := lm.Create(pythonVersion)
	
	// Update from solution
	if err := lockfile.UpdateFromSolution(solution); err != nil {
		return err
	}
	
	// Update hash
	if err := lockfile.UpdateHash(requirementsPath); err != nil {
		return err
	}
	
	// Save lockfile
	return lm.Save(lockfile)
} 