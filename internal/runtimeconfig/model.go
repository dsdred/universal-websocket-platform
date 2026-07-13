package runtimeconfig

// Snapshot is an immutable runtime view of a Published Configuration Version.
type Snapshot struct {
	ConfigurationID uint64
	VersionID       uint64

	Listener       ListenerSnapshot
	Authentication AuthenticationSnapshot
}

// ListenerSnapshot contains values required to configure a runtime Listener.
type ListenerSnapshot struct {
	Host     string
	Port     uint16
	TLS      TLSSnapshot
	Timeouts TimeoutSnapshot
}

// TLSSnapshot contains runtime TLS settings and Secret References.
type TLSSnapshot struct {
	Enabled        bool
	CertificateRef string
	PrivateKeyRef  string
	MinVersion     string
}

// TimeoutSnapshot contains Listener timeout values in seconds.
type TimeoutSnapshot struct {
	HandshakeSeconds uint32
	ReadSeconds      uint32
	WriteSeconds     uint32
	IdleSeconds      uint32
}

// AuthenticationSnapshot contains values required to configure runtime Authentication.
type AuthenticationSnapshot struct {
	Enabled   bool
	Providers []AuthenticationProviderSnapshot
}

// AuthenticationProviderType identifies an Authentication Provider implementation.
type AuthenticationProviderType string

const (
	AuthenticationProviderJWT    AuthenticationProviderType = "jwt"
	AuthenticationProviderAPIKey AuthenticationProviderType = "api-key"
	AuthenticationProviderBasic  AuthenticationProviderType = "basic"
)

// AuthenticationProviderSnapshot contains runtime settings for one Authentication Provider.
type AuthenticationProviderSnapshot struct {
	Name     string
	Type     AuthenticationProviderType
	Enabled  bool
	Priority uint32
	APIKey   *APIKeySnapshot
	JWT      *JWTSnapshot
	Basic    *BasicSnapshot
}

// APIKeySnapshot contains runtime API Key Provider settings.
type APIKeySnapshot struct {
	Header    string
	SecretRef string
}

// BasicSnapshot contains runtime Basic Provider settings.
type BasicSnapshot struct {
	Realm     string
	SecretRef string
}

// JWTSnapshot contains runtime JWT verification policy.
type JWTSnapshot struct {
	SigningKeys       []JWTSigningKeySnapshot
	AllowedAlgorithms []JWTAlgorithm
	AllowedIssuers    []string
	AllowedAudiences  []string
	RequiredClaims    []JWTRequiredClaimSnapshot
	ClockSkewSeconds  uint32
}

// JWTSigningKeySnapshot identifies one JWT Signing Key by Secret Reference.
type JWTSigningKeySnapshot struct {
	Name      string
	SecretRef string
}

// JWTRequiredClaimSnapshot contains one required JWT Claim and value.
type JWTRequiredClaimSnapshot struct {
	Name  string
	Value string
}

// JWTAlgorithm identifies an allowed JWT signing algorithm.
type JWTAlgorithm string

const (
	HS256 JWTAlgorithm = "HS256"
	HS384 JWTAlgorithm = "HS384"
	HS512 JWTAlgorithm = "HS512"
	RS256 JWTAlgorithm = "RS256"
	RS384 JWTAlgorithm = "RS384"
	RS512 JWTAlgorithm = "RS512"
	ES256 JWTAlgorithm = "ES256"
	ES384 JWTAlgorithm = "ES384"
	ES512 JWTAlgorithm = "ES512"
	PS256 JWTAlgorithm = "PS256"
	PS384 JWTAlgorithm = "PS384"
	PS512 JWTAlgorithm = "PS512"
)
