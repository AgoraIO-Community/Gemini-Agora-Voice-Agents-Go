# agora-conversational-ai-demo-go — Repo Card

> Next.js web client + Go (Gin + Agora Agent Server SDK) backend for an Agora Conversational AI voice agent.

## Identity

| Field         | Value                                                              |
| ------------- | ------------------------------------------------------------------ |
| Repo          | `AgoraIO-Conversational-AI/agent-quickstart-go`                    |
| Type          | `distributed-system` (single repo, two co-located processes)       |
| Language      | Go 1.23 (Gin) backend + Next.js 16 / React 19 / TypeScript client  |
| Deploy Target | `client/` as Next.js app with rewrites to a reachable Go service   |
| Owner         | Agora Conversational AI DevEx                                      |
| Last Reviewed | 2026-08-26                                                         |
| Recipe Role   | `base`                                                             |
| Recipe Version| `1.0.0`                                                            |
| Recipe Status | `stable`                                                     |

## L1 — Summaries

The Audience column helps agents prioritise: **Use** = consuming the quickstart's behavior, **Maintain** = modifying internals.

| File                                     | Purpose                                                              | Audience       |
| ---------------------------------------- | -------------------------------------------------------------------- | -------------- |
| [01_setup](L1/01_setup.md)               | pnpm + Go toolchain, env vars, doctor, Makefile + script reference   | Use & Maintain |
| [02_architecture](L1/02_architecture.md) | Two-process topology, `/api/*` rewrite proxy, request lifecycle      | Maintain       |
| [03_code_map](L1/03_code_map.md)         | `client/` and `server/` trees with key file responsibilities         | Maintain       |
| [04_conventions](L1/04_conventions.md)   | Go package layout, Biome, JSON contract style, hook ownership        | Maintain       |
| [05_workflows](L1/05_workflows.md)       | Add a route, change managed agent config, verify, deploy each half   | Use            |
| [06_interfaces](L1/06_interfaces.md)     | Gin route contracts, rewrites, env vars, managed agent payload       | Use & Maintain |
| [07_gotchas](L1/07_gotchas.md)           | `AGENT_BACKEND_URL` dependency, doc drift, missing assets            | Maintain       |
| [08_security](L1/08_security.md)         | Cert handling, CORS, token expiry, server-only env                   | Maintain       |
