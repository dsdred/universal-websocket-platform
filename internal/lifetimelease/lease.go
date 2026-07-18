// Package lifetimelease defines the narrow Owner Lifetime Lease release
// contract shared by Session Manager accounting and Execution Owner wiring.
package lifetimelease

// ReleaseOutcome is the immutable result of one bound Owner Lifetime Lease
// release attempt.
type ReleaseOutcome uint8

const (
	// ReleaseOutcomeReleased indicates that the bound Lease accounting was
	// removed by this attempt.
	ReleaseOutcomeReleased ReleaseOutcome = iota + 1
	// ReleaseOutcomeAccountingAnomaly indicates that the bound Lease was
	// unknown or had already been released.
	ReleaseOutcomeAccountingAnomaly
)

// Lease exposes release for exactly one bound Owner Lifetime Lease without
// exposing a caller-selected identity or Session Manager.
type Lease interface {
	Release() ReleaseOutcome
}
