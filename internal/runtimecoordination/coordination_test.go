package runtimecoordination

import (
	"errors"
	"sync"
	"testing"
	"time"
)

func testScope(t *testing.T, instance string) Scope {
	t.Helper()
	s, err := NewScope("domain", 1, 2, instance)
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func TestScopeAndZeroHandlesReject(t *testing.T) {
	for _, args := range []Scope{{}, {workspace: 1, configuration: 2, instance: "a"}, {domain: "d", configuration: 2, instance: "a"}, {domain: "d", workspace: 1, instance: "a"}, {domain: "d", workspace: 1, configuration: 2}} {
		if args.Valid() {
			t.Fatal("invalid scope accepted")
		}
		if _, err := NewScope(args.domain, args.workspace, args.configuration, args.instance); !errors.Is(err, ErrInvalidScope) {
			t.Fatal(err)
		}
	}
	if _, err := NewParticipant(Scope{}); !errors.Is(err, ErrInvalidScope) {
		t.Fatal(err)
	}
	for _, p := range []*Participant{nil, {}} {
		if _, err := p.AttachAggregate(testScope(t, "a")); !errors.Is(err, ErrInvalidScope) {
			t.Fatal(err)
		}
	}
	if _, err := (AggregateAttachment{}).Observe(testScope(t, "a")); !errors.Is(err, ErrInvalidScope) {
		t.Fatal(err)
	}
	if err := (AggregateAttachment{}).Transition(testScope(t, "a")); !errors.Is(err, ErrInvalidScope) {
		t.Fatal(err)
	}
	var zero ObservationRegion
	zero.Close()
	(*ObservationRegion)(nil).Close()
}

func TestAttachmentNonRebindingThroughParticipantCopies(t *testing.T) {
	scope := testScope(t, "a")
	p, _ := NewParticipant(scope)
	before := *p
	if _, err := before.AttachAggregate(testScope(t, "foreign")); !errors.Is(err, ErrForeignScope) {
		t.Fatal(err)
	}
	a, err := before.AttachAggregate(scope)
	if err != nil {
		t.Fatal(err)
	}
	after := *p
	for _, copy := range []*Participant{p, &before, &after} {
		if _, err := copy.AttachAggregate(scope); !errors.Is(err, ErrAlreadyAttached) {
			t.Fatal(err)
		}
	}
	if _, err := a.Observe(Scope{}); !errors.Is(err, ErrInvalidScope) {
		t.Fatal(err)
	}
	if _, err := a.Observe(testScope(t, "foreign")); !errors.Is(err, ErrForeignScope) {
		t.Fatal(err)
	}
	if err := a.Transition(scope); !errors.Is(err, ErrClosed) {
		t.Fatal(err)
	}
	if err := a.Transition(testScope(t, "foreign")); !errors.Is(err, ErrForeignScope) {
		t.Fatal(err)
	}
}

func TestCopiedRegionReleaseExactlyOnce(t *testing.T) {
	scope := testScope(t, "a")
	p, _ := NewParticipant(scope)
	a, _ := p.AttachAggregate(scope)
	r, err := a.Observe(scope)
	if err != nil {
		t.Fatal(err)
	}
	copy := *r
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() { defer wg.Done(); copy.Close(); r.Close() }()
	}
	wg.Wait()
	r.Close()
	next, err := a.Observe(scope)
	if err != nil {
		t.Fatal(err)
	}
	next.Close()
}

func TestIndependentInstanceDuringRetainedOuter(t *testing.T) {
	scope := testScope(t, "a")
	p, _ := NewParticipant(scope)
	a, _ := p.AttachAggregate(scope)
	_ = a
	// Fixture-only outer retention models external work, never a short owner region.
	p.data.outer.Lock()
	defer p.data.outer.Unlock()
	if !p.data.state.TryLock() {
		t.Fatal("state lock retained during external work")
	}
	p.data.state.Unlock()
	otherScope := testScope(t, "b")
	other, _ := NewParticipant(otherScope)
	b, _ := other.AttachAggregate(otherScope)
	done := make(chan error, 1)
	go func() {
		region, err := b.Observe(otherScope)
		if err == nil {
			region.Close()
		}
		done <- err
	}()
	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("independent Instance blocked")
	}
	if p.data.outer.TryLock() {
		p.data.outer.Unlock()
		t.Fatal("retained outer was lost")
	}
}
