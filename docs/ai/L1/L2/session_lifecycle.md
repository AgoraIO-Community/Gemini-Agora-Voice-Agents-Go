> **When to Read This:** Load this document when you are touching `LandingPage.tsx`, `ConversationComponent.tsx`, RTM bootstrap, or token renewal — anything in the browser orchestration path.

# Session Lifecycle

## Why It Matters

The browser holds the only state that ties RTC, RTM, and the managed agent together. Mistakes here surface as ghost audio sessions, missing transcripts, or silent token-expiry disconnects. The Go server is stateless beyond per-session bookkeeping; correctness lives in the client.

## End-to-End Sequence

```
User clicks "Start"
   │
   ▼
LandingPage.tsx
   │
   ├─▶ getConfig()                       (GET /api/get_config → GET /get_config)
   │     ◀── data: { app_id, token, uid, channel_name, agent_uid }
   │
   ├─▶ startAgent(channel_name, agent_uid, uid)    (POST /api/startAgent)
   │     ◀── data: { agent_id, channel_name, status }
   │
   ├─▶ new AgoraRTM.RTM(app_id, uid)
   │     .login({ token })
   │     .subscribe(channel_name)
   │
   └─▶ setAgoraData({ app_id, token, uid, channel, agentId })
         setRtmClient(rtmClient)
         render <ConversationComponent />

ConversationComponent.tsx
   │
   ├─▶ useJoin({ appid, token, channel, uid })
   │     (StrictMode-safe via useRef-held AgoraRTCProvider)
   │
   ├─▶ useLocalMicrophoneTrack()
   ├─▶ usePublish([track])
   │
   └─▶ new AgoraVoiceAI({
            rtcEngine,
            rtmConfig: { rtmEngine: rtmClient },
        }).subscribeMessage(channel_name)
         Listens: TRANSCRIPT_UPDATED, AGENT_STATE_CHANGED, AGENT_METRICS,
                  MESSAGE_ERROR, MESSAGE_SAL_STATUS, AGENT_ERROR
```

## Why `getConfig` and `startAgent` Run in Parallel

`startAgent` does not need RTM to be live; it only needs the channel name and UIDs. Running it alongside RTM login shaves ~200ms off perceived start latency. If `startAgent` fails (e.g. backend transient), `LandingPage` logs but continues — the user can still see RTC join and may retry from the UI.

## RTM Bootstrap Details

`LandingPage` dynamically imports `agora-rtm` to avoid touching `window` during SSR. After `RTM` is constructed:

1. `await rtm.login({ token })` — uses the **same token string** as RTC (one token, both privileges).
2. Wait for the `CONNECTED` event (helper `waitForRtmConnected` or similar).
3. `await rtm.subscribe(channel_name)` — the channel name from `getConfig`.
4. Pass the client into `ConversationComponent` as `rtmClient`.

If RTM `login` fails, the whole call fails — there is no fallback path; transcripts cannot flow without RTM in this quickstart.

## Token Renewal Sequence

RTC fires `token-privilege-will-expire` ~30 seconds before expiry. `ConversationComponent` attaches a handler that calls `LandingPage.handleTokenWillExpire`:

```ts
async function handleTokenWillExpire(joinedUid: UID) {
  if (!joinedUid) return; // skip if RTC never reported a uid
  const [rtcConfig, rtmConfig] = await Promise.all([
    getConfig({ channel: agoraData.channel, uid: joinedUid }),
    getConfig({ channel: agoraData.channel, uid: agoraData.uid }),
  ]);
  await client.renewToken(rtcConfig.token);
  await rtmClient.renewToken(rtmConfig.token);
}
```

Why two `getConfig` calls?

- The RTC `client.uid` may differ from the original `agoraData.uid` if the SDK assigned one.
- RTM is logged in with `agoraData.uid` and rejects tokens issued for a different UID.

`ConversationComponent` only invokes the handler when `joinedUid` (RTC's reported UID) is truthy. If RTC has not yet reported, renewal is silently skipped — RTC will reconnect on its own when the channel is re-joined.

## Ending the Call

Ending the conversation starts in `ConversationComponent.handleEndConversation()` and then calls `LandingPage.handleEndConversation()`:

1. Unpublish the microphone track from RTC.
2. Stop and close the local microphone track.
3. Call `stopAgent(agoraData.agentId)` if there is one; failures are logged and the UI still tears down.
4. Call `rtmClient.logout()`.
5. Reset state (`setAgoraData(null)`, `setRtmClient(null)`, `setShowConversation(false)`).

The order matters: stopping the agent before RTM logout means the user keeps receiving transcripts of any final agent words.

## StrictMode Considerations

- The `AgoraRTCProvider` is `dynamic`-imported with `{ ssr: false }`.
- The RTC client lives in `useRef`. Switching to `useMemo` recreates the client during StrictMode's fake-unmount and breaks `useJoin`.
- `useJoin` and `useLocalMicrophoneTrack` must be gated until the dynamic import resolves (`isReady`-style pattern in `ConversationComponent`).

## Failure Modes

| Symptom                                              | Cause                                                              |
| ---------------------------------------------------- | ------------------------------------------------------------------ |
| `/api/get_config` 404                                | `AGENT_BACKEND_URL` not set → no rewrites registered.              |
| `400 channelName required` from `/startAgent`        | Browser sent empty or wrong-named field.                           |
| Transcript empty, agent state stuck                  | RTM not subscribed or `subscribeMessage(channel)` never ran.       |
| Agent disconnects exactly at 1 hour                  | Renewal skipped (likely `joinedUid` was 0); check RTC join logs.   |
| Microphone busy on second mount in dev               | RTC client was recreated (probably `useMemo` regression).          |
| RTM rejects renewal                                  | Renewal token was built for the wrong UID — verify two-fetch path. |

## See Also

- [Back to Architecture](../02_architecture.md)
- [Back to Gotchas](../07_gotchas.md)
- [Managed Agent Config](managed_agent_config.md)
