package runtimecoordination

import (
	"errors"
	"math"
	"sync"
	"testing"
	"time"
)

func TestRootZeroCopyAbortAndDisposedHandles(t *testing.T) {
	var registration RegistrationRegion
	var replacement ReplacementRegion
	registration.Close()
	registration.Commit()
	replacement.Close()
	replacement.Fence()
	(*RegistrationRegion)(nil).Close()
	(*ReplacementRegion)(nil).Close()
	if _, err := NewClientEpoch(&AuthorityFence{}); !errors.Is(err, ErrInvalidScope) {
		t.Fatal(err)
	}
	if _, err := (&ClientEpoch{}).BeginReplacement(); !errors.Is(err, ErrInvalidScope) {
		t.Fatal(err)
	}
	authority := NewAuthorityFence()
	authorityCopy := *authority
	client, _ := NewClientEpoch(&authorityCopy)
	clientCopy := *client
	scope := testScope(t, "a")
	p, _ := NewParticipant(scope)
	aborted, r, err := clientCopy.RegisterCommand(p, scope)
	if err != nil {
		t.Fatal(err)
	}
	copy := *r
	copy.Close()
	r.Close()
	copy.Commit()
	if _, err := aborted.Observe(scope); !errors.Is(err, ErrInvalidScope) {
		t.Fatal("aborted attachment became usable", err)
	}
	a, r, err := client.RegisterCommand(p, scope)
	if err != nil {
		t.Fatal(err)
	}
	r.Commit()
	r.Commit()
	copy = *r
	copy.Close()
	r.Close()
	if _, _, err := clientCopy.RegisterCommand(p, scope); !errors.Is(err, ErrAlreadyAttached) {
		t.Fatal(err)
	}
	if _, err := aborted.Observe(scope); !errors.Is(err, ErrInvalidScope) {
		t.Fatal("old aborted attachment revived", err)
	}
	region, err := a.Observe(scope)
	if err != nil {
		t.Fatal(err)
	}
	beforeClient, beforeParticipant := region.Epochs()
	regionCopy := *region
	regionCopy.Close()
	region.Close()
	replace, err := clientCopy.BeginReplacement()
	if err != nil {
		t.Fatal(err)
	}
	replaceCopy := *replace
	replaceCopy.Close()
	replace.Close()
	replaceCopy.Fence()
	if client.data.fenced {
		t.Fatal("disposed replacement fenced root")
	}
	if c, p := region.Epochs(); c != beforeClient || p != beforeParticipant {
		t.Fatal("old region reads mutable epochs")
	}
	current, _ := a.Observe(scope)
	c, pEpoch := current.Epochs()
	current.Close()
	if c != 1 || pEpoch != 1 {
		t.Fatalf("epochs %d/%d", c, pEpoch)
	}
}

func TestRootlessRegionBlocksRegistrationUntilClose(t *testing.T) {
	scope := testScope(t, "a")
	p, _ := NewParticipant(scope)
	a, _ := p.AttachAggregate(scope)
	r, _ := a.Observe(scope)
	if p.data.state.TryLock() {
		p.data.state.Unlock()
		t.Fatal("short observation lost state barrier")
	}
	client, _ := NewClientEpoch(NewAuthorityFence())
	started := make(chan struct{})
	done := make(chan error, 1)
	go func() {
		close(started)
		_, registration, err := client.RegisterCommand(p, scope)
		if err == nil {
			registration.Commit()
			registration.Close()
		}
		done <- err
	}()
	<-started
	select {
	case err := <-done:
		t.Fatal("registration bypassed region", err)
	case <-time.After(20 * time.Millisecond):
	}
	copy := *r
	copy.Close()
	r.Close()
	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(time.Second):
		t.Fatal("registration did not resume")
	}
	if !p.data.state.TryLock() {
		t.Fatal("state retained after close")
	}
	p.data.state.Unlock()
}

func TestReplacementIgnoresRetainedOuterAndInvalidatesEveryAttachment(t *testing.T) {
	client, _ := NewClientEpoch(NewAuthorityFence())
	var participants []*Participant
	var attachments []CommandAttachment
	for _, instance := range []string{"a", "b"} {
		scope := testScope(t, instance)
		p, _ := NewParticipant(scope)
		a, r, err := client.RegisterCommand(p, scope)
		if err != nil {
			t.Fatal(err)
		}
		r.Commit()
		r.Close()
		participants = append(participants, p)
		attachments = append(attachments, a)
	}
	// Fixture-only external work retains outer, never short-region gates.
	participants[0].data.outer.Lock()
	defer participants[0].data.outer.Unlock()
	done := make(chan error, 1)
	go func() {
		r, err := client.BeginReplacement()
		if err == nil {
			r.Close()
		}
		done <- err
	}()
	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(time.Second):
		t.Fatal("replacement waited for outer")
	}
	for _, p := range participants {
		p.data.state.Lock()
		epoch := p.data.epoch
		p.data.state.Unlock()
		if epoch != 1 {
			t.Fatal("attachment missed fanout")
		}
	}
	r, _ := attachments[1].Observe(testScope(t, "b"))
	r.Close()
}

func TestRootExhaustionPermanentlyFencesWithoutEpochPublication(t *testing.T) {
	client, _ := NewClientEpoch(NewAuthorityFence())
	scope := testScope(t, "a")
	p, _ := NewParticipant(scope)
	_, r, _ := client.RegisterCommand(p, scope)
	r.Commit()
	r.Close()
	client.data.epoch = math.MaxUint64
	if _, err := client.BeginReplacement(); !errors.Is(err, ErrClosed) {
		t.Fatal(err)
	}
	if !client.data.fenced || client.data.epoch != math.MaxUint64 || p.data.epoch != math.MaxUint64 {
		t.Fatal("exhaustion published wrapped epoch")
	}
	client.data.epoch = 0
	if _, err := client.BeginReplacement(); !errors.Is(err, ErrClosed) {
		t.Fatal("fence was not sticky", err)
	}
}

func TestRegistrationRacesReplacementWithoutEscape(t *testing.T) {
	client, _ := NewClientEpoch(NewAuthorityFence())
	var wg sync.WaitGroup
	for i := 0; i < 30; i++ {
		wg.Add(2)
		go func(i int) {
			defer wg.Done()
			scope := testScope(t, string(rune('a'+i)))
			p, _ := NewParticipant(scope)
			_, r, err := client.RegisterCommand(p, scope)
			if err != nil {
				t.Error(err)
				return
			}
			r.Commit()
			r.Close()
		}(i)
		go func() {
			defer wg.Done()
			r, err := client.BeginReplacement()
			if err != nil {
				t.Error(err)
				return
			}
			r.Close()
		}()
	}
	wg.Wait()
	client.data.gate.RLock()
	defer client.data.gate.RUnlock()
	for p := range client.data.participants {
		p.state.Lock()
		epoch := p.epoch
		p.state.Unlock()
		if epoch != client.data.epoch {
			t.Fatal("registration escaped current epoch")
		}
	}
}
