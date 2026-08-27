# 02 Architecture

> Split frontend + backend in one repo: Next.js browser app proxies `/api/*` to a Go Gin service that drives the Agora Conversational AI managed agent.

## High-Level Topology

```
   Browser tab                       Next.js (client/)              Go (server/)
   ┌────────────────────┐    fetch  ┌──────────────────────┐ HTTP  ┌─────────────────────┐
   │ LandingPage.tsx    │──────────▶│ /api/get_config      │──────▶│ GET  /get_config    │
   │ ConversationCmp.tsx│           │ /api/startAgent      │──────▶│ POST /startAgent    │
   │ AgoraVoiceAI       │◀──────────│ /api/stopAgent       │──────▶│ POST /stopAgent     │
   │ agora-rtc-react    │  JSON     │ next.config.ts       │       │ Gin + CORS          │
   │ agora-rtm          │           │  rewrites()          │       │ Agent Server SDK    │
   └─────┬──────┬───────┘           └──────────────────────┘       └──────────┬──────────┘
         │      │ RTM data                                                    │ HTTPS
         │ RTC  │                                                             ▼
         ▼      ▼                                                  Agora Conversational AI
   Agora media + RTM cloud                             (GeminiSTT → Gemini → MiniMaxTTS session)
```

## Voice Session Lifecycle

1. `LandingPage` calls `getConfig()` → `GET /api/get_config` → Gin returns `data: { app_id, token, uid, channel_name, agent_uid }`.
2. In parallel:
   - `startAgent(channel_name, Number(agent_uid), Number(uid))` → `POST /api/startAgent` → Gin starts the managed agent.
   - `new AgoraRTM.RTM(appId, uid).login({ token })` → `subscribe(channel_name)`.
3. `ConversationComponent` mounts inside a dynamic `AgoraRTCProvider` (client held in `useRef` for StrictMode safety) and:
   - `useJoin` joins the RTC channel.
   - `AgoraVoiceAI.init({ rtcEngine, rtmConfig: { rtmEngine: rtmClient } })` wires transcripts, state, metrics.
   - `subscribeMessage(channel_name)` opens the toolkit's RTM channel for events.
4. End: `stopAgent(agentId)` → `POST /api/stopAgent` → Gin stops the agent. `rtmClient.logout()` follows.
5. Renewal: on RTC `token-privilege-will-expire`, the client fetches `getConfig()` twice (once for RTC `client.uid`, once for the stored `agoraData.uid`) and renews RTC + RTM separately.

## How `/api/*` Reaches Gin

`client/next.config.ts`:

```ts
async rewrites() {
  const backendUrl = (process.env.AGENT_BACKEND_URL ?? '').trim();
  if (!backendUrl) return [];
  return [
    { source: '/api/get_config',   destination: `${backendUrl}/get_config` },
    { source: '/api/startAgent',   destination: `${backendUrl}/startAgent` },
    { source: '/api/stopAgent',    destination: `${backendUrl}/stopAgent` },
  ];
}
```

If `AGENT_BACKEND_URL` is unset/empty, **no rewrites register** — the client cannot reach the backend. `make dev` exports `AGENT_BACKEND_URL=http://localhost:8000` before starting Next.

`client/scripts/verify-api-contracts.ts` also asserts that **no `client/app/api/**/route.ts` files exist** — this client is rewrite-only.

## Go Backend Shape

`server/main.go`:

- `gin.Default()` with `cors.AllowAll`-style middleware.
- Routes: `GET /get_config`, `POST /startAgent`, `POST /stopAgent`.
- `loadEnvFiles` reads `.env.local` then `.env` from the current working directory.
- Handlers wrap responses as `gin.H{ "code": 0, "data": ..., "msg": "success" }`.

`server/agent.go`:

- `agentService` holds the `agentkit.AgoraClient` built with `option.AreaUS` plus app ID / app certificate credentials.
- `generateConfig(channel, uid)` treats missing, zero, and negative UIDs as "generate a usable UID", then calls `agentkit.GenerateConvoAIToken` with `ExpiresInHours(1)`.
- `start(...)` constructs `agentkit.NewAgent(...)` with GeminiSTT / Gemini / MiniMaxTTS providers sharing `GOOGLE_API_KEY`, `WithTurnDetectionConfig`, server-side RTM/error parameters, then `agent.CreateSession(...)` + `session.Start(ctx)`.
- `stop(agentId)` ends the session via the SDK.

## Managed Agent Defaults (server/agent.go)

| Stage | Vendor       | Config highlights                                            |
| ----- | ------------ | ------------------------------------------------------------ |
| STT   | `GeminiSTT` | `gemini-3.5-transcribe-live`, `LanguageCodes: ["en-US"]` |
| LLM   | `Gemini`    | `Model: "gemini-3.6-flash"`, `MaxHistory: 15`, `MaxOutputTokens: 1024`, `Temperature: 0.7`, `TopP: 0.95` |
| TTS   | `MiniMaxTTS` | `VoiceName: "en-US-Chirp3-HD-Charon"`, `LanguageCode: "en-US"`, sample rate 24000 |
| VAD   | Agora        | `SpeechThreshold: 0.5`, start mode `vad`, `InterruptDurationMs: 160`, `PrefixPaddingMs: 300`, end mode `vad`, `SilenceDurationMs: 480` |

Agent parameters: `DataChannel: "rtm"`, `EnableErrorMessage: true`, `EnableMetrics: true`. Advanced features: `EnableRtm: true`, `EnableTools: true`. Session options: `IdleTimeout: 30`, `ExpiresIn: ExpiresInHours(1)`, `EnableStringUID: false`.

## Why This Shape

- Backend in Go gives a tested production-style host for Agora's Conversational AI calls.
- Next.js rewrites hide backend placement from the browser — `/api/*` is the only URL the client knows.
- Single repo keeps client and backend changes reviewable together while preserving deploy separation.

## Related Deep Dives

- [Managed Agent Config](L2/managed_agent_config.md) — Full `agent.go` chain and tunable fields.
- [Session Lifecycle](L2/session_lifecycle.md) — Detailed client orchestration including renewal.
- [Verification Scripts](L2/verification_scripts.md) — How the contract harness asserts the proxy boundary.
