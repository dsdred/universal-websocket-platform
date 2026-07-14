// Package message defines transport-neutral Runtime messages.
package message

import (
	"errors"
	"time"
)

var ErrInvalidMessageType = errors.New("invalid Runtime message type")

// Type identifies an application message kind.
type Type string

const (
	TypeText   Type = "text"
	TypeBinary Type = "binary"
)

// Message is an immutable transport-neutral application message.
type Message struct {
	messageType Type
	data        []byte
	receivedAt  time.Time
}

// New creates an immutable Runtime Message and copies its payload.
func New(messageType Type, data []byte) (Message, error) {
	if messageType != TypeText && messageType != TypeBinary {
		return Message{}, ErrInvalidMessageType
	}
	return Message{
		messageType: messageType,
		data:        append([]byte(nil), data...),
		receivedAt:  time.Now().UTC(),
	}, nil
}

// Type returns the application message type.
func (message Message) Type() Type {
	return message.messageType
}

// Data returns an independent copy of the payload.
func (message Message) Data() []byte {
	return append([]byte(nil), message.data...)
}

// ReceivedAt returns the UTC receive time.
func (message Message) ReceivedAt() time.Time {
	return message.receivedAt
}
