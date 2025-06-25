package solver

import (
	"fmt"
	"strings"
)

// Term represents a statement about a package that may be true or false
// for a given selection of package versions
type Term struct {
	Package string
	Version VersionConstraint
	Negated bool
}

// VersionConstraint represents a version range or specific version
type VersionConstraint struct {
	Min      string
	Max      string
	Specific string
}

// IsSpecific returns true if this constraint represents a specific version
func (vc VersionConstraint) IsSpecific() bool {
	return vc.Specific != ""
}

// String returns a string representation of the version constraint
func (vc VersionConstraint) String() string {
	if vc.IsSpecific() {
		return vc.Specific
	}
	
	if vc.Min != "" && vc.Max != "" {
		return fmt.Sprintf(">=%s <%s", vc.Min, vc.Max)
	} else if vc.Min != "" {
		return fmt.Sprintf(">=%s", vc.Min)
	} else if vc.Max != "" {
		return fmt.Sprintf("<%s", vc.Max)
	}
	return "any"
}

// String returns a string representation of the term
func (t Term) String() string {
	prefix := ""
	if t.Negated {
		prefix = "not "
	}
	return fmt.Sprintf("%s%s %s", prefix, t.Package, t.Version.String())
}

// Incompatibility represents a set of terms that are not all allowed to be true
type Incompatibility struct {
	Terms []Term
	Cause *Incompatibility // For derived incompatibilities
}

// String returns a string representation of the incompatibility
func (i Incompatibility) String() string {
	terms := make([]string, len(i.Terms))
	for j, term := range i.Terms {
		terms[j] = term.String()
	}
	return fmt.Sprintf("{%s}", strings.Join(terms, ", "))
}

// Assignment represents a term that has been assigned a truth value
type Assignment struct {
	Term          Term
	DecisionLevel int
	IsDecision    bool
	Cause         *Incompatibility // For derivations
}

// PartialSolution represents the current state of the solver
type PartialSolution struct {
	Assignments []Assignment
}

// AddAssignment adds a new assignment to the partial solution
func (ps *PartialSolution) AddAssignment(assignment Assignment) {
	ps.Assignments = append(ps.Assignments, assignment)
}

// GetAssignmentByPackage returns the assignment for a given package, if any
func (ps *PartialSolution) GetAssignmentByPackage(pkg string) *Assignment {
	for i := len(ps.Assignments) - 1; i >= 0; i-- {
		if ps.Assignments[i].Term.Package == pkg {
			return &ps.Assignments[i]
		}
	}
	return nil
}

// GetDecisionLevel returns the current decision level
func (ps *PartialSolution) GetDecisionLevel() int {
	if len(ps.Assignments) == 0 {
		return 0
	}
	return ps.Assignments[len(ps.Assignments)-1].DecisionLevel
}

// Backtrack removes assignments at decision levels higher than the given level
func (ps *PartialSolution) Backtrack(level int) {
	for i := len(ps.Assignments) - 1; i >= 0; i-- {
		if ps.Assignments[i].DecisionLevel > level {
			ps.Assignments = ps.Assignments[:i]
		} else {
			break
		}
	}
} 