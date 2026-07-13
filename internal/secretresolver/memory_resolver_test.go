package secretresolver

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
)

func TestNewMemoryAcceptsNilInitialMap(t *testing.T) {
	resolver, err := NewMemory(nil)
	if err != nil {
		t.Fatalf("NewMemory(nil) error = %v", err)
	}
	if resolver == nil {
		t.Fatal("NewMemory(nil) returned nil Resolver")
	}

	var _ Resolver = resolver
}

func TestNewMemoryCreatesResolverAndNormalizesReferences(t *testing.T) {
	resolver, err := NewMemory(map[string][]byte{"  secrets/api-keys/internal  ": []byte("initial")})
	if err != nil {
		t.Fatalf("NewMemory() error = %v", err)
	}

	secret, err := resolver.Resolve(context.Background(), "secrets/api-keys/internal")
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	if !bytes.Equal(secret.Value, []byte("initial")) {
		t.Errorf("Resolve() returned unexpected value")
	}
}

func TestNewMemoryRejectsInvalidReference(t *testing.T) {
	resolver, err := NewMemory(map[string][]byte{"https://example.com/secret": []byte("value")})
	if resolver != nil || !errors.Is(err, ErrInvalidSecretReference) {
		t.Fatalf("NewMemory() = (%v, %v), want nil and ErrInvalidSecretReference", resolver, err)
	}
}

func TestNewMemoryRejectsEmptyValue(t *testing.T) {
	resolver, err := NewMemory(map[string][]byte{"secrets/api-keys/internal": nil})
	if resolver != nil || !errors.Is(err, ErrEmptySecret) {
		t.Fatalf("NewMemory() = (%v, %v), want nil and ErrEmptySecret", resolver, err)
	}
}

func TestNewMemoryDeepCopiesInitialValues(t *testing.T) {
	value := []byte("initial")
	resolver, err := NewMemory(map[string][]byte{"secrets/api-keys/internal": value})
	if err != nil {
		t.Fatalf("NewMemory() error = %v", err)
	}

	value[0] = 'X'
	secret, err := resolver.Resolve(context.Background(), "secrets/api-keys/internal")
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	if !bytes.Equal(secret.Value, []byte("initial")) {
		t.Errorf("stored value changed with constructor input")
	}
}

func TestResolve(t *testing.T) {
	resolver := mustMemoryResolver(t, map[string][]byte{"secrets/jwt/main": []byte("resolved")})

	secret, err := resolver.Resolve(context.Background(), "  secrets/jwt/main  ")
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	if !bytes.Equal(secret.Value, []byte("resolved")) {
		t.Errorf("Resolve() returned unexpected value")
	}
}

func TestResolveNotFound(t *testing.T) {
	resolver := mustMemoryResolver(t, nil)

	_, err := resolver.Resolve(context.Background(), "secrets/missing")
	if !errors.Is(err, ErrSecretNotFound) {
		t.Fatalf("Resolve() error = %v, want ErrSecretNotFound", err)
	}
}

func TestResolveInvalidReference(t *testing.T) {
	resolver := mustMemoryResolver(t, nil)

	_, err := resolver.Resolve(context.Background(), "/secrets/main")
	if !errors.Is(err, ErrInvalidSecretReference) {
		t.Fatalf("Resolve() error = %v, want ErrInvalidSecretReference", err)
	}
}

func TestResolveCanceledContext(t *testing.T) {
	resolver := mustMemoryResolver(t, map[string][]byte{"secrets/jwt/main": []byte("value")})
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := resolver.Resolve(ctx, "secrets/jwt/main")
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Resolve() error = %v, want context.Canceled", err)
	}
}

func TestResolveReturnsIndependentValues(t *testing.T) {
	resolver := mustMemoryResolver(t, map[string][]byte{"secrets/jwt/main": []byte("stored")})

	first, err := resolver.Resolve(context.Background(), "secrets/jwt/main")
	if err != nil {
		t.Fatalf("Resolve(first) error = %v", err)
	}
	first.Value[0] = 'X'

	second, err := resolver.Resolve(context.Background(), "secrets/jwt/main")
	if err != nil {
		t.Fatalf("Resolve(second) error = %v", err)
	}
	if !bytes.Equal(second.Value, []byte("stored")) {
		t.Errorf("stored value changed with previous Resolve result")
	}
	if len(first.Value) > 0 && &first.Value[0] == &second.Value[0] {
		t.Error("Resolve() results share the same backing array")
	}
}

func TestSetCreatesAndReplacesSecret(t *testing.T) {
	resolver := mustMemoryResolver(t, nil)
	if err := resolver.Set("  secrets/api-keys/internal  ", []byte("first")); err != nil {
		t.Fatalf("Set(create) error = %v", err)
	}
	if err := resolver.Set("secrets/api-keys/internal", []byte("second")); err != nil {
		t.Fatalf("Set(replace) error = %v", err)
	}

	secret, err := resolver.Resolve(context.Background(), "secrets/api-keys/internal")
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	if !bytes.Equal(secret.Value, []byte("second")) {
		t.Errorf("Set(replace) did not replace value")
	}
}

func TestSetRejectsInvalidReferenceAndEmptyValue(t *testing.T) {
	resolver := mustMemoryResolver(t, nil)

	if err := resolver.Set("/secrets/main", []byte("value")); !errors.Is(err, ErrInvalidSecretReference) {
		t.Errorf("Set(invalid reference) error = %v, want ErrInvalidSecretReference", err)
	}
	if err := resolver.Set("secrets/main", nil); !errors.Is(err, ErrEmptySecret) {
		t.Errorf("Set(empty value) error = %v, want ErrEmptySecret", err)
	}
}

func TestSetDeepCopiesValue(t *testing.T) {
	resolver := mustMemoryResolver(t, nil)
	value := []byte("stored")
	if err := resolver.Set("secrets/main", value); err != nil {
		t.Fatalf("Set() error = %v", err)
	}
	value[0] = 'X'

	secret, err := resolver.Resolve(context.Background(), "secrets/main")
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	if !bytes.Equal(secret.Value, []byte("stored")) {
		t.Errorf("stored value changed with Set input")
	}
}

func TestDelete(t *testing.T) {
	resolver := mustMemoryResolver(t, map[string][]byte{"secrets/main": []byte("value")})

	if err := resolver.Delete("  secrets/main  "); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	if _, err := resolver.Resolve(context.Background(), "secrets/main"); !errors.Is(err, ErrSecretNotFound) {
		t.Fatalf("Resolve() after Delete error = %v, want ErrSecretNotFound", err)
	}
}

func TestDeleteNotFoundAndInvalidReference(t *testing.T) {
	resolver := mustMemoryResolver(t, nil)

	if err := resolver.Delete("secrets/missing"); !errors.Is(err, ErrSecretNotFound) {
		t.Errorf("Delete(missing) error = %v, want ErrSecretNotFound", err)
	}
	if err := resolver.Delete("/secrets/main"); !errors.Is(err, ErrInvalidSecretReference) {
		t.Errorf("Delete(invalid reference) error = %v, want ErrInvalidSecretReference", err)
	}
}

func TestMemoryResolverConcurrentOperations(t *testing.T) {
	const operations = 64
	initial := make(map[string][]byte, operations*2)
	for index := 0; index < operations; index++ {
		initial[fmt.Sprintf("secrets/concurrent/resolve-%d", index)] = []byte("resolve")
		initial[fmt.Sprintf("secrets/concurrent/delete-%d", index)] = []byte("delete")
	}
	resolver := mustMemoryResolver(t, initial)

	errorsChannel := make(chan error, operations*3)
	var waitGroup sync.WaitGroup
	for index := 0; index < operations; index++ {
		index := index
		waitGroup.Add(3)
		go func() {
			defer waitGroup.Done()
			_, err := resolver.Resolve(context.Background(), fmt.Sprintf("secrets/concurrent/resolve-%d", index))
			errorsChannel <- err
		}()
		go func() {
			defer waitGroup.Done()
			errorsChannel <- resolver.Set(fmt.Sprintf("secrets/concurrent/set-%d", index), []byte("set"))
		}()
		go func() {
			defer waitGroup.Done()
			errorsChannel <- resolver.Delete(fmt.Sprintf("secrets/concurrent/delete-%d", index))
		}()
	}

	waitGroup.Wait()
	close(errorsChannel)
	for err := range errorsChannel {
		if err != nil {
			t.Errorf("concurrent operation error = %v", err)
		}
	}
}

func TestMemoryResolverSmokeScenario(t *testing.T) {
	const ref = "secrets/api-keys/internal"
	resolver := mustMemoryResolver(t, map[string][]byte{ref: []byte("initial-value")})

	first, err := resolver.Resolve(context.Background(), ref)
	if err != nil {
		t.Fatalf("initial Resolve() error = %v", err)
	}
	t.Logf("initial resolve: ok, length=%d", len(first.Value))
	first.Value[0] = 'X'

	second, err := resolver.Resolve(context.Background(), ref)
	if err != nil {
		t.Fatalf("second Resolve() error = %v", err)
	}
	if second.Value[0] == 'X' {
		t.Fatal("second Resolve() returned mutated value")
	}
	t.Logf("second resolve after result mutation: unchanged, length=%d", len(second.Value))

	if err := resolver.Set(ref, []byte("replacement")); err != nil {
		t.Fatalf("Set() error = %v", err)
	}
	t.Log("set replacement: ok")
	if err := resolver.Delete(ref); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	t.Log("delete: ok")
	if _, err := resolver.Resolve(context.Background(), ref); !errors.Is(err, ErrSecretNotFound) {
		t.Fatalf("Resolve() after Delete error = %v, want ErrSecretNotFound", err)
	}
	t.Log("resolve after delete: ErrSecretNotFound")
}

func mustMemoryResolver(t *testing.T, initial map[string][]byte) *MemoryResolver {
	t.Helper()
	resolver, err := NewMemory(initial)
	if err != nil {
		t.Fatalf("NewMemory() error = %v", err)
	}
	return resolver
}
