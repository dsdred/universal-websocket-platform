package runtimemanagement

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"

	"github.com/dsdred/universal-websocket-platform/internal/configurationloader"
	"github.com/dsdred/universal-websocket-platform/internal/runtimeconfigload"
	"github.com/dsdred/universal-websocket-platform/internal/runtimelaunchflow"
	"github.com/dsdred/universal-websocket-platform/internal/runtimelifecycle"
	"github.com/dsdred/universal-websocket-platform/internal/runtimeorchestrationbinding"
)

const testOperationalDomain = "operations-a"

type invokerContinuation struct {
	noClaimCalls atomic.Int32
	afterCalls   atomic.Int32
	noClaim      func(context.Context, runtimeorchestrationbinding.StartExecutionBinding, runtimeorchestrationbinding.StartNoClaimCause) error
	after        func(context.Context, runtimeorchestrationbinding.StartExecutionBinding, runtimelaunchflow.OwnerClaimView) (runtimelaunchflow.StartClaimOutcome, error)
}

func (c *invokerContinuation) StartNoClaim(
	ctx context.Context,
	binding runtimeorchestrationbinding.StartExecutionBinding,
	cause runtimeorchestrationbinding.StartNoClaimCause,
) error {
	c.noClaimCalls.Add(1)
	if c.noClaim == nil {
		return nil
	}
	return c.noClaim(ctx, binding, cause)
}

func (c *invokerContinuation) AfterOwnerClaim(
	ctx context.Context,
	binding runtimeorchestrationbinding.StartExecutionBinding,
	view runtimelaunchflow.OwnerClaimView,
) (runtimelaunchflow.StartClaimOutcome, error) {
	c.afterCalls.Add(1)
	if c.after == nil {
		return runtimelaunchflow.StartClaimContinue, nil
	}
	return c.after(ctx, binding, view)
}

func TestManagedStartInvokerConstruction(t *testing.T) {
	target := mustTarget(t, testWorkspaceID, testConfigurationID, "runtime-invoker")
	flow := &runtimelaunchflow.ManagedFlow{}
	for name, construct := range map[string]func() (*ManagedStartInvoker, error){
		"empty domain": func() (*ManagedStartInvoker, error) { return NewManagedStartInvoker("", target, flow) },
		"invalid target": func() (*ManagedStartInvoker, error) {
			return NewManagedStartInvoker(testOperationalDomain, Target{}, flow)
		},
		"nil flow": func() (*ManagedStartInvoker, error) {
			return NewManagedStartInvoker(testOperationalDomain, target, nil)
		},
	} {
		t.Run(name, func(t *testing.T) {
			if invoker, err := construct(); invoker != nil || err != ErrInvalidManagedStartInvoker {
				t.Fatalf("NewManagedStartInvoker() = %#v/%v, want nil/bare ErrInvalidManagedStartInvoker", invoker, err)
			}
		})
	}

	invoker, err := NewManagedStartInvoker(testOperationalDomain, target, flow)
	if err != nil || invoker == nil {
		t.Fatalf("NewManagedStartInvoker(valid) = %#v/%v, want non-nil/nil", invoker, err)
	}
}

func TestManagedStartInvokerRejectsInvalidInvocationBeforeFlow(t *testing.T) {
	target := mustTarget(t, testWorkspaceID, testConfigurationID, "runtime-validation")
	var attemptCalls atomic.Int32
	source := &unavailableSource{}
	continuation := &invokerContinuation{}
	flow := mustManagedFlow(t, target, &attemptCalls, source, continuation)
	invoker := mustManagedStartInvoker(t, testOperationalDomain, target, flow)
	request := runtimelifecycle.NewStartRequest(testWorkspaceID, testConfigurationID, testVersionID)
	binding := mustPrimitiveStartBinding(t, testOperationalDomain, target, testVersionID)

	foreignWorkspace := mustTarget(t, 99, testConfigurationID, target.RuntimeInstanceID())
	foreignConfiguration := mustTarget(t, testWorkspaceID, 99, target.RuntimeInstanceID())
	foreignInstance := mustTarget(t, testWorkspaceID, testConfigurationID, "runtime-foreign")
	validContext := context.Background()
	var nilInvoker *ManagedStartInvoker
	invalidInvoker := &ManagedStartInvoker{}
	cases := []struct {
		name     string
		receiver *ManagedStartInvoker
		ctx      context.Context
		request  runtimelifecycle.StartRequest
		binding  runtimeorchestrationbinding.StartExecutionBinding
	}{
		{"invalid receiver", nilInvoker, validContext, request, binding},
		{"invalid receiver construction", invalidInvoker, validContext, request, binding},
		{"nil context", invoker, nil, request, binding},
		{"zero workspace request", invoker, validContext, runtimelifecycle.NewStartRequest(0, testConfigurationID, testVersionID), binding},
		{"zero configuration request", invoker, validContext, runtimelifecycle.NewStartRequest(testWorkspaceID, 0, testVersionID), binding},
		{"zero version request", invoker, validContext, runtimelifecycle.NewStartRequest(testWorkspaceID, testConfigurationID, 0), binding},
		{"invalid binding", invoker, validContext, request, runtimeorchestrationbinding.StartExecutionBinding{}},
		{"domain mismatch", invoker, validContext, request, mustPrimitiveStartBinding(t, "operations-b", target, testVersionID)},
		{"target workspace mismatch", invoker, validContext, runtimelifecycle.NewStartRequest(99, testConfigurationID, testVersionID), mustPrimitiveStartBinding(t, testOperationalDomain, foreignWorkspace, testVersionID)},
		{"target configuration mismatch", invoker, validContext, runtimelifecycle.NewStartRequest(testWorkspaceID, 99, testVersionID), mustPrimitiveStartBinding(t, testOperationalDomain, foreignConfiguration, testVersionID)},
		{"target instance mismatch", invoker, validContext, request, mustPrimitiveStartBinding(t, testOperationalDomain, foreignInstance, testVersionID)},
		{"request workspace mismatch", invoker, validContext, runtimelifecycle.NewStartRequest(99, testConfigurationID, testVersionID), binding},
		{"request configuration mismatch", invoker, validContext, runtimelifecycle.NewStartRequest(testWorkspaceID, 99, testVersionID), binding},
		{"request version mismatch", invoker, validContext, runtimelifecycle.NewStartRequest(testWorkspaceID, testConfigurationID, 99), binding},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if outcome, err := tc.receiver.InvokeManagedStart(tc.ctx, tc.request, tc.binding); outcome != (runtimelifecycle.StartOutcome{}) || err != ErrInvalidManagedStartInvocation {
				t.Fatalf("InvokeManagedStart() = %#v/%v, want zero/bare ErrInvalidManagedStartInvocation", outcome, err)
			}
		})
	}
	if attemptCalls.Load() != 0 || source.calls.Load() != 0 || continuation.noClaimCalls.Load() != 0 || continuation.afterCalls.Load() != 0 {
		t.Fatalf("invalid invocations reached Flow: attempts=%d source=%d no-claim=%d after=%d", attemptCalls.Load(), source.calls.Load(), continuation.noClaimCalls.Load(), continuation.afterCalls.Load())
	}
}

func TestManagedStartInvokerDelegatesPrimitiveAndLinkedOnce(t *testing.T) {
	downstreamErr := errors.New("downstream continuation failure")
	for _, action := range []runtimeorchestrationbinding.OrchestrationAction{
		runtimeorchestrationbinding.OrchestrationActionActivateExactTarget,
		runtimeorchestrationbinding.OrchestrationActionReplaceExactTarget,
		runtimeorchestrationbinding.OrchestrationActionRollbackExactTarget,
	} {
		t.Run(string(action), func(t *testing.T) {
			target := mustTarget(t, testWorkspaceID, testConfigurationID, runtimeconfigload.RuntimeInstanceID("runtime-"+string(action)))
			request := runtimelifecycle.NewStartRequest(testWorkspaceID, testConfigurationID, testVersionID)
			binding := mustStartBinding(t, testOperationalDomain, target, action, testVersionID)
			var attemptCalls atomic.Int32
			source := &unavailableSource{}
			contextKey := struct{}{}
			ctx := context.WithValue(context.Background(), contextKey, "preserved")
			continuation := &invokerContinuation{after: func(gotCtx context.Context, gotBinding runtimeorchestrationbinding.StartExecutionBinding, view runtimelaunchflow.OwnerClaimView) (runtimelaunchflow.StartClaimOutcome, error) {
				if gotCtx.Value(contextKey) != "preserved" || gotBinding != binding {
					t.Fatalf("AfterOwnerClaim() did not receive preserved context value and exact binding")
				}
				if view.WorkspaceID() != request.WorkspaceID() || view.ConfigurationID() != request.ConfigurationID() ||
					view.RuntimeInstanceID() != target.RuntimeInstanceID() || view.TargetConfigurationVersionID() != request.ConfigurationVersionID() {
					t.Fatalf("OwnerClaimView = %#v, want exact request/target identities", view)
				}
				return runtimelaunchflow.StartClaimBlocked, downstreamErr
			}}
			flow := mustManagedFlow(t, target, &attemptCalls, source, continuation)
			invoker := mustManagedStartInvoker(t, testOperationalDomain, target, flow)

			outcome, err := invoker.InvokeManagedStart(ctx, request, binding)
			if outcome != (runtimelifecycle.StartOutcome{}) || err != downstreamErr {
				t.Fatalf("InvokeManagedStart() = %#v/%v, want zero/exact downstream error", outcome, err)
			}
			if attemptCalls.Load() != 1 || continuation.afterCalls.Load() != 1 ||
				continuation.noClaimCalls.Load() != 0 || source.calls.Load() != 0 {
				t.Fatalf("delegation calls: attempts=%d after=%d no-claim=%d source=%d, want 1/1/0/0", attemptCalls.Load(), continuation.afterCalls.Load(), continuation.noClaimCalls.Load(), source.calls.Load())
			}
		})
	}
}

func TestManagedStartInvokerPassesAlreadyCancelledContextToFlow(t *testing.T) {
	target := mustTarget(t, testWorkspaceID, testConfigurationID, "runtime-cancelled")
	binding := mustPrimitiveStartBinding(t, testOperationalDomain, target, testVersionID)
	request := runtimelifecycle.NewStartRequest(testWorkspaceID, testConfigurationID, testVersionID)
	contextKey := struct{}{}
	ctx, cancel := context.WithCancel(context.WithValue(context.Background(), contextKey, "preserved"))
	cancel()
	var attemptCalls atomic.Int32
	source := &unavailableSource{}
	continuation := &invokerContinuation{noClaim: func(gotCtx context.Context, gotBinding runtimeorchestrationbinding.StartExecutionBinding, cause runtimeorchestrationbinding.StartNoClaimCause) error {
		if gotCtx.Value(contextKey) != "preserved" || gotBinding != binding || cause != runtimeorchestrationbinding.StartNoClaimCancelled {
			t.Fatalf("StartNoClaim() did not receive exact binding, context value, and cancelled cause")
		}
		return nil
	}}
	flow := mustManagedFlow(t, target, &attemptCalls, source, continuation)
	invoker := mustManagedStartInvoker(t, testOperationalDomain, target, flow)

	if outcome, err := invoker.InvokeManagedStart(ctx, request, binding); outcome != (runtimelifecycle.StartOutcome{}) || err != context.Canceled {
		t.Fatalf("InvokeManagedStart(cancelled) = %#v/%v, want zero/context.Canceled", outcome, err)
	}
	if continuation.noClaimCalls.Load() != 1 || continuation.afterCalls.Load() != 0 || attemptCalls.Load() != 0 || source.calls.Load() != 0 {
		t.Fatalf("cancelled calls: no-claim=%d after=%d attempts=%d source=%d, want 1/0/0/0", continuation.noClaimCalls.Load(), continuation.afterCalls.Load(), attemptCalls.Load(), source.calls.Load())
	}
}

func TestManagedStartInvokerReturnsExactFlowOutcome(t *testing.T) {
	target := mustTarget(t, testWorkspaceID, testConfigurationID, "runtime-outcome")
	binding := mustPrimitiveStartBinding(t, testOperationalDomain, target, testVersionID)
	request := runtimelifecycle.NewStartRequest(testWorkspaceID, testConfigurationID, testVersionID)
	var attemptCalls atomic.Int32
	source := &unavailableSource{}
	continuation := &invokerContinuation{}
	flow := mustManagedFlow(t, target, &attemptCalls, source, continuation)
	invoker := mustManagedStartInvoker(t, testOperationalDomain, target, flow)

	outcome, err := invoker.InvokeManagedStart(context.Background(), request, binding)
	if err != nil || outcome.Kind() != runtimelifecycle.StartPreparationFailed {
		t.Fatalf("InvokeManagedStart() = %#v/%v, want preparation failure outcome/nil", outcome, err)
	}
	failure, ok := outcome.PreparationFailure()
	if !ok || failure != configurationloader.ErrSourceUnavailable {
		t.Fatalf("PreparationFailure() = %v/%t, want exact ErrSourceUnavailable", failure, ok)
	}
	attempt := outcome.Attempt()
	if attempt.WorkspaceID() != request.WorkspaceID() ||
		attempt.ConfigurationID() != request.ConfigurationID() ||
		attempt.ConfigurationVersionID() != request.ConfigurationVersionID() ||
		attempt.RuntimeInstanceID() != target.RuntimeInstanceID() {
		t.Fatalf("StartOutcome attempt = %#v, want exact request/target identities", attempt)
	}
	if attemptCalls.Load() != 1 || continuation.afterCalls.Load() != 1 || source.calls.Load() != 1 {
		t.Fatalf("outcome calls: attempts=%d after=%d source=%d, want 1/1/1", attemptCalls.Load(), continuation.afterCalls.Load(), source.calls.Load())
	}
}

func mustManagedStartInvoker(t *testing.T, domain string, target Target, flow *runtimelaunchflow.ManagedFlow) *ManagedStartInvoker {
	t.Helper()
	invoker, err := NewManagedStartInvoker(domain, target, flow)
	if err != nil {
		t.Fatalf("NewManagedStartInvoker() error = %v", err)
	}
	return invoker
}

func mustManagedFlow(t *testing.T, target Target, attemptCalls *atomic.Int32, source *unavailableSource, continuation runtimelaunchflow.StartClaimContinuation) *runtimelaunchflow.ManagedFlow {
	t.Helper()
	flow, err := runtimelaunchflow.NewManaged(
		mustOwnerWithCounter(t, target.WorkspaceID(), target.ConfigurationID(), target.RuntimeInstanceID(), attemptCalls),
		configurationloader.New(source),
		continuation,
	)
	if err != nil {
		t.Fatalf("NewManaged() error = %v", err)
	}
	return flow
}

func mustPrimitiveStartBinding(t *testing.T, domain string, target Target, version uint64) runtimeorchestrationbinding.StartExecutionBinding {
	t.Helper()
	return mustStartBinding(t, domain, target, runtimeorchestrationbinding.OrchestrationActionActivateExactTarget, version)
}

func mustStartBinding(t *testing.T, domain string, target Target, action runtimeorchestrationbinding.OrchestrationAction, version uint64) runtimeorchestrationbinding.StartExecutionBinding {
	t.Helper()
	authorization, err := runtimeorchestrationbinding.NewOrchestrationAuthorizationRequest(
		domain, target.WorkspaceID(), target.ConfigurationID(), target.RuntimeInstanceID(), action, version,
	)
	if err != nil {
		t.Fatalf("NewOrchestrationAuthorizationRequest() error = %v", err)
	}
	rendezvous, err := runtimeorchestrationbinding.NewStartRendezvous("rendezvous-" + string(action))
	if err != nil {
		t.Fatalf("NewStartRendezvous() error = %v", err)
	}
	if action == runtimeorchestrationbinding.OrchestrationActionActivateExactTarget {
		binding, bindingErr := runtimeorchestrationbinding.NewPrimitiveStartExecutionBinding(authorization, 1, "generation-1", rendezvous)
		if bindingErr != nil {
			t.Fatalf("NewPrimitiveStartExecutionBinding() error = %v", bindingErr)
		}
		return binding
	}
	parent, err := runtimeorchestrationbinding.NewParentCommandIdentity(authorization, "parent-"+string(action))
	if err != nil {
		t.Fatalf("NewParentCommandIdentity() error = %v", err)
	}
	phase, err := runtimeorchestrationbinding.DeriveStartTargetPhaseIdentity(parent)
	if err != nil {
		t.Fatalf("DeriveStartTargetPhaseIdentity() error = %v", err)
	}
	linked, err := runtimeorchestrationbinding.NewLinkedExecutionIdentity(parent, phase)
	if err != nil {
		t.Fatalf("NewLinkedExecutionIdentity() error = %v", err)
	}
	binding, err := runtimeorchestrationbinding.NewLinkedStartExecutionBinding(authorization, 1, "generation-1", linked, rendezvous)
	if err != nil {
		t.Fatalf("NewLinkedStartExecutionBinding() error = %v", err)
	}
	return binding
}
