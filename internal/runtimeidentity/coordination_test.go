package runtimeidentity

import (
	"errors"
	"maps"
	"reflect"
	"slices"
	"testing"
	"time"

	"github.com/dsdred/universal-websocket-platform/internal/runtimeconfigload"
	"github.com/dsdred/universal-websocket-platform/internal/runtimecoordination"
)

func coordinatedFixture(t *testing.T, instance string) (*Store, runtimecoordination.Scope) {
	t.Helper()
	scope, err := runtimecoordination.NewScope("domain", 1, 2, instance)
	if err != nil {
		t.Fatal(err)
	}
	p, err := runtimecoordination.NewParticipant(scope)
	if err != nil {
		t.Fatal(err)
	}
	s, err := NewCoordinatedStore(scope, p)
	if err != nil {
		t.Fatal(err)
	}
	return s, scope
}

// seedDetachedFixture copies historical legacy facts solely to test complete
// detached observation/staleness. It never opens coordinated admission or proves
// a successful coordinated publication. No production fixture API exists.
func seedDetachedFixture(t *testing.T, s *Store, active bool) {
	t.Helper()
	id := runtimeconfigload.RuntimeInstanceID(s.coordination.scope.InstanceID())
	legacy := NewStore()
	if err := legacy.CreateRuntimeInstance(1, 2, id); err != nil {
		t.Fatal(err)
	}
	claim, err := legacy.ConditionalClaimLaunchAttempt(id, 1, "old", 7)
	if err != nil {
		t.Fatal(err)
	}
	terminal, err := legacy.OwnerTerminalPublisher().ConditionalPublishNoHostProduced(id, claim.Revision(), "old", false)
	if err != nil {
		t.Fatal(err)
	}
	if active {
		claim, err = legacy.ConditionalClaimLaunchAttempt(id, terminal.Revision(), "live", 8)
		if err != nil {
			t.Fatal(err)
		}
		if _, err = legacy.ConditionalBindExecutionGeneration(id, claim.Revision(), "live", "generation"); err != nil {
			t.Fatal(err)
		}
	}
	s.allocated, s.instances, s.presenceRevision = legacy.allocated, legacy.instances, legacy.presenceRevision
}

func TestCoordinatedConstructorRejectsZeroForeignDuplicateAndStoreCopy(t *testing.T) {
	scope, _ := runtimecoordination.NewScope("domain", 1, 2, "instance")
	p, _ := runtimecoordination.NewParticipant(scope)
	for _, participant := range []*runtimecoordination.Participant{nil, {}} {
		if _, err := NewCoordinatedStore(scope, participant); !errors.Is(err, runtimecoordination.ErrInvalidScope) {
			t.Fatal(err)
		}
	}
	if _, err := NewCoordinatedStore(runtimecoordination.Scope{}, p); !errors.Is(err, runtimecoordination.ErrInvalidScope) {
		t.Fatal(err)
	}
	foreign, _ := runtimecoordination.NewScope("foreign", 1, 2, "instance")
	if _, err := NewCoordinatedStore(foreign, p); !errors.Is(err, runtimecoordination.ErrForeignScope) {
		t.Fatal(err)
	}
	s, err := NewCoordinatedStore(scope, p)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := NewCoordinatedStore(scope, p); !errors.Is(err, runtimecoordination.ErrAlreadyAttached) {
		t.Fatal(err)
	}
	// Deliberately model an unsupported value copy without asking vet to allow
	// ordinary copying of mutexes. The original constructor pointer is binding.
	copy := reflect.New(reflect.TypeOf(s).Elem()).Interface().(*Store)
	reflect.ValueOf(copy).Elem().Set(reflect.ValueOf(s).Elem())
	if _, err := copy.ReadRuntimeInstance("instance"); !errors.Is(err, ErrInvalidIdentity) {
		t.Fatal(err)
	}
	if err := copy.AllocateCandidateIdentity("instance"); !errors.Is(err, ErrInvalidIdentity) {
		t.Fatal(err)
	}
	if _, err := copy.CaptureCoordinationSnapshot(scope); !errors.Is(err, ErrInvalidIdentity) {
		t.Fatal(err)
	}
}

func TestEveryCoordinatedDirectAndNarrowWriterClosed(t *testing.T) {
	s, scope := coordinatedFixture(t, "instance")
	seedDetachedFixture(t, s, true)
	before, err := s.CaptureCoordinationSnapshot(scope)
	if err != nil {
		t.Fatal(err)
	}
	owner, recovery := s.OwnerTerminalPublisher(), s.RecoveryTerminalPublisher()
	pub := func(r PublishResult, err error) error {
		if r.Committed() || r.Revision() != 0 {
			t.Fatal("rejected publication returned facts")
		}
		return err
	}
	ops := map[string]func(runtimeconfigload.RuntimeInstanceID) error{
		"allocate": func(id runtimeconfigload.RuntimeInstanceID) error { return s.AllocateCandidateIdentity(id) },
		"create":   func(id runtimeconfigload.RuntimeInstanceID) error { return s.CreateRuntimeInstance(1, 2, id) },
		"claim": func(id runtimeconfigload.RuntimeInstanceID) error {
			r, err := s.ConditionalClaimLaunchAttempt(id, 5, "new", 9)
			if r.Committed() || r.Revision() != 0 {
				t.Fatal("rejected claim returned facts")
			}
			return err
		},
		"bind": func(id runtimeconfigload.RuntimeInstanceID) error {
			return pub(s.ConditionalBindExecutionGeneration(id, 5, "live", "generation"))
		},
		"running": func(id runtimeconfigload.RuntimeInstanceID) error {
			return pub(s.ConditionalPublishRunning(id, 5, "live"))
		},
		"stop": func(id runtimeconfigload.RuntimeInstanceID) error { return pub(s.ConditionalClaimStop(id, 5, "live")) },
		"shutdown": func(id runtimeconfigload.RuntimeInstanceID) error {
			return pub(owner.ConditionalPublishShutdownCompleted(id, 5, "live"))
		},
		"no-host": func(id runtimeconfigload.RuntimeInstanceID) error {
			return pub(owner.ConditionalPublishNoHostProduced(id, 5, "live", true))
		},
		"stop-failure": func(id runtimeconfigload.RuntimeInstanceID) error {
			return pub(owner.ConditionalPublishStopFailure(id, 5, "live"))
		},
		"reconciled": func(id runtimeconfigload.RuntimeInstanceID) error {
			return pub(recovery.ConditionalPublishReconciled(id, 5, "live"))
		},
	}
	for name, op := range ops {
		t.Run(name, func(t *testing.T) {
			if err := op("instance"); !errors.Is(err, runtimecoordination.ErrClosed) {
				t.Fatal(err)
			}
			if err := op("foreign"); !errors.Is(err, runtimecoordination.ErrForeignScope) {
				t.Fatal(err)
			}
			if err := op(""); !errors.Is(err, ErrInvalidIdentity) {
				t.Fatal(err)
			}
			if err := s.RevalidateCoordinationSnapshot(before); err != nil {
				t.Fatal("mutation on rejection:", err)
			}
		})
	}
	if err := s.CreateRuntimeInstance(99, 2, "instance"); !errors.Is(err, runtimecoordination.ErrForeignScope) {
		t.Fatal(err)
	}
	if err := s.CreateRuntimeInstance(1, 99, "instance"); !errors.Is(err, runtimecoordination.ErrForeignScope) {
		t.Fatal(err)
	}
}

func TestCoordinationAbsenceAllocationAndCreateAfterAbsence(t *testing.T) {
	s, scope := coordinatedFixture(t, "instance")
	absent, err := s.CaptureCoordinationSnapshot(scope)
	if err != nil {
		t.Fatal(err)
	}
	if absent.Scope() != scope || absent.Allocated() || absent.Present() || absent.PresenceRevision() != 0 || absent.View().Revision() != 0 || len(absent.History()) != 0 {
		t.Fatal("absence not explicit")
	}
	if err := s.RevalidateCoordinationSnapshot(absent); err != nil {
		t.Fatal(err)
	}
	// Package-local presence fixture only: public allocation remains closed.
	s.allocated["instance"] = struct{}{}
	s.presenceRevision++
	if err := s.RevalidateCoordinationSnapshot(absent); !errors.Is(err, ErrStaleRevision) {
		t.Fatal(err)
	}
	allocated, err := s.CaptureCoordinationSnapshot(scope)
	if err != nil || !allocated.Allocated() || allocated.Present() {
		t.Fatal("allocated absence:", err)
	}
	seedDetachedFixture(t, s, false)
	s.presenceRevision = allocated.presenceRevision + 1
	if err := s.RevalidateCoordinationSnapshot(allocated); !errors.Is(err, ErrStaleRevision) {
		t.Fatal(err)
	}
	if err := s.RevalidateCoordinationSnapshot(absent); !errors.Is(err, ErrStaleRevision) {
		t.Fatal(err)
	}
}

func TestCoordinationCompleteDetachedStaleForeignAndPartialSnapshots(t *testing.T) {
	s, scope := coordinatedFixture(t, "instance")
	seedDetachedFixture(t, s, true)
	snapshot, err := s.CaptureCoordinationSnapshot(scope)
	if err != nil {
		t.Fatal(err)
	}
	if !snapshot.Present() || !snapshot.Allocated() || len(snapshot.History()) != 2 || snapshot.View().Revision() != 5 {
		t.Fatal("incomplete aggregate/history")
	}
	history := snapshot.History()
	history[0].configurationVersionID = 999
	if err := s.RevalidateCoordinationSnapshot(snapshot); err != nil {
		t.Fatal("history alias:", err)
	}
	if snapshot.History()[0].ConfigurationVersionID() != 7 {
		t.Fatal("accessor aliases snapshot")
	}
	other, _ := coordinatedFixture(t, "instance")
	seedDetachedFixture(t, other, true)
	if err := other.RevalidateCoordinationSnapshot(snapshot); !errors.Is(err, runtimecoordination.ErrForeignScope) {
		t.Fatal(err)
	}
	if err := s.RevalidateCoordinationSnapshot(CoordinationSnapshot{}); !errors.Is(err, ErrInvalidIdentity) {
		t.Fatal(err)
	}
	partial := snapshot
	partial.complete = false
	if err := s.RevalidateCoordinationSnapshot(partial); !errors.Is(err, ErrInvalidIdentity) {
		t.Fatal(err)
	}
	partial = snapshot
	partial.history = slices.Clone(snapshot.history[:1])
	if err := s.RevalidateCoordinationSnapshot(partial); !errors.Is(err, ErrStaleRevision) {
		t.Fatal(err)
	}
	partial = snapshot
	partial.scope, _ = runtimecoordination.NewScope("foreign", 1, 2, "instance")
	if err := s.RevalidateCoordinationSnapshot(partial); !errors.Is(err, runtimecoordination.ErrForeignScope) {
		t.Fatal(err)
	}
	// Fixture advances domain revision with coherent facts; no authority opens.
	s.instances["instance"].revision++
	if err := s.RevalidateCoordinationSnapshot(snapshot); !errors.Is(err, ErrStaleRevision) {
		t.Fatal(err)
	}
}

func TestCoordinatedReadersAndExactAttemptConform(t *testing.T) {
	s, scope := coordinatedFixture(t, "instance")
	seedDetachedFixture(t, s, true)
	if v, err := s.ReadRuntimeInstance("foreign"); !errors.Is(err, runtimecoordination.ErrForeignScope) || v != (RuntimeInstanceView{}) {
		t.Fatal(v, err)
	}
	if h, err := s.ReadLaunchAttemptHistory("foreign"); !errors.Is(err, runtimecoordination.ErrForeignScope) || h != nil {
		t.Fatal(h, err)
	}
	foreign, _ := runtimecoordination.NewScope("foreign", 1, 2, "instance")
	if _, err := s.CaptureCoordinationSnapshot(foreign); !errors.Is(err, runtimecoordination.ErrForeignScope) {
		t.Fatal(err)
	}
	if _, err := s.CaptureCoordinationSnapshot(runtimecoordination.Scope{}); !errors.Is(err, ErrInvalidIdentity) {
		t.Fatal(err)
	}
	history, err := s.ReadLaunchAttemptHistory("instance")
	if err != nil || len(history) != 2 {
		t.Fatal(history, err)
	}
	history[0].phase = AttemptPhaseRunning
	if _, err := s.CaptureCoordinationSnapshot(scope); err != nil {
		t.Fatal("read history aliases owner:", err)
	}
	exact, err := ReadExactAttemptSnapshot(s, "instance", "live")
	if err != nil || exact.ConfigurationVersionID() != 8 || exact.ExecutionGeneration() != "generation" {
		t.Fatal(exact, err)
	}
	if err := RevalidateExactAttemptSnapshot(s, exact); err != nil {
		t.Fatal(err)
	}
	other, _ := coordinatedFixture(t, "instance")
	seedDetachedFixture(t, other, true)
	if err := RevalidateExactAttemptSnapshot(other, exact); !errors.Is(err, ErrIncoherentAttemptSnapshot) {
		t.Fatal(err)
	}
	if _, err := ReadExactAttemptSnapshot(s, "foreign", "live"); !errors.Is(err, runtimecoordination.ErrForeignScope) {
		t.Fatal(err)
	}
	s.instances["instance"].revision++
	if err := RevalidateExactAttemptSnapshot(s, exact); !errors.Is(err, ErrStaleRevision) {
		t.Fatal(err)
	}
}

func TestCoordinationCaptureRejectsMalformedFullHistory(t *testing.T) {
	for name, corrupt := range map[string]func(*runtimeInstanceAggregate){
		"foreign-parent":   func(a *runtimeInstanceAggregate) { a.attempts[0].runtimeInstanceID = "foreign" },
		"duplicate":        func(a *runtimeInstanceAggregate) { a.attempts = append(a.attempts, a.attempts[0]) },
		"missing-version":  func(a *runtimeInstanceAggregate) { a.attempts[0].configurationVersionID = 0 },
		"partial-active":   func(a *runtimeInstanceAggregate) { a.attempts = a.attempts[:1] },
		"wrong-binding":    func(a *runtimeInstanceAggregate) { a.workspaceID = 99 },
		"invalid-desired":  func(a *runtimeInstanceAggregate) { a.desired = "unknown" },
		"invalid-actual":   func(a *runtimeInstanceAggregate) { a.actual = "unknown" },
		"claimed-stopped":  func(a *runtimeInstanceAggregate) { a.actual = ActualStateStopped },
		"claimed-stopping": func(a *runtimeInstanceAggregate) { a.desired = DesiredStateStopped },
		"missing-index":    func(a *runtimeInstanceAggregate) { delete(a.attemptIndex, "old") },
		"extra-index":      func(a *runtimeInstanceAggregate) { a.attemptIndex["extra"] = 0 },
		"wrong-index":      func(a *runtimeInstanceAggregate) { a.attemptIndex["old"] = 1 },
	} {
		t.Run(name, func(t *testing.T) {
			s, scope := coordinatedFixture(t, "instance")
			seedDetachedFixture(t, s, true)
			corrupt(s.instances["instance"])
			agg := s.instances["instance"]
			beforeRevision, beforeHistory, beforeIndex := agg.revision, slices.Clone(agg.attempts), maps.Clone(agg.attemptIndex)
			snapshot, err := s.CaptureCoordinationSnapshot(scope)
			if !errors.Is(err, ErrIncoherentAttemptSnapshot) || snapshot.complete || snapshot.owner != nil || len(snapshot.history) != 0 {
				t.Fatal(snapshot, err)
			}
			if beforeRevision != agg.revision || !slices.Equal(beforeHistory, agg.attempts) || !maps.Equal(beforeIndex, agg.attemptIndex) {
				t.Fatal("invalid capture mutated owner")
			}
		})
	}
	s, scope := coordinatedFixture(t, "instance")
	seedDetachedFixture(t, s, true)
	s.instances["instance"].attempts[1].executionGeneration = ""
	if _, err := s.CaptureCoordinationSnapshot(scope); err != nil {
		t.Fatal("coherent unbound history rejected:", err)
	}
	if _, err := ReadExactAttemptSnapshot(s, "instance", "live"); !errors.Is(err, ErrExecutionGenerationNotBound) {
		t.Fatal(err)
	}
}

func TestCoordinationExternalWorkHoldsNoOwnerLocksAndOtherInstanceProgresses(t *testing.T) {
	s, scope := coordinatedFixture(t, "instance")
	seedDetachedFixture(t, s, true)
	if _, err := s.CaptureCoordinationSnapshot(scope); err != nil {
		t.Fatal(err)
	}
	// Retain only the Instance outer token while simulating external work.
	region, err := s.coordination.attachment.Observe(scope)
	if err != nil {
		t.Fatal(err)
	}
	defer region.Close()
	if !s.mu.TryLock() {
		t.Fatal("catalog lock retained")
	}
	s.mu.Unlock()
	agg := s.instances["instance"]
	if !agg.mu.TryLock() {
		t.Fatal("aggregate lock retained")
	}
	agg.mu.Unlock()
	other, otherScope := coordinatedFixture(t, "other")
	done := make(chan error, 1)
	go func() { _, err := other.CaptureCoordinationSnapshot(otherScope); done <- err }()
	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("independent Instance blocked by external work")
	}
}

func TestPresenceRevisionExhaustionDoesNotMutate(t *testing.T) {
	for _, create := range []bool{false, true} {
		s := NewStore()
		s.presenceRevision = ^uint64(0)
		var err error
		if create {
			err = s.CreateRuntimeInstance(1, 2, "instance")
		} else {
			err = s.AllocateCandidateIdentity("instance")
		}
		if !errors.Is(err, ErrRevisionExhausted) || len(s.allocated) != 0 || len(s.instances) != 0 || s.presenceRevision != ^uint64(0) {
			t.Fatal("wrapped/partial presence publication:", err)
		}
	}
}

func TestEveryDomainRevisionExhaustionRejectsBeforeMutation(t *testing.T) {
	type operation struct {
		phase      AttemptPhase
		generation ExecutionGeneration
		run        func(*Store) (PublishResult, error)
	}
	ops := map[string]operation{
		"bind": {AttemptPhaseClaimed, "", func(s *Store) (PublishResult, error) {
			return s.ConditionalBindExecutionGeneration("instance", ^Revision(0), "live", "new")
		}},
		"running": {AttemptPhaseClaimed, "generation", func(s *Store) (PublishResult, error) {
			return s.ConditionalPublishRunning("instance", ^Revision(0), "live")
		}},
		"stop": {AttemptPhaseRunning, "generation", func(s *Store) (PublishResult, error) { return s.ConditionalClaimStop("instance", ^Revision(0), "live") }},
		"shutdown": {AttemptPhaseStopping, "generation", func(s *Store) (PublishResult, error) {
			return s.OwnerTerminalPublisher().ConditionalPublishShutdownCompleted("instance", ^Revision(0), "live")
		}},
		"no-host": {AttemptPhaseClaimed, "", func(s *Store) (PublishResult, error) {
			return s.OwnerTerminalPublisher().ConditionalPublishNoHostProduced("instance", ^Revision(0), "live", false)
		}},
		"stop-failure": {AttemptPhaseStopping, "generation", func(s *Store) (PublishResult, error) {
			return s.OwnerTerminalPublisher().ConditionalPublishStopFailure("instance", ^Revision(0), "live")
		}},
		"reconciled": {AttemptPhaseRunning, "generation", func(s *Store) (PublishResult, error) {
			return s.RecoveryTerminalPublisher().ConditionalPublishReconciled("instance", ^Revision(0), "live")
		}},
	}
	for name, op := range ops {
		t.Run(name, func(t *testing.T) {
			s := NewStore()
			if err := s.CreateRuntimeInstance(1, 2, "instance"); err != nil {
				t.Fatal(err)
			}
			if _, err := s.ConditionalClaimLaunchAttempt("instance", 1, "live", 7); err != nil {
				t.Fatal(err)
			}
			a := s.instances["instance"]
			// Explicit package-local exhaustion/phase fixture, no public opening.
			a.revision = ^Revision(0)
			a.attempts[0].phase, a.attempts[0].executionGeneration = op.phase, op.generation
			beforeView, _ := s.ReadRuntimeInstance("instance")
			beforeHistory, _ := s.ReadLaunchAttemptHistory("instance")
			r, err := op.run(s)
			afterView, _ := s.ReadRuntimeInstance("instance")
			afterHistory, _ := s.ReadLaunchAttemptHistory("instance")
			if !errors.Is(err, ErrRevisionExhausted) || r.Committed() || r.Revision() != 0 || beforeView != afterView || !slices.Equal(beforeHistory, afterHistory) {
				t.Fatal("wrapped/partial domain publication:", err)
			}
		})
	}
	s := NewStore()
	if err := s.CreateRuntimeInstance(1, 2, "instance"); err != nil {
		t.Fatal(err)
	}
	s.instances["instance"].revision = ^Revision(0)
	r, err := s.ConditionalClaimLaunchAttempt("instance", ^Revision(0), "live", 7)
	if !errors.Is(err, ErrRevisionExhausted) || r.Committed() || len(s.instances["instance"].attempts) != 0 || s.instances["instance"].hasActiveAttempt {
		t.Fatal("partial claim:", err)
	}
}

func TestCoordinationCaptureRejectsNilAndInactiveContradictions(t *testing.T) {
	for _, scenario := range []string{"nil", "initial-failed", "terminal-stopped", "terminal-desired-started", "allocation-missing"} {
		t.Run(scenario, func(t *testing.T) {
			s, scope := coordinatedFixture(t, "instance")
			if scenario == "initial-failed" {
				legacy := NewStore()
				if err := legacy.CreateRuntimeInstance(1, 2, "instance"); err != nil {
					t.Fatal(err)
				}
				s.instances, s.allocated, s.presenceRevision = legacy.instances, legacy.allocated, legacy.presenceRevision
				s.instances["instance"].actual = ActualStateFailed
			} else {
				seedDetachedFixture(t, s, false)
			}
			switch scenario {
			case "nil":
				s.instances["instance"] = nil
			case "terminal-stopped":
				s.instances["instance"].actual = ActualStateStopped
			case "terminal-desired-started":
				s.instances["instance"].desired = DesiredStateStarted
			case "allocation-missing":
				delete(s.allocated, "instance")
			}
			snapshot, err := s.CaptureCoordinationSnapshot(scope)
			if !errors.Is(err, ErrIncoherentAttemptSnapshot) || snapshot.complete || snapshot.owner != nil {
				t.Fatal(snapshot, err)
			}
		})
	}
}

func TestCoordinationCapturePreservesLegacyEmittedPhaseFacts(t *testing.T) {
	for _, scenario := range []string{"initial", "claimed-unbound", "running", "stopping", "stop-failure", "shutdown", "no-host-stopped", "reconciled"} {
		t.Run(scenario, func(t *testing.T) {
			s, scope := coordinatedFixture(t, "instance")
			// Legacy public success prepares observation fixtures only. The coordinated
			// owner is never opened; no lifecycle publication through it can succeed.
			legacy := NewStore()
			if err := legacy.CreateRuntimeInstance(1, 2, "instance"); err != nil {
				t.Fatal(err)
			}
			revision := Revision(1)
			if scenario != "initial" {
				claim, err := legacy.ConditionalClaimLaunchAttempt("instance", revision, "live", 7)
				if err != nil {
					t.Fatal(err)
				}
				revision = claim.Revision()
			}
			if scenario != "initial" && scenario != "claimed-unbound" && scenario != "no-host-stopped" && scenario != "reconciled" {
				bind, err := legacy.ConditionalBindExecutionGeneration("instance", revision, "live", "generation")
				if err != nil {
					t.Fatal(err)
				}
				revision = bind.Revision()
				running, err := legacy.ConditionalPublishRunning("instance", revision, "live")
				if err != nil {
					t.Fatal(err)
				}
				revision = running.Revision()
				if scenario != "running" {
					stop, err := legacy.ConditionalClaimStop("instance", revision, "live")
					if err != nil {
						t.Fatal(err)
					}
					revision = stop.Revision()
				}
			}
			var err error
			switch scenario {
			case "stop-failure":
				_, err = legacy.OwnerTerminalPublisher().ConditionalPublishStopFailure("instance", revision, "live")
			case "shutdown":
				_, err = legacy.OwnerTerminalPublisher().ConditionalPublishShutdownCompleted("instance", revision, "live")
			case "no-host-stopped":
				_, err = legacy.OwnerTerminalPublisher().ConditionalPublishNoHostProduced("instance", revision, "live", true)
			case "reconciled":
				_, err = legacy.RecoveryTerminalPublisher().ConditionalPublishReconciled("instance", revision, "live")
			}
			if err != nil {
				t.Fatal(err)
			}
			s.instances, s.allocated, s.presenceRevision = legacy.instances, legacy.allocated, legacy.presenceRevision
			snapshot, err := s.CaptureCoordinationSnapshot(scope)
			if err != nil {
				t.Fatal(err)
			}
			if err = s.RevalidateCoordinationSnapshot(snapshot); err != nil {
				t.Fatal(err)
			}
			if err = s.AllocateCandidateIdentity("instance"); !errors.Is(err, runtimecoordination.ErrClosed) {
				t.Fatal("observation opened admission:", err)
			}
		})
	}
}
