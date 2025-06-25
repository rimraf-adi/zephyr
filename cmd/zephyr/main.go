package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"rimraf-adi.com/zephyr/pkg/buildmeta"
	"rimraf-adi.com/zephyr/pkg/installer"
	"rimraf-adi.com/zephyr/pkg/pypi"
	"rimraf-adi.com/zephyr/pkg/solver"
)

var rootCmd = &cobra.Command{
	Use:   "zephyr",
	Short: "Zephyr - A modern Python package manager",
	Long: `Zephyr is a fast, reliable Python package manager that uses the Pubgrub dependency resolution algorithm.

Features:
- Fast dependency resolution with Pubgrub
- PyPI integration
- Virtual environment management
- Lockfile support
- buildmeta.yaml configuration
- PEP 517/518/621 compliance`,
}

var initCmd = &cobra.Command{
	Use:   "init [project-name]",
	Short: "Initialize a new Python project",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		projectName := "my-python-project"
		if len(args) > 0 {
			projectName = args[0]
		}
		buildMeta := buildmeta.NewBuildMeta(projectName, "0.1.0")
		buildMeta.Description = "A Python project created with Zephyr"
		buildMeta.Author = "Your Name"
		buildMeta.Email = "your.email@example.com"
		buildMeta.License = "MIT"
		if err := buildmeta.WriteToDirectory(".", buildMeta); err != nil {
			fmt.Fprintf(os.Stderr, "[zephyr] Error: Could not create buildmeta.yaml: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("✅ Initialized Python project '%s'\n", projectName)
		fmt.Println("📁 Created buildmeta.yaml")
		fmt.Println("\nNext steps:")
		fmt.Println("  zephyr add <package>     # Add a dependency")
		fmt.Println("  zephyr install           # Install dependencies")
		fmt.Println("  zephyr venv create       # Create virtual environment")
		if pyprojectFlag {
			pyproject := fmt.Sprintf(`[tool.poetry]\nname = "%s"\nversion = "0.1.0"\ndescription = "A Python project created with Zephyr"\nauthors = ["Your Name <your.email@example.com>"]\nreadme = "README.md"\n\n[tool.poetry.dependencies]\npython = "^3.11.4"\n\n[build-system]\nrequires = ["poetry-core>=1.0.0", "poetry>=1.0.0"]\nbuild-backend = "poetry.core.masonry.api"\n`, projectName)
			if err := os.WriteFile("pyproject.toml", []byte(pyproject), 0644); err != nil {
				fmt.Fprintf(os.Stderr, "[zephyr] Error: Could not create pyproject.toml: %v\n", err)
				os.Exit(1)
			}
			fmt.Println("\n📁 Created pyproject.toml")
		}
	},
}

var addCmd = &cobra.Command{
	Use:   "add [package] [constraint]",
	Short: "Add a dependency to the project",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		packageName := args[0]
		constraint := ""
		if len(args) > 1 {
			constraint = args[1]
		}
		buildMeta, err := buildmeta.ParseFromDirectory(".")
		if err != nil {
			fmt.Fprintf(os.Stderr, "[zephyr] Error: Could not load buildmeta.yaml: %v\n", err)
			fmt.Fprintln(os.Stderr, "Run 'zephyr init' to create a new project.")
			os.Exit(1)
		}
		buildMeta.AddDependency(packageName, constraint)
		if err := buildmeta.WriteToDirectory(".", buildMeta); err != nil {
			fmt.Fprintf(os.Stderr, "[zephyr] Error: Could not save buildmeta.yaml: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("✅ Added %s%s to dependencies\n", packageName, constraint)
	},
}

var removeCmd = &cobra.Command{
	Use:   "remove [package]",
	Short: "Remove a dependency from the project",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		packageName := args[0]
		buildMeta, err := buildmeta.ParseFromDirectory(".")
		if err != nil {
			fmt.Fprintf(os.Stderr, "[zephyr] Error: Could not load buildmeta.yaml: %v\n", err)
			os.Exit(1)
		}
		buildMeta.RemoveDependency(packageName)
		if err := buildmeta.WriteToDirectory(".", buildMeta); err != nil {
			fmt.Fprintf(os.Stderr, "[zephyr] Error: Could not save buildmeta.yaml: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("✅ Removed %s from dependencies\n", packageName)
	},
}

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update all dependencies to the latest allowed by constraints",
	Run: func(cmd *cobra.Command, args []string) {
		buildMeta, err := buildmeta.ParseFromDirectory(".")
		if err != nil {
			fmt.Fprintf(os.Stderr, "[zephyr] Error: Could not load buildmeta.yaml: %v\n", err)
			os.Exit(1)
		}
		client := pypi.NewPyPIClient()
		updated := false
		for name, constraint := range buildMeta.GetDependencies() {
			latest, err := client.GetLatestVersion(name)
			if err != nil {
				fmt.Fprintf(os.Stderr, "[zephyr] Warning: Could not fetch latest version for %s: %v\n", name, err)
				continue
			}
			if constraint == "" || constraint == latest || strings.HasSuffix(constraint, latest) {
				continue
			}
			buildMeta.AddDependency(name, latest)
			fmt.Printf("Updated %s to %s\n", name, latest)
			updated = true
		}
		if updated {
			if err := buildmeta.WriteToDirectory(".", buildMeta); err != nil {
				fmt.Fprintf(os.Stderr, "[zephyr] Error: Could not save buildmeta.yaml: %v\n", err)
				os.Exit(1)
			}
			fmt.Println("✅ Dependencies updated. Run 'zephyr install' to apply changes.")
		} else {
			fmt.Println("All dependencies are up to date.")
		}
	},
}

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install project dependencies",
	Run: func(cmd *cobra.Command, args []string) {
		buildMeta, err := buildmeta.ParseFromDirectory(".")
		if err != nil {
			fmt.Fprintf(os.Stderr, "[zephyr] Error: Could not load buildmeta.yaml: %v\n", err)
			os.Exit(1)
		}
		s := solver.NewSolver(buildMeta.Name, buildMeta.Version)
		for name, constraint := range buildMeta.GetDependencies() {
			incompatibility := solver.Incompatibility{
				Terms: []solver.Term{
					{
						Package: buildMeta.Name,
						Version: solver.VersionConstraint{Specific: buildMeta.Version},
						Negated: false,
					},
					{
						Package: name,
						Version: parseVersionConstraint(constraint),
						Negated: true,
					},
				},
			}
			s.AddIncompatibility(incompatibility)
		}
		solution, err := s.Solve()
		if err != nil {
			fmt.Fprintf(os.Stderr, "[zephyr] Dependency resolution failed: %v\n", err)
			if report := s.GenerateErrorReport(s.GetLastConflict()); report != nil {
				fmt.Fprintln(os.Stderr, "\nDependency conflict details:")
				for _, line := range report.Lines {
					fmt.Fprintln(os.Stderr, line)
				}
			}
			os.Exit(1)
		}
		fmt.Println("✅ Dependencies resolved successfully!")
		fmt.Println("\nResolved packages:")
		for _, assignment := range solution.Assignments {
			if assignment.IsDecision {
				fmt.Printf("  %s == %s\n", assignment.Term.Package, assignment.Term.Version.String())
			}
		}
		lockManager := installer.NewLockfileManager(".")
		if err := lockManager.Update("buildmeta.yaml", solution, "3.11"); err != nil {
			fmt.Fprintf(os.Stderr, "[zephyr] Error: Could not create lockfile: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("\n📦 Lockfile updated: zephyr.lock")
	},
}

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Install dependencies from lockfile (no resolution)",
	Run: func(cmd *cobra.Command, args []string) {
		venvPath := ".venv"
		venv := installer.NewVirtualEnvironment(venvPath)
		if !venv.Exists() {
			fmt.Fprintf(os.Stderr, "[zephyr] Error: Virtual environment does not exist at %s\n", venvPath)
			fmt.Fprintln(os.Stderr, "Create it first with: zephyr venv create")
			os.Exit(1)
		}
		lockManager := installer.NewLockfileManager(".")
		lockfile, err := lockManager.Load()
		if err != nil {
			fmt.Fprintf(os.Stderr, "[zephyr] Error: Could not load lockfile: %v\n", err)
			os.Exit(1)
		}
		wheelInstaller := installer.NewWheelInstaller(venvPath)
		for name, pkg := range lockfile.Packages {
			fmt.Printf("Installing %s %s...\n", name, pkg.Version)
			if err := wheelInstaller.InstallWheelFromPyPI(name, pkg.Version); err != nil {
				fmt.Fprintf(os.Stderr, "[zephyr] Error: Could not install %s: %v\n", name, err)
				os.Exit(1)
			}
		}
		fmt.Println("✅ All packages installed from lockfile!")
	},
}

var lockCmd = &cobra.Command{
	Use:   "lock",
	Short: "Generate lockfile without installing",
	Run: func(cmd *cobra.Command, args []string) {
		buildMeta, err := buildmeta.ParseFromDirectory(".")
		if err != nil {
			fmt.Fprintf(os.Stderr, "[zephyr] Error: Could not load buildmeta.yaml: %v\n", err)
			os.Exit(1)
		}
		s := solver.NewSolver(buildMeta.Name, buildMeta.Version)
		for name, constraint := range buildMeta.GetDependencies() {
			incompatibility := solver.Incompatibility{
				Terms: []solver.Term{
					{
						Package: buildMeta.Name,
						Version: solver.VersionConstraint{Specific: buildMeta.Version},
						Negated: false,
					},
					{
						Package: name,
						Version: parseVersionConstraint(constraint),
						Negated: true,
					},
				},
			}
			s.AddIncompatibility(incompatibility)
		}
		solution, err := s.Solve()
		if err != nil {
			fmt.Fprintf(os.Stderr, "[zephyr] Dependency resolution failed: %v\n", err)
			if report := s.GenerateErrorReport(s.GetLastConflict()); report != nil {
				fmt.Fprintln(os.Stderr, "\nDependency conflict details:")
				for _, line := range report.Lines {
					fmt.Fprintln(os.Stderr, line)
				}
			}
			os.Exit(1)
		}
		lockManager := installer.NewLockfileManager(".")
		if err := lockManager.Update("buildmeta.yaml", solution, "3.11"); err != nil {
			fmt.Fprintf(os.Stderr, "[zephyr] Error: Could not create lockfile: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("✅ Lockfile generated: zephyr.lock")
	},
}

var venvCmd = &cobra.Command{
	Use:   "venv",
	Short: "Manage virtual environments",
}

var venvCreateCmd = &cobra.Command{
	Use:   "create [path]",
	Short: "Create a new virtual environment",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		venvPath := ".venv"
		if len(args) > 0 {
			venvPath = args[0]
		}
		venv := installer.NewVirtualEnvironment(venvPath)
		if err := venv.Create(); err != nil {
			fmt.Fprintf(os.Stderr, "[zephyr] Error: Could not create virtual environment: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("✅ Created virtual environment at %s\n", venvPath)
		fmt.Println("\nTo activate:")
		if venvPath == ".venv" {
			fmt.Println("  source .venv/bin/activate  # Linux/macOS")
			fmt.Println("  .venv\\Scripts\\activate     # Windows")
		} else {
			fmt.Printf("  source %s/bin/activate\n", venvPath)
		}
	},
}

var venvInstallCmd = &cobra.Command{
	Use:   "install [venv-path]",
	Short: "Install dependencies into virtual environment",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		venvPath := ".venv"
		if len(args) > 0 {
			venvPath = args[0]
		}
		venv := installer.NewVirtualEnvironment(venvPath)
		if !venv.Exists() {
			fmt.Fprintf(os.Stderr, "[zephyr] Error: Virtual environment does not exist at %s\n", venvPath)
			fmt.Fprintln(os.Stderr, "Create it first with: zephyr venv create")
			os.Exit(1)
		}
		lockManager := installer.NewLockfileManager(".")
		lockfile, err := lockManager.Load()
		if err != nil {
			fmt.Fprintf(os.Stderr, "[zephyr] Error: Could not load lockfile: %v\n", err)
			os.Exit(1)
		}
		wheelInstaller := installer.NewWheelInstaller(venvPath)
		for name, pkg := range lockfile.Packages {
			fmt.Printf("Installing %s %s...\n", name, pkg.Version)
			if err := wheelInstaller.InstallWheelFromPyPI(name, pkg.Version); err != nil {
				fmt.Fprintf(os.Stderr, "[zephyr] Error: Could not install %s: %v\n", name, err)
				os.Exit(1)
			}
		}
		fmt.Println("✅ All packages installed successfully!")
	},
}

var venvListCmd = &cobra.Command{
	Use:   "list",
	Short: "List available virtual environments",
	Run: func(cmd *cobra.Command, args []string) {
		if _, err := os.Stat(".venv"); err == nil {
			fmt.Println(".venv (default)")
		} else {
			fmt.Println("No virtual environments found.")
		}
	},
}

var venvActivateCmd = &cobra.Command{
	Use:   "activate [venv-path]",
	Short: "Print activation instructions for a virtual environment",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		venvPath := ".venv"
		if len(args) > 0 {
			venvPath = args[0]
		}
		if _, err := os.Stat(venvPath); err != nil {
			fmt.Fprintf(os.Stderr, "[zephyr] Error: Virtual environment does not exist at %s\n", venvPath)
			os.Exit(1)
		}
		fmt.Println("To activate:")
		fmt.Printf("  source %s/bin/activate  # Linux/macOS\n", venvPath)
		fmt.Printf("  %s\\Scripts\\activate     # Windows\n", venvPath)
	},
}

var searchCmd = &cobra.Command{
	Use:   "search [query]",
	Short: "Search for packages on PyPI",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		query := args[0]
		client := pypi.NewPyPIClient()
		metadata, err := client.FetchPackageMetadata(query)
		if err != nil {
			fmt.Fprintf(os.Stderr, "[zephyr] Error: Could not search for package: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("📦 %s %s\n", metadata.Info.Name, metadata.Info.Version)
		fmt.Printf("📝 %s\n", metadata.Info.Summary)
		if metadata.Info.Author != "" {
			fmt.Printf("👤 Author: %s\n", metadata.Info.Author)
		}
		if metadata.Info.HomePage != "" {
			fmt.Printf("🌐 Homepage: %s\n", metadata.Info.HomePage)
		}
		fmt.Println("\nAvailable versions:")
		versions, err := client.GetVersions(query)
		if err != nil {
			fmt.Fprintf(os.Stderr, "[zephyr] Error: Could not get versions: %v\n", err)
			os.Exit(1)
		}
		for _, version := range versions {
			fmt.Printf("  %s\n", version)
		}
	},
}

var solveCmd = &cobra.Command{
	Use:   "solve",
	Short: "Solve dependencies using Pubgrub algorithm",
	Run: func(cmd *cobra.Command, args []string) {
		s := solver.NewSolver("example", "1.0.0")
		dependencies := map[string]string{
			"requests": ">=2.25.0",
			"urllib3":  ">=1.26.0",
			"certifi":  ">=2020.12.0",
		}
		for name, constraint := range dependencies {
			incompatibility := solver.Incompatibility{
				Terms: []solver.Term{
					{
						Package: "example",
						Version: solver.VersionConstraint{Specific: "1.0.0"},
						Negated: false,
					},
					{
						Package: name,
						Version: parseVersionConstraint(constraint),
						Negated: true,
					},
				},
			}
			s.AddIncompatibility(incompatibility)
		}
		solution, err := s.Solve()
		if err != nil {
			fmt.Fprintf(os.Stderr, "[zephyr] Dependency resolution failed: %v\n", err)
			if report := s.GenerateErrorReport(s.GetLastConflict()); report != nil {
				fmt.Fprintln(os.Stderr, "\nDependency conflict details:")
				for _, line := range report.Lines {
					fmt.Fprintln(os.Stderr, line)
				}
			}
			os.Exit(1)
		}
		fmt.Println("✅ Dependencies solved successfully!")
		fmt.Println("\nSolution:")
		for _, assignment := range solution.Assignments {
			if assignment.IsDecision {
				fmt.Printf("  %s == %s\n", assignment.Term.Package, assignment.Term.Version.String())
			}
		}
	},
}

var demoCmd = &cobra.Command{
	Use:   "demo",
	Short: "Run Pubgrub algorithm demo",
	Run: func(cmd *cobra.Command, args []string) {
		solver.RunDemo()
	},
}

var examplesCmd = &cobra.Command{
	Use:   "examples",
	Short: "Show Pubgrub algorithm examples",
	Run: func(cmd *cobra.Command, args []string) {
		solver.RunExamples()
	},
}

var importCmd = &cobra.Command{
	Use:   "import [file]",
	Short: "Import dependencies from requirements.txt or pyproject.toml",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		file := args[0]
		if strings.HasSuffix(file, ".txt") {
			reqs, err := buildmeta.ParseRequirementsFile(file)
			if err != nil {
				fmt.Fprintf(os.Stderr, "[zephyr] Error: Could not parse requirements.txt: %v\n", err)
				os.Exit(1)
			}
			buildMeta, err := buildmeta.ParseFromDirectory(".")
			if err != nil {
				buildMeta = buildmeta.NewBuildMeta("imported-project", "0.1.0")
			}
			for name, constraint := range reqs {
				buildMeta.AddDependency(name, constraint)
			}
			if err := buildmeta.WriteToDirectory(".", buildMeta); err != nil {
				fmt.Fprintf(os.Stderr, "[zephyr] Error: Could not save buildmeta.yaml: %v\n", err)
				os.Exit(1)
			}
			fmt.Println("✅ Imported dependencies from requirements.txt into buildmeta.yaml")
		} else if strings.HasSuffix(file, ".toml") {
			pyMeta, err := buildmeta.ParsePyProjectToml(file)
			if err != nil {
				fmt.Fprintf(os.Stderr, "[zephyr] Error: Could not parse pyproject.toml: %v\n", err)
				os.Exit(1)
			}
			buildMeta := buildmeta.NewBuildMeta(pyMeta.Name, pyMeta.Version)
			for name, constraint := range pyMeta.Dependencies {
				buildMeta.AddDependency(name, constraint)
			}
			if err := buildmeta.WriteToDirectory(".", buildMeta); err != nil {
				fmt.Fprintf(os.Stderr, "[zephyr] Error: Could not save buildmeta.yaml: %v\n", err)
				os.Exit(1)
			}
			fmt.Println("✅ Imported dependencies from pyproject.toml into buildmeta.yaml")
		} else {
			fmt.Fprintln(os.Stderr, "[zephyr] Error: Unsupported file type. Use requirements.txt or pyproject.toml.")
			os.Exit(1)
		}
	},
}

var exportCmd = &cobra.Command{
	Use:   "export [file]",
	Short: "Export dependencies to requirements.txt or pyproject.toml",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		file := args[0]
		buildMeta, err := buildmeta.ParseFromDirectory(".")
		if err != nil {
			fmt.Fprintf(os.Stderr, "[zephyr] Error: Could not load buildmeta.yaml: %v\n", err)
			os.Exit(1)
		}
		if strings.HasSuffix(file, ".txt") {
			if err := buildmeta.ExportRequirementsFile(file, buildMeta.GetDependencies()); err != nil {
				fmt.Fprintf(os.Stderr, "[zephyr] Error: Could not write requirements.txt: %v\n", err)
				os.Exit(1)
			}
			fmt.Println("✅ Exported dependencies to requirements.txt")
		} else if strings.HasSuffix(file, ".toml") {
			if err := buildmeta.ExportPyProjectToml(file, buildMeta); err != nil {
				fmt.Fprintf(os.Stderr, "[zephyr] Error: Could not write pyproject.toml: %v\n", err)
				os.Exit(1)
			}
			fmt.Println("✅ Exported dependencies to pyproject.toml")
		} else {
			fmt.Fprintln(os.Stderr, "[zephyr] Error: Unsupported file type. Use requirements.txt or pyproject.toml.")
			os.Exit(1)
		}
	},
}

// Enhance init to optionally create pyproject.toml
var pyprojectFlag bool

func init() {
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(addCmd)
	rootCmd.AddCommand(removeCmd)
	rootCmd.AddCommand(updateCmd)
	rootCmd.AddCommand(installCmd)
	rootCmd.AddCommand(syncCmd)
	rootCmd.AddCommand(lockCmd)
	rootCmd.AddCommand(venvCmd)
	rootCmd.AddCommand(searchCmd)
	rootCmd.AddCommand(solveCmd)
	rootCmd.AddCommand(demoCmd)
	rootCmd.AddCommand(examplesCmd)
	rootCmd.AddCommand(importCmd)
	rootCmd.AddCommand(exportCmd)

	venvCmd.AddCommand(venvCreateCmd)
	venvCmd.AddCommand(venvInstallCmd)
	venvCmd.AddCommand(venvListCmd)
	venvCmd.AddCommand(venvActivateCmd)

	initCmd.Flags().BoolVar(&pyprojectFlag, "pyproject", false, "Also create pyproject.toml")
}

// parseVersionConstraint parses a version constraint string
func parseVersionConstraint(constraint string) solver.VersionConstraint {
	if constraint == "" {
		return solver.VersionConstraint{}
	}
	
	// Simple parsing - in real implementation this would be more robust
	if strings.HasPrefix(constraint, ">=") {
		return solver.VersionConstraint{Min: constraint[2:]}
	} else if strings.HasPrefix(constraint, "<=") {
		return solver.VersionConstraint{Max: constraint[2:]}
	} else if strings.HasPrefix(constraint, "==") {
		return solver.VersionConstraint{Specific: constraint[2:]}
	} else if strings.HasPrefix(constraint, ">") {
		return solver.VersionConstraint{Min: constraint[1:]}
	} else if strings.HasPrefix(constraint, "<") {
		return solver.VersionConstraint{Max: constraint[1:]}
	}
	
	// Default to specific version
	return solver.VersionConstraint{Specific: constraint}
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
} 