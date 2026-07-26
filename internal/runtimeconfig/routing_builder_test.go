package runtimeconfig

import "testing"

func TestRoutingPresenceAndDefaultPresenceAreObservable(t *testing.T) {
	version := validConfigurationVersion()
	version.Routing = nil
	absent, diagnostics := NewBuilder().Build(detachedResult(version))
	if len(diagnostics) != 0 {
		t.Fatalf("Build(absent) diagnostics = %#v", diagnostics)
	}
	if _, present := absent.Routing(); present {
		t.Fatal("absent Routing reported present")
	}

	version.Routing = validConfigurationVersion().Routing
	version.Routing.Routes = nil
	version.Routing.DefaultHandlerRef = ""
	present, diagnostics := NewBuilder().Build(detachedResult(version))
	if len(diagnostics) != 0 {
		t.Fatalf("Build(present empty) diagnostics = %#v", diagnostics)
	}
	routing, ok := present.Routing()
	if !ok || len(routing.Routes()) != 0 {
		t.Fatalf("present-empty Routing = %#v, %t", routing, ok)
	}
	if value, defaultPresent := routing.DefaultHandlerRef(); defaultPresent || value != "" {
		t.Fatalf("DefaultHandlerRef() = (%q, %t)", value, defaultPresent)
	}
}
