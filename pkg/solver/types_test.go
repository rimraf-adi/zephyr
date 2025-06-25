package solver

import (
	"reflect"
	"testing"
)

func TestTermString(t *testing.T) {
	term := Term{Package: "foo", Version: VersionConstraint{Specific: "1.0.0"}, Negated: false}
	if term.String() != "foo 1.0.0" {
		t.Errorf("Expected 'foo 1.0.0', got '%s'", term.String())
	}
	neg := Term{Package: "foo", Version: VersionConstraint{Specific: "1.0.0"}, Negated: true}
	if neg.String() != "not foo 1.0.0" {
		t.Errorf("Expected 'not foo 1.0.0', got '%s'", neg.String())
	}
}

func TestVersionConstraintString(t *testing.T) {
	tests := []struct {
		vc       VersionConstraint
		expected string
	}{
		{VersionConstraint{Specific: "1.0.0"}, "1.0.0"},
		{VersionConstraint{Min: "1.0.0"}, ">=1.0.0"},
		{VersionConstraint{Max: "2.0.0"}, "<2.0.0"},
		{VersionConstraint{Min: "1.0.0", Max: "2.0.0"}, ">=1.0.0 <2.0.0"},
		{VersionConstraint{}, "any"},
	}
	for _, test := range tests {
		if test.vc.String() != test.expected {
			t.Errorf("Expected '%s', got '%s'", test.expected, test.vc.String())
		}
	}
}

func TestIncompatibilityString(t *testing.T) {
	inc := Incompatibility{Terms: []Term{{Package: "foo", Version: VersionConstraint{Specific: "1.0.0"}, Negated: false}, {Package: "bar", Version: VersionConstraint{Specific: "2.0.0"}, Negated: true}}}
	exp := "{foo 1.0.0, not bar 2.0.0}"
	if inc.String() != exp {
		t.Errorf("Expected '%s', got '%s'", exp, inc.String())
	}
}

func TestAssignmentAndPartialSolution(t *testing.T) {
	ps := &PartialSolution{}
	assign := Assignment{Term: Term{Package: "foo", Version: VersionConstraint{Specific: "1.0.0"}}, DecisionLevel: 1, IsDecision: true}
	ps.AddAssignment(assign)
	if len(ps.Assignments) != 1 {
		t.Error("AddAssignment failed")
	}
	if got := ps.GetAssignmentByPackage("foo"); !reflect.DeepEqual(*got, assign) {
		t.Error("GetAssignmentByPackage failed")
	}
	if ps.GetDecisionLevel() != 1 {
		t.Error("GetDecisionLevel failed")
	}
	ps.Backtrack(0)
	if len(ps.Assignments) != 0 {
		t.Error("Backtrack failed")
	}
} 