package executionowner

// CompleteOutcome is the immutable result of one bound Registration
// completion attempt.
type CompleteOutcome uint8

const (
	// CompleteOutcomeCompleted indicates that the bound Registration was
	// removed by this completion attempt.
	CompleteOutcomeCompleted CompleteOutcome = iota + 1
	// CompleteOutcomeAccountingAnomaly indicates that the bound Registration
	// was already removed or could not be found.
	CompleteOutcomeAccountingAnomaly
)

// CompletionAdapter exposes terminal completion for exactly one bound
// Registration without exposing its identity or Session Manager.
type CompletionAdapter interface {
	CompleteBoundRegistration() CompleteOutcome
}
