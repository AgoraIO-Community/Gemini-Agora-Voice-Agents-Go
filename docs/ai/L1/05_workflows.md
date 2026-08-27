# 05 Workflows

> Step-by-step recipes for the tasks contributors actually do in this repo.

## Add a New Backend Endpoint

1. Add a route in `server/main.go` (`router.GET("/path", handler)` or `POST` / `PUT` / `DELETE`).
2. Implement the handler. Validate inputs, return `agentService` errors through `toHTTPError`. Wrap success responses in `gin.H{"code": 0, "data": <payload>, "msg": "success"}`.
3. If the endpoint needs new service logic, extend `agentService` in `server/agent.go` and add a test in `server/main_test.go`.
4. Add the corresponding rewrite in `client/next.config.ts`:
   ```ts
   { source: '/api/<name>', destination: `${backendUrl}/<name>` }
   ```
5. Add a fetch helper in `client/src/services/api.ts`.
6. Extend `client/scripts/verify-api-contracts.ts` with at least one happy-path and one validation case.
7. Run `make verify-backend && make verify-web-api`. Add `make verify-local-go` and `make verify-web-proxy` if the new route is consumed in the proxy smoke path.

## Change Agent Prompt, VAD, Model, or Voice

Edit `server/agent.go`:

- **Prompt:** modify the `adaPrompt` constant.
- **Greeting:** modify the default greeting in `newAgentService`, or set `AGENT_GREETING` in `server/.env.local`.
- **VAD:** edit `TurnDetectionConfig` (`SpeechThreshold`, `InterruptDurationMs`, `PrefixPaddingMs`, `SilenceDurationMs`, start/end mode).
- **LLM:** change `vendors.NewGemini(...)`; the default model is `gemini-3.6-flash`.
- **STT:** change `vendors.NewGeminiSTT(...)`; the GeminiSTT provider uses standard `NewAgoraClient`.
- **TTS:** change `vendors.NewMiniMaxTTS(...)`; the default is `en-US-Chirp3-HD-Charon`, `en-US`, 24000 Hz.
- **Credentials:** all three default providers reuse `GOOGLE_API_KEY`; add another env var only for an intentional alternate provider.
- **Agent parameters:** edit `agentkit.WithParameters(...)` / `WithAdvancedFeatures(...)` for server-side RTM data channel, error messages, and tool flags.
- **Session:** edit `agentkit.CreateSessionOptions{...}` fields like `IdleTimeout`, `ExpiresIn`, and `EnableStringUID`.

After editing, run `make fmt` and `make verify-backend`.

## Deploy the Client and Backend Separately

- **Client (Next.js):** build via `cd client && pnpm build`. Configure `AGENT_BACKEND_URL` on the deploy target to the public URL of your Go service. Serve with `pnpm start` or any Node hosting platform.
- **Backend (Go):** `make build-backend` produces `server/bin/agent-quickstart-go`. Set `AGORA_APP_ID`, `AGORA_APP_CERTIFICATE`, `GOOGLE_API_KEY`, and optionally `AGENT_GREETING` / `PORT` in the runtime env. Expose `PORT` to your reverse proxy / load balancer.
- The two deploys never share env vars. The browser only ever needs `/api/*` to resolve via the rewrite layer.

## Verify Locally

```bash
make doctor              # quick gate
make doctor-local        # adds Go + env checks
make verify-backend      # go test ./...
make verify-web-api      # contract harness on the rewrite shape
make verify-web-proxy    # fake-server smoke through Next rewrites
make verify-local-go     # builds cmd/fake-server, runs full local stack
make verify              # alias for the production-bound web checks (doctor + verify-web-api + verify-web-build)
make verify-local        # doctor-local + verify-backend + verify-local-go + verify-web-proxy + verify-web-build
```

> `make verify-local` deliberately **does not** include `verify-web-api`. The contract harness already runs as part of `make verify`. Run `make verify-web-api` alongside `make verify-local` when you want both gates.

## Run `make dev`

```bash
make dev
# server: go run . in ./server (PORT defaults to 8000)
# client: AGENT_BACKEND_URL=http://localhost:8000 pnpm dev in ./client
```

`make dev` traps `EXIT` to `kill 0`, so Ctrl-C reliably stops both processes.

## Token Renewal

The browser receives `token-privilege-will-expire` from RTC and calls `getConfig()` twice — once with the RTC `client.uid`, once with the stored `agoraData.uid`. The Go backend re-issues RTM-capable RTC tokens via `agentkit.GenerateConvoAIToken`. If a request passes `uid=0`, the backend generates a non-zero UID because Agora RTC treats `0` as auto-assign but RTM token subjects cannot use `0`.

## Update Module Guides After Behavior Changes

If you change runtime behavior, also update:

- `README.md`
- Repo-root `ARCHITECTURE.md` and `AGENTS.md`
- The relevant file in `docs/ai/L1/` (often `02_architecture.md` and `03_code_map.md`) and `Last Reviewed` in `docs/ai/L0_repo_card.md`
- `docs/ai/RECIPE.md` when extension points, invariants, or stable contracts change

## Roll Back a Bad Deploy

- **Client:** redeploy the previous Next build on the host platform.
- **Backend:** redeploy the previous Go binary; `agent-quickstart-go` is a single statically linked binary, so rollback is just running the older artifact.

## Related Deep Dives

- [From-Scratch Bootstrap](L2/from_scratch_bootstrap.md) — Implementation map for recreating the Go-backed recipe.
- [Managed Agent Config](L2/managed_agent_config.md) — Every tunable field on the agent.
- [Session Lifecycle](L2/session_lifecycle.md) — Renewal sequence in detail.
