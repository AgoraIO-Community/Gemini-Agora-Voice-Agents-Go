# Agent Development Guide

This guide is for coding agents making changes in `agent-quickstart-go`.

## How to Load

This repository uses progressive disclosure documentation. Docs live under `docs/ai/` in three levels.

1. Read [docs/ai/L0_repo_card.md](docs/ai/L0_repo_card.md) to identify the repo.
2. Load ALL 8 files in [docs/ai/L1/](docs/ai/L1/). They are small — load all upfront.
3. Follow L2 deep-dive links only when L1 isn't detailed enough. The index is at [docs/ai/L1/L2/_index.md](docs/ai/L1/L2/_index.md).
4. This repo declares `Recipe Role: base`; also read [docs/ai/RECIPE.md](docs/ai/RECIPE.md) before changing extension points or stable contracts.

The sections below (Start Here, Patterns, Anti-Patterns, etc.) remain the canonical contributor handbook for hands-on work; the `docs/ai/` tree is the structured summary used by AI agents.

## Start Here

- Read [README.md](./README.md) for setup, supported run modes, and verification.
- Use [ARCHITECTURE.md](./ARCHITECTURE.md) for system-level request flow.
- For layout and responsibilities inside `client/` vs `server/`, use [docs/ai/L1/03_code_map.md](docs/ai/L1/03_code_map.md) and [docs/ai/L1/02_architecture.md](docs/ai/L1/02_architecture.md).

## Current System Shape

- Frontend: Next.js 16, React 19, TypeScript, Tailwind CSS, `agora-rtc-react`, `agora-rtm`, `agora-agent-client-toolkit`, and `agora-agent-uikit`
- Backend: Go + Gin in `server`
- Web API facade: Next rewrites in `client/next.config.ts`
- Auth: Token007 generated from `AGORA_APP_ID` and `AGORA_APP_CERTIFICATE`
- Default agent config: GeminiSTT (`GOOGLE_API_KEY`) → Gemini `gemini-3.6-flash` → managed MiniMax TTS.

## Supported Modes

### Local Go-Backed Development

- Run from the repo root with `make dev`.
- Root scripts start Gin on `http://localhost:8000` and Next.js on `http://localhost:3000`.
- The web app calls `/api/*`; Next rewrites those requests to the Go service through `AGENT_BACKEND_URL=http://localhost:8000`.

### Deployment

- Deploy `client` as a Next.js app.
- Provide a reachable Go backend or keep deployment routing aligned with the README and architecture docs.
- Set `AGENT_BACKEND_URL` when the deployed web app should forward requests to a separate backend.

## Routing / Ownership

- UI and RTC/RTM client lifecycle live in `client`.
- Browser-facing `/api/*` paths are declared in `client/next.config.ts` rewrites.
- Go agent lifecycle logic lives in `server`.
- For deployability changes, update README and architecture docs when the owner of `/api/*` changes.

## Key Files

- `README.md`: setup, local vs deploy modes, troubleshooting, and verification.
- `ARCHITECTURE.md`: top-level environment model.
- `client/next.config.ts`: `/api/*` rewrite mappings to the Go backend.
- `client/src/components/LandingPage.tsx`: conversation entry point, RTM login, token renewal.
- `client/src/components/ConversationComponent.tsx`: core real-time UI, RTC join, `AgoraVoiceAI`, mic UI.
- `client/src/services/api.ts`: browser API client.
- `server/main.go`: Gin entrypoints.
- `server/agent.go`: Agora agent lifecycle wrapper.

## Patterns

- Keep the web client calling `/api/*`; hide backend placement behind Next routing.
- Keep token expiry and renewal behavior aligned between frontend expectations and the Go backend.
- Keep RTC client creation StrictMode-safe.
- Keep transcript speaker mapping based on actual UIDs, not heuristics.
- Keep `GOOGLE_API_KEY` scoped to the Gemini ASR and LLM stages; keep the managed MiniMax TTS configuration explicit.

## Working Rules

- Prefer the smallest change that keeps local mode and deployed mode aligned.
- Keep Go-specific agent lifecycle changes in `server`.
- Keep browser state and RTC/RTM lifecycle changes in `client`.
- If you change request or response contracts, update the web client, backend, contract checks, and README together.

## Commands

From the repo root:

```bash
make setup
make doctor
make doctor-local
make fmt
make test
make dev
make verify
make verify-local
```

Useful narrower checks:

```bash
make build-backend
make verify-web
make verify-local-go
make verify-web-proxy
make verify-backend
```

## Verification Safety

- Safe without live Agora credentials:
  - `make doctor`
  - `make verify-web-api`
  - `make verify-web-proxy`
  - `make verify-local-go`
- Requires local env setup but not a live Agora session:
  - `make doctor-local`
  - `make verify-local`
- Compile and local-runtime coverage only:
  - `make verify-backend`
  - `make verify-local-go`
- Often blocked inside restricted sandboxes because of port binding or Turbopack process spawning:
  - `make dev`
  - `make verify-web-build`
  - `make verify-local`

## Anti-Patterns / What NOT To Do

- Do not reintroduce `client/proxy.ts`; routing should stay explicit through Next config unless the architecture intentionally changes.
- Do not assume Zustand or a separate client-side store exists.
- MiniMax TTS is managed by Agora in this quickstart and does not require a separate credential variable.
- Do not change `/api/*` ownership without updating README, architecture docs, root `AGENTS.md`, and the relevant `docs/ai/L1/` files.
- Do not let Go backend defaults diverge from documented web expectations.

## Done Criteria

Before finishing a change:

1. Run the narrowest relevant verification command.
2. Run `make fmt` if you changed Go files.
3. If the change affects the Go backend binary or startup path, ensure `make build-backend` passes.
4. If the change affects the deployable web app, ensure `make verify-web` passes.
5. If the change affects local Go-backed development, ensure `make verify-local` or the narrower `make verify-local-go`, `make verify-web-proxy`, or `make verify-backend` command passes as appropriate.
6. Update README or architecture docs when developer workflow, request flow, or deployment guidance changes.
7. If the change touches workflows, interfaces, gotchas, or security details, update the matching file under [docs/ai/L1/](docs/ai/L1/) and bump `Last Reviewed` in [docs/ai/L0_repo_card.md](docs/ai/L0_repo_card.md).

## Git Conventions

### Commit messages — conventional commits

- **Format:** `type: description` or `type(scope): description`
- **Types:** `feat:` (new feature), `fix:` (bug fix), `chore:` (maintenance, version bumps), `test:` (test additions/changes), `docs:` (documentation)
- **Scoped variant:** `feat(scope):`, `fix(scope):` — e.g. `feat(server): add greeting env override`
- **Lowercase after prefix** — `feat: add feature`, not `feat: Add feature`
- **Present tense** — "add feature", not "added feature"

### Branch names

- **Format:** `type/short-description` — lowercase, hyphen-separated
- **Types match commit types:** `feat/`, `fix/`, `chore/`, `test/`, `docs/`
- **Examples:** `feat/agent-greeting`, `fix/proxy-rewrite`, `docs/progressive-disclosure`

### General rules

- **No AI tool names** — never mention claude, cursor, copilot, cody, aider, gemini, codex, chatgpt, or gpt-3/4 in commit messages or PR descriptions.
- **No Co-Authored-By trailers** — omit AI attribution lines.
- **No `--no-verify`** — let git hooks run normally.
- **No git config changes** — do not modify `user.name` or `user.email`.

## Doc Commands

| Command         | When to use                                                  |
| --------------- | ------------------------------------------------------------ |
| generate docs   | No `docs/ai/` directory exists yet                           |
| update docs     | Code changed since the `Last Reviewed` date in L0            |
| test docs       | Verify docs give agents the right context (writes `docs/ai/test-results.md`) |
| fix docs        | Close findings from a docs review or test run                |

The generator and tester live in the [AgoraIO-Community/ai-devkit](https://github.com/AgoraIO-Community/ai-devkit) skill set. See the [progressive disclosure standard](https://github.com/AgoraIO-Community/ai-devkit/blob/main/docs/standard/progressive-disclosure-standard.md) for the full specification.
