# Deep Dives Index

| Document                                              | Summary                                                                          | Load When                                                          |
| ----------------------------------------------------- | -------------------------------------------------------------------------------- | ------------------------------------------------------------------ |
| [from_scratch_bootstrap.md](from_scratch_bootstrap.md) | Implementation map for recreating the Go-backed quickstart recipe                | Implementing a new baseline repo from this recipe                  |
| [managed_agent_config.md](managed_agent_config.md)    | Full `server/agent.go` chain: vendors, VAD, session options                       | Changing prompt / VAD / model / voice or wiring a BYOK vendor      |
| [session_lifecycle.md](session_lifecycle.md)          | Browser orchestration of `getConfig` + `startAgent`, RTC + RTM, token renewal     | Touching client-side join, renewal, or mid-call control            |
| [verification_scripts.md](verification_scripts.md)    | What each `client/scripts/*.ts` harness asserts and how to extend it              | Adding a route, changing the proxy boundary, or debugging `verify` |

Recipe profile details live in `docs/ai/RECIPE.md`; load it when changing extension points, invariants, or stable contracts.
