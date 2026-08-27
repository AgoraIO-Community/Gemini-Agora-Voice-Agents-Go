# Agora Conversational AI Go Quickstart

Build and run a real-time voice agent with **GeminiSTT**, Gemini 3.6, and Agora-managed MiniMax TTS.

The project includes a Next.js web client and a Go backend. The browser connects to Agora RTC and RTM, while the backend creates and manages the Conversational AI agent session.

## Pipeline

```text
Microphone -> GeminiSTT -> Gemini 3.6 LLM -> MiniMax TTS -> Browser
```

The backend uses the published Agora Agent Server SDK for Go. Gemini ASR and Gemini LLM use the same Google API key. MiniMax TTS is managed by Agora and does not require a separate TTS credential.

## Prerequisites

- Go 1.23 or newer
- Node.js and [pnpm](https://pnpm.io/installation)
- [Agora CLI](https://www.npmjs.com/package/agoraio-cli)
- An Agora project with Conversational AI access enabled
- A Google API key with access to Gemini

## Quickstart

### 1. Clone the repository

```bash
git clone https://github.com/AgoraIO-Conversational-AI/agent-quickstart-go.git
cd agent-quickstart-go
```

### 2. Install and sign in to the Agora CLI

Skip installation if `agora` is already available on your path.

```bash
curl -fsSL https://raw.githubusercontent.com/AgoraIO/cli/main/install.sh | sh -s -- --add-to-path
agora login
```

Select the Agora project to use for this checkout:

```bash
agora project use <project-id-or-name>
```

### 3. Create the local environment

Install the Go and frontend dependencies and create `server/.env.local` from the example:

```bash
make setup
```

Write the Agora App ID and App Certificate from the selected Agora project:

```bash
agora project env write server/.env.local --with-secrets
```

Open `server/.env.local` and add your Google API key:

```env
GOOGLE_API_KEY=your_google_api_key
```

Keep `server/.env.local` private. It is ignored by Git and is never sent to the browser.

### 4. Run the application

Check the local configuration, then start the Go backend and Next.js client:

```bash
make doctor-local
make dev
```

Open [http://localhost:3000](http://localhost:3000) and select **Start conversation**.

The services run at:

| Service | URL |
| --- | --- |
| Web client | [http://localhost:3000](http://localhost:3000) |
| Go backend | [http://localhost:8000](http://localhost:8000) |

## Configuration

The Go server reads configuration from `server/.env.local`.

| Variable | Required | Description |
| --- | :---: | --- |
| `AGORA_APP_ID` | Yes | Agora project App ID. |
| `AGORA_APP_CERTIFICATE` | Yes | Server-only certificate used to create RTC and RTM tokens. |
| `GOOGLE_API_KEY` | Yes | Google API key used by GeminiSTT and Gemini 3.6. |
| `AGENT_GREETING` | No | Overrides the default opening message. |
| `PORT` | No | Go backend port. Defaults to `8000`. |

The source template is [`server/.env.example`](server/.env.example).

## How it works

1. The browser requests a channel, UID, and RTC/RTM token from the Go backend.
2. The browser joins Agora RTC and RTM and publishes microphone audio.
3. The backend starts a Conversational AI agent in the channel.
4. GeminiSTT transcribes the user, Gemini 3.6 generates the response, and MiniMax TTS produces the agent audio.
5. Transcript, agent state, and pipeline metrics are delivered to the browser over RTM.

The browser uses stable `/api/*` paths. In local development, Next.js rewrites those requests to the Go backend through `AGENT_BACKEND_URL`.

## Commands

```bash
make setup          # Install dependencies and create server/.env.local
make doctor-local   # Check Go, dependencies, and required environment variables
make dev            # Run the Go backend and Next.js client
make backend        # Run only the Go backend
make frontend       # Run only the Next.js client
make fmt            # Format Go source
make test           # Run Go tests
make verify         # Run the verification suite
make clean          # Remove generated build artifacts
```

## Troubleshooting

### The agent does not join

Run `make doctor-local` and confirm that `AGORA_APP_ID`, `AGORA_APP_CERTIFICATE`, and `GOOGLE_API_KEY` are non-empty. Also confirm that the selected Agora project has Conversational AI access enabled.

### The browser cannot reach the backend

Confirm that the Go service is listening on port `8000`. When running the frontend separately, set:

```bash
cd client
AGENT_BACKEND_URL=http://localhost:8000 pnpm dev
```

### No transcript or audio appears

Check the browser console for the agent connection and confirm that microphone permission was granted. The pipeline panel displays the latest Gemini ASR, Gemini LLM, and MiniMax TTS latency metrics when those events arrive.

## Project structure

```text
client/                  Next.js voice conversation UI
server/                  Go backend and Agora Agent Server SDK integration
server/.env.example      Safe configuration template
ARCHITECTURE.md          Runtime architecture and request flow
docs/ai/RECIPE.md        Implementation recipe and design constraints
```

## Documentation

- [Architecture](ARCHITECTURE.md)
- [Implementation recipe](docs/ai/RECIPE.md)
- [Server README](server/README.md)

## License

Released under the [MIT License](LICENSE).
