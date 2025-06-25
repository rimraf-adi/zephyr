package solver

import (
	"testing"
)

func TestNewSolver(t *testing.T) {
	s := NewSolver("test", "1.0.0")
	
	if s.rootPackage != "test" {
		t.Errorf("Expected root package 'test', got '%s'", s.rootPackage)
	}
	
	if s.rootVersion != "1.0.0" {
		t.Errorf("Expected root version '1.0.0', got '%s'", s.rootVersion)
	}
}

func TestTermString(t *testing.T) {
	term := Term{
		Package: "foo",
		Version: VersionConstraint{Specific: "1.0.0"},
		Negated: false,
	}
	
	expected := "foo 1.0.0"
	if term.String() != expected {
		t.Errorf("Expected '%s', got '%s'", expected, term.String())
	}
	
	negatedTerm := Term{
		Package: "foo",
		Version: VersionConstraint{Specific: "1.0.0"},
		Negated: true,
	}
	
	expectedNegated := "not foo 1.0.0"
	if negatedTerm.String() != expectedNegated {
		t.Errorf("Expected '%s', got '%s'", expectedNegated, negatedTerm.String())
	}
}

func TestIncompatibilityString(t *testing.T) {
	incompatibility := Incompatibility{
		Terms: []Term{
			{Package: "foo", Version: VersionConstraint{Specific: "1.0.0"}, Negated: false},
			{Package: "bar", Version: VersionConstraint{Specific: "2.0.0"}, Negated: true},
		},
	}
	
	expected := "{foo 1.0.0, not bar 2.0.0}"
	if incompatibility.String() != expected {
		t.Errorf("Expected '%s', got '%s'", expected, incompatibility.String())
	}
}

func TestVersionConstraintString(t *testing.T) {
	tests := []struct {
		constraint VersionConstraint
		expected   string
	}{
		{
			VersionConstraint{Specific: "1.0.0"},
			"1.0.0",
		},
		{
			VersionConstraint{Min: "1.0.0"},
			">=1.0.0",
		},
		{
			VersionConstraint{Max: "2.0.0"},
			"<2.0.0",
		},
		{
			VersionConstraint{Min: "1.0.0", Max: "2.0.0"},
			">=1.0.0 <2.0.0",
		},
		{
			VersionConstraint{},
			"any",
		},
	}
	
	for _, test := range tests {
		if test.constraint.String() != test.expected {
			t.Errorf("Expected '%s', got '%s'", test.expected, test.constraint.String())
		}
	}
}

func TestPartialSolutionAddAssignment(t *testing.T) {
	ps := &PartialSolution{}
	
	assignment := Assignment{
		Term:          Term{Package: "foo", Version: VersionConstraint{Specific: "1.0.0"}, Negated: false},
		DecisionLevel: 0,
		IsDecision:    true,
	}
	
	ps.AddAssignment(assignment)
	
	if len(ps.Assignments) != 1 {
		t.Errorf("Expected 1 assignment, got %d", len(ps.Assignments))
	}
	
	if ps.Assignments[0].Term.Package != "foo" {
		t.Errorf("Expected package 'foo', got '%s'", ps.Assignments[0].Term.Package)
	}
}

func TestPartialSolutionGetAssignmentByPackage(t *testing.T) {
	ps := &PartialSolution{}
	
	assignment := Assignment{
		Term:          Term{Package: "foo", Version: VersionConstraint{Specific: "1.0.0"}, Negated: false},
		DecisionLevel: 0,
		IsDecision:    true,
	}
	
	ps.AddAssignment(assignment)
	
	found := ps.GetAssignmentByPackage("foo")
	if found == nil {
		t.Error("Expected to find assignment for package 'foo'")
	}
	
	if found.Term.Package != "foo" {
		t.Errorf("Expected package 'foo', got '%s'", found.Term.Package)
	}
	
	notFound := ps.GetAssignmentByPackage("bar")
	if notFound != nil {
		t.Error("Expected not to find assignment for package 'bar'")
	}
} 