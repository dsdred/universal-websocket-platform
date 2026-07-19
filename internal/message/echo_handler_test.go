package message

import (
	"context"
	"errors"
	"testing"
)

func TestEchoHandlerSendsTextAndBinaryMessagesUnchanged(t *testing.T) {
	for _, messageType := range []Type{TypeText, TypeBinary} {
		t.Run(string(messageType), func(t *testing.T) {
			runtimeMessage, err := New(messageType, []byte("payload"))
			if err != nil {
				t.Fatalf("New() error = %v", err)
			}
			sender := &recordingSender{}
			runtimeContext, err := NewContext(&runtimeMessage, sender, "session", true, false, "jwt", "provider")
			if err != nil {
				t.Fatalf("NewContext() error = %v", err)
			}
			handler := NewEchoHandler()
			if err := handler.Handle(context.Background(), runtimeContext); err != nil {
				t.Fatalf("Handle() error = %v", err)
			}
			if sender.calls != 1 || sender.message.Type() != runtimeMessage.Type() ||
				string(sender.message.Data()) != string(runtimeMessage.Data()) ||
				!sender.message.ReceivedAt().Equal(runtimeMessage.ReceivedAt()) {
				t.Fatalf("sent Message = %+v, want original Message", sender.message)
			}
		})
	}
}

func TestEchoHandlerReturnsSendError(t *testing.T) {
	wantErr := errors.New("send failed")
	runtimeMessage, err := New(TypeText, []byte("payload"))
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	sender := &recordingSender{err: wantErr}
	runtimeContext, err := NewContext(&runtimeMessage, sender, "session", true, false, "jwt", "provider")
	if err != nil {
		t.Fatalf("NewContext() error = %v", err)
	}
	err = NewEchoHandler().Handle(context.Background(), runtimeContext)
	if !errors.Is(err, wantErr) {
		t.Fatalf("Handle() error = %v, want Send error", err)
	}
}

type recordingSender struct {
	message Message
	err     error
	calls   int
}

func (sender *recordingSender) Send(_ context.Context, runtimeMessage Message) error {
	sender.calls++
	sender.message = runtimeMessage
	return sender.err
}
