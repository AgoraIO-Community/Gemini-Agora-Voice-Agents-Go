# 07 Gotchas

> Concrete pitfalls that have been hit before. Read this before refactoring across the proxy boundary.

## `AGENT_BACKEND_URL` Is Mandatory for the Client

`client/next.config.ts` only registers `/api/*` rewrites when `AGENT_BACKEND_URL` is a non-empty trimmed string. If it is unset:

- `client/scripts/doctor.ts` exits with an error.
- `make dev` starts, but every `/api/*` request 404s.
- The browser silently shows "failed to start agent" with a network-tab 404.

`make dev` always exports `AGENT_BACKEND_URL=http://localhost:8000`. Ad-hoc client runs and deploy hosts must set it manually.

## No `client/app/api/**/route.ts`

`client/scripts/verify-api-contracts.ts` asserts that no `app/api` route handlers exist. The client must be rewrite-only. Adding a Next route handler would:

- Shadow the rewrite path (Next matches local routes before applying rewrites).
- Diverge behavior between local dev and deployed environments.
- Fail `make verify-web-api` immediately.

## Deploy Must Stay Rewrite-Backed

- Deploy the Next app and Go service separately, then set `AGENT_BACKEND_URL` on the Next host to the public Go service URL.
- A single-target Next deployment with in-process `/api/*` handlers does **not** exist in this repo today; adding it requires an intentional architecture change and docs updates.
- Per-module `AGENTS.md` / `ARCHITECTURE.md` under `client/` and `server/` were removed. Use repo-root `ARCHITECTURE.md`, `AGENTS.md`, and `docs/ai/L1/` instead of looking for copies next to each crate.

## StrictMode + RTC

- `client/next.config.ts` sets `reactStrictMode: true`. RTC client creation must stay inside the dynamically imported `AgoraRTCProvider` with a `useRef`-held instance.
- `useJoin`, `useLocalMicrophoneTrack`, `usePublish` own their lifecycles; do not call `.leave()`, `.close()`, or `unpublish` manually.

## `uid="0"` Sentinel

`normalizeTranscript` in `client/src/lib/conversation.ts` maps `uid === '0'` to the local UID before rendering. New transcript renderers must use the normalized turn list — bypassing the helper puts the user on the wrong side.

## Token Renewal Uses Two `getConfig` Calls

`handleTokenWillExpire` in `LandingPage.tsx` issues two `getConfig()` requests:

- One with the RTC `client.uid` for the RTC renewal.
- One with the stored `agoraData.uid` for the RTM renewal.

If RTC's `client.uid` is `0` (never reported), the handler skips. Swapping the two UIDs silently breaks RTM renewal.

## Server `uid <= 0` Means "Generate One"

`generateConfig` treats any `uid <= 0` query value as "give me a random user UID." `main_test.go` asserts the generated UID is not `"0"`. If you change this behavior, update the test in the same commit.

## One Token String, Concrete UID

The quickstart deliberately uses `GenerateConvoAIToken` for both RTC and RTM to keep setup simple. That only works after the backend has resolved a concrete user UID. Do not pass `0` through to RTM login or RTM renewal.

## CORS Is Wide Open

`server/main.go` uses `cors.AllowAllOrigins: true` with `AllowCredentials: true`. This is fine for a local quickstart but is **not** appropriate for production behind a reverse proxy that already enforces CORS. Tighten this before exposing the Go service publicly.

## `make verify-local-go` Builds a Binary

`verify-local-go.ts` produces `server/bin/fake-server-smoke` and execs it. The binary is gitignored but is not cleaned by `make clean` (which removes `server/bin` entirely — see `clean-backend`). If you run the script from a clean checkout, expect ~10 seconds the first time while the binary builds.

## No `Co-Authored-By` in Commit Messages

The repo's git history is short and entirely human-authored. Keep it that way — see `AGENTS.md` "Git Conventions."

## `server/.env.local` Is CWD-Sensitive

`loadEnvFiles` reads `.env.local` then `.env` from the **current working directory**. `make dev` and `make backend` run `go run .` from inside `server/`, so they find the file. Running `go run ./server` from the repo root would silently skip it because the CWD is the repo root, not `server/`. Always cd into `server/` for ad-hoc commands.

## Verify Scripts Spawn Subprocesses

`verify-local-proxy.ts` and `verify-local-go.ts` spawn child processes (Next dev, the fake-server binary). In a restricted sandbox these may fail with `EACCES` or port-binding errors. The pure contract checks (`verify-web-api`, `verify-backend`) run without spawning servers and are safe in sandboxed CI.

## Server-Side Agent Flags

RTM delivery, tool enablement, error messages, and data channel settings belong in `server/agent.go`. The browser only consumes the resulting events; it should not invent server-side agent parameters.

## Related Deep Dives

- [Managed Agent Config](L2/managed_agent_config.md) — Server-side defaults.
- [Verification Scripts](L2/verification_scripts.md) — Each verify script in detail.
