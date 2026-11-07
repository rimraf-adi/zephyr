# 03: LLM Client Layer, Protocols & Multi-Model Routing

> **Focus Area:** OpenAI-Compatible Client, Server-Sent Events (SSE), Protocols (MCP, ACP), Rate Limiting, Exponential Backoff, Fallback Chains.

---

## 1. LLM Client Architecture: Thin Custom Client vs. Heavy SDK

When designing an agent harness for diverse backends (such as FreeLLMAPI, OmniRoute, self-hosted vLLM, or proprietary APIs), avoiding heavy framework abstractions is critical:

| Approach | Trade-offs | Decision |
|---|---|---|
| **High-level Frameworks (`rig-core`, `langchain-rs`)** | High abstraction, opinionated agent loops, inflexible delta handling. | Avoid for custom harnesses. |
| **Official SDKs (`async-openai`)** | Good typed wrappers, but assume standard OpenAI behavior. | Viable baseline, but struggles with router quirks. |
| **Thin Custom Client (`reqwest` + SSE)** | ~300 lines of code; full control over streaming deltas, usage token parsing, and custom retry logic. | **Recommended.** |

---

## 2. Server-Sent Events (SSE) Streaming & Delta Assembly

Streaming is essential for fast perceived latency and live UI feedback. The client consumes an SSE stream and aggregates deltas in real time:

```
Provider Stream ──► SSE Parser ──► Delta Accumulator ──► Event Broadcast
                                       │
                      ┌────────────────┴────────────────┐
                      ▼                                 ▼
             Content Chunk (Text)             Tool Call Delta (JSON)
                      │                                 │
                      ▼                                 ▼
              RunEvent::TextDelta              Assemble JSON Buffer
                                                        │
                                                        ▼
                                               RunEvent::ToolStart
```

### Critical Gotcha: Fragmented Tool-Call JSON
Free and routed models often stream tool call arguments across multiple tiny chunks (`{"name": "read"`, `", "path": "src/main.rs"}`). The client must buffer and assemble arguments until `finish_reason: "tool_calls"` before attempting JSON parsing.

---

## 3. Extensible Protocols: MCP & ACP

### Model Context Protocol (MCP)
- **Role:** Industry standard for connecting agents to external databases, documentation searchers, and third-party tools.
- **Implementation in Rust:** Use the official `rmcp` SDK.
- **Client Mode:** Spawns local child processes via stdio or connects over HTTP SSE.
- **Server Mode:** Expose Zephyr's internal code index (Loupe) as an MCP server for external agents.
- **Context Economy Rule:** **Never preload all MCP schemas by default.** Preload only a minimal catalog; load full schemas dynamically on demand.

### Agent Client Protocol (ACP)
- Standardized protocol allowing code editors (such as Zed) to communicate with and drive external agent harnesses.
- Implemented as an optional adapter over Zephyr's internal event stream.

---

## 4. Multi-Model Routing, Rate Limits & Fallback

```
                    ┌─────────────────────────┐
                    │      Turn Request       │
                    └────────────┬────────────┘
                                 │
                                 ▼
                    ┌─────────────────────────┐
         ┌──────────┤   Primary Model (Fast)  ├──────────┐
         │          └─────────────────────────┘          │
    Success                                            Failure (429, 5xx, Timeout)
         │                                               │
         ▼                                               ▼
┌─────────────────┐                             ┌─────────────────────────┐
│ Continue Loop   │                             │ Exponential Backoff     │
└─────────────────┘                             │ + Jitter (Retry-After)  │
                                                └────────────┬────────────┘
                                                             │ Max retries exceeded
                                                             ▼
                                                ┌─────────────────────────┐
                                                │ Secondary Fallback Model│
                                                │ (Log Fallback Event)    │
                                                └─────────────────────────┘
```

1. **Backoff & Jitter:** Handle HTTP `429` (Too Many Requests) and `503` (Service Unavailable) by honoring the `Retry-After` header with full jitter.
2. **Idle Stream Timeouts:** Free tiers occasionally drop TCP connections silently mid-stream. Enforce a per-chunk idle timeout (e.g., 15 seconds) to trigger immediate reconnection.
3. **Usage Field Normalization:** Different OpenAI-compatible endpoints format token usage differently (some omit cached tokens, others place usage in the final chunk). The client normalizes all formats into a single `TokenUsage` struct.
