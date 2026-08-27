# 04 Conventions

> How code is structured in this repo and what patterns to preserve when editing.

## Languages & Tooling

| Concern        | Toolchain                                                            |
| -------------- | -------------------------------------------------------------------- |
| Go formatting  | `gofmt` via `make fmt` / `make fmt-backend`                          |
| Go testing     | `go test ./...` via `make test` / `make verify-backend`              |
| Go modules     | `agent-quickstart-go/server` (`server/go.mod`); `go 1.23.0`          |
| TypeScript     | `strict: true` in `client/tsconfig.json`; path alias `@/* → ./src/*` |
| Linter         | Biome (`client/biome.json`); `noExplicitAny` off, `useExhaustiveDependencies` off |
| Format         | Biome (`pnpm lint:fix` writes)                                       |

There is **no ESLint config file** in the client — Biome is the only TS/JS linter.

## Go Package Layout

- Single `package main` under `server/` exposing the HTTP service.
- A second `package main` under `server/cmd/fake-server/` used only for verification.
- No internal subpackages today. New cohesive surfaces should live in new files under `server/`, not new packages, until there is at least one cross-cutting consumer.

## Go Patterns

- HTTP handlers are thin wrappers around `agentService` methods. The service returns `error`; handlers map errors via `toHTTPError` + `isValidationError`.
- `ValueError`-like cases use `errors.New(...)` with messages that pass `isValidationError`; these become `400`. Anything else becomes `500`.
- Logging uses the standard library `log` package; no structured logger is wired.
- JSON: response envelope is always `gin.H{"code": 0, "msg": "success", "data": ...}` (the `data` field is omitted for endpoints that have no payload).
- Pointers `intPtr`, `float64Ptr` exist as small helpers; do not introduce a generic package for one-line wrappers.

## JSON Contract Style

| Side       | Style                                                                                 |
| ---------- | ------------------------------------------------------------------------------------- |
| Go structs | snake_case JSON tags for response payloads (e.g. `app_id`, `channel_name`).            |
| Browser    | camelCase request bodies (`channelName`, `rtcUid`, `userUid`, `agentId`).             |
| Envelope   | `code` / `msg` numeric + string fields; `data` optional.                              |

When adding a new field, mirror its name in both `server/main.go` (or `agent.go`) and `client/src/services/api.ts` to keep the proxy boundary symmetric.

## TypeScript / React Patterns

- Components are PascalCase `.tsx` files under `client/src/components/`. Shared primitives live under `components/ui/` in lowercase files (`button.tsx`, `dropdown-menu.tsx`).
- The RTC client is held in a `useRef` inside a dynamically imported `AgoraRTCProvider` to survive React StrictMode double-mount.
- `useJoin`, `useLocalMicrophoneTrack`, and `usePublish` from `agora-rtc-react` own normal mount/unmount lifecycles. The current end-call handler explicitly unpublishes and closes the microphone track before parent cleanup; avoid adding extra cleanup paths elsewhere.
- `normalizeTranscript` in `client/src/lib/conversation.ts` remaps `uid === '0'` to the local UID. New transcript renderers must keep this remap upstream of any side-of-screen heuristic.

## Hook Ownership Quick Reference

| Hook                       | Owns                          | Anti-pattern                                       |
| -------------------------- | ----------------------------- | -------------------------------------------------- |
| `useJoin`                  | `client.leave()`              | Manual `client.leave()` calls in component cleanup  |
| `useLocalMicrophoneTrack`  | Track creation                | Closing the same track from multiple cleanup paths  |
| `usePublish`               | Publish state                 | Using `unpublish` for mute toggles (use `setEnabled`) |

## Testing

- Go: `server/main_test.go` covers `parseOptionalInt`, `isValidationError`, `newAgentService` env requirement, `generateConfig`, `start`/`stop` validation, and `newRouter` validation routes. Add new helpers' tests alongside.
- Client: no Vitest harness. The three `client/scripts/*.ts` files are the regression suite — extend them when contracts change.

## File Naming

- Components: PascalCase `.tsx` (e.g. `ConversationComponent.tsx`).
- UI primitives: lowercase under `ui/` (e.g. `ui/button.tsx`).
- Scripts: lowercase kebab/snake (`verify-api-contracts.ts`).
- Go: lowercase package files (`main.go`, `agent.go`).

## Module Discipline

- `server/agent.go` is the only place that imports `github.com/AgoraIO/agora-agents-go/v2`.
- `client/src/services/api.ts` is the only place that hard-codes `/api/...` paths (apart from `next.config.ts`).
- `client/scripts/` must remain dependency-free of `client/src/` — they run before `pnpm build` and stand alone.

## Error Handling Shapes

- Go: callers receive an `error`. Validation-shaped errors bubble up to `toHTTPError`, which returns `400` with `{ "detail": "<message>" }`. Anything else becomes `500` with the same `detail` shape.
- TS: `api.ts` helpers throw on non-2xx HTTP; callers (`LandingPage`) catch with `try/catch` and surface a user-friendly message via the existing `ConnectionStatusPanel` issue list.

## Related Deep Dives

- [Verification Scripts](L2/verification_scripts.md) — Implementation details of the three Node verification harnesses.
