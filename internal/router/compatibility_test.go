package router

import (
	"context"
	"errors"
	"reflect"
	"sync"
	"testing"

	"github.com/dsdred/universal-websocket-platform/internal/configurationversion"
	"github.com/dsdred/universal-websocket-platform/internal/message"
)

func TestNewCompatibilityCreatesImmutableDefaultOnlyRouter(t *testing.T) {
	defaultHandler := &countingHandler{}
	compiled := NewCompatibility(defaultHandler)

	if compiled == nil {
		t.Fatal("NewCompatibility() returned nil")
	}
	if len(compiled.routes) != 0 {
		t.Fatalf("compiled Routes = %d, want 0", len(compiled.routes))
	}
	if compiled.defaultHandler == nil ||
		compiled.defaultHandler.reference != legacyHandlerRef ||
		compiled.defaultHandler.handler != defaultHandler {
		t.Fatalf("compiled Default Handler = %#v", compiled.defaultHandler)
	}
	if err := validateCompiled(compiled); err != nil {
		t.Fatalf("validateCompiled() error = %v", err)
	}
}

func TestNewCompatibilityNilHandlerProducesNoMatch(t *testing.T) {
	for _, test := range []struct {
		name    string
		handler message.Handler
	}{
		{name: "nil"},
		{name: "typed nil", handler: (*countingHandler)(nil)},
	} {
		t.Run(test.name, func(t *testing.T) {
			compiled := NewCompatibility(test.handler)
			if compiled == nil || len(compiled.routes) != 0 || compiled.defaultHandler != nil {
				t.Fatalf("NewCompatibility() = %#v, want empty Router", compiled)
			}
			if err := compiled.Handle(context.Background(), authenticatedContext(t, message.TypeText)); err != nil {
				t.Fatalf("Handle() error = %v, want nil No Match", err)
			}
			if err := validateCompiled(compiled); err != nil {
				t.Fatalf("validateCompiled() error = %v", err)
			}
		})
	}
}

func TestCompatibilityRouterInvokesDefaultExactlyOnce(t *testing.T) {
	defaultHandler := &countingHandler{}
	compiled := NewCompatibility(defaultHandler)

	if err := compiled.Handle(context.Background(), authenticatedContext(t, message.TypeBinary)); err != nil {
		t.Fatalf("Handle() error = %v", err)
	}
	if defaultHandler.calls.Load() != 1 {
		t.Fatalf("Default Handler calls = %d, want 1", defaultHandler.calls.Load())
	}
}

func TestCompatibilityRouterReturnsDefaultHandlerErrorUnchanged(t *testing.T) {
	wantErr := errors.New("compatibility Handler failure")
	defaultHandler := &countingHandler{err: wantErr}
	compiled := NewCompatibility(defaultHandler)

	if err := compiled.Handle(context.Background(), authenticatedContext(t, message.TypeText)); err != wantErr {
		t.Fatalf("Handle() error = %v, want unchanged Handler error", err)
	}
	if defaultHandler.calls.Load() != 1 {
		t.Fatalf("Default Handler calls = %d, want 1", defaultHandler.calls.Load())
	}
}

func TestCompatibilityRouterSupportsConcurrentHandle(t *testing.T) {
	defaultHandler := &countingHandler{}
	compiled := NewCompatibility(defaultHandler)
	runtimeContext := authenticatedContext(t, message.TypeText)
	wantDefault := compiled.defaultHandler
	start := make(chan struct{})

	const callers = 64
	results := make(chan error, callers)
	var waitGroup sync.WaitGroup
	waitGroup.Add(callers)
	for range callers {
		go func() {
			defer waitGroup.Done()
			<-start
			results <- compiled.Handle(context.Background(), runtimeContext)
		}()
	}
	close(start)
	waitGroup.Wait()
	close(results)

	for err := range results {
		if err != nil {
			t.Fatalf("Handle() error = %v", err)
		}
	}
	if defaultHandler.calls.Load() != callers {
		t.Fatalf("Default Handler calls = %d, want %d", defaultHandler.calls.Load(), callers)
	}
	if len(compiled.routes) != 0 || compiled.defaultHandler != wantDefault {
		t.Fatal("concurrent Handle mutated compatibility Router")
	}
}

func TestCompatibilityAndCompiledRouterDispatchDefaultIdentically(t *testing.T) {
	compatibilityErr := errors.New("default result")
	compatibilityHandler := &countingHandler{err: compatibilityErr}
	compiledHandler := &countingHandler{err: compatibilityErr}
	compatibility := NewCompatibility(compatibilityHandler)
	snapshot := buildRoutingSnapshot(t, &configurationversion.RoutingSettings{DefaultHandlerRef: legacyHandlerRef})
	configured, err := New(snapshot, map[string]message.Handler{legacyHandlerRef: compiledHandler})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	runtimeContext := authenticatedContext(t, message.TypeText)

	compatibilityResult := compatibility.Handle(context.Background(), runtimeContext)
	configuredResult := configured.Handle(context.Background(), runtimeContext)
	if compatibilityResult != compatibilityErr || configuredResult != compatibilityErr {
		t.Fatalf("Handle() errors = (%v, %v), want unchanged shared result", compatibilityResult, configuredResult)
	}
	if compatibilityHandler.calls.Load() != 1 || compiledHandler.calls.Load() != 1 {
		t.Fatalf("Default Handler calls = (%d, %d), want (1, 1)", compatibilityHandler.calls.Load(), compiledHandler.calls.Load())
	}
}

func TestCompatibilityRouterRemainsImmutableAcrossRepeatedDispatch(t *testing.T) {
	defaultHandler := &countingHandler{}
	compiled := NewCompatibility(defaultHandler)
	want := *compiled.defaultHandler
	runtimeContext := authenticatedContext(t, message.TypeText)

	const calls = 100
	for range calls {
		if err := compiled.Handle(context.Background(), runtimeContext); err != nil {
			t.Fatalf("Handle() error = %v", err)
		}
	}
	if len(compiled.routes) != 0 || compiled.defaultHandler == nil || !reflect.DeepEqual(*compiled.defaultHandler, want) {
		t.Fatal("repeated Handle mutated compatibility Router")
	}
	if defaultHandler.calls.Load() != calls {
		t.Fatalf("Default Handler calls = %d, want %d", defaultHandler.calls.Load(), calls)
	}
}

func TestCompatibilityRouterHandleAllocations(t *testing.T) {
	compiled := NewCompatibility(&handlerStub{})
	runtimeContext := authenticatedContext(t, message.TypeText)
	callContext := context.Background()
	var handleErr error

	allocations := testing.AllocsPerRun(1000, func() {
		handleErr = compiled.Handle(callContext, runtimeContext)
	})
	if handleErr != nil {
		t.Fatalf("Handle() error = %v", handleErr)
	}
	if allocations != 0 {
		t.Fatalf("Handle() allocations = %v, want 0", allocations)
	}
}
