> **When to Read This:** Load this when an agent needs to implement a baseline Go-backed Agora Conversational AI quickstart in a new repo from this recipe.

# From-Scratch Bootstrap

## Baseline Rule

This repo is the Go-backed Agora quickstart baseline for the recipe. Do not implement an Agora ConvoAI quickstart from memory. Start from this repo's source and docs, then adapt only after the token, Gin start/stop, RTC, RTM, and transcript flow is understood.

Why: provider schemas, SDK builder fields, token behavior, and RTM event details can drift. The source files in this repo are the implementation reference for this recipe version.

## Implementation Map

| Need | Read First | Deep Detail | Source Reference |
| --- | --- | --- | --- |
| Project setup, commands, env vars | [../01_setup.md](../01_setup.md) | none | `Makefile`, `package.json`, `server/.env.example`, `client/.env.local.example` |
| End-to-end architecture and data flow | [../02_architecture.md](../02_architecture.md) | [session_lifecycle.md](session_lifecycle.md) | `client/src/components/LandingPage.tsx`, `client/src/components/ConversationComponent.tsx`, `server/main.go` |
| File/module responsibilities | [../03_code_map.md](../03_code_map.md) | none | `client/`, `server/`, `client/scripts/` |
| API payloads and response shapes | [../06_interfaces.md](../06_interfaces.md) | [verification_scripts.md](verification_scripts.md) | `server/main.go`, `client/src/services/api.ts`, `client/next.config.ts` |
| Managed agent configuration | [../05_workflows.md](../05_workflows.md) | [managed_agent_config.md](managed_agent_config.md) | `server/agent.go` |
| Browser RTC/RTM/toolkit lifecycle | [../04_conventions.md](../04_conventions.md), [../07_gotchas.md](../07_gotchas.md) | [session_lifecycle.md](session_lifecycle.md) | `LandingPage.tsx`, `ConversationComponent.tsx`, `client/src/lib/conversation.ts` |
| Security and secret boundaries | [../08_security.md](../08_security.md) | none | `server/main.go`, `server/agent.go`, `client/next.config.ts` |
| Validation expectations | [../05_workflows.md](../05_workflows.md) | [verification_scripts.md](verification_scripts.md) | `client/scripts/*.ts`, `server/main_test.go`, `server/cmd/fake-server/main.go` |

## Minimum Implementation Checklist

Implement these pieces in order:

1. Create a pnpm workspace with `client` as a workspace member and root Make targets that orchestrate backend, frontend, setup, doctor, verify, and clean tasks.
2. Create `server/` as a Go 1.23+ module with Gin, gin-contrib CORS, godotenv, and `github.com/AgoraIO/agora-agents-go/v2`.
3. Add `server/.env.example` with `AGORA_APP_ID`, `AGORA_APP_CERTIFICATE`, required `GOOGLE_API_KEY`, optional `AGENT_GREETING`, and optional `PORT`.
4. Implement `server/agent.go` with `agentService` that reads env once, constructs standard `NewAgoraClient`, generates one-hour RTC+RTM ConvoAI tokens, builds GeminiSTT (`gemini-3.5-transcribe-live`), Gemini (`gemini-3.6-flash`), and MiniMaxTTS (`en-US-Chirp3-HD-Charon`, `en-US`, 24000 Hz), reuses `GOOGLE_API_KEY` for every provider, starts sessions, stores sessions by `agent_id`, and stops only through the retained session.
5. Implement `server/main.go` with `GET /get_config`, `POST /startAgent`, and `POST /stopAgent`; load `.env.local` then `.env` from the `server/` working directory.
6. In `GET /get_config`, replace missing, zero, or negative UIDs with a generated non-zero UID, generate a one-hour token with `agentkit.GenerateConvoAIToken`, and return `{ app_id, token, uid, channel_name, agent_uid }`.
7. Create a Next.js App Router web app under `client/` with React, TypeScript, Tailwind, `agora-rtc-react`, `agora-rtm`, `agora-agent-client-toolkit`, and `agora-agent-uikit`.
8. Implement `client/next.config.ts` rewrites for `/api/get_config`, `/api/startAgent`, and `/api/stopAgent` to `${AGENT_BACKEND_URL}/...`; return no rewrites when the env var is missing.
9. Implement `client/src/services/api.ts` as the only browser API facade for `getConfig`, `startAgent`, and `stopAgent`.
10. Implement `LandingPage.tsx` to fetch config, start the agent, create/login/subscribe RTM, mount the conversation, renew RTC and RTM tokens with UID-specific `getConfig` calls, and log out RTM on end.
11. Implement `ConversationComponent.tsx` with StrictMode-safe RTC provider usage, `useJoin`, `useLocalMicrophoneTrack`, `usePublish`, `AgoraVoiceAI.init`, toolkit event subscriptions, raw RTM fallback parsing, token renewal, and explicit end-call media release.
12. Implement transcript helpers in `client/src/lib/conversation.ts` that remap toolkit `uid="0"` to the local UID before side-of-screen or speaker mapping logic.
13. Add verification scripts for no `client/app/api/**/route.ts`, rewrite/fetch contract checks, local rewrite stub checks, Go fake-server smoke checks, and Go unit tests.

## Required Copy-Forward Invariants

- Browser code calls `/api/*`; Gin owns the real `/get_config`, `/startAgent`, and `/stopAgent` routes.
- Do not add Next Route Handlers or `client/proxy.ts` for agent or token logic.
- `AGORA_APP_CERTIFICATE`, `GOOGLE_API_KEY`, and  never enter `client/`.
- `AGENT_BACKEND_URL` is a Next server-time env var, not a `NEXT_PUBLIC_*` value.
- Token generation produces a token usable for both RTC and RTM.
- RTM login identity must match the token subject; renewal uses separate RTC and RTM `getConfig` calls when UIDs differ.
- `agentService.sessions` is process-local; production multi-instance deployments need external lifecycle state or a stateless stop strategy.
- `useJoin`, `useLocalMicrophoneTrack`, and `usePublish` own normal mount/unmount lifecycle cleanup.
- Transcript normalization remaps toolkit `uid="0"` before rendering.
- `AdvancedFeatures.EnableRtm`, `SessionParams.DataChannel="rtm"`, `EnableErrorMessage`, and `EnableMetrics` stay enabled for transcript/state/metrics delivery.

## Official Reference Links

Use local source first for this recipe version. For current Agora details that can change, use these official references:

- Official Go quickstart baseline: `https://github.com/AgoraIO-Conversational-AI/agent-quickstart-go`
- Next.js quickstart baseline: `https://github.com/AgoraIO-Conversational-AI/agent-quickstart-nextjs`
- Python quickstart baseline: `https://github.com/AgoraIO-Conversational-AI/agent-quickstart-python`
- ConvoAI OpenAPI spec: `https://docs-md.agora.io/api/conversational-ai-api-v2.x.yaml`
- Current docs index: `https://docs.agora.io/en/llms.txt`

Fetch the OpenAPI spec or current docs before changing direct REST payloads, provider matrices, or vendor-specific config fields.

## Verification

Run narrow checks while building:

```bash
make verify-backend
make verify-web-api
make verify-web-proxy
make verify-local-go
```

Before publishing a derivative baseline, run:

```bash
make verify-web
make verify-local
```

## See Also

- [Back to Workflows](../05_workflows.md)
- [Back to Code Map](../03_code_map.md)
- [Managed Agent Config](managed_agent_config.md)
- [Session Lifecycle](session_lifecycle.md)
- [Verification Scripts](verification_scripts.md)
