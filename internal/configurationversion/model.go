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
}

// APIKeySettings describes API Key Provider metadata.
type APIKeySettings struct {
	Header    string `json:"header"`
	SecretRef string `json:"secretRef"`
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
	CreatedAt       time.Time              `json:"createdAt"`
	UpdatedAt       time.Time              `json:"updatedAt"`
}
