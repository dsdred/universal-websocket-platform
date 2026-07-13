package secretresolver

import (
	"errors"
	"strings"
	"testing"
)

func TestValidateReferenceAcceptsValidReferences(t *testing.T) {
	references := []string{
		"secrets/api-keys/internal",
		"secrets/jwt/main",
		"certificates/default/server",
		"workspace/main/oauth/client-secret",
		"  secrets/jwt/main  ",
	}

	for _, ref := range references {
		t.Run(ref, func(t *testing.T) {
			if err := ValidateReference(ref); err != nil {
				t.Errorf("ValidateReference(%q) error = %v", ref, err)
			}
		})
	}
}

func TestValidateReferenceRejectsInvalidReferences(t *testing.T) {
	tests := []struct {
		name string
		ref  string
	}{
		{name: "empty", ref: ""},
		{name: "whitespace", ref: "   "},
		{name: "leading slash", ref: "/secrets/main"},
		{name: "trailing slash", ref: "secrets/main/"},
		{name: "double slash", ref: "secrets//main"},
		{name: "URL", ref: "https://example.com/secret"},
		{name: "Windows path", ref: `C:\secrets\key`},
		{name: "Unix absolute path", ref: "/tmp/key"},
		{name: "PEM", ref: "-----BEGIN PRIVATE KEY-----"},
		{name: "spaces", ref: "actual secret with spaces"},
		{name: "invalid character", ref: "secrets/main@current"},
		{name: "too long", ref: strings.Repeat("a", 256)},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := ValidateReference(test.ref)
			if !errors.Is(err, ErrInvalidSecretReference) {
				t.Errorf("ValidateReference(%q) error = %v, want ErrInvalidSecretReference", test.ref, err)
			}
		})
	}
}
