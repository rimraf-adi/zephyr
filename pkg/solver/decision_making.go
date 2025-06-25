package solver

import (
	"fmt"
)

// DecisionResult represents the result of decision making
type DecisionResult struct {
	Success bool
	NextPackage string
	Error string
}

// DecisionMaking performs decision making to choose the next package version
func (s *Solver) DecisionMaking() DecisionResult {
	// Find a package with a positive derivation but no decision
	packageName := s.findPackageForDecision()
	if packageName == "" {
		// No more decisions to make - we have a solution
		return DecisionResult{Success: true}
	}
	
	// Get the term for this package from the partial solution
	term := s.getTermForPackage(packageName)
	if term == nil {
		return DecisionResult{
			Success: false,
			Error:   fmt.Sprintf("no term found for package %s", packageName),
		}
	}
	
	// Find a version that matches the term
	version := s.findMatchingVersion(packageName, *term)
	if version == "" {
		// No matching version found - add an incompatibility
		incompatibility := Incompatibility{
			Terms: []Term{*term},
		}
		s.incompatibilities = append(s.incompatibilities, incompatibility)
		return DecisionResult{NextPackage: packageName}
	}
	
	// Add dependencies for this version
	s.addDependenciesForVersion(packageName, version)
	
	// Create the decision assignment
	decisionTerm := Term{
		Package: packageName,
		Version: VersionConstraint{Specific: version},
		Negated: false,
	}
	
	assignment := Assignment{
		Term:          decisionTerm,
		DecisionLevel: s.partialSolution.GetDecisionLevel() + 1,
		IsDecision:    true,
		Cause:         nil,
	}
	
	s.partialSolution.AddAssignment(assignment)
	
	return DecisionResult{NextPackage: packageName}
}

// findPackageForDecision finds a package that needs a decision
func (s *Solver) findPackageForDecision() string {
	// Look for packages that have positive derivations but no decisions
	for _, assignment := range s.partialSolution.Assignments {
		if !assignment.IsDecision && !assignment.Term.Negated {
			// Check if we already have a decision for this package
			hasDecision := false
			for _, otherAssignment := range s.partialSolution.Assignments {
				if otherAssignment.IsDecision && otherAssignment.Term.Package == assignment.Term.Package {
					hasDecision = true
					break
				}
			}
			
			if !hasDecision {
				return assignment.Term.Package
			}
		}
	}
	
	return ""
}

// getTermForPackage gets the term for a package from the partial solution
func (s *Solver) getTermForPackage(packageName string) *Term {
	// Find all assignments for this package
	var terms []Term
	for _, assignment := range s.partialSolution.Assignments {
		if assignment.Term.Package == packageName {
			terms = append(terms, assignment.Term)
		}
	}
	
	if len(terms) == 0 {
		return nil
	}
	
	// For now, just return the first term
	// In a full implementation, we would intersect all terms
	return &terms[0]
}

// findMatchingVersion finds a version that matches the given term
func (s *Solver) findMatchingVersion(packageName string, term Term) string {
	// This is a simplified implementation
	// In a real implementation, this would query the package registry
	// and find a version that satisfies the term
	
	// For now, just return a dummy version
	if term.Version.IsSpecific() {
		return term.Version.Specific
	}
	
	// Return a default version
	return "1.0.0"
}

// addDependenciesForVersion adds dependencies for a specific version
func (s *Solver) addDependenciesForVersion(packageName, version string) {
	// This is a simplified implementation
	// In a real implementation, this would:
	// 1. Query the package registry for dependencies
	// 2. Convert dependencies to incompatibilities
	// 3. Add them to the solver
	
	// For now, just add a dummy incompatibility
	dependency := Incompatibility{
		Terms: []Term{
			{
				Package: packageName,
				Version: VersionConstraint{Specific: version},
				Negated: false,
			},
			{
				Package: "dependency",
				Version: VersionConstraint{Min: "1.0.0"},
				Negated: true,
			},
		},
	}
	
	s.incompatibilities = append(s.incompatibilities, dependency)
} 