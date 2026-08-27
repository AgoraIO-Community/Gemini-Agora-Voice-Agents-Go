# PD Documentation Test Results

Tested: 2026-05-28
Agent: Codex with delegated explorer sub-agents
Repo: `agent-quickstart-go`

## Summary

- Total questions: 6
- Passed: 6
- L1 gaps: 0
- L2 gaps: 0
- Cross-ref issues: 0

Structural checks passed: L0 exists and is under 50 lines, all 8 L1 files exist, combined L1 line count is under 1,600, L2 files start with the required callout, links resolve, `AGENTS.md` includes How to Load / Git Conventions / Doc Commands, and `CLAUDE.md` references `@AGENTS.md`.

## Results

### Setup & Build

| # | Question | Answer Correct? | Files Read | Level Loaded | Result |
| - | -------- | --------------- | ---------- | ------------ | ------ |
| 1 | How do I install dependencies, configure env, and run both services locally? | Yes | `docs/ai/L0_repo_card.md`, all 8 L1 files, `README.md`, `ARCHITECTURE.md` | L0+L1 sufficient | Pass |

### Test & Run

| # | Question | Answer Correct? | Files Read | Level Loaded | Result |
| - | -------- | --------------- | ---------- | ------------ | ------ |
| 2 | What does the verification suite cover and which commands should I run for web, local Go-backed, and backend-only changes? | Yes | `docs/ai/L0_repo_card.md`, all 8 L1 files, `docs/ai/RECIPE.md`, `docs/ai/L1/L2/_index.md`, `docs/ai/L1/L2/verification_scripts.md`, `docs/ai/L1/L2/session_lifecycle.md` | L2 required and correctly used | Pass |

### Conventions

| # | Question | Answer Correct? | Files Read | Level Loaded | Result |
| - | -------- | --------------- | ---------- | ------------ | ------ |
| 3 | Why is this repo rewrite-only instead of using `client/app/api` route handlers, and what breaks if `AGENT_BACKEND_URL` is missing? | Yes | `docs/ai/L0_repo_card.md`, all 8 L1 files, `docs/ai/RECIPE.md`, `docs/ai/L1/L2/_index.md`, `docs/ai/L1/L2/verification_scripts.md`, `docs/ai/L1/L2/session_lifecycle.md` | L0+L1 sufficient | Pass |

### Development

| # | Question | Answer Correct? | Files Read | Level Loaded | Result |
| - | -------- | --------------- | ---------- | ------------ | ------ |
| 4 | How would I add a new backend endpoint exposed to the browser? | Yes | `docs/ai/L0_repo_card.md`, `docs/ai/L1/02_architecture.md`, `docs/ai/L1/03_code_map.md`, `docs/ai/L1/04_conventions.md`, `docs/ai/L1/05_workflows.md`, `docs/ai/L1/06_interfaces.md`, `docs/ai/L1/07_gotchas.md`, `server/main.go`, `server/agent.go`, `client/next.config.ts`, `client/src/services/api.ts`, `ARCHITECTURE.md` | L0+L1 sufficient | Pass |

### Deep Dive

| # | Question | Answer Correct? | Files Read | Level Loaded | Result |
| - | -------- | --------------- | ---------- | ------------ | ------ |
| 5 | What recipe extension points and invariants must be preserved when changing the agent provider pipeline? | Yes | `docs/ai/L0_repo_card.md`, relevant L1 files, `docs/ai/L1/L2/_index.md`, `docs/ai/L1/L2/managed_agent_config.md`, `docs/ai/RECIPE.md`, `server/agent.go` | L2 + RECIPE required and correctly used | Pass |
| 6 | How does token renewal keep RTC and RTM UIDs straight? | Yes | `docs/ai/L0_repo_card.md`, all 8 L1 files, `docs/ai/RECIPE.md`, `docs/ai/L1/L2/_index.md`, `docs/ai/L1/L2/verification_scripts.md`, `docs/ai/L1/L2/session_lifecycle.md` | L2 required and correctly used | Pass |

## Recommended Fixes

- None.

## Review Fix Retest

Retested: 2026-05-28

| Finding | Source checked | Docs changed | Result | Notes |
| ------- | -------------- | ------------ | ------ | ----- |
| Stale single-target deployment claims | `README.md`, `ARCHITECTURE.md`, `client/next.config.ts`, `client/scripts/verify-api-contracts.ts` | `README.md`, `ARCHITECTURE.md`, `docs/ai/L1/07_gotchas.md`, `server/README.md` | Pass | Docs now describe rewrite-backed deployment to a reachable Go service. |
| Missing recipe profile artifact | `docs/standard/recipe-profile.md`, `docs/ai/L0_repo_card.md`, `AGENTS.md`, `../agent-quickstart-nextjs/docs/ai/RECIPE.md`, `../agent-quickstart-python/docs/ai/RECIPE.md` | `docs/ai/RECIPE.md`, `docs/ai/L0_repo_card.md`, `AGENTS.md`, `docs/ai/L1/03_code_map.md`, `docs/ai/L1/05_workflows.md`, `docs/ai/L1/L2/_index.md`, `docs/ai/L1/L2/from_scratch_bootstrap.md` | Pass | L0 declares base recipe metadata, `RECIPE.md` defines extension points/invariants/stable contracts, and the L2 bootstrap map gives agents a scratch implementation path. |
| Stale L2 session and managed-agent details | `client/src/services/api.ts`, `client/src/components/LandingPage.tsx`, `client/src/components/ConversationComponent.tsx`, `server/agent.go` | `docs/ai/L1/L2/session_lifecycle.md`, `docs/ai/L1/L2/managed_agent_config.md`, `docs/ai/L1/02_architecture.md`, `docs/ai/L1/06_interfaces.md` | Pass | Corrected `GET /api/get_config`, options-object `getConfig`, `EnableMetrics`, stop behavior, and end-call cleanup. |
