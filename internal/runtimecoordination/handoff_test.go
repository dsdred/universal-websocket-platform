package runtimecoordination

import (
	"errors"
	"reflect"
	"sync"
	"testing"
	"time"
)

func TestHandoffInvalidAndRootOnlySources(t *testing.T) {
	if _, _, err := NewAuthoritySource(""); !errors.Is(err, ErrInvalidScope) {
		t.Fatal(err)
	}
	for _, issuer := range []*SourceIssuer{nil, {}, {authority: NewAuthorityFence().data}} {
		if err := issuer.CommitGeneration("g"); !errors.Is(err, ErrInvalidScope) {
			t.Fatal(err)
		}
		if _, err := issuer.Observe(); !errors.Is(err, ErrInvalidScope) {
			t.Fatal(err)
		}
	}
	root, issuer, _ := NewAuthoritySource("domain")
	if _, err := issuer.Observe(); !errors.Is(err, ErrInvalidScope) {
		t.Fatal(err)
	}
	if err := issuer.CommitGeneration(""); !errors.Is(err, ErrInvalidScope) {
		t.Fatal(err)
	}
	if err := issuer.CommitGeneration("g"); err != nil {
		t.Fatal(err)
	}
	o, _ := issuer.Observe()
	legacy, _ := NewSourceAuthorityFence("domain")
	for _, r := range []*AuthorityFence{nil, {}, NewAuthorityFence(), legacy} {
		if err := r.RevalidateCurrentObservation(o); !errors.Is(err, ErrInvalidScope) {
			t.Fatal(err)
		}
	}
	if err := root.RevalidateCurrentObservation(CurrentObservation{}); !errors.Is(err, ErrInvalidScope) {
		t.Fatal(err)
	}
}

func TestHandoffCopiesBindExactlyOnce(t *testing.T) {
	root, issuer, _ := NewAuthoritySource("domain")
	copyIssuer, copyRoot := *issuer, *root
	if err := copyIssuer.CommitGeneration("g"); err != nil {
		t.Fatal(err)
	}
	for _, g := range []string{"g", "different", ""} {
		if err := issuer.CommitGeneration(g); !errors.Is(err, ErrAlreadyCommitted) {
			t.Fatal(err)
		}
	}
	o, err := issuer.Observe()
	if err != nil {
		t.Fatal(err)
	}
	copyObservation := o
	if err := copyRoot.RevalidateCurrentObservation(copyObservation); err != nil {
		t.Fatal(err)
	}
	root.BeginFatalPublication().Close()
	if _, err := copyIssuer.Observe(); !errors.Is(err, ErrClosed) {
		t.Fatal(err)
	}
	if err := copyRoot.RevalidateCurrentObservation(o); !errors.Is(err, ErrClosed) {
		t.Fatal(err)
	}
	if err := issuer.CommitGeneration("g"); !errors.Is(err, ErrClosed) {
		t.Fatal(err)
	}
}

func TestHandoffExactForeignAffinity(t *testing.T) {
	root, issuer, _ := NewAuthoritySource("domain")
	other, otherIssuer, _ := NewAuthoritySource("domain")
	_ = issuer.CommitGeneration("g")
	_ = otherIssuer.CommitGeneration("g")
	o, _ := otherIssuer.Observe()
	if err := root.RevalidateCurrentObservation(o); !errors.Is(err, ErrForeignScope) {
		t.Fatal(err)
	}
	o, _ = issuer.Observe()
	for _, field := range []string{"domain", "generation", "marker"} {
		wrong := o
		switch field {
		case "domain":
			wrong.domain = "other"
		case "generation":
			wrong.generation = "other"
		case "marker":
			wrong.marker = other.data.issuer
		}
		if err := root.RevalidateCurrentObservation(wrong); !errors.Is(err, ErrForeignScope) {
			t.Fatal(field, err)
		}
	}
}

func TestHandoffConcurrentBindingCopies(t *testing.T) {
	_, issuer, _ := NewAuthoritySource("domain")
	results := make(chan error, 16)
	var wg sync.WaitGroup
	for i := 0; i < cap(results); i++ {
		wg.Add(1)
		go func() { defer wg.Done(); copy := *issuer; results <- copy.CommitGeneration("g") }()
	}
	wg.Wait()
	close(results)
	successes := 0
	for err := range results {
		if err == nil {
			successes++
		} else if !errors.Is(err, ErrAlreadyCommitted) {
			t.Fatal(err)
		}
	}
	if successes != 1 {
		t.Fatal(successes)
	}
}

func TestHandoffFatalFirstBlocksAndRejectsObservation(t *testing.T) {
	root, issuer, _ := NewAuthoritySource("domain")
	_ = issuer.CommitGeneration("g")
	o, _ := issuer.Observe()
	fatal := root.BeginFatalPublication()
	mint, revalidate := make(chan error, 1), make(chan error, 1)
	started := make(chan struct{}, 2)
	go func() { started <- struct{}{}; _, err := issuer.Observe(); mint <- err }()
	go func() { started <- struct{}{}; revalidate <- root.RevalidateCurrentObservation(o) }()
	<-started
	<-started
	select {
	case <-mint:
		t.Fatal("mint escaped exclusive publication")
	case <-revalidate:
		t.Fatal("revalidation escaped exclusive publication")
	case <-time.After(20 * time.Millisecond):
	}
	fatal.Close()
	if err := <-mint; !errors.Is(err, ErrClosed) {
		t.Fatal(err)
	}
	if err := <-revalidate; !errors.Is(err, ErrClosed) {
		t.Fatal(err)
	}
}

func TestHandoffReadFirstOrdersFatalAndLaterInvalidates(t *testing.T) {
	root, issuer, _ := NewAuthoritySource("domain")
	_ = issuer.CommitGeneration("g")
	o, _ := issuer.Observe()
	if err := root.RevalidateCurrentObservation(o); err != nil {
		t.Fatal(err)
	}
	root.data.gate.RLock()
	started, done := make(chan struct{}), make(chan struct{})
	go func() { close(started); root.BeginFatalPublication().Close(); close(done) }()
	<-started
	select {
	case <-done:
		t.Fatal("fatal bypassed read gate")
	case <-time.After(20 * time.Millisecond):
	}
	root.data.gate.RUnlock()
	<-done
	if err := root.RevalidateCurrentObservation(o); !errors.Is(err, ErrClosed) {
		t.Fatal(err)
	}
}

func TestHandoffOpaqueSurface(t *testing.T) {
	for _, typ := range []reflect.Type{reflect.TypeOf(SourceIssuer{}), reflect.TypeOf(CurrentObservation{})} {
		for i := 0; i < typ.NumField(); i++ {
			if typ.Field(i).IsExported() {
				t.Fatal("public construction field", typ.Field(i).Name)
			}
		}
	}
	if reflect.TypeOf(CurrentObservation{}).NumMethod() != 0 {
		t.Fatal("observation exposes capability")
	}
	for _, name := range []string{"Issuer", "SourceIssuer", "CommitGeneration", "Observe", "Open", "Release"} {
		if _, ok := reflect.TypeOf(&AuthorityFence{}).MethodByName(name); ok {
			t.Fatal("root recovers issuer", name)
		}
	}
}
