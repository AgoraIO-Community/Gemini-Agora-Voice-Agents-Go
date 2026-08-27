package main

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/AgoraIO/agora-agents-go/v2/agentkit"
	"github.com/AgoraIO/agora-agents-go/v2/option"
)

// Captures the start request this demo actually sends.
type wireRecorder struct {
	req  *http.Request
	body []byte
}

func (w *wireRecorder) Do(req *http.Request) (*http.Response, error) {
	b, _ := io.ReadAll(req.Body)
	w.req, w.body = req, b
	return &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(strings.NewReader(`{"agent_id":"a1"}`)),
		Header:     make(http.Header),
	}, nil
}

func TestStandardWireShape(t *testing.T) {
	if os.Getenv("WIRE_CHECK") == "" {
		t.Skip("set WIRE_CHECK=1")
	}
	os.Setenv("AGORA_APP_ID", "81190c52971d4004b7244bdcd93e2f34")
	os.Setenv("AGORA_APP_CERTIFICATE", "0123456789abcdef0123456789abcdef")
	os.Setenv("GOOGLE_API_KEY", "FAKEKEY")

	svc, err := newAgentService()
	if err != nil {
		t.Fatal(err)
	}
	// Same package: swap in a standard client whose transport records the request.
	rec := &wireRecorder{}
	svc.sessionClient = agentkit.NewAgoraClient(agentkit.AgoraClientOptions{
		Area:           option.AreaUS,
		AppID:          os.Getenv("AGORA_APP_ID"),
		AppCertificate: os.Getenv("AGORA_APP_CERTIFICATE"),
		HTTPClient:     rec,
	})

	if _, err := svc.start("ch", 123456, 100); err != nil {
		t.Logf("start returned: %v", err)
	}
	if rec.req == nil {
		t.Fatal("no request captured")
	}
	var payload map[string]interface{}
	if err := json.Unmarshal(rec.body, &payload); err != nil {
		t.Fatalf("decode request body: %v", err)
	}
	props, ok := payload["properties"].(map[string]interface{})
	if !ok {
		t.Fatal("request body has no properties object")
	}

	if got := rec.req.Header.Get("agora-feature"); got != "" {
		t.Errorf("agora-feature = %q, want no preview gate", got)
	}
	if host := rec.req.URL.Host; host != "api.agora.io" {
		t.Errorf("request host = %q, want the standard gateway", host)
	}
	if token, _ := props["token"].(string); token == "" {
		t.Error("token is empty; the agent cannot join without one")
	}

	asr, ok := props["asr"].(map[string]interface{})
	if !ok {
		t.Fatal("no asr config in the request")
	}
	if got, _ := asr["vendor"].(string); got != "gemini" {
		t.Errorf("asr.vendor = %v, want gemini", asr["vendor"])
	}
	// The top-level asr.language comes from turn detection, not the vendor.
	if got, _ := asr["language"].(string); got != "en-US" {
		t.Errorf("asr.language = %v, want en-US", asr["language"])
	}

	params, ok := asr["params"].(map[string]interface{})
	if !ok {
		t.Fatal("no asr.params in the request")
	}
	if got, _ := params["model"].(string); got != "gemini-3.5-transcribe-live" {
		t.Errorf("asr.params.model = %v, want gemini-3.5-transcribe-live", params["model"])
	}
	codes, _ := params["language_codes"].([]interface{})
	if len(codes) != 1 || codes[0] != "en-US" {
		t.Errorf("asr.params.language_codes = %v, want [en-US]", params["language_codes"])
	}
	if vocab, _ := params["custom_vocabulary"].([]interface{}); len(vocab) != 2 {
		t.Errorf("asr.params.custom_vocabulary = %v, want 2 terms", params["custom_vocabulary"])
	}
	if timestamp, found := params["word_timestamp"]; !found || timestamp != false {
		t.Errorf("asr.params.word_timestamp = %v, want explicit false", timestamp)
	}
	// This vendor has no `language` option — a value here would be a no-op the
	// builder overwrites.
	if _, found := params["language"]; found {
		t.Error("asr.params.language must not be set by this vendor")
	}

	if _, found := props["mllm"]; found {
		t.Error("a cascading pipeline must not send an mllm stage")
	}
	// The vendor sets max_history; the agent level must not also set one, or the
	// agent-level value silently loses.
	llm, ok := props["llm"].(map[string]interface{})
	if !ok {
		t.Fatal("no llm config in the request")
	}
	if got, _ := llm["max_history"].(float64); got != 15 {
		t.Errorf("llm.max_history = %v, want 15 (the vendor value)", llm["max_history"])
	}
}
