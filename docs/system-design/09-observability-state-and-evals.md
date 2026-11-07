# 09: Observability, State Persistence & Evaluation Harnesses

> **Focus Area:** Structured Tracing, JSONL Event Sourcing, Session Replay, Benchmarking (Terminal-Bench, SWE-bench), Agent Scorecard Metrics.

---

## 1. Observability: Why Tracing is Mandatory

In an agentic system, silent failures are impossible to debug without deep instrumentation:
- **`tracing` Spans:** Wrap every major operation: `span!(Level::INFO, "turn", turn_id = turn_num)`, `span!(Level::DEBUG, "tool_execution", tool = name)`.
- **Latency Attribution:** Explicitly record:
  - Time spent waiting for LLM completion tokens (TTFT + streaming time).
  - Time spent executing local tools (disk I/O, compilation, shell commands).
- **Token Accounting:** Log prompt tokens, completion tokens, and cache-hit tokens per turn.

---

## 2. State & Session Storage: Append-Only JSONL

Rather than complex relational database schemas, state persistence is best modeled as an **append-only event log (JSONL)**:

```jsonl
{"timestamp": "2026-11-07T12:00:01Z", "event": "session_started", "session_id": "ses_01", "model": "gpt-4o"}
{"timestamp": "2026-11-07T12:00:02Z", "event": "user_message", "content": "Fix compile error in src/main.rs"}
{"timestamp": "2026-11-07T12:00:05Z", "event": "tool_call", "tool": "cargo_check", "args": {}}
{"timestamp": "2026-11-07T12:00:06Z", "event": "tool_result", "output_ref": ".zephyr/scratch/call_1.log"}
```

### Advantages for System Design:
- **Crash Safety:** Appends are atomic; a crash never corrupts historical turns.
- **Time Travel & Branching:** Easily fork a session from turn $N$ by replaying up to line $N$.
- **Offline Inspection:** Greppable with standard unix tools (`jq`, `grep`).

---

## 3. Evaluation & Benchmarking Strategy

Relying on subjective human prompting to judge an agent leads to regressions. A rigorous harness requires automated evaluation:

### Public Benchmarks
1. **Terminal-Bench (Stanford/Laude):** Containerized real-world terminal tasks tested via the Harbor harness. Evaluates CLI manipulation and tool usage.
2. **SWE-bench (Verified / Lite):** GitHub pull-request resolution benchmark. Tests bug resolution across large Python repositories.

### The Mock Provider & Trajectory Replay
To test the harness without incurring LLM API costs:
- **Mock Provider:** Returns deterministic, pre-scripted tool calls to test loop logic, circuit breakers, and compaction.
- **Cassette Replay:** Record real HTTP streams once and replay them in CI using `wiremock` or snapshot testing (`insta`).

---

## 4. The Agent Scorecard

Track these objective metrics across a benchmark suite to prove that harness improvements actually produce results:

| Metric | What it Measures | Target Direction |
|---|---|---|
| **Task Resolve Rate** | Percentage of tasks passing validation tests | $\uparrow$ Maximize |
| **Turns per Task** | Number of model invocations to reach resolution | $\downarrow$ Minimize |
| **Token Cost ($)** | Total prompt + completion tokens consumed | $\downarrow$ Minimize |
| **KV Cache-Hit Rate** | Ratio of prompt tokens served from prefix cache | $\uparrow$ Maximize ($>80\%$) |
| **Malformed Tool Calls** | Count of schema parse failures | $\downarrow$ Zero |
| **Edit Re-read Rate** | Frequency of re-reading files already in context | $\downarrow$ Minimize |
| **Wall-Clock Time** | End-to-end task duration | $\downarrow$ Minimize |
