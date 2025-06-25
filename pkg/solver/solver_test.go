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

func TestSolver_AddIncompatibilityAndGetters(t *testing.T) {
	s := NewSolver("foo", "1.0.0")
	inc := Incompatibility{Terms: []Term{{Package: "foo", Version: VersionConstraint{Specific: "1.0.0"}}}}
	s.AddIncompatibility(inc)
	incs := s.GetIncompatibilities()
	if len(incs) == 0 || incs[0].Terms[0].Package != "foo" {
		t.Error("AddIncompatibility or GetIncompatibilities failed")
	}
}

func TestSolver_Solve_Success(t *testing.T) {
	s := NewSolver("foo", "1.0.0")
	// Add a simple incompatibility that should not cause conflict
	inc := Incompatibility{Terms: []Term{{Package: "foo", Version: VersionConstraint{Specific: "1.0.0"}, Negated: false}}}
	s.AddIncompatibility(inc)
	_, err := s.Solve()
	if err != nil {
		t.Errorf("Solve failed: %v", err)
	}
}

func TestSolver_Solve_Conflict(t *testing.T) {
	s := NewSolver("foo", "1.0.0")
	// Add a conflict: foo 1.0.0 and not foo 1.0.0
	inc1 := Incompatibility{Terms: []Term{{Package: "foo", Version: VersionConstraint{Specific: "1.0.0"}, Negated: false}}}
	inc2 := Incompatibility{Terms: []Term{{Package: "foo", Version: VersionConstraint{Specific: "1.0.0"}, Negated: true}}}
	s.AddIncompatibility(inc1)
	s.AddIncompatibility(inc2)
	_, err := s.Solve()
	if err == nil {
		t.Error("Expected conflict error, got nil")
	}
}

func TestSolver_ErrorReporting(t *testing.T) {
	s := NewSolver("foo", "1.0.0")
	inc := Incompatibility{Terms: []Term{{Package: "foo", Version: VersionConstraint{Specific: "1.0.0"}}}}
	report := s.GenerateErrorReport(inc)
	if report == nil || len(report.Lines) == 0 {
		t.Error("GenerateErrorReport failed")
	}
} 