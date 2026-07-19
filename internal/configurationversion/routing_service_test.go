package configurationversion

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestServiceUpdateRouting(t *testing.T) {
	repository := NewMemoryConfigurationVersionRepository()
	service := NewService(repository, configurationCheckerStub{exists: true}, time.Now)
	created, err := service.Create(context.Background(), 1, 1)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	input := &RoutingSettings{
		DefaultHandlerRef: "\u2003legacy\u2003",
		Routes: []Route{
			{
				ID:         "\u2003route.one\u2003",
				Enabled:    true,
				Priority:   10,
				HandlerRef: "\u2003legacy\u2003",
				Matchers: []Matcher{
					{Type: MatcherTypePrincipalKind, Value: " authenticated "},
					{Type: MatcherTypeMessageType, Value: " text "},
				},
			},
			{ID: "disabled", Priority: 20, HandlerRef: "future"},
		},
	}
	updated, err := service.UpdateRouting(context.Background(), 1, 1, created.ID, input)
	if err != nil {
		t.Fatalf("UpdateRouting() error = %v", err)
	}
	if updated.Routing == nil || updated.Routing.DefaultHandlerRef != "legacy" {
		t.Fatalf("Routing = %#v", updated.Routing)
	}
	if route := updated.Routing.Routes[0]; route.ID != "route.one" || route.HandlerRef != "legacy" || route.Matchers[0].Type != MatcherTypeMessageType {
		t.Errorf("normalized Route = %#v", route)
	}
	if updated.Routing.Routes[1].Enabled || updated.Routing.Routes[1].HandlerRef != "future" {
		t.Errorf("disabled Route = %#v", updated.Routing.Routes[1])
	}

	input.Routes[0].ID = "caller.changed"
	input.Routes[0].Matchers[0].Value = "anonymous"
	updated.Routing.Routes[0].ID = "result.changed"
	stored, err := repository.Get(created.ID)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if stored.Routing.Routes[0].ID != "route.one" || stored.Routing.Routes[0].Matchers[1].Value != "authenticated" {
		t.Fatalf("stored Routing shares caller memory: %#v", stored.Routing)
	}
}

func TestServiceUpdateRoutingPresenceSemantics(t *testing.T) {
	service := NewService(NewMemoryConfigurationVersionRepository(), configurationCheckerStub{exists: true}, time.Now)
	created, err := service.Create(context.Background(), 1, 1)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	initial, err := service.UpdateRouting(context.Background(), 1, 1, created.ID, &RoutingSettings{})
	if err != nil || initial.Routing == nil {
		t.Fatalf("UpdateRouting(empty) = (%#v, %v), want present Routing", initial.Routing, err)
	}

	absent, err := service.UpdateRouting(context.Background(), 1, 1, created.ID, nil)
	if err != nil {
		t.Fatalf("UpdateRouting(nil) error = %v", err)
	}
	if absent.Routing != nil {
		t.Fatalf("UpdateRouting(nil) Routing = %#v, want nil", absent.Routing)
	}

	present, err := service.UpdateRouting(context.Background(), 1, 1, created.ID, &RoutingSettings{})
	if err != nil {
		t.Fatalf("UpdateRouting(empty) error = %v", err)
	}
	if present.Routing == nil {
		t.Fatal("UpdateRouting(empty) lost present Routing")
	}
}

func TestServiceUpdateRoutingLifecycleRestriction(t *testing.T) {
	for _, state := range []VersionState{Validated, Published, Archived} {
		t.Run(string(state), func(t *testing.T) {
			repository := NewMemoryConfigurationVersionRepository()
			version, err := repository.Create(ConfigurationVersion{ConfigurationID: 1, Number: 1, State: state})
			if err != nil {
				t.Fatalf("Create() error = %v", err)
			}
			service := NewService(repository, configurationCheckerStub{exists: true}, time.Now)
			_, err = service.UpdateRouting(context.Background(), 1, 1, version.ID, &RoutingSettings{})
			if !errors.Is(err, ErrVersionNotEditable) {
				t.Fatalf("UpdateRouting(%s) error = %v, want ErrVersionNotEditable", state, err)
			}
		})
	}
}

func TestServiceUpdateRoutingRejectsBeforePersistence(t *testing.T) {
	repository := NewMemoryConfigurationVersionRepository()
	service := NewService(repository, configurationCheckerStub{exists: true}, time.Now)
	created, err := service.Create(context.Background(), 1, 1)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	invalid := &RoutingSettings{Routes: []Route{{ID: "route", Enabled: true, Priority: 1, HandlerRef: "legacy"}}}
	if _, err := service.UpdateRouting(context.Background(), 1, 1, created.ID, invalid); err == nil {
		t.Fatal("UpdateRouting(invalid) error = nil")
	}
	stored, err := repository.Get(created.ID)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if stored.Routing != nil {
		t.Fatalf("invalid Routing persisted as %#v", stored.Routing)
	}
}

func TestServicePublishRejectsMalformedRoutingBeforeLifecycleMutation(t *testing.T) {
	repository := NewMemoryConfigurationVersionRepository()
	current, err := repository.Create(ConfigurationVersion{ConfigurationID: 1, Number: 1, State: Published})
	if err != nil {
		t.Fatalf("Create(current) error = %v", err)
	}
	target, err := repository.Create(ConfigurationVersion{
		ConfigurationID: 1,
		Number:          2,
		State:           Draft,
		Routing:         &RoutingSettings{Routes: []Route{{ID: "invalid", Enabled: true, Priority: 1, HandlerRef: "legacy"}}},
	})
	if err != nil {
		t.Fatalf("Create(target) error = %v", err)
	}
	service := NewService(repository, configurationCheckerStub{exists: true}, time.Now)

	if _, err := service.Publish(context.Background(), 1, 1, target.ID); err == nil {
		t.Fatal("Publish(malformed Routing) error = nil")
	}
	storedCurrent, _ := repository.Get(current.ID)
	storedTarget, _ := repository.Get(target.ID)
	if storedCurrent.State != Published || storedTarget.State != Draft {
		t.Fatalf("states after rejected Publish = [%s %s], want [Published Draft]", storedCurrent.State, storedTarget.State)
	}
}

func TestServicePublishNormalizesProgrammaticRouting(t *testing.T) {
	repository := NewMemoryConfigurationVersionRepository()
	target, err := repository.Create(ConfigurationVersion{
		ConfigurationID: 1,
		Number:          1,
		State:           Draft,
		Routing: &RoutingSettings{Routes: []Route{{
			ID:         "\u2003route\u2003",
			Enabled:    true,
			Priority:   1,
			HandlerRef: " legacy ",
			Matchers:   []Matcher{{Type: MatcherTypeMessageType, Value: " text "}},
		}}},
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	service := NewService(repository, configurationCheckerStub{exists: true}, time.Now)

	published, err := service.Publish(context.Background(), 1, 1, target.ID)
	if err != nil {
		t.Fatalf("Publish() error = %v", err)
	}
	if published.Routing.Routes[0].ID != "route" || published.Routing.Routes[0].Matchers[0].Value != "text" {
		t.Fatalf("published Routing = %#v", published.Routing)
	}
	stored, _ := repository.Get(target.ID)
	if stored.Routing.Routes[0].ID != "route" || stored.State != Published {
		t.Fatalf("stored published Version = %#v", stored)
	}
}
