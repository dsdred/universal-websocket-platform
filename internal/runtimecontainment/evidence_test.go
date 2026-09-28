package runtimecontainment

import (
	"bytes"
	"context"
	"encoding/hex"
	"runtime"
	"strings"
	"sync"
	"testing"

	bolt "go.etcd.io/bbolt"
)

func TestGenerationEvidenceCurrentPriorAndNoLedgerMutation(t *testing.T) {
	requireWindows(t)
	domain := testDomain(t)
	config := provisionForTest(t, domain, t.TempDir())
	prior, err := newGeneration()
	if err != nil {
		t.Fatal(err)
	}
	descriptor, _ := parseDescriptor(domain, config)
	db, err := openExistingProvisioned(domain, descriptor)
	if err != nil {
		t.Fatal(err)
	}
	if err = appendExpected(db, domain, generation{}, true, prior); err != nil {
		t.Fatal(err)
	}
	if err = db.Close(); err != nil {
		t.Fatal(err)
	}
	authority := acquiredAuthority(t, domain, config)
	before := ledgerRevision(t, authority.keeper.db)
	for _, row := range []struct {
		name string
		want EvidenceKind
	}{
		{authority.CurrentGeneration(), EvidenceGenerationLive},
		{hexGeneration(prior), EvidenceGenerationTerminated},
	} {
		got := authority.ReadGeneration(context.Background(), domain, row.name)
		if got.Kind() != row.want || got.Reason() != "" {
			t.Fatalf("read %q: kind=%v reason=%q", row.name, got.Kind(), got.Reason())
		}
	}
	if after := ledgerRevision(t, authority.keeper.db); after != before {
		t.Fatalf("read changed ledger transaction: %d -> %d", before, after)
	}
	if !authority.IsAuthoritative() {
		t.Fatal("read fenced valid authority")
	}
}

func TestGenerationEvidenceFailsClosedWithoutFencingCallerErrors(t *testing.T) {
	requireWindows(t)
	domain := testDomain(t)
	authority := acquiredAuthority(t, domain, provisionForTest(t, domain, t.TempDir()))
	other := testDomain(t)
	missing, _ := newGeneration()
	cancelled, cancel := context.WithCancel(context.Background())
	cancel()
	for _, row := range []struct {
		name       string
		ctx        context.Context
		domain     Domain
		generation string
		want       UnknownReason
	}{
		{"missing", context.Background(), domain, hexGeneration(missing), UnknownScopeMismatch},
		{"malformed", context.Background(), domain, "BAD", UnknownScopeMismatch},
		{"cross-domain", context.Background(), other, authority.CurrentGeneration(), UnknownScopeMismatch},
		{"cancelled", cancelled, domain, authority.CurrentGeneration(), UnknownCancelled},
		{"nil-context", nil, domain, authority.CurrentGeneration(), UnknownIndeterminate},
	} {
		got := authority.ReadGeneration(row.ctx, row.domain, row.generation)
		if got.Kind() != EvidenceUnknown || got.Reason() != row.want {
			t.Fatalf("%s: kind=%v reason=%q, want %q", row.name, got.Kind(), got.Reason(), row.want)
		}
		if !authority.IsAuthoritative() {
			t.Fatalf("%s fenced authority", row.name)
		}
	}
}

func TestGenerationEvidenceCorruptionFencesBeforeCancelledReturn(t *testing.T) {
	requireWindows(t)
	domain := testDomain(t)
	authority := acquiredAuthority(t, domain, provisionForTest(t, domain, t.TempDir()))
	err := authority.keeper.db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket(stateBucket).Put(tailKey, bytes.Repeat([]byte{0xff}, identitySize))
	})
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	got := authority.ReadGeneration(ctx, domain, authority.CurrentGeneration())
	if got.Kind() != EvidenceUnknown || got.Reason() != UnknownContradictory {
		t.Fatalf("corruption/cancellation: kind=%v reason=%q", got.Kind(), got.Reason())
	}
	if !authority.keeper.fenced.Load() || authority.IsAuthoritative() {
		t.Fatal("corruption failed to fence authority")
	}
	got = authority.ReadGeneration(context.Background(), domain, authority.CurrentGeneration())
	if got.Kind() != EvidenceUnknown || got.Reason() != UnknownUnavailable {
		t.Fatalf("fenced read: kind=%v reason=%q", got.Kind(), got.Reason())
	}
}

func TestGenerationEvidenceConcurrentReaders(t *testing.T) {
	requireWindows(t)
	domain := testDomain(t)
	authority := acquiredAuthority(t, domain, provisionForTest(t, domain, t.TempDir()))
	var wait sync.WaitGroup
	errors := make(chan GenerationEvidence, 32)
	for i := 0; i < cap(errors); i++ {
		wait.Add(1)
		go func() {
			defer wait.Done()
			for j := 0; j < 12; j++ {
				got := authority.ReadGeneration(context.Background(), domain, authority.CurrentGeneration())
				if got.Kind() != EvidenceGenerationLive {
					errors <- got
					return
				}
			}
		}()
	}
	wait.Wait()
	close(errors)
	for got := range errors {
		t.Fatalf("concurrent read: kind=%v reason=%q", got.Kind(), got.Reason())
	}
}

func TestGenerationEvidenceUnavailablePlatformAndNoAuthority(t *testing.T) {
	domain := testDomain(t)
	var authority *ActiveAuthority
	got := authority.ReadGeneration(context.Background(), domain, strings.Repeat("01", identitySize))
	want := UnknownGuaranteeNotDeclared
	if runtime.GOOS != "windows" {
		want = UnknownUnsupportedTopology
	}
	if got.Kind() != EvidenceUnknown || got.Reason() != want {
		t.Fatalf("absent authority: kind=%v reason=%q, want %q", got.Kind(), got.Reason(), want)
	}
	if runtime.GOOS != "windows" {
		result := AcquireProcessContainment(domain, LocalConfig{})
		if !result.IsUnavailable() {
			t.Fatal("unsupported platform issued authority")
		}
	}
}

func acquiredAuthority(t *testing.T, domain Domain, config LocalConfig) *ActiveAuthority {
	t.Helper()
	result := AcquireProcessContainment(domain, config)
	authority, ok := result.Authority()
	if !ok {
		t.Fatalf("acquire: unavailable=%v fatal=%v", result.IsUnavailable(), result.IsFatalFenced())
	}
	t.Cleanup(func() {
		// The package retains the process capability; the test closes only its
		// private fixture database after all assertions, so TempDir can clean up.
		authority.keeper.fenced.Store(true)
		if err := authority.keeper.db.Close(); err != nil {
			t.Errorf("close fixture database: %v", err)
		}
	})
	return authority
}

func ledgerRevision(t *testing.T, db *bolt.DB) int {
	t.Helper()
	var id int
	if err := db.View(func(tx *bolt.Tx) error { id = tx.ID(); return nil }); err != nil {
		t.Fatal(err)
	}
	return id
}

func hexGeneration(value generation) string {
	return hex.EncodeToString(value.value[:])
}
