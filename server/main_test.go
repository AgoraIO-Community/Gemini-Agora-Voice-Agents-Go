package main

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

)

type fakeSession struct {
	stopErr   error
	stopCalls int
}

func (s *fakeSession) Stop(context.Context) error {
	s.stopCalls++
	return s.stopErr
}

func TestParseOptionalInt(t *testing.T) {
	value, err := parseOptionalInt("1234")
	if err != nil {
		t.Fatalf("parseOptionalInt returned error: %v", err)
	}
	if value != 1234 {
		t.Fatalf("expected 1234, got %d", value)
	}

	value, err = parseOptionalInt("")
	if err != nil {
		t.Fatalf("parseOptionalInt returned error for empty string: %v", err)
	}
	if value != 0 {
		t.Fatalf("expected 0 for empty string, got %d", value)
	}
}

func TestIsValidationError(t *testing.T) {
	if !isValidationError("channel_name is required and cannot be empty") {
		t.Fatal("expected channel_name validation error to be recognized")
	}
	if isValidationError("some other error") {
		t.Fatal("unexpected validation error match")
	}
}

func TestNewAgentServiceRequiresEnv(t *testing.T) {
	originalAppID := os.Getenv("AGORA_APP_ID")
	originalCertificate := os.Getenv("AGORA_APP_CERTIFICATE")
	t.Cleanup(func() {
		_ = os.Setenv("AGORA_APP_ID", originalAppID)
		_ = os.Setenv("AGORA_APP_CERTIFICATE", originalCertificate)
	})

	_ = os.Unsetenv("AGORA_APP_ID")
	_ = os.Unsetenv("AGORA_APP_CERTIFICATE")

	if _, err := newAgentService(); err == nil {
		t.Fatal("expected newAgentService to fail without AGORA_APP_ID and AGORA_APP_CERTIFICATE")
	}
}

func TestGenerateConfigPreservesRequestedValues(t *testing.T) {
	service := &agentService{
		appID:       strings.Repeat("a", 32),
		certificate: strings.Repeat("b", 32),
	}

	config, err := service.generateConfig("room-123", 4321)
	if err != nil {
		t.Fatalf("generateConfig returned error: %v", err)
	}
	if config.AppID != service.appID {
		t.Fatalf("expected app id %q, got %q", service.appID, config.AppID)
	}
	if config.UID != "4321" {
		t.Fatalf("expected uid 4321, got %q", config.UID)
	}
	if config.ChannelName != "room-123" {
		t.Fatalf("expected channel room-123, got %q", config.ChannelName)
	}
	if config.Token == "" {
		t.Fatal("expected non-empty token")
	}
	if config.AgentUID == "" {
		t.Fatal("expected non-empty agent uid")
	}
}

func TestGenerateConfigCreatesDefaults(t *testing.T) {
	service := &agentService{
		appID:       strings.Repeat("c", 32),
		certificate: strings.Repeat("d", 32),
	}

	config, err := service.generateConfig("", 0)
	if err != nil {
		t.Fatalf("generateConfig returned error: %v", err)
	}
	if config.UID == "" || config.UID == "0" {
		t.Fatalf("expected generated uid, got %q", config.UID)
	}
	if !strings.HasPrefix(config.ChannelName, "ai-conversation-") {
		t.Fatalf("expected generated channel name, got %q", config.ChannelName)
	}
	if config.Token == "" {
		t.Fatal("expected generated token")
	}
	if config.AgentUID == "" {
		t.Fatal("expected generated agent uid")
	}
}

func TestAgentServiceStartValidation(t *testing.T) {
	service := &agentService{}

	if _, err := service.start("", 1, 1); err == nil || err.Error() != "channel_name is required and cannot be empty" {
		t.Fatalf("expected channel validation error, got %v", err)
	}
	if _, err := service.start("room", 0, 1); err == nil || err.Error() != "agent_uid is required and cannot be empty" {
		t.Fatalf("expected agent uid validation error, got %v", err)
	}
	if _, err := service.start("room", 1, 0); err == nil || err.Error() != "user_uid is required and cannot be empty" {
		t.Fatalf("expected user uid validation error, got %v", err)
	}
}

func TestAgentServiceStopUsesSessionFirst(t *testing.T) {
	session := &fakeSession{}
	service := &agentService{
		sessions: map[string]sessionStopper{
			"agent-1": session,
		},
	}

	if err := service.stop("agent-1"); err != nil {
		t.Fatalf("stop returned error: %v", err)
	}
	if session.stopCalls != 1 {
		t.Fatalf("expected session stop to be called once, got %d", session.stopCalls)
	}
}

func TestAgentServiceStopReturnsSessionError(t *testing.T) {
	session := &fakeSession{stopErr: errors.New("session stale")}
	service := &agentService{
		sessions: map[string]sessionStopper{
			"agent-2": session,
		},
	}

	if err := service.stop("agent-2"); err == nil {
		t.Fatal("expected retained session stop error")
	}
	if session.stopCalls != 1 {
		t.Fatalf("expected session stop to be called once, got %d", session.stopCalls)
	}
}

func TestAgentServiceStopRequiresAgentID(t *testing.T) {
	service := &agentService{sessions: map[string]sessionStopper{}}
	if err := service.stop(""); err == nil || err.Error() != "agent_id is required and cannot be empty" {
		t.Fatalf("expected agent id validation error, got %v", err)
	}
}

func TestAgentServiceStopIsIdempotentForUnknownSession(t *testing.T) {
	service := &agentService{sessions: map[string]sessionStopper{}}

	if err := service.stop("agent-3"); err != nil {
		t.Fatalf("stop returned error: %v", err)
	}
}

func TestStartAgentRouteValidation(t *testing.T) {
	router := newRouter(&agentService{})
	request := httptest.NewRequest(http.MethodPost, "/startAgent", strings.NewReader(`{"channelName":"room"}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", response.Code)
	}
	var body map[string]any
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if body["detail"] != "agent_uid is required and cannot be empty" {
		t.Fatalf("unexpected response detail: %#v", body["detail"])
	}
}

func TestStopAgentRouteValidation(t *testing.T) {
	service := &agentService{sessions: map[string]sessionStopper{}}
	router := newRouter(service)
	request := httptest.NewRequest(http.MethodPost, "/stopAgent", strings.NewReader(`{}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", response.Code)
	}
	var body map[string]any
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if body["detail"] != "agent_id is required and cannot be empty" {
		t.Fatalf("unexpected response detail: %#v", body["detail"])
	}
}
