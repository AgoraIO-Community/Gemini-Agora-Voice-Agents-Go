# 03 Code Map

> Where to find things. Paths are relative to the repo root.

## Top-Level Tree (curated)

```
Makefile                  # Canonical command runner
package.json              # pnpm workspaces ["client"]; mirror scripts
pnpm-workspace.yaml
README.md                 # Setup + commands
ARCHITECTURE.md           # Top-level environment model
AGENTS.md                 # Contributor entry point
CLAUDE.md                 # Pointer to AGENTS.md
docs/ai/RECIPE.md         # Base recipe contract: extension points, invariants, stable contracts

client/                   # Next.js 16 web app
  app/
    layout.tsx            # Fonts, metadata, viewport, imports @/index.css
    page.tsx              # Renders <LandingPage />
  src/
    components/           # UI + RTC/RTM lifecycle
      LandingPage.tsx
      ConversationComponent.tsx
      QuickstartConversationLayout.tsx
      QuickstartTranscriptPanel.tsx
      QuickstartPipelineMetrics.tsx
      QuickstartPreCallCard.tsx
      ConnectionStatusPanel.tsx
      ConversationErrorCard.tsx
      MicrophoneSelector.tsx
      ErrorBoundary.tsx
      LoadingSkeleton.tsx
      share-button.tsx
      ui/
        button.tsx
        dropdown-menu.tsx
    lib/
      agora.ts            # DEFAULT_AGENT_UID = 123456
      conversation.ts     # Transcript normalization + visualizer mapping
      utils.ts            # cn() (clsx + tailwind-merge)
    services/
      api.ts              # getConfig / startAgent / stopAgent fetch helpers
    types/
      conversation.ts     # AgoraTokenData, AgoraRenewalTokens, ConversationComponentProps
  public/                 # favicon.svg, agora logos, site.webmanifest
  scripts/
    doctor.ts             # Requires .env.local.example + valid AGENT_BACKEND_URL
    verify-api-contracts.ts
    verify-local-proxy.ts
    verify-local-go.ts
  biome.json
  next.config.ts          # rewrites() (see 02_architecture)
  tsconfig.json
  docs/                   # Workflow + review + project state templates

server/                   # Go (Gin) backend
  go.mod                  # module agent-quickstart-go/server, go 1.23.0
  go.sum
  main.go                 # newRouter(), CORS, env loading, route registration
  agent.go                # agentService, generateConfig, start, stop, helpers
  main_test.go            # Go unit tests covering routes + agentService
  cmd/fake-server/
    main.go               # Fake /get_config/startAgent/stopAgent for smoke tests
  .env.example            # AGORA_APP_ID, AGORA_APP_CERTIFICATE, GOOGLE_API_KEY, greeting, port
  README.md
```

## Core Files Table

| File                                                | Purpose                                                                  |
| --------------------------------------------------- | ------------------------------------------------------------------------ |
| `Makefile`                                          | All top-level workflows; `make dev` orchestrates both processes.         |
| `client/next.config.ts`                             | Rewrites `/api/*` to `${AGENT_BACKEND_URL}/...` when env is set.         |
| `client/src/services/api.ts`                        | Browser API client: `getConfig`, `startAgent`, `stopAgent`.              |
| `client/src/components/LandingPage.tsx`             | Session bootstrap, RTM login, renewal handler, provider wiring.          |
| `client/src/components/ConversationComponent.tsx`   | RTC join, `AgoraVoiceAI` init, transcript/state/metrics, mic UI.         |
| `client/src/lib/conversation.ts`                    | `normalizeTranscript` (uid `"0"` remap), visualizer state mapping.       |
| `client/scripts/verify-api-contracts.ts`            | Asserts no `app/api` route handlers + browser-side fetch shapes.         |
| `client/scripts/verify-local-proxy.ts`              | Smoke test: fake server + Next rewrites round-trip.                      |
| `client/scripts/verify-local-go.ts`                 | Builds `cmd/fake-server` binary, runs full local Go path.                |
| `server/main.go`                                    | Router, env load, handlers, validation error mapping.                    |
| `server/agent.go`                                   | `agentService`: standard client, shared-key GeminiSTT → Gemini → MiniMaxTTS, retained-session lifecycle. |
| `server/cmd/fake-server/main.go`                    | In-process stand-in for the real backend during verification.            |
| `server/main_test.go`                               | `go test ./...` coverage for routes + helpers.                           |

## Module Boundaries

- `client/` owns React UI, RTC/RTM lifecycle, and the proxy contract to the backend.
- `server/` owns Gin handlers and all Agora SDK calls; secrets never leave this process.
- `client/scripts/` owns verification harnesses that gate `make verify*`.
- Module-specific `AGENTS.md` / `ARCHITECTURE.md` under `client/` and `server/` were removed — use repo-root `ARCHITECTURE.md`, `AGENTS.md`, and this L1 tree.

## What's Not in the Repo

- **No `client/src/hooks/`** and **no `useAgoraConnection.ts`** — RTC/RTM orchestration lives in `LandingPage.tsx` and `ConversationComponent.tsx`.
- **No `client/README.md`** in the current tree.
- **No `*_test.ts`** in the client; verification is via the three Node-based scripts above.
- **No `app/api/**/route.ts`** — `verify-api-contracts.ts` enforces this.

## Related Deep Dives

- [From-Scratch Bootstrap](L2/from_scratch_bootstrap.md) — Implementation map for recreating the Go-backed recipe.
- [Session Lifecycle](L2/session_lifecycle.md) — Concrete walk through `LandingPage` + `ConversationComponent`.
