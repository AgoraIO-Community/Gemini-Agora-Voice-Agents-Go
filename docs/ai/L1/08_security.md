# 08 Security

> Trust boundaries, secret handling, and security-relevant invariants for the two-process Go quickstart.

## Trust Model

- The browser is untrusted. It may only see Agora tokens issued by the Go server.
- The Go server is the only process that holds `AGORA_APP_CERTIFICATE`, required `GOOGLE_API_KEY`, and any optional alternate-provider keys.
- The Next.js client process sees `AGENT_BACKEND_URL` (a server-side build env var) but does not see any Agora secret.
- The Go server has no per-user authentication today; the threat model assumes the backend URL is gated upstream.

## Environment Variable Boundaries

| Boundary       | Variables                                                              |
| -------------- | ---------------------------------------------------------------------- |
| Browser        | `NEXT_PUBLIC_AGENT_UID` (optional)                                     |
| Next build/run | `AGENT_BACKEND_URL`                                                    |
| Go server      | `AGORA_APP_ID`, `AGORA_APP_CERTIFICATE`, `GOOGLE_API_KEY`, `AGENT_GREETING`, `PORT` |

Mark `AGORA_APP_CERTIFICATE` and `GOOGLE_API_KEY` as secrets in whichever deploy host runs the Go service. The certificate value never appears in `client/`.

## Token Issuance

- `agentService.generateConfig` calls `agentkit.GenerateConvoAIToken(GenerateConvoAITokenOptions{..., TokenExpire: ExpiresInHours(1)})`.
- The same token grants RTC and RTM privileges; `uid <= 0` is replaced with a generated UID before minting so RTM can log in with the token subject.
- Sessions also carry `ExpiresIn: ExpiresInHours(1)` in `CreateSessionOptions`, so an idle session aligns with token expiry.

## Token Renewal

- The client listens for `token-privilege-will-expire` on the RTC engine.
- It calls `getConfig()` twice (RTC UID + stored UID) and renews each client.
- If renewal fails, the next failure surfaces through `MESSAGE_ERROR` on RTM or RTC disconnect events.

## CORS

`server/main.go` uses `gin-contrib/cors` with:

- `AllowAllOrigins: true`
- `AllowMethods: GET, POST, OPTIONS`
- `AllowHeaders: Origin, Content-Type, Accept, Authorization`
- `AllowCredentials: true`

This is suitable for a local-only quickstart. For a public deploy:

1. Restrict `AllowOrigins` to your deployed client origin(s).
2. Reconsider `AllowCredentials: true` if you do not need cookies.
3. Front the Go service with a reverse proxy that enforces its own CORS policy if needed.

## Authentication

- No bearer-token or API-key middleware on the Gin routes.
- No auth in the Next.js rewrites; the browser hits the rewrite directly.
- Anyone with the deployed client URL can start an agent session.

If you need real auth, add a middleware step in `server/main.go` that validates a header before reaching `agentService`. Update `client/src/services/api.ts` and `verify-api-contracts.ts` to send and assert the header.

## Input Validation

- `parseOptionalInt`, `toHTTPError`, and `isValidationError` (in `server/main.go`) translate missing/invalid fields into `400` responses with explicit messages.
- `startAgentRequest` and `stopAgentRequest` are bound with Gin's `ShouldBindJSON`, so malformed JSON yields `400` automatically.
- Validation paths are covered by `server/main_test.go`.

## Secret Handling Rules

- `server/.env.local` is the developer's secret store; do not commit it.
- `server/.env.example` documents shape only — never put real values there.
- `loadEnvFiles` reads `.env.local` then `.env` from the current working directory; running `go run .` from anywhere except `server/` will skip the env file.
- Do not log full env. `log.Printf("failed: %s", err)` is fine; logging `os.Getenv("AGORA_APP_CERTIFICATE")` is not.

## CSP / Security Headers

- No CSP or HSTS headers are set on Gin responses today.
- No security headers are configured in `client/next.config.ts`.
- If you put the Go service behind a reverse proxy, set the headers there.

## Known Limitations

- No rate limiting on `/get_config`, `/startAgent`, `/stopAgent`. A determined client can rapidly issue tokens — bound this upstream if exposed publicly.
- The fake server (`cmd/fake-server`) accepts the same routes with no validation. Do not deploy it.
- The client does not encrypt or sign the browser → Next → Go path beyond TLS at the host level.

## Related Deep Dives

- [Managed Agent Config](L2/managed_agent_config.md) — Where to plug BYOK vendor keys.
