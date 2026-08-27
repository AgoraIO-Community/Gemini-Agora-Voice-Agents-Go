> **When to Read This:** Load this document when you are adding a route, changing the proxy boundary, debugging a failing `make verify*` command, or expanding the contract harness.

# Verification Scripts

## The Four Verification Layers

| Layer                   | Script / Tool                              | Make target          | What it asserts                                          |
| ----------------------- | ------------------------------------------ | -------------------- | -------------------------------------------------------- |
| Go unit tests           | `go test ./...`                            | `make verify-backend`| Handlers, validation, helpers in `server/`               |
| Web → Rewrite contract  | `client/scripts/verify-api-contracts.ts`   | `make verify-web-api`| No `app/api` routes; `/api/*` rewrites are correct        |
| Web → rewrite stub      | `client/scripts/verify-local-proxy.ts`     | `make verify-web-proxy` | Imports `next.config.ts`, resolves rewrites, fetches an in-process stub directly |
| Web → real Go binary    | `client/scripts/verify-local-go.ts`        | `make verify-local-go`  | Builds `cmd/fake-server`, drives the full local stack |

`make verify` is the production-bound chain (`doctor` → `verify-web-api` → `verify-web-build`). `make verify-local` is the dev-bound chain (`doctor-local` → `verify-backend` → `verify-local-go` → `verify-web-proxy` → `verify-web-build`).

## `verify-api-contracts.ts`

Purpose: lock the **shape** of the proxy boundary without standing up any backend.

What it does:

1. Globs `client/app/api/**/route.ts` and fails if any exist.
2. Imports `client/next.config.ts` and asserts `rewrites()` returns the expected `source` → `destination` triples.
3. Imports `client/src/services/api.ts` and asserts each helper hits the correct URL with the correct body shape (via a mock `fetch`).

When you add a route:

- Add the new rewrite entry.
- Add the new fetch helper.
- Extend this script with the new expectation.
- Run `make verify-web-api`.

## `verify-local-proxy.ts`

Purpose: smoke test the **rewrite mapping** without spawning Next dev or the real Go service.

What it does:

1. Starts an in-process stub backend via Node `createServer` that responds to `/get_config`, `/startAgent`, `/stopAgent` with canned JSON.
2. Imports `next.config.ts` directly and calls its `rewrites()` async function to get the rewrite triples.
3. For each browser-side path (e.g. `/api/get_config`), resolves the matching `rewrite.destination`, copies the query string, and `fetch`es the stub backend URL directly — no Next process is involved.
4. Asserts canned payloads round-trip cleanly.

This catches rewrite typos and body-shape regressions instantly. It does **not** catch Next-runtime issues (middleware, headers, edge runtime).

If you change rewrite paths or body shapes, this is the first script that breaks.

## `verify-local-go.ts`

Purpose: exercise the **real Go binary** locally (no managed cloud calls — `cmd/fake-server` stands in).

What it does:

1. Runs `go build -o server/bin/fake-server-smoke ./server/cmd/fake-server`.
2. Spawns the binary listening on a random port.
3. Sets `AGENT_BACKEND_URL=http://localhost:<port>` and runs the same browser-side fetch helpers through Next.
4. Asserts canned responses round-trip cleanly.

This is the closest CI gets to a full integration test. It runs `go build` only — never the real `agentkit` calls — so it is safe in any environment.

## Go Unit Tests

`server/main_test.go` covers:

- `parseOptionalInt` — query param parsing edge cases.
- `isValidationError` — error classification used by `toHTTPError`.
- `newAgentService` — refuses to start without `AGORA_APP_ID`, `AGORA_APP_CERTIFICATE`, and `GOOGLE_API_KEY`.
- `generateConfig` — random UID generation (never `"0"`), echoes channel.
- `start` / `stop` — validation cases (empty `channelName`, missing `agentId`).
- `newRouter` — wires the three routes and rejects malformed inputs.

Add tests for any new helper or handler in the same commit; failure to do so will not break CI (`go test` only runs what is there) but will leak regressions.

## Adding a New Route — Checklist

1. **Go:** handler in `main.go`, optional service method in `agent.go`, test in `main_test.go`.
2. **Client rewrite:** new entry in `next.config.ts`.
3. **Client fetch helper:** new function in `services/api.ts`.
4. **Contract harness:** new assertion in `verify-api-contracts.ts`.
5. **Smoke harness (if needed):** new case in `verify-local-proxy.ts` and/or `verify-local-go.ts` if the route is part of the smoke flow.
6. **Run:** `make verify-backend && make verify-web-api && make verify-local-go`.

## What These Scripts Do NOT Cover

- They do not call the real Agora Conversational AI API. Vendor model changes will not be caught by `make verify`.
- They do not exercise RTC or RTM at the wire level. Browser regression testing requires `make dev` plus a real Agora project.
- They do not run lint/format. Run `pnpm lint` and `make fmt` separately.

## Failure Modes

| Symptom                                                       | Cause                                                                |
| ------------------------------------------------------------- | -------------------------------------------------------------------- |
| `verify-api-contracts` fails on "app/api should not exist"    | Someone added a Next route handler — remove it.                       |
| `verify-local-proxy` hangs on `fetch`                         | Fake server failed to start; check the script's stderr.              |
| `verify-local-go` errors on `go build`                        | Likely `server/go.mod` drift; run `cd server && go mod tidy`.        |
| `verify-backend` test failure on `generateConfig`             | UID logic changed without updating the "not zero" assertion.         |

## See Also

- [Back to Setup](../01_setup.md)
- [Back to Workflows](../05_workflows.md)
- [Back to Interfaces](../06_interfaces.md)
