package session

import "github.com/dsdred/universal-websocket-platform/internal/executionowner"

// provisionalSession is the structurally complete, transaction-local
// Session-side preparation for a future Commit. Its component identities are
// fixed at construction, but the caller retains transport, cancellation, and
// execution ownership until that boundary.
type provisionalSession struct {
	core    *sessionCore
	session *DefaultSession
	owner   *executionowner.Owner
	cleanup *sessionCleanup
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

	return &provisionalSession{
		core:    core,
		session: runtimeSession,
		owner:   owner,
		cleanup: cleanup,
	}, nil
}
