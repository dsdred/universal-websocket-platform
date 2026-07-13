package secretresolver

import (
	"context"
	"errors"
	"strings"
	"unicode/utf8"
)

const maxReferenceLength = 255

var (
	// ErrSecretNotFound indicates that a Secret Reference has no stored value.
	ErrSecretNotFound = errors.New("secret not found")
	// ErrInvalidSecretReference indicates that a Secret Reference has an invalid format.
	ErrInvalidSecretReference = errors.New("invalid secret reference")
	// ErrEmptySecret indicates that a Secret value is empty.
	ErrEmptySecret = errors.New("secret value is empty")
)

// Secret contains resolved secret bytes owned by the caller.
type Secret struct {
	Value []byte
}

// Resolver resolves Secret References without exposing a storage implementation.
type Resolver interface {
	Resolve(ctx context.Context, ref string) (Secret, error)
}

// ValidateReference validates a Secret Reference after trimming surrounding whitespace.
func ValidateReference(ref string) error {
	_, err := normalizeReference(ref)
	return err
}

func normalizeReference(ref string) (string, error) {
	ref = strings.TrimSpace(ref)
	if ref == "" || utf8.RuneCountInString(ref) > maxReferenceLength {
		return "", ErrInvalidSecretReference
	}
	if strings.HasPrefix(ref, "/") || strings.HasSuffix(ref, "/") ||
		strings.Contains(ref, "//") || strings.Contains(ref, "://") ||
		strings.Contains(ref, "-----BEGIN") {
		return "", ErrInvalidSecretReference
	}

	for _, character := range ref {
		if !isReferenceCharacter(character) {
			return "", ErrInvalidSecretReference
		}
	}

	return ref, nil
}

func isReferenceCharacter(character rune) bool {
	return (character >= 'a' && character <= 'z') ||
		(character >= 'A' && character <= 'Z') ||
		(character >= '0' && character <= '9') ||
		character == '/' || character == '-' || character == '_' || character == '.'
}
