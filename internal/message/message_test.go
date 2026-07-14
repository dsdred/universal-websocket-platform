package message

import (
	"errors"
	"reflect"
	"testing"
	"time"
)

func TestNewTextAndBinaryMessages(t *testing.T) {
	for _, messageType := range []Type{TypeText, TypeBinary} {
		t.Run(string(messageType), func(t *testing.T) {
			createdBefore := time.Now().UTC()
			message, err := New(messageType, []byte("payload"))
			if err != nil {
				t.Fatalf("New() error = %v", err)
			}
			if message.Type() != messageType {
				t.Fatalf("Type() = %q, want %q", message.Type(), messageType)
			}
			if string(message.Data()) != "payload" {
				t.Fatalf("Data() = %q", message.Data())
			}
			if message.ReceivedAt().Before(createdBefore) || message.ReceivedAt().Location() != time.UTC {
				t.Fatalf("ReceivedAt() = %v", message.ReceivedAt())
			}
		})
	}
}

func TestNewRejectsInvalidMessageType(t *testing.T) {
	message, err := New(Type("ping"), nil)
	if !errors.Is(err, ErrInvalidMessageType) {
		t.Fatalf("New() error = %v, want ErrInvalidMessageType", err)
	}
	if !reflect.DeepEqual(message, Message{}) {
		t.Fatalf("New() Message = %+v, want zero value", message)
	}
}

func TestMessageCopiesInputAndReturnedData(t *testing.T) {
	payload := []byte("original")
	message, err := New(TypeBinary, payload)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	payload[0] = 'X'
	firstCopy := message.Data()
	firstCopy[1] = 'X'
	if got := string(message.Data()); got != "original" {
		t.Fatalf("Data() = %q, want original", got)
	}
}

func TestMessageContainsOnlyRuntimeData(t *testing.T) {
	messageType := reflect.TypeOf(Message{})
	if messageType.NumField() != 3 {
		t.Fatalf("Message fields = %d, want 3", messageType.NumField())
	}
	for index := 0; index < messageType.NumField(); index++ {
		if messageType.Field(index).IsExported() {
			t.Fatalf("Message field %q is exported", messageType.Field(index).Name)
		}
	}
}
