package sessionmanager

import "github.com/dsdred/universal-websocket-platform/internal/lifetimelease"

// CommitResult is the immutable result of one successful Reservation Commit.
// It publishes Registration identity and its bound Owner Lifetime Lease
// capability together.
type CommitResult struct {
	registrationID RegistrationID
	lifetimeLease  lifetimelease.Lease
}

// RegistrationID returns the committed Registration identity.
func (result CommitResult) RegistrationID() RegistrationID {
	return result.registrationID
}

// LifetimeLease returns the release capability bound to the committed
// Registration's Owner Lifetime Lease accounting.
func (result CommitResult) LifetimeLease() lifetimelease.Lease {
	if result.lifetimeLease == nil {
		return invalidLifetimeLease{}
	}
	return result.lifetimeLease
}

type boundLifetimeLease struct {
	manager        *Manager
	registrationID RegistrationID
}

var _ lifetimelease.Lease = (*boundLifetimeLease)(nil)

func (lease *boundLifetimeLease) Release() lifetimelease.ReleaseOutcome {
	if lease == nil || lease.manager == nil || lease.registrationID == (RegistrationID{}) {
		return lifetimelease.ReleaseOutcomeAccountingAnomaly
	}

	return lease.manager.releaseLifetimeLease(lease.registrationID)
}

type invalidLifetimeLease struct{}

func (invalidLifetimeLease) Release() lifetimelease.ReleaseOutcome {
	return lifetimelease.ReleaseOutcomeAccountingAnomaly
}

func (manager *Manager) releaseLifetimeLease(registrationID RegistrationID) lifetimelease.ReleaseOutcome {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	if _, exists := manager.lifetimeLeases[registrationID]; !exists {
		return lifetimelease.ReleaseOutcomeAccountingAnomaly
	}

	delete(manager.lifetimeLeases, registrationID)
	manager.closeIfAccountingEmptyLocked()
	return lifetimelease.ReleaseOutcomeReleased
}
