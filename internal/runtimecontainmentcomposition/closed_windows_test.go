//go:build windows

package runtimecontainmentcomposition

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/dsdred/universal-websocket-platform/internal/runtimecommandidempotency"
	"github.com/dsdred/universal-websocket-platform/internal/runtimecontainment"
	"github.com/dsdred/universal-websocket-platform/internal/runtimecoordination"
	"github.com/dsdred/universal-websocket-platform/internal/runtimemanagement"
)

const closedProofHelper = "UWP_TASK087_CLOSED_PROOF_HELPER"

func TestClosedAssemblyLocalProofs(t *testing.T) {
	if root := os.Getenv(closedProofHelper); root != "" {
		exerciseClosedAssembly(t, root)
		return
	}
	root, err := os.MkdirTemp("", "uwp-closed-assembly-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(root)
	cmd := exec.Command(os.Args[0], "-test.run=^TestClosedAssemblyLocalProofs$", "-test.v")
	cmd.Env = append(os.Environ(), closedProofHelper+"="+root)
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("closed helper %v\n%s", err, output)
	}
}

func exerciseClosedAssembly(t *testing.T, root string) {
	f := newConcreteFixture(t, filepath.Join(root, "fixture"), nil)
	assembly, err := NewClosedAssembly(f.domainText, f.domain, f.containment, f.target)
	if err != nil || assembly == nil {
		t.Fatal(assembly, err)
	}
	scope := assembly.data.scope

	t.Run("complete detached exact-scope surface", func(t *testing.T) {
		copy := *assembly
		observed, err := copy.Capture(scope)
		if err != nil {
			t.Fatal(err)
		}
		if observed.Identities().Scope() != scope || observed.Commands().Scope() != scope || observed.Identities().Present() || observed.Commands().Present() {
			t.Fatal("incomplete owner supply")
		}
		if observed.Identities().Allocated() || observed.Identities().PresenceRevision() != 0 || observed.Commands().PresenceRevision() != 0 {
			t.Fatal("construction allocated facts")
		}
		c, p := observed.Commands().Epochs()
		if c == 0 || c != p {
			t.Fatal("shared participant epoch missing")
		}
		foreign, _ := runtimecoordination.NewScope(scope.Domain(), scope.WorkspaceID(), scope.ConfigurationID(), "foreign")
		if o, err := assembly.Capture(foreign); !errors.Is(err, runtimecoordination.ErrForeignScope) || o.Identities().Present() || o.Commands().Present() {
			t.Fatal("foreign scope disclosed facts", o, err)
		}
		for _, a := range []*ClosedAssembly{nil, {}} {
			if _, err := a.Capture(scope); !errors.Is(err, runtimecoordination.ErrInvalidScope) {
				t.Fatal(err)
			}
		}
		if _, err := assembly.Capture(runtimecoordination.Scope{}); !errors.Is(err, runtimecoordination.ErrInvalidScope) {
			t.Fatal(err)
		}
		typ := reflect.TypeOf(assembly)
		if typ.NumMethod() != 1 || typ.Method(0).Name != "Capture" {
			t.Fatal("assembly exposes non-read operation", typ)
		}
		for _, typ := range []reflect.Type{reflect.TypeOf(*assembly), reflect.TypeOf(ClosedObservation{})} {
			for i := 0; i < typ.NumField(); i++ {
				if typ.Field(i).IsExported() {
					t.Fatal("opaque field exported", typ.Field(i))
				}
			}
		}
		// Prove exact topology in the real owners without obtaining writable
		// owner/participant handles from the assembly's public surface.
		aggregate := reflect.ValueOf(assembly.data.identities).Elem().FieldByName("coordination").Elem().FieldByName("attachment").FieldByName("participant").Pointer()
		commands := reflect.ValueOf(assembly.data.commands).Elem().FieldByName("coord").Elem().FieldByName("attachments")
		values := commands.MapRange()
		if !values.Next() {
			t.Fatal("command participant absent")
		}
		command := values.Value().FieldByName("participant").Pointer()
		if aggregate == 0 || aggregate != command || values.Next() {
			t.Fatal("owners did not share one exact participant")
		}
	})

	t.Run("closed real writers and absent replay zero callbacks", func(t *testing.T) {
		before, err := assembly.Capture(scope)
		if err != nil {
			t.Fatal(err)
		}
		if err := assembly.data.identities.AllocateCandidateIdentity(f.target.RuntimeInstanceID()); !errors.Is(err, runtimecoordination.ErrClosed) {
			t.Fatal(err)
		}
		if err := assembly.data.identities.CreateRuntimeInstance(scope.WorkspaceID(), scope.ConfigurationID(), f.target.RuntimeInstanceID()); !errors.Is(err, runtimecoordination.ErrClosed) {
			t.Fatal(err)
		}
		if result, err := assembly.data.identities.ConditionalClaimLaunchAttempt(f.target.RuntimeInstanceID(), 0, "attempt", 1); !errors.Is(err, runtimecoordination.ErrClosed) || result.Committed() {
			t.Fatal(result, err)
		}
		commandScope, err := runtimecommandidempotency.NewScope(f.domainText, scope.WorkspaceID(), scope.ConfigurationID(), f.target.RuntimeInstanceID(), runtimecommandidempotency.OperationStart)
		if err != nil {
			t.Fatal(err)
		}
		intent, err := runtimecommandidempotency.NewStartIntent(1)
		if err != nil {
			t.Fatal(err)
		}
		var callbacks atomic.Int32
		_, err = assembly.data.boundary.Execute(context.Background(), commandScope, "closed", intent, func(context.Context, runtimecommandidempotency.Scope, runtimecommandidempotency.Intent) error {
			return nil
		}, func() (runtimecommandidempotency.TerminalOutcome, error) {
			callbacks.Add(1)
			return runtimecommandidempotency.TerminalOutcome{}, nil
		})
		if !errors.Is(err, runtimecoordination.ErrClosed) || callbacks.Load() != 0 {
			t.Fatal("closed command delegated", err, callbacks.Load())
		}
		if err := assembly.data.identities.RevalidateCoordinationSnapshot(before.Identities()); err != nil {
			t.Fatal(err)
		}
		if err := assembly.data.commands.RevalidateCoordinationSnapshot(before.Commands()); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("foreign sources invalid target and fatal partial construction", func(t *testing.T) {
		otherText, _ := randomCompositionIdentity(t)
		otherDomain, _ := runtimecontainment.ParseDomain(otherText)
		for _, row := range []struct {
			text   string
			domain runtimecontainment.Domain
			result runtimecontainment.Result
			target runtimemanagement.Target
		}{
			{otherText, f.domain, f.containment, f.target}, {otherText, otherDomain, f.containment, f.target}, {f.domainText, f.domain, runtimecontainment.Result{}, f.target}, {f.domainText, f.domain, f.containment, runtimemanagement.Target{}},
		} {
			if a, err := NewClosedAssembly(row.text, row.domain, row.result, row.target); a != nil || err == nil {
				t.Fatal("partial/foreign assembly exposed", a, err)
			}
		}
	})

	t.Run("construction races fatal and both owners terminal", func(t *testing.T) {
		root, err := f.authority.CoordinationRoot(f.domain)
		if err != nil {
			t.Fatal(err)
		}
		var wg sync.WaitGroup
		for i := 0; i < 12; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				a, err := NewClosedAssembly(f.domainText, f.domain, f.containment, f.target)
				if err != nil && a != nil {
					t.Error("partial assembly returned")
				}
				if a != nil && (a.data == nil || a.data.identities == nil || a.data.commands == nil || a.data.boundary == nil) {
					t.Error("incomplete assembly")
				}
			}()
		}
		publication := root.BeginFatalPublication()
		publication.Invalidate()
		publication.Close()
		wg.Wait()
		if _, err := assembly.data.identities.CaptureCoordinationSnapshot(scope); !errors.Is(err, runtimecoordination.ErrClosed) {
			t.Fatal("aggregate owner escaped source", err)
		}
		if _, err := assembly.data.commands.CaptureCoordinationSnapshot(scope); err == nil {
			t.Fatal("command owner escaped source")
		}
		if o, err := assembly.Capture(scope); err == nil || o.Identities().Present() || o.Commands().Present() {
			t.Fatal("partial observations escaped", o, err)
		}
		if a, err := NewClosedAssembly(f.domainText, f.domain, f.containment, f.target); a != nil || err == nil {
			t.Fatal("fatal construction accepted", a, err)
		}
	})
}
