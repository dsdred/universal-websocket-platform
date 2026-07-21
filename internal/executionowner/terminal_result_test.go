package executionowner_test

import (
	"errors"
	"reflect"
	"testing"

	"github.com/dsdred/universal-websocket-platform/internal/executionowner"
)

func TestNewTerminalResultPreservesImmutableCategories(t *testing.T) {
	input := validTerminalResultInput()
	result, err := input.build()
	if err != nil {
		t.Fatalf("NewTerminalResult() error = %v", err)
	}

	if result.StartCategory() != input.start ||
		result.RunCategory() != input.run ||
		result.CleanupCategory() != input.cleanup ||
		result.CancellationCategory() != input.cancellation ||
		result.CompletionCategory() != input.completion ||
		result.RecoveredPanicPhase() != input.panicPhase ||
		result.PrimaryCause() != input.primary ||
		result.SecondaryCauses() != input.secondary ||
		result.RuntimeObservationAnomalies() != input.observation {
		t.Fatalf("Terminal Result accessors do not preserve construction input: %+v", result)
	}
	if !result.SecondaryCauses().Contains(executionowner.TerminationCauseExplicitStop) {
		t.Fatal("SecondaryCauses() lost ExplicitStop")
	}
	if !result.RuntimeObservationAnomalies().Has(
		executionowner.RuntimeObservationAnomalyInstallFailure |
			executionowner.RuntimeObservationAnomalyInvocationPanic,
	) {
		t.Fatal("RuntimeObservationAnomalies() lost the bounded anomaly set")
	}
}

func TestTerminalResultHasStructuralEqualityAndNativeComparability(t *testing.T) {
	input := validTerminalResultInput()
	first, err := input.build()
	if err != nil {
		t.Fatalf("first NewTerminalResult() error = %v", err)
	}
	second, err := input.build()
	if err != nil {
		t.Fatalf("second NewTerminalResult() error = %v", err)
	}
	copy := first
	if first != second || first != copy {
		t.Fatalf("equal construction values differ: first=%+v second=%+v copy=%+v", first, second, copy)
	}

	differentInput := input
	differentInput.completion = executionowner.CompletionCategoryAccountingAnomaly
	different, err := differentInput.build()
	if err != nil {
		t.Fatalf("different NewTerminalResult() error = %v", err)
	}
	if first == different {
		t.Fatal("Terminal Results with different bounded fields compare equal")
	}

	values := map[executionowner.TerminalResult]string{first: "observed"}
	if got := values[copy]; got != "observed" {
		t.Fatalf("native map lookup through copied Terminal Result = %q", got)
	}
	if !reflect.TypeOf(first).Comparable() {
		t.Fatal("TerminalResult is not natively comparable")
	}
}

func TestNewTerminalResultRejectsInvalidInput(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*terminalResultInput)
	}{
		{name: "unknown Start", mutate: func(input *terminalResultInput) { input.start = 255 }},
		{name: "unknown Run", mutate: func(input *terminalResultInput) { input.run = 255 }},
		{name: "unknown Cleanup", mutate: func(input *terminalResultInput) { input.cleanup = 255 }},
		{name: "Cleanup blocked", mutate: func(input *terminalResultInput) {
			input.cleanup = executionowner.CleanupCategoryBlocked
		}},
		{name: "missing Cancellation", mutate: func(input *terminalResultInput) { input.cancellation = 0 }},
		{name: "unknown Cancellation", mutate: func(input *terminalResultInput) { input.cancellation = 255 }},
		{name: "missing Completion", mutate: func(input *terminalResultInput) { input.completion = 0 }},
		{name: "unknown Completion", mutate: func(input *terminalResultInput) { input.completion = 255 }},
		{name: "unknown panic phase", mutate: func(input *terminalResultInput) { input.panicPhase = 255 }},
		{name: "missing primary cause", mutate: func(input *terminalResultInput) { input.primary = 0 }},
		{name: "unknown primary cause", mutate: func(input *terminalResultInput) { input.primary = 255 }},
		{name: "unknown secondary bit", mutate: func(input *terminalResultInput) { input.secondary |= 1 << 7 }},
		{name: "unknown observation bit", mutate: func(input *terminalResultInput) { input.observation |= 1 << 7 }},
		{name: "primary duplicated as secondary", mutate: func(input *terminalResultInput) {
			input.secondary |= executionowner.SecondaryCauseNaturalCompletion
		}},
		{name: "Run returned without successful Start", mutate: func(input *terminalResultInput) {
			input.start = executionowner.StartCategoryFailed
			input.primary = executionowner.TerminationCauseExecutionFailure
			input.secondary = executionowner.SecondaryCauseNaturalCompletion
		}},
		{name: "failed Start without ExecutionFailure", mutate: func(input *terminalResultInput) {
			input.start = executionowner.StartCategoryFailed
			input.run = executionowner.RunCategoryNotStarted
			input.primary = executionowner.TerminationCauseExplicitStop
			input.secondary = 0
		}},
		{name: "failed Run without ExecutionFailure", mutate: func(input *terminalResultInput) {
			input.run = executionowner.RunCategoryFailed
			input.primary = executionowner.TerminationCauseExplicitStop
			input.secondary = 0
		}},
		{name: "returned Run without NaturalCompletion", mutate: func(input *terminalResultInput) {
			input.primary = executionowner.TerminationCauseExplicitStop
			input.secondary = 0
		}},
		{name: "NaturalCompletion without returned Run", mutate: func(input *terminalResultInput) {
			input.run = executionowner.RunCategoryNotStarted
		}},
		{name: "panic cause without phase", mutate: func(input *terminalResultInput) {
			input.primary = executionowner.TerminationCauseRecoveredPanic
			input.secondary = 0
			input.run = executionowner.RunCategoryPanicked
		}},
		{name: "Start panic with wrong phase", mutate: func(input *terminalResultInput) {
			input.start = executionowner.StartCategoryPanicked
			input.run = executionowner.RunCategoryNotStarted
			input.primary = executionowner.TerminationCauseRecoveredPanic
			input.secondary = 0
			input.panicPhase = executionowner.RecoveredPanicPhaseRun
		}},
		{name: "Run panic with wrong phase", mutate: func(input *terminalResultInput) {
			input.run = executionowner.RunCategoryPanicked
			input.primary = executionowner.TerminationCauseRecoveredPanic
			input.secondary = 0
			input.panicPhase = executionowner.RecoveredPanicPhaseStart
		}},
		{name: "phase without panic cause", mutate: func(input *terminalResultInput) {
			input.run = executionowner.RunCategoryPanicked
			input.panicPhase = executionowner.RecoveredPanicPhaseRun
		}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			input := validTerminalResultInput()
			test.mutate(&input)
			result, err := input.build()
			if result != (executionowner.TerminalResult{}) {
				t.Fatalf("NewTerminalResult() result = %+v, want zero value", result)
			}
			if !errors.Is(err, executionowner.ErrInvalidTerminalResult) {
				t.Fatalf("NewTerminalResult() error = %v, want ErrInvalidTerminalResult", err)
			}
		})
	}
}

func TestNewTerminalResultAcceptsConsistentExecutionPaths(t *testing.T) {
	tests := []struct {
		name  string
		input terminalResultInput
	}{
		{name: "normal Run return", input: validTerminalResultInput()},
		{name: "explicit Stop before Start", input: terminalResultInput{
			start:        executionowner.StartCategoryNotAttempted,
			run:          executionowner.RunCategoryNotStarted,
			cleanup:      executionowner.CleanupCategorySucceeded,
			cancellation: executionowner.CancellationCategoryConfirmed,
			completion:   executionowner.CompletionCategoryCompleted,
			primary:      executionowner.TerminationCauseExplicitStop,
		}},
		{name: "Start failure", input: terminalResultInput{
			start:        executionowner.StartCategoryFailed,
			run:          executionowner.RunCategoryNotStarted,
			cleanup:      executionowner.CleanupCategoryFailed,
			cancellation: executionowner.CancellationCategoryConfirmed,
			completion:   executionowner.CompletionCategoryAccountingAnomaly,
			primary:      executionowner.TerminationCauseExecutionFailure,
		}},
		{name: "recovered Start panic", input: terminalResultInput{
			start:        executionowner.StartCategoryPanicked,
			run:          executionowner.RunCategoryNotStarted,
			cleanup:      executionowner.CleanupCategoryPanicked,
			cancellation: executionowner.CancellationCategoryAnomaly,
			completion:   executionowner.CompletionCategoryPanicked,
			panicPhase:   executionowner.RecoveredPanicPhaseStart,
			primary:      executionowner.TerminationCauseRecoveredPanic,
		}},
		{name: "recovered Run panic", input: terminalResultInput{
			start:        executionowner.StartCategorySucceeded,
			run:          executionowner.RunCategoryPanicked,
			cleanup:      executionowner.CleanupCategorySucceeded,
			cancellation: executionowner.CancellationCategoryConfirmed,
			completion:   executionowner.CompletionCategoryCompleted,
			panicPhase:   executionowner.RecoveredPanicPhaseRun,
			primary:      executionowner.TerminationCauseRecoveredPanic,
		}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if _, err := test.input.build(); err != nil {
				t.Fatalf("NewTerminalResult() error = %v", err)
			}
		})
	}
}

func TestTerminalResultAPIContainsOnlyReadOnlyBoundedValues(t *testing.T) {
	resultType := reflect.TypeOf(executionowner.TerminalResult{})
	if !resultType.Comparable() {
		t.Fatal("TerminalResult is not comparable")
	}
	for index := 0; index < resultType.NumField(); index++ {
		field := resultType.Field(index)
		if field.IsExported() {
			t.Errorf("TerminalResult field %q is exported", field.Name)
		}
		switch field.Type.Kind() {
		case reflect.Bool, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		default:
			t.Errorf("TerminalResult field %q has non-bounded kind %s", field.Name, field.Type.Kind())
		}
	}

	wantMethods := map[string]struct{}{
		"CancellationCategory":        {},
		"CleanupCategory":             {},
		"CompletionCategory":          {},
		"PrimaryCause":                {},
		"RecoveredPanicPhase":         {},
		"RunCategory":                 {},
		"RuntimeObservationAnomalies": {},
		"SecondaryCauses":             {},
		"StartCategory":               {},
	}
	if resultType.NumMethod() != len(wantMethods) {
		t.Fatalf("TerminalResult method count = %d, want %d", resultType.NumMethod(), len(wantMethods))
	}
	for index := 0; index < resultType.NumMethod(); index++ {
		delete(wantMethods, resultType.Method(index).Name)
	}
	if len(wantMethods) != 0 {
		t.Fatalf("TerminalResult missing methods: %v", wantMethods)
	}
}

type terminalResultInput struct {
	start        executionowner.StartCategory
	run          executionowner.RunCategory
	cleanup      executionowner.CleanupCategory
	cancellation executionowner.CancellationCategory
	completion   executionowner.CompletionCategory
	panicPhase   executionowner.RecoveredPanicPhase
	primary      executionowner.TerminationCause
	secondary    executionowner.SecondaryCauses
	observation  executionowner.RuntimeObservationAnomalies
}

func validTerminalResultInput() terminalResultInput {
	return terminalResultInput{
		start:        executionowner.StartCategorySucceeded,
		run:          executionowner.RunCategoryReturned,
		cleanup:      executionowner.CleanupCategorySucceeded,
		cancellation: executionowner.CancellationCategoryConfirmed,
		completion:   executionowner.CompletionCategoryCompleted,
		primary:      executionowner.TerminationCauseNaturalCompletion,
		secondary:    executionowner.SecondaryCauseExplicitStop,
		observation: executionowner.RuntimeObservationAnomalyInstallFailure |
			executionowner.RuntimeObservationAnomalyInvocationPanic,
	}
}

func (input terminalResultInput) build() (executionowner.TerminalResult, error) {
	return executionowner.NewTerminalResult(
		input.start,
		input.run,
		input.cleanup,
		input.cancellation,
		input.completion,
		input.panicPhase,
		input.primary,
		input.secondary,
		input.observation,
	)
}
