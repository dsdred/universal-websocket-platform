package secretresolver

import (
	"context"
	"sync"
)

// MemoryResolver is a thread-safe in-memory Resolver for tests and local development.
type MemoryResolver struct {
	mu      sync.RWMutex
	secrets map[string][]byte
}

// NewMemory creates an in-memory Resolver containing independent copies of initial values.
func NewMemory(initial map[string][]byte) (*MemoryResolver, error) {
	secrets := make(map[string][]byte, len(initial))
	for ref, value := range initial {
		normalized, err := normalizeReference(ref)
		if err != nil {
			return nil, err
		}
		if len(value) == 0 {
			return nil, ErrEmptySecret
		}
		secrets[normalized] = cloneBytes(value)
	}

	return &MemoryResolver{secrets: secrets}, nil
}

// Resolve returns an independent copy of the Secret identified by ref.
func (resolver *MemoryResolver) Resolve(ctx context.Context, ref string) (Secret, error) {
	if err := ctx.Err(); err != nil {
		return Secret{}, err
	}

	normalized, err := normalizeReference(ref)
	if err != nil {
		return Secret{}, err
	}

	resolver.mu.RLock()
	value, exists := resolver.secrets[normalized]
	if exists {
		value = cloneBytes(value)
	}
	resolver.mu.RUnlock()

	if err := ctx.Err(); err != nil {
		return Secret{}, err
	}
	if !exists {
		return Secret{}, ErrSecretNotFound
	}

	return Secret{Value: value}, nil
}

// Set stores an independent copy of value under a normalized Secret Reference.
func (resolver *MemoryResolver) Set(ref string, value []byte) error {
	normalized, err := normalizeReference(ref)
	if err != nil {
		return err
	}
	if len(value) == 0 {
		return ErrEmptySecret
	}

	value = cloneBytes(value)
	resolver.mu.Lock()
	resolver.secrets[normalized] = value
	resolver.mu.Unlock()
	return nil
}

// Delete removes the Secret stored under a normalized Secret Reference.
func (resolver *MemoryResolver) Delete(ref string) error {
	normalized, err := normalizeReference(ref)
	if err != nil {
		return err
	}

	resolver.mu.Lock()
	defer resolver.mu.Unlock()
	if _, exists := resolver.secrets[normalized]; !exists {
		return ErrSecretNotFound
	}
	delete(resolver.secrets, normalized)
	return nil
}

func cloneBytes(value []byte) []byte {
	result := make([]byte, len(value))
	copy(result, value)
	return result
}
