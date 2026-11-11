package runtimelifecycle

import (
	"context"

	"github.com/dsdred/universal-websocket-platform/internal/runtime"
	"github.com/dsdred/universal-websocket-platform/internal/runtimeconfigload"
	"github.com/dsdred/universal-websocket-platform/internal/runtimediagnostics"
	"github.com/dsdred/universal-websocket-platform/internal/runtimediagnosticspull"
)

type diagnosticsLifetime struct {
	attemptID runtimeconfigload.LaunchAttemptID
	view      runtimediagnosticspull.View
	active    bool
}

func diagnosticsBaseDisabled(dependencies *runtime.DependencyBindings) bool {
	return dependencies != nil &&
		dependencies.DiagnosticsProducer == (runtimediagnostics.Producer{})
}

func newDiagnosticsLifetime(
	attemptID runtimeconfigload.LaunchAttemptID,
	base *runtime.DependencyBindings,
) (*diagnosticsLifetime, *runtime.DependencyBindings) {
	producer, reader := runtimediagnostics.New()
	lifetime := &diagnosticsLifetime{
		attemptID: attemptID,
		view:      runtimediagnosticspull.New(*reader),
		active:    true,
	}
	dependencies := *base
	dependencies.DiagnosticsProducer = *producer
	return lifetime, &dependencies
}

func retireDiagnosticsLocked(attempt *attemptState) {
	if attempt == nil || attempt.diagnostics == nil {
		return
	}
	attempt.diagnostics.active = false
	attempt.diagnostics.view = runtimediagnosticspull.View{}
	attempt.diagnostics = nil
}

// PollDiagnosticsExpectedAttempt performs one bounded diagnostics Poll only for
// the exact active Launch Attempt. It captures the private View under the Owner
// mutex and releases that mutex before Poll samples context or reads the mailbox.
// False is the normalized unavailable result and performs no mailbox read.
func (o *Owner) PollDiagnosticsExpectedAttempt(
	ctx context.Context,
	expectedAttemptID runtimeconfigload.LaunchAttemptID,
	budget uint8,
) (runtimediagnosticspull.PollResult, bool) {
	if o == nil || ctx == nil || expectedAttemptID == "" || budget == 0 || budget > 64 {
		return runtimediagnosticspull.PollResult{}, false
	}
	if ctx.Err() != nil {
		return runtimediagnosticspull.PollResult{}, false
	}

	o.mu.Lock()
	attempt := o.active
	if attempt == nil ||
		attempt.fact.launchAttemptID != expectedAttemptID ||
		attempt.diagnostics == nil ||
		!attempt.diagnostics.active ||
		attempt.diagnostics.attemptID != expectedAttemptID {
		o.mu.Unlock()
		return runtimediagnosticspull.PollResult{}, false
	}
	view := attempt.diagnostics.view
	o.mu.Unlock()

	return view.Poll(ctx, budget), true
}
