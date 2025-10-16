package gql

import "fmt"

// Milestone represents the implementation stage of the experimental GQL front-end.
type Milestone int

const (
	// MilestoneUnknown is used when no milestone was specified.
	MilestoneUnknown Milestone = iota
	// Milestone1Readiness tracks the architecture and catalog readiness stage.
	Milestone1Readiness
	// Milestone2ParserValidation captures the parser and semantic validation milestone.
	Milestone2ParserValidation
	// Milestone3PlanningExecution indicates planner and executor integration work.
	Milestone3PlanningExecution
	// Milestone4EcosystemTooling contains ecosystem and tooling enablement efforts.
	Milestone4EcosystemTooling
	// Milestone5LaunchPreparation denotes verification and launch preparation tasks.
	Milestone5LaunchPreparation
)

// String implements fmt.Stringer for Milestone.
func (m Milestone) String() string {
	switch m {
	case Milestone1Readiness:
		return "milestone1-readiness"
	case Milestone2ParserValidation:
		return "milestone2-parser-validation"
	case Milestone3PlanningExecution:
		return "milestone3-planning-execution"
	case Milestone4EcosystemTooling:
		return "milestone4-ecosystem-tooling"
	case Milestone5LaunchPreparation:
		return "milestone5-launch-preparation"
	default:
		return "milestone-unknown"
	}
}

// Capability represents a functional capability gated by milestones.
type Capability int

const (
	// CapabilityParsing covers syntactic analysis and statement handling.
	CapabilityParsing Capability = iota
	// CapabilitySemanticValidation covers catalog lookups and semantic checks.
	CapabilitySemanticValidation
	// CapabilityExecution covers query planning and execution.
	CapabilityExecution
)

// String implements fmt.Stringer for Capability.
func (c Capability) String() string {
	switch c {
	case CapabilityParsing:
		return "parsing"
	case CapabilitySemanticValidation:
		return "semantic validation"
	case CapabilityExecution:
		return "execution"
	default:
		return fmt.Sprintf("capability(%d)", c)
	}
}

// Supports reports whether the milestone enables the provided capability.
func (m Milestone) Supports(cap Capability) bool {
	switch cap {
	case CapabilityParsing:
		return m >= Milestone2ParserValidation
	case CapabilitySemanticValidation:
		return m >= Milestone2ParserValidation
	case CapabilityExecution:
		return m >= Milestone3PlanningExecution
	default:
		return false
	}
}

// MilestoneError is returned when a capability is requested that is not enabled
// for the configured milestone.
type MilestoneError struct {
	Milestone  Milestone
	Capability Capability
}

func (e *MilestoneError) Error() string {
	return fmt.Sprintf("gql: %s is unavailable for %s", e.Capability, e.Milestone)
}

// Is allows errors.Is checks against ErrNotImplemented for forward compatibility.
func (e *MilestoneError) Is(target error) bool {
	return target == ErrNotImplemented
}
