package configurationversion

import (
	"encoding/json"
	"net/http"
	"reflect"
	"testing"
)

func TestHandlerUpdateRoutingRoundTrip(t *testing.T) {
	router := newTestRouter(t, true)
	versionsPath := "/api/v1/workspaces/1/configurations/1/versions"
	performRequest(router, http.MethodPost, versionsPath)
	body := `{"routes":[{"id":"\u2003route.one\u2003","enabled":true,"priority":10,"matchers":[{"type":"principal-kind","value":" authenticated "},{"type":"message-type","value":" text "}],"handlerRef":" legacy "},{"id":"disabled","enabled":false,"priority":20,"matchers":[],"handlerRef":"future"}],"defaultHandlerRef":" legacy "}`

	response := performRequestWithBody(router, http.MethodPut, versionsPath+"/1/routing", body)
	assertStatus(t, response, http.StatusOK)
	assertContentType(t, response)
	updated := decodeVersion(t, response)
	if updated.Routing == nil || updated.Routing.DefaultHandlerRef != "legacy" || len(updated.Routing.Routes) != 2 {
		t.Fatalf("Routing = %#v", updated.Routing)
	}
	if route := updated.Routing.Routes[0]; route.ID != "route.one" || route.Matchers[0].Type != MatcherTypeMessageType || route.Matchers[1].Value != "authenticated" {
		t.Errorf("normalized Route = %#v", route)
	}
	if updated.Routing.Routes[1].Enabled || updated.Routing.Routes[1].HandlerRef != "future" {
		t.Errorf("disabled Route = %#v", updated.Routing.Routes[1])
	}

	list := performRequest(router, http.MethodGet, versionsPath)
	assertStatus(t, list, http.StatusOK)
	var versions []ConfigurationVersion
	decodeResponse(t, list, &versions)
	if len(versions) != 1 || !reflect.DeepEqual(versions[0].Routing, updated.Routing) {
		t.Fatalf("listed Routing = %#v, want %#v", versions, updated.Routing)
	}
}

func TestHandlerUpdateRoutingPresenceSemantics(t *testing.T) {
	tests := []struct {
		name        string
		body        string
		wantPresent bool
	}{
		{name: "absent", body: "null", wantPresent: false},
		{name: "present empty", body: `{}`, wantPresent: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			router := newTestRouter(t, true)
			versionsPath := "/api/v1/workspaces/1/configurations/1/versions"
			performRequest(router, http.MethodPost, versionsPath)
			if !test.wantPresent {
				setup := performRequestWithBody(router, http.MethodPut, versionsPath+"/1/routing", `{}`)
				assertStatus(t, setup, http.StatusOK)
			}
			response := performRequestWithBody(router, http.MethodPut, versionsPath+"/1/routing", test.body)
			assertStatus(t, response, http.StatusOK)

			var object map[string]json.RawMessage
			decodeResponse(t, response, &object)
			_, present := object["routing"]
			if present != test.wantPresent {
				t.Fatalf("routing present = %t, want %t; JSON = %#v", present, test.wantPresent, object)
			}
			if test.wantPresent {
				var routing RoutingSettings
				if err := json.Unmarshal(object["routing"], &routing); err != nil {
					t.Fatalf("decode routing: %v", err)
				}
				if len(routing.Routes) != 0 || routing.DefaultHandlerRef != "" {
					t.Fatalf("present-empty Routing = %#v", routing)
				}
			}
		})
	}
}

func TestHandlerUpdateRoutingValidationAndJSONErrors(t *testing.T) {
	tests := []struct {
		name     string
		body     string
		wantCode string
	}{
		{name: "enabled without matchers", body: `{"routes":[{"id":"route","enabled":true,"priority":1,"matchers":[],"handlerRef":"legacy"}]}`, wantCode: "validation_failed"},
		{name: "unicode identifier", body: `{"routes":[{"id":"route\u0416","enabled":false,"priority":1,"matchers":[],"handlerRef":"legacy"}]}`, wantCode: "validation_failed"},
		{name: "unknown matcher", body: `{"routes":[{"id":"route","enabled":true,"priority":1,"matchers":[{"type":"payload","value":"value"}],"handlerRef":"legacy"}]}`, wantCode: "validation_failed"},
		{name: "unknown field", body: `{"routes":[],"router":"future"}`, wantCode: "invalid_request"},
		{name: "unknown nested field", body: `{"routes":[{"id":"route","enabled":false,"priority":1,"matchers":[],"handlerRef":"legacy","runtime":true}]}`, wantCode: "invalid_request"},
		{name: "malformed", body: `{"routes":`, wantCode: "invalid_request"},
		{name: "empty", body: "", wantCode: "invalid_request"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			router := newTestRouter(t, true)
			path := "/api/v1/workspaces/1/configurations/1/versions"
			performRequest(router, http.MethodPost, path)
			response := performRequestWithBody(router, http.MethodPut, path+"/1/routing", test.body)
			assertStatus(t, response, http.StatusBadRequest)
			assertErrorCode(t, response, test.wantCode)
		})
	}
}

func TestHandlerUpdateRoutingLifecycleRestriction(t *testing.T) {
	for _, transition := range []string{"publish", "archive"} {
		t.Run(transition, func(t *testing.T) {
			router := newTestRouter(t, true)
			path := "/api/v1/workspaces/1/configurations/1/versions"
			performRequest(router, http.MethodPost, path)
			performRequest(router, http.MethodPost, path+"/1/"+transition)
			response := performRequestWithBody(router, http.MethodPut, path+"/1/routing", `{}`)
			assertStatus(t, response, http.StatusConflict)
			assertErrorCode(t, response, "version_not_editable")
		})
	}
}

func TestRoutingRequestDTOIsolation(t *testing.T) {
	request := &routingSettingsRequest{
		DefaultHandlerRef: "legacy",
		Routes: []routeRequest{{
			ID:         "route",
			Enabled:    true,
			Priority:   1,
			HandlerRef: "legacy",
			Matchers:   []matcherRequest{{Type: "message-type", Value: "text"}},
		}},
	}
	settings := request.settings()

	request.Routes[0].ID = "request.changed"
	request.Routes[0].Matchers[0].Value = "binary"
	if settings.Routes[0].ID != "route" || settings.Routes[0].Matchers[0].Value != "text" {
		t.Fatalf("settings share request DTO memory: %#v", settings)
	}
	settings.Routes[0].ID = "settings.changed"
	settings.Routes[0].Matchers[0].Value = "authenticated"
	if request.Routes[0].ID != "request.changed" || request.Routes[0].Matchers[0].Value != "binary" {
		t.Fatalf("request DTO changed through settings: %#v", request)
	}
}

func TestNilRoutingRequestDTO(t *testing.T) {
	var request *routingSettingsRequest
	if request.settings() != nil {
		t.Fatal("nil request settings must preserve absent Routing")
	}
}
