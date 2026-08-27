import { spawn, spawnSync } from 'node:child_process'
import { once } from 'node:events'
import { existsSync, mkdirSync, rmSync } from 'node:fs'
import path from 'node:path'
import { setTimeout as sleep } from 'node:timers/promises'

import nextConfig from '../next.config'

type Rewrite = {
  source: string
  destination: string
}

type ChildProcessHandle = {
  kill: () => void
  exited: Promise<number>
  readStderr: () => string
  getExitCode: () => number | null
}

function assert(condition: unknown, message: string): asserts condition {
  if (!condition) {
    throw new Error(message)
  }
}

function getJson(response: Response) {
  return response.json() as Promise<Record<string, unknown>>
}

async function getRewrites(): Promise<Rewrite[]> {
  const rewrites = nextConfig.rewrites
  assert(typeof rewrites === 'function', 'next.config.ts should define async rewrites()')

  const result = await rewrites()
  if (Array.isArray(result)) {
    return result as Rewrite[]
  }

  return [
    ...((result.beforeFiles ?? []) as Rewrite[]),
    ...((result.afterFiles ?? []) as Rewrite[]),
    ...((result.fallback ?? []) as Rewrite[]),
  ]
}

async function requestViaRewrite(sourceUrl: string, init?: RequestInit) {
  const source = new URL(sourceUrl, 'http://localhost:3000')
  const rewrites = await getRewrites()
  const rewrite = rewrites.find((candidate) => candidate.source === source.pathname)
  assert(rewrite, `Missing rewrite for ${source.pathname}`)

  const target = new URL(rewrite.destination)
  target.search = source.search
  return fetch(target, init)
}

async function waitForHealthyBackend(baseUrl: string, timeoutMs: number) {
  const deadline = Date.now() + timeoutMs
  let lastError = 'backend did not start'

  while (Date.now() < deadline) {
    try {
      const response = await fetch(`${baseUrl}/get_config?uid=4321&channel=go-smoke`)
      if (response.ok) {
        return
      }
      lastError = `backend returned ${response.status}`
    } catch (error) {
      lastError = error instanceof Error ? error.message : String(error)
    }

    await sleep(250)
  }

  throw new Error(`Timed out waiting for Go backend: ${lastError}`)
}

function spawnProcess(cmd: string[], cwd: string, env: NodeJS.ProcessEnv): ChildProcessHandle {
  const child = spawn(cmd[0], cmd.slice(1), {
    cwd,
    env,
    stdio: ['ignore', 'ignore', 'pipe'],
  })

  let stderr = ''
  child.stderr.on('data', (chunk: Buffer | string) => {
    stderr += chunk.toString()
  })

  const exited = (async () => {
    const [code] = (await once(child, 'exit')) as [number | null]
    return code ?? 1
  })()

  return {
    kill: () => {
      if (!child.killed) {
        child.kill('SIGTERM')
      }
    },
    exited,
    readStderr: () => stderr.trim(),
    getExitCode: () => child.exitCode,
  }
}

async function main() {
  const projectRoot = process.cwd()
  const serverRoot = path.resolve(projectRoot, '..', 'server')
  const fakeServerMain = path.join(serverRoot, 'cmd', 'fake-server', 'main.go')
  const tempDir = path.join(serverRoot, 'bin')
  const fakeServerBinary = path.join(tempDir, 'fake-server-smoke')

  if (!existsSync(fakeServerMain)) {
    throw new Error('Missing server/cmd/fake-server/main.go. Expected Go fake server for local verification.')
  }
  mkdirSync(tempDir, { recursive: true })

  const buildResult = spawnSync('go', ['build', '-o', fakeServerBinary, './cmd/fake-server'], {
    cwd: serverRoot,
    stdio: ['ignore', 'ignore', 'pipe'],
  })
  if (buildResult.status !== 0) {
    const stderr = buildResult.stderr?.toString().trim() ?? ''
    throw new Error(`Failed to build Go fake server.${stderr ? ` go build said: ${stderr}` : ''}`)
  }

  const port = 43120 + Math.floor(Math.random() * 20)
  const backendUrl = `http://127.0.0.1:${port}`
  const originalBackendUrl = process.env.AGENT_BACKEND_URL

  const serverProcess = spawnProcess([fakeServerBinary], serverRoot, {
    ...process.env,
    PORT: String(port),
  })

  try {
    await waitForHealthyBackend(backendUrl, 10_000)

    process.env.AGENT_BACKEND_URL = backendUrl

    const response = await requestViaRewrite('/api/get_config?uid=4321&channel=go-smoke')
    const body = await getJson(response)

    assert(response.status === 200, 'GET /api/get_config should proxy to the Go backend')
    assert(body.code === 0, 'GET /api/get_config should preserve the Go success payload')

    const data = body.data as Record<string, unknown> | undefined
    assert(data?.uid === '4321', 'GET /api/get_config should preserve the requested uid through Go')
    assert(
      data?.channel_name === 'go-smoke',
      'GET /api/get_config should preserve the requested channel through Go',
    )
    assert(
      typeof data?.token === 'string' && data.token.length > 0,
      'GET /api/get_config should return a token from Go',
    )
    assert(
      typeof data?.agent_uid === 'string' && data.agent_uid.length > 0,
      'GET /api/get_config should return an agent uid from Go',
    )

    const startResponse = await requestViaRewrite('/api/startAgent', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        channelName: 'go-smoke',
        rtcUid: 9999,
        userUid: 4321,
      }),
    })
    const startBody = await getJson(startResponse)
    assert(startResponse.status === 200, 'POST /api/startAgent should proxy to the Go backend')
    assert(startBody.code === 0, 'POST /api/startAgent should preserve the Go success payload')
    assert(
      (startBody.data as Record<string, unknown> | undefined)?.agent_id === 'fake-agent-9999',
      'POST /api/startAgent should return the agent id from Go',
    )

    const stopResponse = await requestViaRewrite('/api/stopAgent', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ agentId: 'fake-agent-9999' }),
    })
    const stopBody = await getJson(stopResponse)
    assert(stopResponse.status === 200, 'POST /api/stopAgent should proxy to the Go backend')
    assert(stopBody.code === 0, 'POST /api/stopAgent should preserve the Go success payload')

    console.log('Local Go app proxy smoke check passed')
  } finally {
    if (originalBackendUrl) {
      process.env.AGENT_BACKEND_URL = originalBackendUrl
    } else {
      process.env.AGENT_BACKEND_URL = ''
    }

    serverProcess.kill()
    await Promise.race([serverProcess.exited, sleep(2_000)])
    rmSync(fakeServerBinary, { force: true })

    const exitCode = serverProcess.getExitCode()
    if (exitCode && exitCode !== 0) {
      const stderr = serverProcess.readStderr()
      if (stderr) {
        console.error(stderr)
      }
    }
  }
}

main().catch((error) => {
  console.error(error)
  process.exitCode = 1
})
