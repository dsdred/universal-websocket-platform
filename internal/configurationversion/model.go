package configurationversion

import "time"

// VersionState represents the lifecycle state of a Configuration Version.
type VersionState string

const (
	Draft     VersionState = "Draft"
	Validated VersionState = "Validated"
	Published VersionState = "Published"
	Archived  VersionState = "Archived"
)

// JWTAlgorithm identifies a permitted JWT signing algorithm.
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

// AuthenticationProviderType identifies the configured Authentication mechanism.
type AuthenticationProviderType string

const (
	AuthenticationProviderJWT    AuthenticationProviderType = "jwt"
	AuthenticationProviderAPIKey AuthenticationProviderType = "api-key"
	AuthenticationProviderBasic  AuthenticationProviderType = "basic"
)

// AuthenticationSettings describes Authentication metadata for a Configuration Version.
type AuthenticationSettings struct {
	Enabled   bool                     `json:"enabled"`
	Providers []AuthenticationProvider `json:"providers"`
}

// AuthenticationProvider describes one configured Authentication Provider.
type AuthenticationProvider struct {
	Name     string                     `json:"name"`
	Type     AuthenticationProviderType `json:"type"`
	Enabled  bool                       `json:"enabled"`
	Priority uint32                     `json:"priority"`
	APIKey   *APIKeySettings            `json:"apiKey,omitempty"`
	JWT      *JWTSettings               `json:"jwt,omitempty"`
	Basic    *BasicSettings             `json:"basic,omitempty"`
}

// MatcherType identifies one supported transport-neutral routing predicate.
type MatcherType string

const (
	MatcherTypeMessageType            MatcherType = "message-type"
	MatcherTypePrincipalKind          MatcherType = "principal-kind"
	MatcherTypeAuthenticationType     MatcherType = "authentication-type"
	MatcherTypeAuthenticationProvider MatcherType = "authentication-provider"
)

// RoutingSettings describes optional declarative Runtime Message routing metadata.
type RoutingSettings struct {
	Routes            []Route `json:"routes"`
	DefaultHandlerRef string  `json:"defaultHandlerRef,omitempty"`
}

// Route describes one ordered Handler selection rule.
type Route struct {
	ID         string    `json:"id"`
	Enabled    bool      `json:"enabled"`
	Priority   uint32    `json:"priority"`
	Matchers   []Matcher `json:"matchers"`
	HandlerRef string    `json:"handlerRef"`
}

// Matcher describes one exact equality predicate over Runtime Message Context.
type Matcher struct {
	Type  MatcherType `json:"type"`
	Value string      `json:"value"`
}

// APIKeySettings describes API Key Provider metadata.
type APIKeySettings struct {
	Header    string `json:"header"`
	SecretRef string `json:"secretRef"`
}

// BasicSettings describes Basic Authentication Provider metadata.
type BasicSettings struct {
	Realm     string `json:"realm"`
	SecretRef string `json:"secretRef"`
}

// JWTSettings describes JWT Provider verification-policy metadata.
type JWTSettings struct {
	SigningKeys       []JWTSigningKey    `json:"signingKeys"`
	AllowedAlgorithms []JWTAlgorithm     `json:"allowedAlgorithms"`
	AllowedIssuers    []string           `json:"allowedIssuers"`
	AllowedAudiences  []string           `json:"allowedAudiences"`
	RequiredClaims    []JWTRequiredClaim `json:"requiredClaims"`
	ClockSkewSeconds  uint32             `json:"clockSkewSeconds"`
}

// JWTSigningKey identifies one signing key by Secret Reference.
type JWTSigningKey struct {
	Name      string `json:"name"`
	SecretRef string `json:"secretRef"`
}

// JWTRequiredClaim declares a required JWT Claim and value.
type JWTRequiredClaim struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// ListenerSettings describes where a future WebSocket Listener will accept connections.
type ListenerSettings struct {
	Host     string          `json:"host"`
	Port     uint16          `json:"port"`
	TLS      TLSSettings     `json:"tls"`
	Timeouts TimeoutSettings `json:"timeouts"`
}

// TimeoutSettings describes timeout metadata in seconds for a future Listener.
type TimeoutSettings struct {
	HandshakeSeconds uint32 `json:"handshakeSeconds"`
	ReadSeconds      uint32 `json:"readSeconds"`
	WriteSeconds     uint32 `json:"writeSeconds"`
	IdleSeconds      uint32 `json:"idleSeconds"`
}

// TLSSettings describes TLS metadata for a future secure WebSocket Listener.
type TLSSettings struct {
	Enabled        bool   `json:"enabled"`
	CertificateRef string `json:"certificateRef"`
	PrivateKeyRef  string `json:"privateKeyRef"`
	MinVersion     string `json:"minVersion"`
}

// ConfigurationVersion contains metadata for a Configuration Version.
type ConfigurationVersion struct {
	ID              uint64                 `json:"id"`
	ConfigurationID uint64                 `json:"configurationId"`
	Number          uint32                 `json:"number"`
	State           VersionState           `json:"state"`
	Listener        ListenerSettings       `json:"listener"`
	Authentication  AuthenticationSettings `json:"authentication"`
	Routing         *RoutingSettings       `json:"routing,omitempty"`
	CreatedAt       time.Time              `json:"createdAt"`
	UpdatedAt       time.Time              `json:"updatedAt"`
}
