package runtimeconfig

import (
	"net"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/dsdred/universal-websocket-platform/internal/configurationversion"
	"github.com/dsdred/universal-websocket-platform/internal/runtimeconfigload"
)

var (
	secretReferencePattern  = regexp.MustCompile(`^[A-Za-z0-9/._-]+$`)
	handlerReferencePattern = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9._-]*$`)
	httpTokenPattern        = regexp.MustCompile(`^[A-Za-z0-9!#$%&'*+\-.^_` + "`" + `|~]+$`)
)

func stringsTrim(value string) string { return strings.TrimSpace(value) }

func validateHandoff(
	input runtimeconfigload.DetachedLoadResult,
	version configurationversion.ConfigurationVersion,
	diagnostics *diagnosticCollector,
) {
	requiredUint64(input.WorkspaceID(), diagnostics,
		"snapshot.provenance.workspace.required", "$.provenance.workspaceId", "workspace identity is required")
	requiredUint64(input.ConfigurationID(), diagnostics,
		"snapshot.provenance.configuration.required", "$.provenance.configurationId", "configuration identity is required")
	requiredUint64(input.ConfigurationVersionID(), diagnostics,
		"snapshot.provenance.configuration_version.required", "$.provenance.configurationVersionId", "configuration version identity is required")
	requiredUint32(input.ConfigurationVersionNumber(), diagnostics,
		"snapshot.provenance.configuration_number.required", "$.provenance.configurationVersionNumber", "configuration version number is required")
	if input.RuntimeInstanceID() == "" {
		diagnostics.add("snapshot.provenance.runtime_instance.required", "$.provenance.runtimeInstanceId", "runtime instance identity is required")
	}
	if input.LaunchAttemptID() == "" {
		diagnostics.add("snapshot.provenance.launch_attempt.required", "$.provenance.launchAttemptId", "launch attempt identity is required")
	}
	if input.SchemaIdentity() == "" {
		diagnostics.add("snapshot.schema.identity.required", "$.provenance.schemaIdentity", "configuration schema identity is required")
	} else if input.SchemaIdentity() != supportedSchemaIdentity {
		diagnostics.add("snapshot.schema.identity.unsupported", "$.provenance.schemaIdentity", "configuration schema identity is unsupported")
	}
	if input.SchemaVersion() == 0 {
		diagnostics.add("snapshot.schema.version.required", "$.provenance.schemaVersion", "configuration schema version is required")
	} else if input.SchemaVersion() != supportedSchemaVersion {
		diagnostics.add("snapshot.schema.version.unsupported", "$.provenance.schemaVersion", "configuration schema version is unsupported")
	}
	if !input.Published() {
		diagnostics.add("snapshot.handoff.not_published", "$.handoff.published", "detached load result is not published")
	}
	if version.State != configurationversion.Published {
		diagnostics.add("snapshot.configuration_version.state.not_published", "$.configurationVersion.state", "configuration version state is not published")
	}
	requiredUint64(version.ConfigurationID, diagnostics,
		"snapshot.configuration_version.configuration.required", "$.configurationVersion.configurationId", "configuration version payload configuration identity is required")
	requiredUint64(version.ID, diagnostics,
		"snapshot.configuration_version.identity.required", "$.configurationVersion.id", "configuration version payload identity is required")
	requiredUint32(version.Number, diagnostics,
		"snapshot.configuration_version.number.required", "$.configurationVersion.number", "configuration version payload number is required")
	if input.ConfigurationID() != 0 && version.ConfigurationID != 0 && input.ConfigurationID() != version.ConfigurationID {
		diagnostics.add("snapshot.configuration_version.configuration.inconsistent", "$.configurationVersion.configurationId", "configuration identity conflicts with provenance")
	}
	if input.ConfigurationVersionID() != 0 && version.ID != 0 && input.ConfigurationVersionID() != version.ID {
		diagnostics.add("snapshot.configuration_version.identity.inconsistent", "$.configurationVersion.id", "configuration version identity conflicts with provenance")
	}
	if input.ConfigurationVersionNumber() != 0 && version.Number != 0 && input.ConfigurationVersionNumber() != version.Number {
		diagnostics.add("snapshot.configuration_version.number.inconsistent", "$.configurationVersion.number", "configuration version number conflicts with provenance")
	}
}

func requiredUint64(value uint64, diagnostics *diagnosticCollector, code, location, message string) {
	if value == 0 {
		diagnostics.add(code, location, message)
	}
}

func requiredUint32(value uint32, diagnostics *diagnosticCollector, code, location, message string) {
	if value == 0 {
		diagnostics.add(code, location, message)
	}
}

func validateConfiguration(version configurationversion.ConfigurationVersion, diagnostics *diagnosticCollector) {
	validateListener(version.Listener, diagnostics)
	validateAuthentication(version.Authentication, diagnostics)
	validateRouting(version.Routing, diagnostics)
}

func validateListener(listener configurationversion.ListenerSettings, diagnostics *diagnosticCollector) {
	host := stringsTrim(listener.Host)
	switch {
	case host == "":
		diagnostics.add("snapshot.listener.host.required", "$.listener.host", "listener host is required")
	default:
		if utf8.RuneCountInString(host) > 255 {
			diagnostics.add("snapshot.listener.host.too_long", "$.listener.host", "listener host exceeds 255 characters")
		}
		if !validListenerHost(host) {
			diagnostics.add("snapshot.listener.host.invalid", "$.listener.host", "listener host is not a valid IP address or hostname")
		}
	}
	if listener.Port == 0 {
		diagnostics.add("snapshot.listener.port.out_of_range", "$.listener.port", "listener port must be between 1 and 65535")
	}

	minVersion := stringsTrim(listener.TLS.MinVersion)
	if minVersion == "" {
		diagnostics.add("snapshot.listener.tls.min_version.required", "$.listener.tls.minVersion", "TLS minimum version is required")
	} else if minVersion != "1.2" && minVersion != "1.3" {
		diagnostics.add("snapshot.listener.tls.min_version.unsupported", "$.listener.tls.minVersion", "TLS minimum version is unsupported")
	}
	validateTLSReference(
		listener.TLS.CertificateRef,
		listener.TLS.Enabled,
		"certificate",
		"$.listener.tls.certificateRef",
		diagnostics,
	)
	validateTLSReference(
		listener.TLS.PrivateKeyRef,
		listener.TLS.Enabled,
		"private_key",
		"$.listener.tls.privateKeyRef",
		diagnostics,
	)
	if listener.Timeouts.HandshakeSeconds < 1 || listener.Timeouts.HandshakeSeconds > 300 {
		diagnostics.add("snapshot.listener.timeout.handshake.out_of_range", "$.listener.timeouts.handshakeSeconds", "handshake timeout must be between 1 and 300 seconds")
	}
	if listener.Timeouts.ReadSeconds > 86400 {
		diagnostics.add("snapshot.listener.timeout.read.out_of_range", "$.listener.timeouts.readSeconds", "read timeout must be between 0 and 86400 seconds")
	}
	if listener.Timeouts.WriteSeconds < 1 || listener.Timeouts.WriteSeconds > 300 {
		diagnostics.add("snapshot.listener.timeout.write.out_of_range", "$.listener.timeouts.writeSeconds", "write timeout must be between 1 and 300 seconds")
	}
	if listener.Timeouts.IdleSeconds > 86400 {
		diagnostics.add("snapshot.listener.timeout.idle.out_of_range", "$.listener.timeouts.idleSeconds", "idle timeout must be between 0 and 86400 seconds")
	}
}

func validateTLSReference(raw string, required bool, name, location string, diagnostics *diagnosticCollector) {
	value := stringsTrim(raw)
	if value == "" {
		if required {
			diagnostics.add("snapshot.listener.tls."+name+".required", location, "TLS "+strings.ReplaceAll(name, "_", " ")+" reference is required when TLS is enabled")
		}
		return
	}
	if len(value) > 255 {
		diagnostics.add("snapshot.listener.tls."+name+".too_long", location, "TLS "+strings.ReplaceAll(name, "_", " ")+" reference exceeds 255 characters")
	}
	if !validSecretReference(value) {
		diagnostics.add("snapshot.listener.tls."+name+".invalid", location, "TLS "+strings.ReplaceAll(name, "_", " ")+" reference is invalid")
	}
}

func validListenerHost(value string) bool {
	if strings.Contains(value, ":") {
		return !strings.ContainsAny(value, "[]%") && net.ParseIP(value) != nil
	}
	if looksLikeIPv4(value) {
		parts := strings.Split(value, ".")
		if len(parts) != 4 {
			return false
		}
		for _, part := range parts {
			if part == "" || (len(part) > 1 && part[0] == '0') {
				return false
			}
			number, err := strconv.Atoi(part)
			if err != nil || number > 255 {
				return false
			}
		}
		return true
	}
	if len(value) > 255 || !isASCII(value) {
		return false
	}
	hostname := strings.TrimSuffix(value, ".")
	if hostname == "" {
		return false
	}
	for _, label := range strings.Split(hostname, ".") {
		if len(label) == 0 || len(label) > 63 || !isASCIIAlphanumeric(label[0]) || !isASCIIAlphanumeric(label[len(label)-1]) {
			return false
		}
		for index := 1; index < len(label)-1; index++ {
			if !isASCIIAlphanumeric(label[index]) && label[index] != '-' {
				return false
			}
		}
	}
	return true
}

func looksLikeIPv4(value string) bool {
	for _, character := range value {
		if character != '.' && (character < '0' || character > '9') {
			return false
		}
	}
	return strings.Contains(value, ".")
}

func isASCII(value string) bool {
	for index := 0; index < len(value); index++ {
		if value[index] > 127 {
			return false
		}
	}
	return true
}

func isASCIIAlphanumeric(value byte) bool {
	return value >= 'a' && value <= 'z' ||
		value >= 'A' && value <= 'Z' ||
		value >= '0' && value <= '9'
}

func validSecretReference(value string) bool {
	return len(value) >= 1 && len(value) <= 255 &&
		isASCII(value) &&
		secretReferencePattern.MatchString(value) &&
		!strings.HasPrefix(value, "/") &&
		!strings.HasSuffix(value, "/") &&
		!strings.Contains(value, "//") &&
		!strings.Contains(value, "://") &&
		!strings.Contains(value, "-----BEGIN")
}

func validateAuthentication(authentication configurationversion.AuthenticationSettings, diagnostics *diagnosticCollector) {
	if authentication.Enabled && len(authentication.Providers) == 0 {
		diagnostics.add("snapshot.authentication.providers.required", "$.authentication.providers", "authentication requires at least one provider")
	} else if authentication.Enabled {
		foundEnabled := false
		for _, provider := range authentication.Providers {
			foundEnabled = foundEnabled || provider.Enabled
		}
		if !foundEnabled {
			diagnostics.add("snapshot.authentication.enabled_provider.required", "$.authentication.providers", "authentication requires at least one enabled provider")
		}
	}

	names := make(map[string]int)
	priorities := make(map[uint32]int)
	for index, provider := range authentication.Providers {
		base := "$.authentication.providers[" + strconv.Itoa(index) + "]"
		name := stringsTrim(provider.Name)
		if name == "" {
			diagnostics.add("snapshot.authentication.provider.name.required", base+".name", "authentication provider name is required")
		} else {
			nameTooLong := utf8.RuneCountInString(name) > 255
			if nameTooLong {
				diagnostics.add("snapshot.authentication.provider.name.too_long", base+".name", "authentication provider name exceeds 255 characters")
			}
			if !nameTooLong {
				if _, exists := names[name]; exists {
					diagnostics.add("snapshot.authentication.provider.name.duplicate", base+".name", "authentication provider name is duplicated")
				} else {
					names[name] = index
				}
			}
		}
		if _, exists := priorities[provider.Priority]; exists {
			diagnostics.add("snapshot.authentication.provider.priority.duplicate", base+".priority", "authentication provider priority is duplicated")
		} else {
			priorities[provider.Priority] = index
		}

		providerType := string(provider.Type)
		if providerType == "" {
			diagnostics.add("snapshot.authentication.provider.type.required", base+".type", "authentication provider type is required")
			continue
		}
		if providerType != "api-key" && providerType != "basic" && providerType != "jwt" {
			diagnostics.add("snapshot.authentication.provider.type.unsupported", base+".type", "authentication provider type is unsupported")
			continue
		}
		validateProviderSettings(provider, providerType, base, diagnostics)
	}
}

func validateProviderSettings(provider configurationversion.AuthenticationProvider, providerType, base string, diagnostics *diagnosticCollector) {
	type branch struct {
		name    string
		present bool
	}
	branches := []branch{{"apiKey", provider.APIKey != nil}, {"basic", provider.Basic != nil}, {"jwt", provider.JWT != nil}}
	expected := map[string]string{"api-key": "apiKey", "basic": "basic", "jwt": "jwt"}[providerType]
	for _, candidate := range branches {
		if candidate.name == expected {
			if !candidate.present {
				diagnostics.add("snapshot.authentication.provider.settings.required", base+"."+candidate.name, "authentication provider settings are required for its type")
			}
			continue
		}
		if candidate.present {
			diagnostics.add("snapshot.authentication.provider.settings.forbidden", base+"."+candidate.name, "authentication provider settings are forbidden for its type")
		}
	}
	switch providerType {
	case "api-key":
		if provider.APIKey != nil {
			validateAPIKey(*provider.APIKey, base+".apiKey", diagnostics)
		}
	case "basic":
		if provider.Basic != nil {
			validateBasic(*provider.Basic, base+".basic", diagnostics)
		}
	case "jwt":
		if provider.JWT != nil {
			validateJWT(*provider.JWT, base+".jwt", diagnostics)
		}
	}
}

func validateAPIKey(settings configurationversion.APIKeySettings, base string, diagnostics *diagnosticCollector) {
	header := stringsTrim(settings.Header)
	if header == "" {
		diagnostics.add("snapshot.authentication.api_key.header.required", base+".header", "API key header is required")
	} else {
		if utf8.RuneCountInString(header) > 255 {
			diagnostics.add("snapshot.authentication.api_key.header.too_long", base+".header", "API key header exceeds 255 characters")
		}
		if !isASCII(header) || !httpTokenPattern.MatchString(header) {
			diagnostics.add("snapshot.authentication.api_key.header.invalid", base+".header", "API key header is invalid")
		}
	}
	validateProviderSecret(settings.SecretRef, "api_key", "API key", base+".secretRef", diagnostics)
}

func validateBasic(settings configurationversion.BasicSettings, base string, diagnostics *diagnosticCollector) {
	realm := stringsTrim(settings.Realm)
	if realm == "" {
		diagnostics.add("snapshot.authentication.basic.realm.required", base+".realm", "Basic realm is required")
	} else if utf8.RuneCountInString(realm) > 255 {
		diagnostics.add("snapshot.authentication.basic.realm.too_long", base+".realm", "Basic realm exceeds 255 characters")
	}
	validateProviderSecret(settings.SecretRef, "basic", "Basic", base+".secretRef", diagnostics)
}

func validateProviderSecret(raw, namespace, label, location string, diagnostics *diagnosticCollector) {
	value := stringsTrim(raw)
	if value == "" {
		diagnostics.add("snapshot.authentication."+namespace+".secret_ref.required", location, label+" secret reference is required")
		return
	}
	if len(value) > 255 {
		diagnostics.add("snapshot.authentication."+namespace+".secret_ref.too_long", location, label+" secret reference exceeds 255 characters")
	}
	if !validSecretReference(value) {
		diagnostics.add("snapshot.authentication."+namespace+".secret_ref.invalid", location, label+" secret reference is invalid")
	}
}

func validateJWT(settings configurationversion.JWTSettings, base string, diagnostics *diagnosticCollector) {
	if len(settings.SigningKeys) == 0 {
		diagnostics.add("snapshot.authentication.jwt.signing_keys.required", base+".signingKeys", "JWT requires at least one signing key")
	}
	keyNames := make(map[string]int)
	for index, key := range settings.SigningKeys {
		item := base + ".signingKeys[" + strconv.Itoa(index) + "]"
		name := stringsTrim(key.Name)
		if name == "" {
			diagnostics.add("snapshot.authentication.jwt.signing_key.name.required", item+".name", "JWT signing key name is required")
		} else if _, exists := keyNames[name]; exists {
			diagnostics.add("snapshot.authentication.jwt.signing_key.name.duplicate", item+".name", "JWT signing key name is duplicated")
		} else {
			keyNames[name] = index
		}
		validateProviderSecret(key.SecretRef, "jwt.signing_key", "JWT signing key", item+".secretRef", diagnostics)
	}

	if len(settings.AllowedAlgorithms) == 0 {
		diagnostics.add("snapshot.authentication.jwt.algorithms.required", base+".allowedAlgorithms", "JWT requires at least one allowed algorithm")
	}
	algorithms := make(map[string]int)
	for index, raw := range settings.AllowedAlgorithms {
		location := base + ".allowedAlgorithms[" + strconv.Itoa(index) + "]"
		value := string(raw)
		if !supportedJWTAlgorithm(value) {
			diagnostics.add("snapshot.authentication.jwt.algorithm.unsupported", location, "JWT algorithm is unsupported")
			continue
		}
		if _, exists := algorithms[value]; exists {
			diagnostics.add("snapshot.authentication.jwt.algorithm.duplicate", location, "JWT algorithm is duplicated")
		} else {
			algorithms[value] = index
		}
	}
	validateUniqueTrimmedStrings(settings.AllowedIssuers, base+".allowedIssuers", "issuer", diagnostics)
	validateUniqueTrimmedStrings(settings.AllowedAudiences, base+".allowedAudiences", "audience", diagnostics)
	claims := make(map[string]int)
	for index, claim := range settings.RequiredClaims {
		item := base + ".requiredClaims[" + strconv.Itoa(index) + "]"
		name := stringsTrim(claim.Name)
		if name == "" {
			diagnostics.add("snapshot.authentication.jwt.claim.name.required", item+".name", "JWT required claim name is required")
		} else if _, exists := claims[name]; exists {
			diagnostics.add("snapshot.authentication.jwt.claim.name.duplicate", item+".name", "JWT required claim name is duplicated")
		} else {
			claims[name] = index
		}
		if stringsTrim(claim.Value) == "" {
			diagnostics.add("snapshot.authentication.jwt.claim.value.required", item+".value", "JWT required claim value is required")
		}
	}
	if settings.ClockSkewSeconds > 300 {
		diagnostics.add("snapshot.authentication.jwt.clock_skew.out_of_range", base+".clockSkewSeconds", "JWT clock skew must be between 0 and 300 seconds")
	}
}

func supportedJWTAlgorithm(value string) bool {
	switch value {
	case "HS256", "HS384", "HS512", "RS256", "RS384", "RS512", "ES256", "ES384", "ES512", "PS256", "PS384", "PS512":
		return true
	default:
		return false
	}
}

func validateUniqueTrimmedStrings(values []string, base, field string, diagnostics *diagnosticCollector) {
	seen := make(map[string]int)
	for index, raw := range values {
		value := stringsTrim(raw)
		location := base + "[" + strconv.Itoa(index) + "]"
		if value == "" {
			diagnostics.add("snapshot.authentication.jwt."+field+".required", location, "JWT "+field+" must not be empty")
			continue
		}
		if _, exists := seen[value]; exists {
			diagnostics.add("snapshot.authentication.jwt."+field+".duplicate", location, "JWT "+field+" is duplicated")
		} else {
			seen[value] = index
		}
	}
}

func validateRouting(routing *configurationversion.RoutingSettings, diagnostics *diagnosticCollector) {
	if routing == nil {
		return
	}
	if len(routing.Routes) > 256 {
		diagnostics.add("snapshot.routing.routes.too_many", "$.routing.routes", "routing contains more than 256 routes")
	}
	validateDefaultHandler(routing.DefaultHandlerRef, diagnostics)
	routeIDs := make(map[string]int)
	priorities := make(map[uint32]int)
	matcherSets := make(map[string]int)
	for index, route := range routing.Routes {
		base := "$.routing.routes[" + strconv.Itoa(index) + "]"
		validateRouteScalar(route.ID, "id", base+".id", diagnostics)
		id := stringsTrim(route.ID)
		if validHandlerReference(id) {
			if _, exists := routeIDs[id]; exists {
				diagnostics.add("snapshot.routing.route.id.duplicate", base+".id", "route identity is duplicated")
			} else {
				routeIDs[id] = index
			}
		}
		if route.Priority == 0 {
			diagnostics.add("snapshot.routing.route.priority.out_of_range", base+".priority", "route priority must be positive")
		} else if _, exists := priorities[route.Priority]; exists {
			diagnostics.add("snapshot.routing.route.priority.duplicate", base+".priority", "route priority is duplicated")
		} else {
			priorities[route.Priority] = index
		}
		validateRouteScalar(route.HandlerRef, "handler", base+".handlerRef", diagnostics)
		validMatchers, key := validateMatchers(route, base, diagnostics)
		if route.Enabled && validMatchers {
			if _, exists := matcherSets[key]; exists {
				diagnostics.add("snapshot.routing.matcher_set.duplicate", base+".matchers", "enabled route duplicates an earlier normalized matcher set")
			} else {
				matcherSets[key] = index
			}
		}
	}
}

func validateDefaultHandler(raw string, diagnostics *diagnosticCollector) {
	if raw == "" {
		return
	}
	value := stringsTrim(raw)
	if value == "" {
		diagnostics.add("snapshot.routing.default_handler.required", "$.routing.defaultHandlerRef", "default handler reference is required when present")
		return
	}
	if len(value) > 128 {
		diagnostics.add("snapshot.routing.default_handler.too_long", "$.routing.defaultHandlerRef", "default handler reference exceeds 128 characters")
	}
	if !validHandlerReference(value) {
		diagnostics.add("snapshot.routing.default_handler.invalid", "$.routing.defaultHandlerRef", "default handler reference is invalid")
	}
}

func validateRouteScalar(raw, field, location string, diagnostics *diagnosticCollector) {
	value := stringsTrim(raw)
	label := "route identity"
	if field == "handler" {
		label = "route handler reference"
	}
	if value == "" {
		diagnostics.add("snapshot.routing.route."+field+".required", location, label+" is required")
		return
	}
	if len(value) > 128 {
		diagnostics.add("snapshot.routing.route."+field+".too_long", location, label+" exceeds 128 characters")
	}
	if !validHandlerReference(value) {
		diagnostics.add("snapshot.routing.route."+field+".invalid", location, label+" is invalid")
	}
}

func validHandlerReference(value string) bool {
	return len(value) >= 1 && len(value) <= 128 && isASCII(value) && handlerReferencePattern.MatchString(value)
}

func validateMatchers(route configurationversion.Route, base string, diagnostics *diagnosticCollector) (bool, string) {
	if len(route.Matchers) > 4 {
		diagnostics.add("snapshot.routing.route.matchers.too_many", base+".matchers", "route contains more than four matchers")
	}
	if route.Enabled && len(route.Matchers) == 0 {
		diagnostics.add("snapshot.routing.route.matchers.required", base+".matchers", "enabled route requires at least one matcher")
	}
	validCandidate := route.Enabled && len(route.Matchers) >= 1 && len(route.Matchers) <= 4
	types := make(map[string]int)
	pairs := make([]string, 0, len(route.Matchers))
	for index, matcher := range route.Matchers {
		item := base + ".matchers[" + strconv.Itoa(index) + "]"
		matcherType := stringsTrim(string(matcher.Type))
		value := stringsTrim(matcher.Value)
		typeSupported := supportedMatcherType(matcherType)
		if matcherType == "" {
			diagnostics.add("snapshot.routing.matcher.type.required", item+".type", "matcher type is required")
			validCandidate = false
		} else if !typeSupported {
			diagnostics.add("snapshot.routing.matcher.type.unsupported", item+".type", "matcher type is unsupported")
			validCandidate = false
		} else if _, exists := types[matcherType]; exists {
			diagnostics.add("snapshot.routing.matcher.type.duplicate", item+".type", "matcher type is duplicated within the route")
			validCandidate = false
		} else {
			types[matcherType] = index
		}
		if value == "" {
			diagnostics.add("snapshot.routing.matcher.value.required", item+".value", "matcher value is required")
			validCandidate = false
		} else if typeSupported && !supportedMatcherValue(matcherType, value) {
			diagnostics.add("snapshot.routing.matcher.value.unsupported", item+".value", "matcher value is unsupported for its type")
			validCandidate = false
		} else if typeSupported {
			pairs = append(pairs,
				strconv.Itoa(len(matcherType))+":"+matcherType+
					strconv.Itoa(len(value))+":"+value,
			)
		}
	}
	sort.Strings(pairs)
	return validCandidate, strings.Join(pairs, "")
}

func supportedMatcherType(value string) bool {
	switch value {
	case "message-type", "principal-kind", "authentication-type", "authentication-provider":
		return true
	default:
		return false
	}
}

func supportedMatcherValue(matcherType, value string) bool {
	switch matcherType {
	case "message-type":
		return value == "text" || value == "binary"
	case "principal-kind":
		return value == "authenticated" || value == "anonymous"
	case "authentication-type":
		return value == "jwt" || value == "api-key" || value == "basic"
	case "authentication-provider":
		return value != ""
	default:
		return false
	}
}

func sortMatchers(values []MatcherSnapshot) {
	sort.Slice(values, func(i, j int) bool {
		if values[i].matcherType != values[j].matcherType {
			return values[i].matcherType < values[j].matcherType
		}
		return values[i].value < values[j].value
	})
}
