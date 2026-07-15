package runtime

import (
	"context"
	"testing"
)

func TestHandshakeCapabilitiesAreStableLiveReadOnlyViews(t *testing.T) {
	runtimeListener := newControlledListener(nil, false)
	host := newTestHost(t, fixedComposer(runtimeListener))
	capabilities := host.capabilities
	if capabilities == nil {
		t.Fatal("Handshake capabilities are nil")
	}
	if capabilities.CanAccept() || capabilities.RuntimeContext() != nil {
		t.Fatal("capabilities are active before startup commit")
	}
	if err := host.Build(); err != nil {
		t.Fatalf("Build() error = %v", err)
	}
	if err := host.Start(context.Background()); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	if host.capabilities != capabilities {
		t.Fatal("startup replaced stable Handshake capabilities")
	}
	if !capabilities.CanAccept() || capabilities.RuntimeContext() != host.RuntimeContext() {
		t.Fatal("capabilities did not observe successful startup commit")
	}
	runtimeContext := capabilities.RuntimeContext()
	if err := host.Stop(context.Background()); err != nil {
		t.Fatalf("Stop() error = %v", err)
	}
	if capabilities.CanAccept() {
		t.Fatal("capabilities still permit admission after Stop")
	}
	assertContextCanceled(t, runtimeContext, "capability Runtime context after Stop")
}
