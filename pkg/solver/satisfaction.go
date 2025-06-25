package solver

// SatisfactionResult represents the result of checking satisfaction
type SatisfactionResult int

const (
	Inconclusive SatisfactionResult = iota
	Satisfied
	Contradicted
)

// Satisfies checks if a set of terms satisfies another term
func (ps *PartialSolution) Satisfies(term Term) SatisfactionResult {
	// Check if any assignment contradicts the term
	for _, assignment := range ps.Assignments {
		if assignment.Term.Package == term.Package {
			if assignment.Term.Negated != term.Negated {
				// One is positive, one is negative - check if they're compatible
				if !areCompatible(assignment.Term.Version, term.Version) {
					return Contradicted
				}
			} else {
				// Both are positive or both are negative - check if they're compatible
				if !areCompatible(assignment.Term.Version, term.Version) {
					return Contradicted
				}
			}
		}
	}
	
	// Check if any assignment satisfies the term
	for _, assignment := range ps.Assignments {
		if assignment.Term.Package == term.Package {
			if assignment.Term.Negated == term.Negated {
				// Both are positive or both are negative - check if assignment satisfies term
				if satisfies(assignment.Term.Version, term.Version) {
					return Satisfied
				}
			}
		}
	}
	
	return Inconclusive
}

// SatisfiesIncompatibility checks if the partial solution satisfies an incompatibility
func (ps *PartialSolution) SatisfiesIncompatibility(incompatibility Incompatibility) SatisfactionResult {
	satisfiedCount := 0
	contradictedCount := 0
	
	for _, term := range incompatibility.Terms {
		result := ps.Satisfies(term)
		switch result {
		case Satisfied:
			satisfiedCount++
		case Contradicted:
			contradictedCount++
		}
	}
	
	if contradictedCount > 0 {
		return Contradicted
	}
	
	if satisfiedCount == len(incompatibility.Terms) {
		return Satisfied
	}
	
	return Inconclusive
}

// AlmostSatisfies checks if the partial solution almost satisfies an incompatibility
// Returns the unsatisfied term if so, otherwise nil
func (ps *PartialSolution) AlmostSatisfies(incompatibility Incompatibility) *Term {
	satisfiedCount := 0
	var unsatisfiedTerm *Term
	
	for _, term := range incompatibility.Terms {
		result := ps.Satisfies(term)
		if result == Satisfied {
			satisfiedCount++
		} else if result == Inconclusive {
			if unsatisfiedTerm == nil {
				unsatisfiedTerm = &term
			}
		}
	}
	
	if satisfiedCount == len(incompatibility.Terms)-1 && unsatisfiedTerm != nil {
		return unsatisfiedTerm
	}
	
	return nil
}

// areCompatible checks if two version constraints are compatible
func areCompatible(v1, v2 VersionConstraint) bool {
	// If either is "any", they're compatible
	if v1.String() == "any" || v2.String() == "any" {
		return true
	}
	
	// For now, assume they're compatible if they're not explicitly incompatible
	// This is a simplified implementation
	return true
}

// satisfies checks if v1 satisfies v2
func satisfies(v1, v2 VersionConstraint) bool {
	// If v2 is "any", v1 always satisfies it
	if v2.String() == "any" {
		return true
	}
	
	// If v1 is "any", it satisfies everything
	if v1.String() == "any" {
		return true
	}
	
	// For now, assume they satisfy if they're the same
	// This is a simplified implementation
	return v1.String() == v2.String()
} 