package runtimecoordination

import (
	"errors"
	"math"
	"sync"
	"testing"
	"time"
)

func sourceFixture(t *testing.T, instance string) (*AuthorityFence, *Participant, Scope) {
	t.Helper()
	s := testScope(t, instance)
	a, err := NewSourceAuthorityFence(s.Domain())
	if err != nil {
		t.Fatal(err)
	}
	p, err := NewSourceParticipant(a, s)
	if err != nil {
		t.Fatal(err)
	}
	return a, p, s
}

func TestSourceRootForeignAndLegacyBindings(t *testing.T) {
	root, p, scope := sourceFixture(t, "a")
	if _, err := NewSourceAuthorityFence(""); !errors.Is(err, ErrInvalidScope) {
		t.Fatal(err)
	}
	for _, a := range []*AuthorityFence{nil, {}} {
		if _, err := NewSourceParticipant(a, scope); !errors.Is(err, ErrInvalidScope) {
			t.Fatal(err)
		}
	}
	foreign, _ := NewSourceAuthorityFence("other")
	if _, err := NewSourceParticipant(foreign, scope); !errors.Is(err, ErrForeignScope) {
		t.Fatal(err)
	}
	sameDomain, _ := NewSourceAuthorityFence(scope.Domain())
	for _, a := range []*AuthorityFence{NewAuthorityFence(), sameDomain} {
		c, _ := NewClientEpoch(a)
		if _, _, err := c.RegisterCommand(p, scope); !errors.Is(err, ErrForeignScope) {
			t.Fatal("source rebound", err)
		}
	}
	c, _ := NewClientEpoch(root)
	legacy, _ := NewParticipant(scope)
	if _, _, err := c.RegisterCommand(legacy, scope); !errors.Is(err, ErrForeignScope) {
		t.Fatal("unassociated participant accepted", err)
	}
	copy := *root
	c, _ = NewClientEpoch(&copy)
	_, r, err := c.RegisterCommand(p, scope)
	if err != nil {
		t.Fatal(err)
	}
	r.Commit()
	r.Close()
	if p.data.roots.authority != root.data || p.data.source != root.data {
		t.Fatal("copy lost association")
	}
}

func TestSourceFatalCoversAllClientsPartialAndAbortedParticipants(t *testing.T) {
	root, p, scope := sourceFixture(t, "aggregate-only")
	aggregate, _ := p.AttachAggregate(scope)
	participants := []*Participant{p}
	var clients []*ClientEpoch
	for i, name := range []string{"a", "b"} {
		c, _ := NewClientEpoch(root)
		s := testScope(t, name)
		p, _ := NewSourceParticipant(root, s)
		_, r, err := c.RegisterCommand(p, s)
		if err != nil {
			t.Fatal(err)
		}
		if i == 0 {
			r.Commit()
		} // second is deliberately aborted.
		r.Close()
		participants = append(participants, p)
		clients = append(clients, c)
	}
	// Retained outer models external work; fatal must never acquire it.
	p.data.outer.Lock()
	done := make(chan struct{})
	go func() { r := root.BeginFatalPublication(); r.Invalidate(); r.Close(); close(done) }()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("fatal waited for outer")
	}
	p.data.outer.Unlock()
	for _, p := range participants {
		if !p.data.terminal || p.data.epoch != math.MaxUint64 {
			t.Fatal("partial source participant escaped")
		}
		if _, err := p.AttachAggregate(p.data.scope); !errors.Is(err, ErrClosed) {
			t.Fatal(err)
		}
	}
	for _, c := range clients {
		if !c.data.fenced || c.data.epoch != math.MaxUint64 {
			t.Fatal("client escaped")
		}
		if _, err := c.BeginReplacement(); !errors.Is(err, ErrClosed) {
			t.Fatal("replacement resurrected", err)
		}
	}
	if _, err := aggregate.Observe(scope); !errors.Is(err, ErrClosed) {
		t.Fatal(err)
	}
	if _, err := NewClientEpoch(root); !errors.Is(err, ErrClosed) {
		t.Fatal(err)
	}
	if _, err := NewSourceParticipant(root, testScope(t, "late")); !errors.Is(err, ErrClosed) {
		t.Fatal(err)
	}
}

func TestSourceFatalCopyZeroAndCompletionOrdering(t *testing.T) {
	var zero FatalPublication
	zero.Invalidate()
	zero.Close()
	(*FatalPublication)(nil).Close()
	var nilRoot *AuthorityFence
	nilRoot.BeginFatalPublication().Close()
	nilRoot.JoinFatalPublication()
	root, p, scope := sourceFixture(t, "a")
	aggregate, _ := p.AttachAggregate(scope)
	read, _ := aggregate.Observe(scope)
	started, done := make(chan struct{}), make(chan struct{})
	go func() { close(started); r := root.BeginFatalPublication(); r.Invalidate(); r.Close(); close(done) }()
	<-started
	select {
	case <-done:
		t.Fatal("fatal bypassed aggregate-only read")
	case <-time.After(20 * time.Millisecond):
	}
	read.Close()
	<-done
	r := root.BeginFatalPublication()
	copy := *r
	joined := make(chan struct{})
	go func() { root.JoinFatalPublication(); close(joined) }()
	select {
	case <-joined:
		t.Fatal("completion bypassed publisher")
	case <-time.After(20 * time.Millisecond):
	}
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() { defer wg.Done(); copy.Invalidate(); copy.Close(); r.Close() }()
	}
	wg.Wait()
	<-joined
	if !root.data.terminal || !p.data.terminal {
		t.Fatal("copy lost terminal disposition")
	}
}

func TestSourceFatalRegistrationAndReplacementRaces(t *testing.T) {
	for i := 0; i < 40; i++ {
		root, p, scope := sourceFixture(t, "a")
		c, _ := NewClientEpoch(root)
		var wg sync.WaitGroup
		wg.Add(3)
		go func() {
			defer wg.Done()
			_, r, err := c.RegisterCommand(p, scope)
			if err == nil {
				r.Commit()
				r.Close()
			} else if !errors.Is(err, ErrClosed) {
				t.Error(err)
			}
		}()
		go func() {
			defer wg.Done()
			r, err := c.BeginReplacement()
			if err == nil {
				r.Close()
			} else if !errors.Is(err, ErrClosed) {
				t.Error(err)
			}
		}()
		go func() { defer wg.Done(); r := root.BeginFatalPublication(); r.Invalidate(); r.Close() }()
		wg.Wait()
		if !p.data.terminal || p.data.epoch != math.MaxUint64 || !c.data.fenced || c.data.epoch != math.MaxUint64 {
			t.Fatal("race resurrected source state")
		}
	}
}

func TestSourceConcurrentRegistryAndExhaustedReplacement(t *testing.T) {
	root, _, _ := sourceFixture(t, "unused")
	var wg sync.WaitGroup
	for i := 0; i < 30; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			c, err := NewClientEpoch(root)
			if err != nil {
				t.Error(err)
				return
			}
			p, err := NewSourceParticipant(root, testScope(t, "registered"))
			if err != nil {
				t.Error(err)
				return
			}
			_, r, err := c.RegisterCommand(p, p.data.scope)
			if err != nil {
				t.Error(err)
				return
			}
			r.Commit()
			r.Close()
			c.data.epoch = math.MaxUint64
			if _, err = c.BeginReplacement(); !errors.Is(err, ErrClosed) {
				t.Error(err)
			}
		}()
	}
	wg.Wait()
	r := root.BeginFatalPublication()
	r.Close()
	if len(root.data.clients) != 30 || len(root.data.participants) != 31 {
		t.Fatal("incomplete source registry")
	}
}
