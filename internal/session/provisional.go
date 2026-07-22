package session

import (
	"context"

	"github.com/dsdred/universal-websocket-platform/internal/executionowner"
)

// provisionalSession is the structurally complete, transaction-local
// Session-side preparation for a future Commit. Its component identities are
// fixed at construction, but the caller retains transport, cancellation, and
// execution ownership until that boundary.
type provisionalSession struct {
	core      *sessionCore
	session   *DefaultSession
	owner     *executionowner.Owner
	cleanup   *sessionCleanup
	lifecycle *ownerSessionLifecycle
}

type ownerSessionLifecycle struct {
	session *DefaultSession
	cleanup *sessionCleanup
}

func (lifecycle *ownerSessionLifecycle) Start(ctx context.Context) error {
	return lifecycle.session.Start(ctx)
}

func (lifecycle *ownerSessionLifecycle) Run(ctx context.Context) error {
	return lifecycle.session.Run(ctx)
}

func (lifecycle *ownerSessionLifecycle) Cleanup(
	ctx context.Context,
) (executionowner.CleanupCategory, executionowner.CancellationCategory) {
	acknowledgement := lifecycle.cleanup.run(ctx)
	cleanupCategory := executionowner.CleanupCategorySucceeded
	if acknowledgement.panics != cleanupPanicNone {
		cleanupCategory = executionowner.CleanupCategoryPanicked
	} else {
		switch acknowledgement.transport {
		case transportCleanupFailed:
			cleanupCategory = executionowner.CleanupCategoryFailed
		case transportCleanupPanicked:
			cleanupCategory = executionowner.CleanupCategoryPanicked
		}
	}
	cancellationCategory := executionowner.CancellationCategoryAnomaly
	if acknowledgement.cancellation == cancellationConfirmed {
		cancellationCategory = executionowner.CancellationCategoryConfirmed
	}
	return cleanupCategory, cancellationCategory
}

func prepareProvisionalSession(
	core *sessionCore,
	connection websocketConnection,
	cancellation cancellationDependency,
) (*provisionalSession, error) {
	cancellationCell, err := newCancellationCell(cancellation)
	if err != nil {
		return nil, err
	}
	runtimeSession, err := newSessionFromCore(core, connection)
	if err != nil {
		return nil, err
	}
	owner := executionowner.New()
	cleanup := newSessionCleanup(runtimeSession, cancellationCell)
	lifecycle := &ownerSessionLifecycle{session: runtimeSession, cleanup: cleanup}

	return &provisionalSession{
		core:      core,
		session:   runtimeSession,
		owner:     owner,
		cleanup:   cleanup,
		lifecycle: lifecycle,
	}, nil
}
