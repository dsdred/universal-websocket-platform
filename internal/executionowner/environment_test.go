package executionowner

import (
	"context"
	"errors"
	"testing"
)

func TestNewExecutionEnvironmentValidatesCompleteContextBoundary(t *testing.T) {
	root := context.Background()
	execution, cancel := context.WithCancel(root)
	defer cancel()

	tests := []struct {
		name      string
		root      context.Context
		execution context.Context
		cancel    context.CancelFunc
	}{
		{name: "nil root", execution: execution, cancel: cancel},
		{name: "nil execution", root: root, cancel: cancel},
		{name: "nil cancellation", root: root, execution: execution},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			environment, err := NewExecutionEnvironment(test.root, test.execution, test.cancel)
			if environment.valid() || !errors.Is(err, ErrInvalidExecutionEnvironment) {
				t.Fatalf("NewExecutionEnvironment() = (%+v, %v), want invalid and %v", environment, err, ErrInvalidExecutionEnvironment)
			}
		})
	}
}

func TestExecutionEnvironmentCopyPreservesPreparedIdentity(t *testing.T) {
	root := context.Background()
	execution, cancel := context.WithCancel(root)
	environment, err := NewExecutionEnvironment(root, execution, cancel)
	if err != nil {
		t.Fatalf("NewExecutionEnvironment() error = %v", err)
	}
	copied := environment

	if !environment.valid() || !copied.valid() || copied.root != environment.root ||
		copied.execution != environment.execution {
		t.Fatal("copied ExecutionEnvironment does not preserve prepared context identity")
	}
	copied.cancel()
	if !errors.Is(execution.Err(), context.Canceled) {
		t.Fatalf("execution context error = %v, want context.Canceled", execution.Err())
	}
}
