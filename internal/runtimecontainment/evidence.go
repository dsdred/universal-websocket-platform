package runtimecontainment

import (
	"context"
	"encoding/hex"
	"errors"
	"runtime"
	"strings"
)

// EvidenceKind is the closed result of an exact generation-fact read.
type EvidenceKind uint8

const (
	// EvidenceUnknown carries a closed reason and proves no generation fact.
	EvidenceUnknown EvidenceKind = iota + 1
	// EvidenceGenerationLive proves this authority holds the queried generation.
	EvidenceGenerationLive
	// EvidenceGenerationTerminated proves the queried recorded prior generation ended.
	EvidenceGenerationTerminated
)

// UnknownReason identifies why a generation read proved no positive fact.
type UnknownReason string

const (
	// UnknownUnavailable reports lost or unavailable authority or storage.
	UnknownUnavailable UnknownReason = "Unavailable"
	// UnknownScopeMismatch reports malformed, foreign or unrecorded identity.
	UnknownScopeMismatch UnknownReason = "ScopeMismatch"
	// UnknownContradictory reports inconsistent authoritative facts.
	UnknownContradictory UnknownReason = "Contradictory"
	// UnknownIndeterminate reports an inspection with no definite conclusion.
	UnknownIndeterminate UnknownReason = "Indeterminate"
	// UnknownCancelled reports a caller-cancelled read.
	UnknownCancelled UnknownReason = "Cancelled"
	// UnknownUnsupportedTopology reports a platform without this adapter.
	UnknownUnsupportedTopology UnknownReason = "UnsupportedTopology"
	// UnknownGuaranteeNotDeclared reports the absence of an active declared level.
	UnknownGuaranteeNotDeclared UnknownReason = "GuaranteeNotDeclared"
)

// GenerationEvidence reports one domain-scoped generation fact, not an
// attempt-bound DP-017 execution-evidence result.
type GenerationEvidence struct {
	kind   EvidenceKind
	reason UnknownReason
}

// Kind returns exactly one closed generation-fact outcome.
func (e GenerationEvidence) Kind() EvidenceKind {
	if e.kind == 0 {
		return EvidenceUnknown
	}
	return e.kind
}

// Reason returns the closed reason only when Kind is EvidenceUnknown.
func (e GenerationEvidence) Reason() UnknownReason {
	if e.Kind() != EvidenceUnknown {
		return ""
	}
	if e.reason == "" {
		return UnknownIndeterminate
	}
	return e.reason
}

func unknown(reason UnknownReason) GenerationEvidence {
	return GenerationEvidence{kind: EvidenceUnknown, reason: reason}
}

// ReadGeneration reads one exact domain and generation from this authority.
// Detected fatal faults permanently fence the authority before return.
func (a *ActiveAuthority) ReadGeneration(ctx context.Context, domain Domain, named string) GenerationEvidence {
	generation, valid := parseEvidenceGeneration(named)
	if !valid || zeroIdentity(domain.value) {
		return unknown(UnknownScopeMismatch)
	}
	if runtime.GOOS != "windows" {
		return unknown(UnknownUnsupportedTopology)
	}
	if a == nil || a.keeper == nil {
		return unknown(UnknownGuaranteeNotDeclared)
	}
	if domain != a.keeper.domain {
		return unknown(UnknownScopeMismatch)
	}
	if ctx == nil {
		return unknown(UnknownIndeterminate)
	}
	if reason := a.validateForEvidence(); reason != "" {
		return unknown(reason)
	}
	if ctx.Err() != nil {
		return unknown(UnknownCancelled)
	}
	tail, member, err := inspectGenerationMember(a.keeper.db, domain, generation)
	if err != nil {
		return unknown(a.fenceEvidenceError(err))
	}
	if tail != a.generation {
		a.keeper.fenced.Store(true)
		return unknown(UnknownContradictory)
	}
	if reason := a.validateForEvidence(); reason != "" {
		return unknown(reason)
	}
	if ctx.Err() != nil {
		return unknown(UnknownCancelled)
	}
	if !member {
		return unknown(UnknownScopeMismatch)
	}
	if generation == a.generation {
		return GenerationEvidence{kind: EvidenceGenerationLive}
	}
	return GenerationEvidence{kind: EvidenceGenerationTerminated}
}

func (a *ActiveAuthority) validateForEvidence() UnknownReason {
	keeper := a.keeper
	if keeper.fenced.Load() {
		return UnknownUnavailable
	}
	if !validateHeld(keeper.capability) || keeper.db == nil {
		keeper.fenced.Store(true)
		return UnknownUnavailable
	}
	if err := revalidateExistingProvisioned(keeper.db, keeper.domain, keeper.descriptor); err != nil {
		return a.fenceEvidenceError(err)
	}
	if keeper.fenced.Load() {
		return UnknownUnavailable
	}
	return ""
}

func (a *ActiveAuthority) fenceEvidenceError(err error) UnknownReason {
	a.keeper.fenced.Store(true)
	if errors.Is(err, errLedgerInvalid) {
		return UnknownContradictory
	}
	return UnknownUnavailable
}

func parseEvidenceGeneration(value string) (generation, bool) {
	var parsed generation
	if len(value) != hex.EncodedLen(identitySize) || value != strings.ToLower(value) {
		return parsed, false
	}
	decoded, err := hex.DecodeString(value)
	if err != nil {
		return parsed, false
	}
	copy(parsed.value[:], decoded)
	return parsed, !zeroIdentity(parsed.value)
}
