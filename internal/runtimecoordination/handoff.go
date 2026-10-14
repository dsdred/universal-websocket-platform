package runtimecoordination

import "errors"

// ErrAlreadyCommitted rejects every second source generation binding.
var ErrAlreadyCommitted = errors.New("source generation already committed")

type issuerMarker struct{ value byte }

// SourceIssuer is the opaque non-rebinding facet retained by the source keeper.
// Copies share the same generation binding and fatal disposition.
type SourceIssuer struct {
	authority *authorityState
	marker    *issuerMarker
}

// CurrentObservation is a detached exact source/root/domain/generation binding.
// It grants no continuing authority; every later local use needs revalidation.
type CurrentObservation struct {
	authority  *authorityState
	marker     *issuerMarker
	domain     string
	generation string
}

// CommitGeneration binds one immutable nonempty generation under source ordering.
// Invalid initial input does not consume the binding; all later attempts reject.
func (s *SourceIssuer) CommitGeneration(generation string) error {
	if s == nil || s.authority == nil || s.marker == nil {
		return ErrInvalidScope
	}
	a := s.authority
	a.gate.Lock()
	defer a.gate.Unlock()
	if a.domain == "" || a.issuer != s.marker {
		return ErrInvalidScope
	}
	if a.terminal {
		return ErrClosed
	}
	if a.generation != "" {
		return ErrAlreadyCommitted
	}
	if generation == "" {
		return ErrInvalidScope
	}
	a.generation = generation
	return nil
}

// Observe copies only committed local source affinity under the source read gate.
// It performs no external inspection or authorization.
func (s *SourceIssuer) Observe() (CurrentObservation, error) {
	if s == nil || s.authority == nil || s.marker == nil {
		return CurrentObservation{}, ErrInvalidScope
	}
	a := s.authority
	a.gate.RLock()
	defer a.gate.RUnlock()
	if a.domain == "" || a.issuer != s.marker {
		return CurrentObservation{}, ErrInvalidScope
	}
	if a.terminal {
		return CurrentObservation{}, ErrClosed
	}
	if a.generation == "" {
		return CurrentObservation{}, ErrInvalidScope
	}
	return CurrentObservation{a, s.marker, a.domain, a.generation}, nil
}

// RevalidateCurrentObservation checks exact detached affinity and unfenced state.
// Successful comparison is local consistency, never continuing external authority.
func (root *AuthorityFence) RevalidateCurrentObservation(o CurrentObservation) error {
	if root == nil || root.data == nil || o.authority == nil || o.marker == nil || o.domain == "" || o.generation == "" {
		return ErrInvalidScope
	}
	a := root.data
	a.gate.RLock()
	defer a.gate.RUnlock()
	if a.domain == "" || a.issuer == nil || a.generation == "" {
		return ErrInvalidScope
	}
	if o.authority != a || o.marker != a.issuer || o.domain != a.domain || o.generation != a.generation {
		return ErrForeignScope
	}
	if a.terminal {
		return ErrClosed
	}
	return nil
}
