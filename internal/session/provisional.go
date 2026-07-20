package session

import "github.com/dsdred/universal-websocket-platform/internal/executionowner"

// provisionalSession is transaction-local preparation for a future Commit.
// It owns no transport or execution authority before that boundary.
type provisionalSession struct {
	core    *sessionCore
	session *DefaultSession
	owner   *executionowner.Owner
}

func prepareProvisionalSession(
	core *sessionCore,
	connection websocketConnection,
) (*provisionalSession, error) {
	runtimeSession, err := newSessionFromCore(core, connection)
	if err != nil {
		return nil, err
	}

	return &provisionalSession{
		core:    core,
		session: runtimeSession,
		owner:   executionowner.New(),
	}, nil
}
