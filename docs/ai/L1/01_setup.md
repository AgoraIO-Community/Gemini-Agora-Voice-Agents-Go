# 01 Setup

> Local environment for the two-process Go + Next.js quickstart: prerequisites, env vars, and verification commands wired through the `Makefile`, root `package.json`, and `client/scripts/`.

## Prerequisites

- **Go** 1.23.* / 1.24.* / 1.25.* / 1.26.* / 2.* (matched by `make doctor-local`).
- **Node**: any version the pinned **`pnpm@10.23.0`** supports (root `package.json` `packageManager`).
- **pnpm** 10.23.0 (enforced by the workspaces config; the root `package.json` is a pnpm workspace).
- Agora project with App ID + App Certificate.

## Install

```bash
make setup
# or:
pnpm install
cd server && go mod tidy
```

`make setup` chains `setup-env` (copies `server/.env.example` → `server/.env.local` if missing), `setup-backend` (`go mod tidy`), and `setup-frontend` (`pnpm install` when `node_modules/` is missing).

## Environment Variables

`server/.env.example`:

```
AGORA_APP_ID=your_agora_app_id
AGORA_APP_CERTIFICATE=your_agora_app_certificate
GOOGLE_API_KEY=your_google_api_key
AGENT_GREETING=Hi there! I'm Ada, your virtual assistant from Agora. How can I help?
PORT=8000
```

`client/.env.local.example`:

```
AGENT_BACKEND_URL=http://localhost:8000
```

| Variable                 | Process     | Required | Notes                                                                 |
| ------------------------ | ----------- | -------- | --------------------------------------------------------------------- |
| `AGORA_APP_ID`           | Go (server) | Yes      | Loaded in `newAgentService`.                                          |
| `AGORA_APP_CERTIFICATE`  | Go (server) | Yes      | Loaded in `newAgentService`; never exposed to the browser.            |
| `GOOGLE_API_KEY`         | Go (server) | Yes      | Reused by GeminiSTT, Gemini LLM, and MiniMaxTTS; server-only.          |
| `AGENT_GREETING`         | Go (server) | No       | Optional first utterance.                                             |
| `PORT`                   | Go (server) | No       | Default `8000` (`main.go`).                                            |
| `AGENT_BACKEND_URL`      | Next build  | Yes for rewrites | Empty/missing means no `/api/*` rewrites are registered.       |
| `NEXT_PUBLIC_AGENT_UID`  | Browser     | No       | Optional override read in `ConversationComponent.tsx`.                |

## Quick Commands

```bash
make setup            # one-time bootstrap
make doctor           # pnpm + node_modules presence
make doctor-local     # adds Go + .env.local + Agora credential presence
make dev              # spawns Gin + Next dev with AGENT_BACKEND_URL set
make fmt              # gofmt server *.go and cmd/fake-server/*.go
make build            # build-backend + build-web
make test             # alias for test-backend → verify-backend → go test ./...
make verify           # alias for verify-web (doctor + verify-web-api + verify-web-build)
make verify-local     # doctor-local + backend + local-go + web-proxy + web-build
make clean            # remove backend bin, node_modules, .next, client/dist
```

Useful narrower targets:

```bash
make verify-web-api      # client/scripts/verify-api-contracts.ts
make verify-web-proxy    # client/scripts/verify-local-proxy.ts
make verify-local-go     # builds cmd/fake-server, runs verify-local-go.ts
make verify-backend      # cd server && go test ./...
make build-backend       # produces server/bin/agent-quickstart-go
```

The root `package.json` exposes the same workflows under `pnpm run setup`, `pnpm run dev`, `pnpm run verify`, etc. — the Makefile remains the canonical entry point.

## Verification Safety

| Command                | Live Agora? | Notes                                                |
| ---------------------- | ----------- | ---------------------------------------------------- |
| `make doctor`          | No          | pnpm + node_modules sanity                           |
| `make doctor-local`    | No          | Adds Go + env presence                               |
| `make verify-web-api`  | No          | Contract harness with mocked SDK                     |
| `make verify-web-proxy`| No          | Static fake-server smoke                             |
| `make verify-local-go` | No          | Boots `cmd/fake-server`, exercises web → proxy → Go  |
| `make verify-backend`  | No          | `go test ./...`                                      |
| `make verify-web-build`| No          | `pnpm build`                                         |
| `make dev`             | Yes (for use) | Often blocked in sandboxes due to port binding     |

## Common Setup Failures

- `make doctor-local` reports **"Go version not supported"** → install Go 1.23+.
- Doctor fails on missing `server/.env.local` → run `make setup-env` or copy from `server/.env.example`.
- `make verify-web-api` fails on a new route → extend `client/scripts/verify-api-contracts.ts` to cover it.
- `make dev` exits with port-in-use → either Gin or Next is already running; check ports 8000 and 3000.

## Related Deep Dives

- [Verification Scripts](L2/verification_scripts.md) — How the four contract checks differ and when to run each.
