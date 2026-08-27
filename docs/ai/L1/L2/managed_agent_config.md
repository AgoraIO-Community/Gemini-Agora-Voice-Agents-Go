> **When to Read This:** Load this document when you are changing the agent's prompt, voice, VAD behavior, model selection, session options, or wiring a bring-your-own-key (BYOK) provider on the Go side.

# Managed Agent Config

## Where It Lives

All managed agent configuration is in `server/agent.go`. The browser sends `{ channelName, rtcUid, userUid }` to `POST /startAgent`, which calls `agentService.start(...)`. That function builds an SDK-driven agent and starts a session.

## The Agent Builder Chain

The standard `NewAgoraClient` starts the GeminiSTT provider session. All three provider stages reuse `GOOGLE_API_KEY`; Gemini LLM and MiniMaxTTS are Google-backed rather than Agora-managed.

```go
agent := agentkit.NewAgent(
    s.sessionClient,
    agentkit.WithInstructions(adaPrompt),
    agentkit.WithGreeting(s.greeting),
    agentkit.WithFailureMessage("Please wait a moment."),
    agentkit.WithTurnDetectionConfig(&agentkit.TurnDetectionConfig{
        Language: Agora.AsrLanguageEnUs.Ptr(),
        Config: &agentkit.TurnDetectionNestedConfig{/* VAD settings */},
    }),
    agentkit.WithAdvancedFeatures(&agentkit.AdvancedFeatures{
        EnableRtm: &enableRTM, EnableTools: &enableTools,
    }),
    agentkit.WithParameters(&agentkit.SessionParams{
        DataChannel: &dataChannel, EnableErrorMessage: &enableErrorMessage,
        EnableMetrics: &enableMetrics,
    }),
    agentkit.WithAudioScenario(agentkit.ParametersAudioScenario("chorus")),
).
    WithLlm(vendors.NewGemini(vendors.GeminiOptions{
        APIKey: s.googleAPIKey, Model: "gemini-3.6-flash",
        MaxHistory: intPtr(15), MaxOutputTokens: intPtr(1024),
    })).
    WithStt(vendors.NewGeminiSTT(vendors.GeminiSTTOptions{
        APIKey: s.googleAPIKey, LanguageCodes: []string{"en-US"},
        CustomVocabulary: []string{"Agora", "Gemini"},
        WordTimestamp: Agora.Bool(false),
    })).
    WithTts(vendors.NewMiniMaxTTS(vendors.MiniMaxTTSOptions{
        Key: s.googleAPIKey, VoiceName: "en-US-Chirp3-HD-Charon",
        LanguageCode: "en-US", SampleRate: sampleRatePtr(vendors.SampleRate24kHz),
    }))
```

## Session Options

`CreateSession` takes the `AgoraClient` as its first argument, then the session options struct. `session.Start(ctx)` is called separately and returns the `agentID`.

```go
session := agent.CreateSession(s.sessionClient, agentkit.CreateSessionOptions{
    Channel:         channelName,
    AgentUID:        strconv.Itoa(agentUID),
    RemoteUIDs:      []string{strconv.Itoa(userUID)},
    EnableStringUID: &enableStringUID,  // false
    IdleTimeout:     &idleTimeout,      // 30
    ExpiresIn:       expiresIn,
})

agentID, err := session.Start(ctx)
```

| Option            | Effect                                                                      |
| ----------------- | --------------------------------------------------------------------------- |
| `Channel`         | The RTC channel the agent joins.                                            |
| `AgentUID`        | UID the agent occupies; the backend returns this as `agent_uid` for the web client. |
| `RemoteUIDs`      | Restricts the agent to the requester's UID; prevents cross-channel sniping. |
| `EnableStringUID` | `false` keeps UIDs numeric for both RTC and RTM.                            |
| `IdleTimeout`     | Seconds of silence before the session ends.                                 |
| `ExpiresIn`       | Hard ceiling on session length, mirrors the 1-hour token.                   |

`DataChannel`, `EnableRtm`, `EnableTools`, and `EnableErrorMessage` are **not** session options — they live on the agent via `agentkit.WithAdvancedFeatures` and `agentkit.WithParameters`.

## Editing Each Surface

### Change the prompt

Edit the `adaPrompt` string constant at the top of `agent.go`. Keep it concise — long prompts amplify LLM latency.

### Change the greeting

Set `AGENT_GREETING` in `server/.env.local`, or edit the fallback default in `newAgentService`.

### Change VAD

Edit the `TurnDetectionConfig` passed to `agentkit.WithTurnDetectionConfig`. The struct uses a `Config` wrapper field with nested `StartOfSpeech` and `EndOfSpeech` sub-structs — do **not** use a flat struct shape. Tuning notes:

- `SpeechThreshold` (on `TurnDetectionNestedConfig`) — VAD activation sensitivity (0.0–1.0). Lower values trigger on quieter audio.
- `InterruptDurationMs` (on `StartOfSpeechVadConfig`) — minimum user speech before the agent yields. Lower = more responsive interruptions.
- `PrefixPaddingMs` (on `StartOfSpeechVadConfig`) — audio captured before VAD triggers; raise this if early phonemes are clipped.
- `SilenceDurationMs` (on `EndOfSpeechVadConfig`) — silence after speech before VAD ends the turn. Raise this for slow speakers.

### Swap STT / LLM / TTS

The defaults are `NewGeminiSTT` (`gemini-3.5-transcribe-live`), `NewGemini` (`gemini-3.6-flash`), and `NewMiniMaxTTS` (`en-US-Chirp3-HD-Charon`, `en-US`, 24000 Hz). They all reuse `GOOGLE_API_KEY`. Replace a constructor only for an intentional provider change, and document any new credential in `server/.env.example`.

Gemini custom vocabulary and word timestamps are incompatible. Keep `WordTimestamp` explicitly set to a false pointer whenever `CustomVocabulary` is configured; the SDK panics during configuration if it is true.

### Session-Level Tuning

- Lower `IdleTimeout` (e.g. 15) for short demos. It is a pointer field (`&idleTimeout`).
- `DataChannel`, `EnableErrorMessage`, and `EnableMetrics` are set via `agentkit.WithParameters(&agentkit.SessionParams{...})` on the agent, not in `CreateSessionOptions`.
- Switch `DataChannel` to `"sct"` only if you are not relying on RTM transcripts or the current RTM event handlers.

## Response Contract

`startAgent` returns:

```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "agent_id": "string",
    "channel_name": "string",
    "status": "started"
  }
}
```

The client stores `agent_id` in `agoraData` and later passes it to `/api/stopAgent`.

Stop is best-effort from the browser perspective: `LandingPage` catches and logs stop failures. On the server, `agentService.stop` removes the mutex-protected retained session and calls `session.Stop(ctx)`. Unknown or repeated IDs are idempotent no-ops; standalone `StopAgent` is never used.

## Verification

`server/main_test.go` exercises:

- `newAgentService` env requirement (returns error when required Agora or Google credentials are missing).
- `generateConfig` UID generation behavior.
- `start` / `stop` validation behavior.

After editing `agent.go`, run `make fmt && make verify-backend`.

## Failure Modes

| Symptom                                              | Cause                                                                  |
| ---------------------------------------------------- | ---------------------------------------------------------------------- |
| `500 Service not properly configured`                | Missing `AGORA_APP_ID`, `AGORA_APP_CERTIFICATE`, or `GOOGLE_API_KEY` in `server/.env.local`. |
| Agent joins but never speaks                         | `GOOGLE_API_KEY` missing/invalid or MiniMaxTTS voice settings changed incorrectly. |
| Agent state stuck in `IDLE`                          | `EnableRtm` is `false` in `WithAdvancedFeatures`, or RTM subscribed before login. |
| Metrics events missing                               | `EnableMetrics` is not true, `DataChannel` is not `"rtm"`, or the service did not emit metrics. |
| Build fails: `unknown field`                         | SDK version mismatch; run `go mod tidy` and check `server/go.mod`.      |

## Parity With the Python Quickstart

The sibling [`agent-quickstart-python`](https://github.com/AgoraIO-Conversational-AI/agent-quickstart-python) repo builds the same provider pipeline in `server/src/agent.py`. When you change a shared agent field here:

- Mirror the change in the Python repo's `agent.py` if the field is part of the family-wide product surface (model, voice, VAD, session options, advanced features, parameters).
- Keep the field name identical wherever the SDK exposes it under the same name (e.g. `data_channel`, `idle_timeout`, `expires_in`, `enable_metrics`).
- Cosmetic differences (snake_case Python vs CamelCase Go) are expected; semantic differences are not.

There is no automated cross-repo check today — review by diffing `server/agent.go` against `server/src/agent.py` before merging.

## See Also

- [Back to Architecture](../02_architecture.md)
- [Back to Workflows](../05_workflows.md)
- [Session Lifecycle](session_lifecycle.md)
