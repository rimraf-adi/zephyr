package solver

import (
	"fmt"
	"strings"
)

// ErrorReport represents a human-readable error report
type ErrorReport struct {
	Lines []string
}

// GenerateErrorReport generates a human-readable error report from a derivation graph
func (s *Solver) GenerateErrorReport(rootIncompatibility Incompatibility) *ErrorReport {
	report := &ErrorReport{
		Lines: []string{},
	}
	
	// Build the derivation graph
	graph := s.buildDerivationGraph(rootIncompatibility)
	
	// Generate the report
	s.generateReportLines(graph, report)
	
	return report
}

// DerivationNode represents a node in the derivation graph
type DerivationNode struct {
	Incompatibility Incompatibility
	Causes          []*DerivationNode
	OutgoingEdges   int
	LineNumber      int
}

// buildDerivationGraph builds the derivation graph for an incompatibility
func (s *Solver) buildDerivationGraph(root Incompatibility) *DerivationNode {
	// This is a simplified implementation
	// In the full algorithm, this would traverse the cause relationships
	
	node := &DerivationNode{
		Incompatibility: root,
		Causes:          []*DerivationNode{},
		OutgoingEdges:   0,
		LineNumber:      0,
	}
	
	// Count outgoing edges
	if root.Cause != nil {
		node.OutgoingEdges++
	}
	
	return node
}

// generateReportLines generates the lines of the error report
func (s *Solver) generateReportLines(node *DerivationNode, report *ErrorReport) {
	// This is a simplified implementation of the error reporting algorithm
	// In the full algorithm, this would follow the complex rules described in the paper
	
	if len(node.Causes) == 0 {
		// External incompatibility
		line := s.formatExternalIncompatibility(node.Incompatibility)
		report.Lines = append(report.Lines, line)
		return
	}
	
	if len(node.Causes) == 1 {
		// Single cause
		s.generateReportLines(node.Causes[0], report)
		line := s.formatDerivedIncompatibility(node.Incompatibility, node.Causes[0].Incompatibility)
		report.Lines = append(report.Lines, line)
	} else if len(node.Causes) == 2 {
		// Two causes
		s.generateReportLines(node.Causes[0], report)
		s.generateReportLines(node.Causes[1], report)
		line := s.formatTwoCauseIncompatibility(node.Incompatibility, node.Causes[0].Incompatibility, node.Causes[1].Incompatibility)
		report.Lines = append(report.Lines, line)
	}
	
	// Add line number if this incompatibility causes multiple others
	if node.OutgoingEdges > 1 {
		node.LineNumber = len(report.Lines)
	}
}

// formatExternalIncompatibility formats an external incompatibility
func (s *Solver) formatExternalIncompatibility(incompatibility Incompatibility) string {
	// This is a simplified implementation
	// In the full algorithm, this would format based on the type of external incompatibility
	
	if len(incompatibility.Terms) == 1 {
		term := incompatibility.Terms[0]
		if term.Package == s.rootPackage {
			return fmt.Sprintf("The root package %s cannot be selected", term.Version.String())
		}
		return fmt.Sprintf("Package %s %s cannot be selected", term.Package, term.Version.String())
	}
	
	// Format dependency incompatibility
	var terms []string
	for _, term := range incompatibility.Terms {
		if term.Negated {
			terms = append(terms, fmt.Sprintf("not %s %s", term.Package, term.Version.String()))
		} else {
			terms = append(terms, fmt.Sprintf("%s %s", term.Package, term.Version.String()))
		}
	}
	
	return fmt.Sprintf("Dependency conflict: %s", strings.Join(terms, " and "))
}

// formatDerivedIncompatibility formats a derived incompatibility with one cause
func (s *Solver) formatDerivedIncompatibility(incompatibility, cause Incompatibility) string {
	// This is a simplified implementation
	return fmt.Sprintf("Because %s, %s", s.formatIncompatibility(cause), s.formatIncompatibility(incompatibility))
}

// formatTwoCauseIncompatibility formats a derived incompatibility with two causes
func (s *Solver) formatTwoCauseIncompatibility(incompatibility, cause1, cause2 Incompatibility) string {
	// This is a simplified implementation
	return fmt.Sprintf("Because %s and %s, %s", 
		s.formatIncompatibility(cause1), 
		s.formatIncompatibility(cause2), 
		s.formatIncompatibility(incompatibility))
}

// formatIncompatibility formats an incompatibility for display
func (s *Solver) formatIncompatibility(incompatibility Incompatibility) string {
	if len(incompatibility.Terms) == 1 {
		term := incompatibility.Terms[0]
		if term.Package == s.rootPackage {
			return fmt.Sprintf("root %s", term.Version.String())
		}
		return fmt.Sprintf("%s %s", term.Package, term.Version.String())
	}
	
	var terms []string
	for _, term := range incompatibility.Terms {
		if term.Negated {
			terms = append(terms, fmt.Sprintf("not %s %s", term.Package, term.Version.String()))
		} else {
			terms = append(terms, fmt.Sprintf("%s %s", term.Package, term.Version.String()))
		}
	}
	
	return fmt.Sprintf("{%s}", strings.Join(terms, ", "))
}

// String returns the error report as a string
func (er *ErrorReport) String() string {
	return strings.Join(er.Lines, "\n")
} 