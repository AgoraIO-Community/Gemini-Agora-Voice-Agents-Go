# Architecture

## 1. Tech Stack & Project Scope

- Project type: pnpm workspace package inside the Go quickstart
- Frontend framework: Next.js App Router
- Language: TypeScript
- Build tool: Next.js
- UI/Styling: Tailwind CSS
- State management: Local React state and hooks
- Data fetching: Fetch calls through `src/services/api.ts`
- Backend integration: Next rewrites `/api/*` to the Go service when `AGENT_BACKEND_URL` is set

## 2. Package Management & Toolchain

- Runtime: Node.js through pnpm scripts
- Package manager: pnpm from the repo root workspace
- Lockfile: root `pnpm-lock.yaml`
- Lint/Format: Biome
- Verification: Make targets and TypeScript scripts run with `pnpm node --import tsx`
- Code generation: None

## 3. Runtime & Environment

- Environment variables: `.env.local` files and process env
- Required local proxy variable:
  - `AGENT_BACKEND_URL` - Go backend URL for Next rewrites, usually `http://localhost:8000`
- Entry point: `app/page.tsx`
- Default port: 3000
- Secrets handling: Agora credentials stay in `server/.env.local`; the browser receives short-lived tokens from the backend

## 4. Dependencies & External Services

- Backend services:
  - Agora RTC SDK - Real-time communication
  - Agora Conversational AI - Voice/text conversation
- Private dependencies: None
- Third-party services:
  - Agora (authentication via App ID/Token)

## 5. Module Responsibilities & Directory Structure

```
app/
├── page.tsx        # Root page; renders LandingPage
src/
├── components/     # Conversation UI components
├── lib/            # Conversation helpers
└── services/       # Browser API client for /api/* paths
```

## 6. Data Flow & Routing

- Routing: Next.js App Router
- Data flow:
  1. User starts a conversation in the web UI
  2. `src/services/api.ts` calls stable browser paths under `/api/*`
  3. `next.config.ts` rewrites those paths to the Go backend in local proxy mode
  4. RTM bootstrap stays in `LandingPage`; RTC join, publish, transcript, and metrics state stay in `ConversationComponent`
- Error handling: Component-level error state surfaced in the conversation UI
- Loading states: Local React state

## 7. Build & Deployment

- Build command: `pnpm build` from `client/`, or `make build-web` from the repo root
- Output: `.next/`
- Deployment target: Next.js hosting for `client`
- Local mode: `make dev` starts both the Go backend and Next frontend
- Rollback: Redeploy previous version

## 8. Constraints & Known Issues

- Known constraints:
  - Agora SDK requires HTTPS in production
  - WebRTC requires user permission for microphone/camera
- Known pitfalls:
  - Token expiration handling
  - Network reconnection logic
- Behaviors that must not break:
  - Real-time audio/video streaming
  - Conversation state reset on end-call cleanup

## 9. Update Log

- 2026-01-21: Initial architecture document created
