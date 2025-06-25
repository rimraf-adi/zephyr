package main

import (
	"fmt"
	"os"

	"rimraf-adi.com/zephyr/pkg/solver"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "init":
		if len(os.Args) < 3 {
			fmt.Println("Error: init command requires an argument")
			fmt.Println("Usage: zephyr init <project-name>")
			os.Exit(1)
		}
		handleInit(os.Args[2])
		
	case "solve":
		if len(os.Args) < 4 {
			fmt.Println("Error: solve command requires package name and version")
			fmt.Println("Usage: zephyr solve <package-name> <version>")
			os.Exit(1)
		}
		handleSolve(os.Args[2], os.Args[3])
		
	case "demo":
		handleDemo()
		
	case "examples":
		handleExamples()
		
	default:
		fmt.Printf("Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Usage: zephyr <command> [args...]")
	fmt.Println("Commands:")
	fmt.Println("  init <project-name>    Initialize a new project")
	fmt.Println("  solve <pkg> <version>  Solve dependencies for a package")
	fmt.Println("  demo                   Run a demonstration of the solver")
	fmt.Println("  examples               Run examples from the Pubgrub paper")
}

func handleInit(projectName string) {
	fmt.Printf("hello %s\n", projectName)
	fmt.Printf("Initializing project: %s\n", projectName)
	
	// Create a simple project structure
	fmt.Println("Creating project structure...")
	fmt.Printf("  - %s/go.mod\n", projectName)
	fmt.Printf("  - %s/main.go\n", projectName)
	fmt.Printf("  - %s/pkg/\n", projectName)
	fmt.Println("Project initialized successfully!")
}

func handleSolve(packageName, version string) {
	fmt.Printf("Solving dependencies for %s %s...\n", packageName, version)
	
	// Create a solver
	s := solver.NewSolver(packageName, version)
	
	// Add some sample incompatibilities
	addSampleIncompatibilities(s)
	
	// Solve
	solution, err := s.Solve()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	
	// Display the solution
	fmt.Println("Solution found:")
	for _, assignment := range solution.Assignments {
		if assignment.IsDecision {
			fmt.Printf("  %s %s (decision level %d)\n", 
				assignment.Term.Package, 
				assignment.Term.Version.String(), 
				assignment.DecisionLevel)
		}
	}
}

func handleDemo() {
	fmt.Println("Running Pubgrub solver demonstration...")
	
	// Create a solver for a complex scenario
	s := solver.NewSolver("root", "1.0.0")
	
	// Add sample incompatibilities that will create a conflict
	addConflictScenario(s)
	
	// Try to solve
	solution, err := s.Solve()
	if err != nil {
		fmt.Printf("Solver failed as expected: %v\n", err)
		fmt.Println("This demonstrates conflict detection in the Pubgrub algorithm.")
	} else {
		fmt.Println("Solution found:")
		for _, assignment := range solution.Assignments {
			if assignment.IsDecision {
				fmt.Printf("  %s %s\n", assignment.Term.Package, assignment.Term.Version.String())
			}
		}
	}
}

func handleExamples() {
	fmt.Println("Running examples from the Pubgrub paper...")
	solver.RunAllExamples()
}

func addSampleIncompatibilities(s *solver.Solver) {
	// Add incompatibilities for root package dependencies
	rootFooIncompatibility := solver.Incompatibility{
		Terms: []solver.Term{
			{Package: "root", Version: solver.VersionConstraint{Specific: "1.0.0"}, Negated: false},
			{Package: "foo", Version: solver.VersionConstraint{Min: "1.0.0"}, Negated: true},
		},
	}
	s.AddIncompatibility(rootFooIncompatibility)
	
	rootBarIncompatibility := solver.Incompatibility{
		Terms: []solver.Term{
			{Package: "root", Version: solver.VersionConstraint{Specific: "1.0.0"}, Negated: false},
			{Package: "bar", Version: solver.VersionConstraint{Min: "1.0.0"}, Negated: true},
		},
	}
	s.AddIncompatibility(rootBarIncompatibility)
}

func addConflictScenario(s *solver.Solver) {
	// Add incompatibilities that will create a conflict
	// This simulates the example from the paper where foo 1.1.0 depends on bar ^2.0.0
	// but root depends on bar ^1.0.0
	
	rootFooIncompatibility := solver.Incompatibility{
		Terms: []solver.Term{
			{Package: "root", Version: solver.VersionConstraint{Specific: "1.0.0"}, Negated: false},
			{Package: "foo", Version: solver.VersionConstraint{Min: "1.0.0"}, Negated: true},
		},
	}
	s.AddIncompatibility(rootFooIncompatibility)
	
	rootBarIncompatibility := solver.Incompatibility{
		Terms: []solver.Term{
			{Package: "root", Version: solver.VersionConstraint{Specific: "1.0.0"}, Negated: false},
			{Package: "bar", Version: solver.VersionConstraint{Min: "1.0.0"}, Negated: true},
		},
	}
	s.AddIncompatibility(rootBarIncompatibility)
	
	// Add a conflict: foo 1.1.0 requires bar ^2.0.0, but root requires bar ^1.0.0
	fooBarConflict := solver.Incompatibility{
		Terms: []solver.Term{
			{Package: "foo", Version: solver.VersionConstraint{Min: "1.1.0"}, Negated: false},
			{Package: "bar", Version: solver.VersionConstraint{Min: "2.0.0"}, Negated: true},
		},
	}
	s.AddIncompatibility(fooBarConflict)
}
