// Package runtimecontainment establishes one process-scoped execution
// generation over a trusted, pre-provisioned containment store.
package runtimecontainment

import (
	"crypto/rand"
	"encoding/hex"
	"strings"
	"sync"
	"sync/atomic"

	bolt "go.etcd.io/bbolt"
)

const identitySize = 32

// Domain is an opaque stable containment namespace. Its zero value is invalid.
type Domain struct{ value [identitySize]byte }

// ParseDomain parses the canonical 64-character hexadecimal representation of
// a non-zero 32-byte containment domain.
func ParseDomain(value string) (Domain, error) {
	var domain Domain
	if len(value) != hex.EncodedLen(len(domain.value)) || value != strings.ToLower(value) {
		return domain, errInvalidDescriptor
	}
	decoded, err := hex.DecodeString(value)
	if err != nil {
		return domain, errInvalidDescriptor
	}
	copy(domain.value[:], decoded)
	if zeroIdentity(domain.value) {
		return Domain{}, errInvalidDescriptor
	}
	return domain, nil
}

func (d Domain) text() string { return hex.EncodeToString(d.value[:]) }

// LocalConfig is the independently supplied trusted provisioning descriptor.
// RootDir locates candidate storage; the remaining fields are expected facts
// supplied by the deployment boundary, not by that storage.
type LocalConfig struct {
	ExpectedDomain               string
	RootDir                      string
	ExpectedCanonicalRoot        string
	ExpectedPhysicalRootIdentity string
	StorageAuthorityID           string
}

type resultKind uint8

const (
	resultUnavailable resultKind = iota + 1
	resultFatalFenced
	resultActive
)

// Result is the closed outcome of containment acquisition.
type Result struct {
	kind      resultKind
	authority *ActiveAuthority
}

// Authority returns the authority only for a successfully committed result.
func (r Result) Authority() (*ActiveAuthority, bool) {
	return r.authority, r.kind == resultActive && r.authority != nil
}

// IsUnavailable reports that no capability was acquired and no authority was
// created.
func (r Result) IsUnavailable() bool { return r.kind == resultUnavailable }

// IsFatalFenced reports that the process retained the acquired guard but could
// not safely expose authority.
func (r Result) IsFatalFenced() bool { return r.kind == resultFatalFenced }

type generation struct{ value [identitySize]byte }

type keeper struct {
	capability heldCapability
	db         *bolt.DB
	domain     Domain
	descriptor descriptor
	fenced     atomic.Bool
}

// ActiveAuthority exposes only the committed generation and guarantee facts.
// The capability and database handles remain private and strongly reachable.
type ActiveAuthority struct {
	generation generation
	keeper     *keeper
}

// CurrentGeneration returns the canonical opaque committed generation.
func (a *ActiveAuthority) CurrentGeneration() string {
	if a == nil {
		return ""
	}
	return hex.EncodeToString(a.generation.value[:])
}

// GuaranteeLevel returns the only guarantee this package can establish.
func (a *ActiveAuthority) GuaranteeLevel() string { return "ProcessContainment" }

// IsAuthoritative reports whether the private process capability remains valid.
func (a *ActiveAuthority) IsAuthoritative() bool {
	if a == nil || a.keeper == nil || a.keeper.fenced.Load() {
		return false
	}
	if !validateHeld(a.keeper.capability) || a.keeper.db == nil ||
		revalidateExistingProvisioned(a.keeper.db, a.keeper.domain, a.keeper.descriptor) != nil {
		a.keeper.fenced.Store(true)
		return false
	}
	return !a.keeper.fenced.Load()
}

var retained struct {
	sync.Mutex
	keepers []*keeper
}

// AcquireProcessContainment performs the safety-atomic existing-only bootstrap.
// It never provisions, repairs, releases, transfers, or reacquires authority.
func AcquireProcessContainment(domain Domain, config LocalConfig) Result {
	descriptor, ok := parseDescriptor(domain, config)
	if !ok {
		return Result{kind: resultUnavailable}
	}
	capability, acquired, valid := tryAcquire(domain)
	if !acquired {
		return Result{kind: resultUnavailable}
	}
	k := &keeper{capability: capability, domain: domain, descriptor: descriptor}
	retained.Lock()
	retained.keepers = append(retained.keepers, k)
	retained.Unlock()
	if !valid {
		k.fenced.Store(true)
		return Result{kind: resultFatalFenced}
	}

	db, err := openExistingProvisioned(domain, descriptor)
	k.db = db
	if err != nil || !validateHeld(capability) {
		k.fenced.Store(true)
		return Result{kind: resultFatalFenced}
	}
	candidate, err := newGeneration()
	if err != nil || bootstrapSuccessor(db, domain, candidate) != nil ||
		revalidateExistingProvisioned(db, domain, descriptor) != nil || !validateHeld(capability) {
		k.fenced.Store(true)
		return Result{kind: resultFatalFenced}
	}
	authority := &ActiveAuthority{generation: candidate, keeper: k}
	return Result{kind: resultActive, authority: authority}
}

func newGeneration() (generation, error) {
	var candidate generation
	for zeroIdentity(candidate.value) {
		if _, err := rand.Read(candidate.value[:]); err != nil {
			return generation{}, err
		}
	}
	return candidate, nil
}

func zeroIdentity(value [identitySize]byte) bool {
	var zero [identitySize]byte
	return value == zero
}
