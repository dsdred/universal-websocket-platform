//go:build windows

package runtimecontainment

import (
	"bytes"
	"context"
	"errors"
	"os"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/dsdred/universal-websocket-platform/internal/runtimecoordination"
	bolt "go.etcd.io/bbolt"
	"golang.org/x/sys/windows"
)

func TestProvisionedCoordinationHandoffExactCommittedAuthority(t *testing.T) {
	a, _, _, _ := associatedFixture(t)
	before := ledgerRevision(t, a.keeper.db)
	o, err := a.InspectCoordination(context.Background(), a.keeper.domain)
	if err != nil {
		t.Fatal(err)
	}
	if err := a.keeper.root.RevalidateCurrentObservation(o); err != nil {
		t.Fatal(err)
	}
	if a.keeper.committedGeneration != a.generation || zeroIdentity(a.generation.value) {
		t.Fatal("uncommitted generation")
	}
	copy := *a
	if _, err := copy.InspectCoordination(context.Background(), a.keeper.domain); err != nil {
		t.Fatal(err)
	}
	foreign, issuer, _ := runtimecoordination.NewAuthoritySource(a.keeper.domain.text())
	_ = issuer.CommitGeneration(a.CurrentGeneration())
	if err := foreign.RevalidateCurrentObservation(o); !errors.Is(err, runtimecoordination.ErrForeignScope) {
		t.Fatal(err)
	}
	if ledgerRevision(t, a.keeper.db) != before {
		t.Fatal("inspection mutated ledger")
	}
	a.keeper.publishFatal()
	if err := a.keeper.root.RevalidateCurrentObservation(o); !errors.Is(err, runtimecoordination.ErrClosed) {
		t.Fatal(err)
	}
	if _, err := copy.InspectCoordination(context.Background(), a.keeper.domain); !errors.Is(err, runtimecoordination.ErrClosed) {
		t.Fatal(err)
	}
}

func TestCoordinationInspectionCallerErrorsAndLegacy(t *testing.T) {
	a, _, _, _ := associatedFixture(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	for _, row := range []struct {
		a      *ActiveAuthority
		ctx    context.Context
		domain Domain
		want   error
	}{
		{nil, context.Background(), a.keeper.domain, runtimecoordination.ErrInvalidScope},
		{a, nil, a.keeper.domain, runtimecoordination.ErrInvalidScope},
		{a, context.Background(), Domain{}, runtimecoordination.ErrInvalidScope},
		{a, context.Background(), testDomain(t), runtimecoordination.ErrForeignScope},
		{a, ctx, a.keeper.domain, context.Canceled},
	} {
		got, err := row.a.InspectCoordination(row.ctx, row.domain)
		if !errors.Is(err, row.want) || !reflect.DeepEqual(got, runtimecoordination.CurrentObservation{}) {
			t.Fatal(err, row.want)
		}
		if a.keeper.fenced.Load() {
			t.Fatal("caller error fabricated corruption")
		}
	}
	legacy := &ActiveAuthority{generation: a.generation, keeper: &keeper{domain: a.keeper.domain}}
	if _, err := legacy.InspectCoordination(context.Background(), a.keeper.domain); !errors.Is(err, runtimecoordination.ErrClosed) || legacy.keeper.fenced.Load() {
		t.Fatal("legacy adopted or corrupted source", err)
	}
}

func TestCoordinationInspectionFatalSourcesOutrankCancellation(t *testing.T) {
	for _, name := range []string{"capability", "ledger", "missing", "replaced", "generation", "issuer"} {
		t.Run(name, func(t *testing.T) {
			a, s, p, c := associatedFixture(t)
			switch name {
			case "capability":
				if err := windows.SetEvent(a.keeper.capability.handle); err != nil {
					t.Fatal(err)
				}
			case "ledger":
				if err := a.keeper.db.Update(func(tx *bolt.Tx) error {
					return tx.Bucket(stateBucket).Put(tailKey, bytes.Repeat([]byte{255}, identitySize))
				}); err != nil {
					t.Fatal(err)
				}
			case "missing":
				old := a.keeper.descriptor.root
				a.keeper.descriptor.root = t.TempDir()
				t.Cleanup(func() { a.keeper.descriptor.root = old })
			case "replaced":
				// Replace candidate provenance while retaining the independent
				// trusted descriptor, so storage cannot attest its own binding.
				replacement, err := newGeneration()
				if err != nil {
					t.Fatal(err)
				}
				if err := a.keeper.db.Update(func(tx *bolt.Tx) error {
					return tx.Bucket(anchorBucket).Put(storageAuthorityKey, replacement.value[:])
				}); err != nil {
					t.Fatal(err)
				}
			case "generation":
				next, _ := newGeneration()
				a.generation = next
			case "issuer":
				_, a.keeper.issuer, _ = runtimecoordination.NewAuthoritySource(a.keeper.domain.text())
				_ = a.keeper.issuer.CommitGeneration(a.CurrentGeneration())
			}
			ctx, cancel := context.WithCancel(context.Background())
			cancel()
			// Source affinity failures require the local mint gate, so reach that
			// gate uncancelled; actual external faults still beat cancellation.
			if name == "issuer" {
				ctx = context.Background()
			}
			if _, err := a.InspectCoordination(ctx, a.keeper.domain); !errors.Is(err, runtimecoordination.ErrClosed) {
				t.Fatal(err)
			}
			assertSourceFatal(t, a, s, p, c)
		})
	}
}

func TestCoordinationBindingFailureRetainsFatalKeeper(t *testing.T) {
	d := testDomain(t)
	config := provisionForTest(t, d, t.TempDir())
	desc, _ := parseDescriptor(d, config)
	capability, acquired, valid := tryAcquire(d)
	if !acquired || !valid {
		t.Fatal("capability unavailable")
	}
	result := finishAcquisitionWithGeneration(d, desc, capability, true, func() (generation, error) {
		retained.Lock()
		k := retained.keepers[len(retained.keepers)-1]
		retained.Unlock()
		if _, err := k.issuer.Observe(); !errors.Is(err, runtimecoordination.ErrInvalidScope) {
			t.Fatal("provisional source minted", err)
		}
		if err := k.issuer.CommitGeneration("already-bound-fixture"); err != nil {
			t.Fatal(err)
		}
		return newGeneration()
	})
	if _, ok := result.Authority(); ok || !result.IsFatalFenced() {
		t.Fatal("binding failure exposed authority")
	}
	retained.Lock()
	k := retained.keepers[len(retained.keepers)-1]
	retained.Unlock()
	t.Cleanup(func() {
		if k.db != nil {
			_ = k.db.Close()
		}
	})
	if !k.fenced.Load() {
		t.Fatal("binding failure not fenced")
	}
	if _, err := k.issuer.Observe(); !errors.Is(err, runtimecoordination.ErrClosed) {
		t.Fatal(err)
	}
	if _, err := os.Stat(ledgerPath(config.RootDir, d)); err != nil {
		t.Fatal("durable store lost", err)
	}
}

type inspectionPauseContext struct {
	context.Context
	once            sync.Once
	entered, resume chan struct{}
}

func (c *inspectionPauseContext) Err() error {
	c.once.Do(func() { close(c.entered); <-c.resume })
	return c.Context.Err()
}

func TestActualCoordinationInspectionPauseAllowsFatalPublication(t *testing.T) {
	a, s, p, c := associatedFixture(t)
	ctx := &inspectionPauseContext{Context: context.Background(), entered: make(chan struct{}), resume: make(chan struct{})}
	done := make(chan error, 1)
	go func() { _, err := a.InspectCoordination(ctx, a.keeper.domain); done <- err }()
	<-ctx.entered // Inside actual ReadGeneration after actual provisioning validation.
	fatal := make(chan struct{})
	go func() { a.keeper.publishFatal(); close(fatal) }()
	select {
	case <-fatal:
	case <-time.After(time.Second):
		close(ctx.resume)
		t.Fatal("external read retained a short gate")
	}
	close(ctx.resume)
	if err := <-done; !errors.Is(err, runtimecoordination.ErrClosed) {
		t.Fatal(err)
	}
	assertSourceFatal(t, a, s, p, c)
}
