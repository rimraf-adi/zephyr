package solver

import (
	"fmt"
)

// Solver represents the Pubgrub version solver
type Solver struct {
	partialSolution PartialSolution
	incompatibilities []Incompatibility
	rootPackage string
	rootVersion string
}

// NewSolver creates a new solver instance
func NewSolver(rootPackage, rootVersion string) *Solver {
	return &Solver{
		partialSolution: PartialSolution{},
		incompatibilities: []Incompatibility{},
		rootPackage: rootPackage,
		rootVersion: rootVersion,
	}
}

// Solve performs version solving using the Pubgrub algorithm
func (s *Solver) Solve() (*PartialSolution, error) {
	// Initialize the solver with the root package
	s.initializeRootPackage()
	
	// Set the next package to process
	nextPackage := s.rootPackage
	
	// Main solving loop
	for {
		// Perform unit propagation
		result := s.UnitPropagation(nextPackage)
		if !result.Success {
			// Version solving has failed
			return nil, fmt.Errorf("version solving failed: conflict detected")
		}
		
		// Perform decision making
		decisionResult := s.DecisionMaking()
		if decisionResult.Success {
			// We have found a solution
			return &s.partialSolution, nil
		}
		
		if decisionResult.Error != "" {
			return nil, fmt.Errorf("decision making failed: %s", decisionResult.Error)
		}
		
		// Set the next package to process
		nextPackage = decisionResult.NextPackage
	}
}

// initializeRootPackage initializes the solver with the root package
func (s *Solver) initializeRootPackage() {
	// Add the root package as a decision
	rootTerm := Term{
		Package: s.rootPackage,
		Version: VersionConstraint{Specific: s.rootVersion},
		Negated: false,
	}
	
	rootAssignment := Assignment{
		Term:          rootTerm,
		DecisionLevel: 0,
		IsDecision:    true,
		Cause:         nil,
	}
	
	s.partialSolution.AddAssignment(rootAssignment)
	
	// Add incompatibilities for the root package
	// In a real implementation, these would come from the package's dependencies
	s.addRootIncompatibilities()
}

// addRootIncompatibilities adds incompatibilities for the root package
func (s *Solver) addRootIncompatibilities() {
	// This is a simplified implementation
	// In a real implementation, this would:
	// 1. Read the root package's dependencies
	// 2. Convert them to incompatibilities
	// 3. Add them to the solver
	
	// For now, just add a dummy incompatibility
	rootIncompatibility := Incompatibility{
		Terms: []Term{
			{
				Package: s.rootPackage,
				Version: VersionConstraint{Specific: s.rootVersion},
				Negated: false,
			},
			{
				Package: "dependency",
				Version: VersionConstraint{Min: "1.0.0"},
				Negated: true,
			},
		},
	}
	
	s.incompatibilities = append(s.incompatibilities, rootIncompatibility)
}

// AddIncompatibility adds an incompatibility to the solver
func (s *Solver) AddIncompatibility(incompatibility Incompatibility) {
	s.incompatibilities = append(s.incompatibilities, incompatibility)
}

// GetSolution returns the current partial solution
func (s *Solver) GetSolution() *PartialSolution {
	return &s.partialSolution
}

// GetIncompatibilities returns all incompatibilities in the solver
func (s *Solver) GetIncompatibilities() []Incompatibility {
	return s.incompatibilities
} 