package runtimeconfig

import "github.com/dsdred/universal-websocket-platform/internal/runtimeconfigload"

// Snapshot is the immutable, detached Runtime view of one loaded Published
// ConfigurationVersion for one Launch Attempt.
type Snapshot struct {
	provenance     Provenance
	listener       ListenerSnapshot
	authentication AuthenticationSnapshot
	routing        *RoutingSnapshot
}

// Provenance identifies the declarative and operational origin of a Snapshot.
type Provenance struct {
	WorkspaceID                uint64
	ConfigurationID            uint64
	ConfigurationVersionID     uint64
	ConfigurationVersionNumber uint32
	SchemaIdentity             string
	SchemaVersion              uint32
	RuntimeInstanceID          runtimeconfigload.RuntimeInstanceID
	LaunchAttemptID            runtimeconfigload.LaunchAttemptID
}

// Provenance returns a detached copy of the Snapshot provenance.
func (s Snapshot) Provenance() Provenance { return s.provenance }

// Listener returns a detached copy of the Listener configuration.
func (s Snapshot) Listener() ListenerSnapshot { return s.listener }

// Authentication returns a recursively detached copy of Authentication.
func (s Snapshot) Authentication() AuthenticationSnapshot {
	return cloneAuthenticationSnapshot(s.authentication)
}

// Routing returns presence and a recursively detached copy of Routing.
func (s Snapshot) Routing() (RoutingSnapshot, bool) {
	if s.routing == nil {
		return RoutingSnapshot{}, false
	}
	return cloneRoutingSnapshot(*s.routing), true
}

// RoutingSnapshot is an immutable Runtime view of declarative Routing metadata.
type RoutingSnapshot struct {
	routes                   []RouteSnapshot
	defaultHandlerRef        string
	defaultHandlerRefPresent bool
}

// Routes returns a detached copy of configured Routes in declaration order.
func (s RoutingSnapshot) Routes() []RouteSnapshot {
	return copySlice(s.routes, cloneRouteSnapshot)
}

// DefaultHandlerRef returns presence and the normalized default Handler reference.
func (s RoutingSnapshot) DefaultHandlerRef() (string, bool) {
	return s.defaultHandlerRef, s.defaultHandlerRefPresent
}

// RouteSnapshot is an immutable runtime view of one declarative Route.
type RouteSnapshot struct {
	id         string
	enabled    bool
	priority   uint32
	matchers   []MatcherSnapshot
	handlerRef string
}

func (s RouteSnapshot) ID() string                  { return s.id }
func (s RouteSnapshot) Enabled() bool               { return s.enabled }
func (s RouteSnapshot) Priority() uint32            { return s.priority }
func (s RouteSnapshot) Matchers() []MatcherSnapshot { return cloneSlice(s.matchers) }
func (s RouteSnapshot) HandlerRef() string          { return s.handlerRef }

type MatcherType string

const (
	MatcherTypeMessageType            MatcherType = "message-type"
	MatcherTypePrincipalKind          MatcherType = "principal-kind"
	MatcherTypeAuthenticationType     MatcherType = "authentication-type"
	MatcherTypeAuthenticationProvider MatcherType = "authentication-provider"
)

type MatcherSnapshot struct {
	matcherType MatcherType
	value       string
}

func (s MatcherSnapshot) Type() MatcherType { return s.matcherType }
func (s MatcherSnapshot) Value() string     { return s.value }

type ListenerSnapshot struct {
	Host     string
	Port     uint16
	TLS      TLSSnapshot
	Timeouts TimeoutSnapshot
}

type TLSSnapshot struct {
	Enabled        bool
	CertificateRef string
	PrivateKeyRef  string
	MinVersion     string
}

type TimeoutSnapshot struct {
	HandshakeSeconds uint32
	ReadSeconds      uint32
	WriteSeconds     uint32
	IdleSeconds      uint32
}

type AuthenticationSnapshot struct {
	Enabled   bool
	Providers []AuthenticationProviderSnapshot
}

type AuthenticationProviderType string

const (
	AuthenticationProviderJWT    AuthenticationProviderType = "jwt"
	AuthenticationProviderAPIKey AuthenticationProviderType = "api-key"
	AuthenticationProviderBasic  AuthenticationProviderType = "basic"
)

type AuthenticationProviderSnapshot struct {
	Name     string
	Type     AuthenticationProviderType
	Enabled  bool
	Priority uint32
	APIKey   *APIKeySnapshot
	JWT      *JWTSnapshot
	Basic    *BasicSnapshot
}

type APIKeySnapshot struct {
	Header    string
	SecretRef string
}

type BasicSnapshot struct {
	Realm     string
	SecretRef string
}

type JWTSnapshot struct {
	SigningKeys       []JWTSigningKeySnapshot
	AllowedAlgorithms []JWTAlgorithm
	AllowedIssuers    []string
	AllowedAudiences  []string
	RequiredClaims    []JWTRequiredClaimSnapshot
	ClockSkewSeconds  uint32
}

type JWTSigningKeySnapshot struct {
	Name      string
	SecretRef string
}

type JWTRequiredClaimSnapshot struct {
	Name  string
	Value string
}

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

func cloneAuthenticationSnapshot(source AuthenticationSnapshot) AuthenticationSnapshot {
	result := source
	result.Providers = copySlice(source.Providers, cloneAuthenticationProviderSnapshot)
	return result
}

func cloneAuthenticationProviderSnapshot(source AuthenticationProviderSnapshot) AuthenticationProviderSnapshot {
	result := source
	if source.APIKey != nil {
		value := *source.APIKey
		result.APIKey = &value
	}
	if source.Basic != nil {
		value := *source.Basic
		result.Basic = &value
	}
	if source.JWT != nil {
		value := *source.JWT
		value.SigningKeys = cloneSlice(source.JWT.SigningKeys)
		value.AllowedAlgorithms = cloneSlice(source.JWT.AllowedAlgorithms)
		value.AllowedIssuers = cloneSlice(source.JWT.AllowedIssuers)
		value.AllowedAudiences = cloneSlice(source.JWT.AllowedAudiences)
		value.RequiredClaims = cloneSlice(source.JWT.RequiredClaims)
		result.JWT = &value
	}
	return result
}

func cloneRoutingSnapshot(source RoutingSnapshot) RoutingSnapshot {
	source.routes = copySlice(source.routes, cloneRouteSnapshot)
	return source
}

func cloneRouteSnapshot(route RouteSnapshot) RouteSnapshot {
	route.matchers = cloneSlice(route.matchers)
	return route
}

func cloneSlice[T any](source []T) []T {
	if len(source) == 0 {
		return []T{}
	}
	result := make([]T, len(source))
	copy(result, source)
	return result
}

func copySlice[S, T any](source []S, convert func(S) T) []T {
	if len(source) == 0 {
		return []T{}
	}
	result := make([]T, len(source))
	for index, value := range source {
		result[index] = convert(value)
	}
	return result
}
