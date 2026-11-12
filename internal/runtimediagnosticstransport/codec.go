// Package runtimediagnosticstransport provides isolated diagnostics HTTP projection.
// It supplies no listener, authentication, production authority or Runtime wiring.
package runtimediagnosticstransport

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"strconv"
)

var errInvalid = errors.New("invalid diagnostics projection")

// Request carries canonical caller-known routing and a construction-owned capability.
type Request[P comparable] struct {
	Principal                                  P
	WorkspaceID, ConfigurationID               uint64
	RuntimeInstanceID, ExpectedLaunchAttemptID string
	Budget                                     uint8
}

// object requires exactly the named, non-null fields, rejecting duplicate keys.
func object(data []byte, names ...string) (map[string]json.RawMessage, error) {
	d := json.NewDecoder(bytes.NewReader(data))
	token, err := d.Token()
	if err != nil || token != json.Delim('{') {
		return nil, errInvalid
	}
	fields := make(map[string]json.RawMessage, len(names))
	ok := false
	defer func() {
		if !ok {
			clearFields(fields)
		}
	}()
	for d.More() {
		key, err := d.Token()
		if err != nil {
			return nil, errInvalid
		}
		name, valid := key.(string)
		allowed := false
		for _, expected := range names {
			allowed = allowed || name == expected
		}
		if !valid || !allowed || fields[name] != nil {
			return nil, errInvalid
		}
		var value json.RawMessage
		if d.Decode(&value) != nil {
			return nil, errInvalid
		}
		fields[name] = value
		if bytes.Equal(bytes.TrimSpace(value), []byte("null")) {
			return nil, errInvalid
		}
	}
	if token, err = d.Token(); err != nil || token != json.Delim('}') || len(fields) != len(names) {
		return nil, errInvalid
	}
	if _, err = d.Token(); err != io.EOF {
		return nil, errInvalid
	}
	ok = true
	return fields, nil
}

func clearFields(fields map[string]json.RawMessage) {
	for _, value := range fields {
		clear(value)
	}
	clear(fields)
}

func decimal(value string) (uint64, bool) {
	if len(value) == 0 || len(value) > 20 || value[0] < '1' || value[0] > '9' {
		return 0, false
	}
	for _, c := range value {
		if c < '0' || c > '9' {
			return 0, false
		}
	}
	n, err := strconv.ParseUint(value, 10, 64)
	return n, err == nil
}

func opaque(value string) bool {
	if len(value) < 1 || len(value) > 128 {
		return false
	}
	for _, c := range value {
		if !(c >= 'A' && c <= 'Z' || c >= 'a' && c <= 'z' || c >= '0' && c <= '9' || c == '_' || c == '-') {
			return false
		}
	}
	return true
}

func decode[P comparable](data []byte, principal P) (Request[P], error) {
	r := Request[P]{Principal: principal}
	f, err := object(data, "workspaceId", "configurationId", "runtimeInstanceId", "expectedLaunchAttemptId", "budget")
	if err != nil {
		return Request[P]{}, errInvalid
	}
	defer clearFields(f)
	var workspace, configuration string
	if json.Unmarshal(f["workspaceId"], &workspace) != nil || json.Unmarshal(f["configurationId"], &configuration) != nil ||
		json.Unmarshal(f["runtimeInstanceId"], &r.RuntimeInstanceID) != nil || json.Unmarshal(f["expectedLaunchAttemptId"], &r.ExpectedLaunchAttemptID) != nil ||
		json.Unmarshal(f["budget"], &r.Budget) != nil {
		return Request[P]{}, errInvalid
	}
	var valid bool
	if r.WorkspaceID, valid = decimal(workspace); !valid {
		return Request[P]{}, errInvalid
	}
	if r.ConfigurationID, valid = decimal(configuration); !valid || !opaque(r.RuntimeInstanceID) || !opaque(r.ExpectedLaunchAttemptID) || r.Budget < 1 || r.Budget > 64 {
		return Request[P]{}, errInvalid
	}
	return r, nil
}
