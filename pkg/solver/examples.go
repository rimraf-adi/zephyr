package solver

import (
	"fmt"
)

// ExampleNoConflicts demonstrates the "No Conflicts" example from the paper
func ExampleNoConflicts() {
	fmt.Println("=== Example: No Conflicts ===")
	
	// Create solver for root 1.0.0
	s := NewSolver("root", "1.0.0")
	
	// Add the incompatibilities from the paper example:
	// root 1.0.0 depends on foo ^1.0.0
	// foo 1.0.0 depends on bar ^1.0.0
	// bar 1.0.0 and 2.0.0 have no dependencies
	
	// Add incompatibility: {root 1.0.0, not foo ^1.0.0}
	rootFooIncompatibility := Incompatibility{
		Terms: []Term{
			{Package: "root", Version: VersionConstraint{Specific: "1.0.0"}, Negated: false},
			{Package: "foo", Version: VersionConstraint{Min: "1.0.0", Max: "2.0.0"}, Negated: true},
		},
	}
	s.AddIncompatibility(rootFooIncompatibility)
	
	// Add incompatibility: {foo any, not bar ^1.0.0}
	fooBarIncompatibility := Incompatibility{
		Terms: []Term{
			{Package: "foo", Version: VersionConstraint{}, Negated: false}, // foo any
			{Package: "bar", Version: VersionConstraint{Min: "1.0.0", Max: "2.0.0"}, Negated: true},
		},
	}
	s.AddIncompatibility(fooBarIncompatibility)
	
	// Solve
	solution, err := s.Solve()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	
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

// ExampleConflictResolution demonstrates the conflict resolution example from the paper
func ExampleConflictResolution() {
	fmt.Println("=== Example: Conflict Resolution ===")
	
	// Create solver for root 1.0.0
	s := NewSolver("root", "1.0.0")
	
	// Add incompatibilities from the paper example:
	// root 1.0.0 depends on foo >=1.0.0
	// foo 2.0.0 depends on bar ^1.0.0
	// foo 1.0.0 has no dependencies
	// bar 1.0.0 depends on foo ^1.0.0
	
	// Add incompatibility: {root 1.0.0, not foo >=1.0.0}
	rootFooIncompatibility := Incompatibility{
		Terms: []Term{
			{Package: "root", Version: VersionConstraint{Specific: "1.0.0"}, Negated: false},
			{Package: "foo", Version: VersionConstraint{Min: "1.0.0"}, Negated: true},
		},
	}
	s.AddIncompatibility(rootFooIncompatibility)
	
	// Add incompatibility: {foo >=2.0.0, not bar ^1.0.0}
	fooBarIncompatibility := Incompatibility{
		Terms: []Term{
			{Package: "foo", Version: VersionConstraint{Min: "2.0.0"}, Negated: false},
			{Package: "bar", Version: VersionConstraint{Min: "1.0.0", Max: "2.0.0"}, Negated: true},
		},
	}
	s.AddIncompatibility(fooBarIncompatibility)
	
	// Add incompatibility: {bar any, not foo ^1.0.0}
	barFooIncompatibility := Incompatibility{
		Terms: []Term{
			{Package: "bar", Version: VersionConstraint{}, Negated: false}, // bar any
			{Package: "foo", Version: VersionConstraint{Min: "1.0.0", Max: "2.0.0"}, Negated: true},
		},
	}
	s.AddIncompatibility(barFooIncompatibility)
	
	// Solve
	solution, err := s.Solve()
	if err != nil {
		fmt.Printf("Solver failed as expected: %v\n", err)
		fmt.Println("This demonstrates conflict resolution in the Pubgrub algorithm.")
		return
	}
	
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

// ExampleLinearErrorReporting demonstrates linear error reporting from the paper
func ExampleLinearErrorReporting() {
	fmt.Println("=== Example: Linear Error Reporting ===")
	
	// Create solver for root 1.0.0
	s := NewSolver("root", "1.0.0")
	
	// Add incompatibilities from the paper example:
	// root 1.0.0 depends on foo ^1.0.0 and baz ^1.0.0
	// foo 1.0.0 depends on bar ^2.0.0
	// bar 2.0.0 depends on baz ^3.0.0
	// baz 1.0.0 and 3.0.0 have no dependencies
	
	// Add incompatibilities
	rootFooIncompatibility := Incompatibility{
		Terms: []Term{
			{Package: "root", Version: VersionConstraint{Specific: "1.0.0"}, Negated: false},
			{Package: "foo", Version: VersionConstraint{Min: "1.0.0", Max: "2.0.0"}, Negated: true},
		},
	}
	s.AddIncompatibility(rootFooIncompatibility)
	
	rootBazIncompatibility := Incompatibility{
		Terms: []Term{
			{Package: "root", Version: VersionConstraint{Specific: "1.0.0"}, Negated: false},
			{Package: "baz", Version: VersionConstraint{Min: "1.0.0", Max: "2.0.0"}, Negated: true},
		},
	}
	s.AddIncompatibility(rootBazIncompatibility)
	
	fooBarIncompatibility := Incompatibility{
		Terms: []Term{
			{Package: "foo", Version: VersionConstraint{}, Negated: false}, // foo any
			{Package: "bar", Version: VersionConstraint{Min: "2.0.0", Max: "3.0.0"}, Negated: true},
		},
	}
	s.AddIncompatibility(fooBarIncompatibility)
	
	barBazIncompatibility := Incompatibility{
		Terms: []Term{
			{Package: "bar", Version: VersionConstraint{}, Negated: false}, // bar any
			{Package: "baz", Version: VersionConstraint{Min: "3.0.0", Max: "4.0.0"}, Negated: true},
		},
	}
	s.AddIncompatibility(barBazIncompatibility)
	
	// Solve
	solution, err := s.Solve()
	if err != nil {
		fmt.Printf("Solver failed as expected: %v\n", err)
		fmt.Println("This demonstrates linear error reporting in the Pubgrub algorithm.")
		
		// Generate error report
		rootIncompatibility := Incompatibility{
			Terms: []Term{
				{Package: "root", Version: VersionConstraint{Specific: "1.0.0"}, Negated: false},
			},
		}
		report := s.GenerateErrorReport(rootIncompatibility)
		fmt.Println("Error report:")
		fmt.Println(report.String())
		return
	}
	
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

// RunAllExamples runs all the examples from the paper
func RunAllExamples() {
	fmt.Println("Running Pubgrub solver examples from the paper...")
	
	ExampleNoConflicts()
	fmt.Println()
	
	ExampleConflictResolution()
	fmt.Println()
	
	ExampleLinearErrorReporting()
	fmt.Println()
	
	fmt.Println("All examples completed!")
} 