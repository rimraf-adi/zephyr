package solver

// UnitPropagationResult represents the result of unit propagation
type UnitPropagationResult struct {
	Success bool
	Conflict *Incompatibility
}

// UnitPropagation performs unit propagation on the given package
func (s *Solver) UnitPropagation(packageName string) UnitPropagationResult {
	changed := map[string]bool{packageName: true}
	
	for len(changed) > 0 {
		// Remove an element from changed
		var currentPackage string
		for pkg := range changed {
			currentPackage = pkg
			delete(changed, pkg)
			break
		}
		
		// Get incompatibilities that refer to this package
		incompatibilities := s.getIncompatibilitiesForPackage(currentPackage)
		
		// Process incompatibilities from newest to oldest
		for i := len(incompatibilities) - 1; i >= 0; i-- {
			incompatibility := incompatibilities[i]
			
			result := s.partialSolution.SatisfiesIncompatibility(incompatibility)
			
			if result == Satisfied {
				// We have a conflict
				resolvedIncompatibility := s.resolveConflict(incompatibility)
				if resolvedIncompatibility == nil {
					// Version solving has failed
					return UnitPropagationResult{
						Success: false,
						Conflict: &incompatibility,
					}
				}
				
				// Add the negation of the unsatisfied term
				unsatisfiedTerm := s.partialSolution.AlmostSatisfies(*resolvedIncompatibility)
				if unsatisfiedTerm != nil {
					negatedTerm := *unsatisfiedTerm
					negatedTerm.Negated = !negatedTerm.Negated
					
					assignment := Assignment{
						Term:          negatedTerm,
						DecisionLevel: s.partialSolution.GetDecisionLevel(),
						IsDecision:    false,
						Cause:         resolvedIncompatibility,
					}
					
					s.partialSolution.AddAssignment(assignment)
					
					// Replace changed with only the package from the unsatisfied term
					changed = map[string]bool{unsatisfiedTerm.Package: true}
				}
				
			} else if result == Inconclusive {
				// Check if we almost satisfy this incompatibility
				unsatisfiedTerm := s.partialSolution.AlmostSatisfies(incompatibility)
				if unsatisfiedTerm != nil {
					// Add the negation of the unsatisfied term
					negatedTerm := *unsatisfiedTerm
					negatedTerm.Negated = !negatedTerm.Negated
					
					assignment := Assignment{
						Term:          negatedTerm,
						DecisionLevel: s.partialSolution.GetDecisionLevel(),
						IsDecision:    false,
						Cause:         &incompatibility,
					}
					
					s.partialSolution.AddAssignment(assignment)
					
					// Add the package to changed
					changed[unsatisfiedTerm.Package] = true
				}
			}
		}
	}
	
	return UnitPropagationResult{Success: true}
}

// getIncompatibilitiesForPackage returns incompatibilities that refer to the given package
func (s *Solver) getIncompatibilitiesForPackage(packageName string) []Incompatibility {
	var result []Incompatibility
	
	for _, incompatibility := range s.incompatibilities {
		for _, term := range incompatibility.Terms {
			if term.Package == packageName {
				result = append(result, incompatibility)
				break
			}
		}
	}
	
	return result
} 