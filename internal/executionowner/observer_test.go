package executionowner_test

import (
	"reflect"
	"testing"

	"github.com/dsdred/universal-websocket-platform/internal/executionowner"
)

func TestTerminalObserverReceivesOnlyTerminalResult(t *testing.T) {
	observerType := reflect.TypeOf((*executionowner.TerminalObserver)(nil)).Elem()
	if observerType.NumMethod() != 1 {
		t.Fatalf("TerminalObserver method count = %d, want 1", observerType.NumMethod())
	}
	method := observerType.Method(0)
	if method.Name != "Observe" {
		t.Fatalf("TerminalObserver method = %q, want Observe", method.Name)
	}
	if method.Type.NumIn() != 1 || method.Type.In(0) != reflect.TypeOf(executionowner.TerminalResult{}) {
		t.Fatalf("Observe input = %v, want TerminalResult only", method.Type)
	}
	if method.Type.NumOut() != 0 {
		t.Fatalf("Observe output count = %d, want 0", method.Type.NumOut())
	}
}

func TestTerminalObserverCanRetainDetachedComparableValue(t *testing.T) {
	result, err := validTerminalResultInput().build()
	if err != nil {
		t.Fatalf("NewTerminalResult() error = %v", err)
	}
	observer := &recordingTerminalObserver{}
	var contract executionowner.TerminalObserver = observer

	contract.Observe(result)
	if observer.result != result {
		t.Fatalf("Observer result = %+v, want %+v", observer.result, result)
	}
}

type recordingTerminalObserver struct {
	result executionowner.TerminalResult
}

func (observer *recordingTerminalObserver) Observe(result executionowner.TerminalResult) {
	observer.result = result
}
