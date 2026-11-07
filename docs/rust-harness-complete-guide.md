# Building a custom coding-agent harness in Rust: complete guide

One file, two parts.

- **Part A: Toolkit.** Existing harnesses, crates, tool implementations, the agent loop and context management, tool-call reliability, security, testing, already-published techniques, candidate differentiators, and a scorecard.
- **Part B: Borrow vs. build.** What to depend on, what to reimplement, what to build yourself, which design decisions to integrate from Pi, mini-SWE-agent, OpenCode, Codex, Goose, OpenHands and Aider, what to avoid, a crate map, and a build order.

Numbering: Part A sections are A0-A25 (A21 is unused). Part B sections are B1-B8. Inside Part A, "section 10" means A10. Inside Part B, an unlabeled "section N" means B(N) and references to Part A are labeled explicitly.

Legend: **[searched]** = seen in web results during research. **[known]** = from my own knowledge, not re-verified. **[inference]** = my conclusion, not a published claim.

---

## Contents

- A0. TL;DR recommended stack
- A1. Whole harnesses to read, borrow from, or fork
- A2. LLM client layer
- A3. Protocols
- A4. Tool implementations
- A5. Code understanding (where Loupe fits)
- A6. Terminal UI
- A7. Config, state, observability
- A8. Suggested workspace layout
- A9. Build order
- A10. The agent loop and context management (the part most hobby harnesses get wrong)
- A11. Tool-call reliability with free, small, or routed models
- A12. Routing, rate limits, and fallback
- A13. Git integration, checkpoints, and undo
- A14. Edit feedback: diagnostics, tests, formatters
- A15. Project instructions, skills, hooks, subagents
- A16. Security and safety
- A17. Testing and evaluation
- A18. Observability and session storage
- A19. Web tools, config, packaging
- A20. Open design decisions to settle early
- A22. Already published: adopt these, do not reinvent them
- A23. Candidate differentiators (no exact prior art found, but not proof of novelty)
- A24. Scorecard: how you will know it is the best you have built
- A25. Caveats
- B1. The rule of thumb
- B2. Component-by-component: borrow or build
- B3. Design decisions to integrate from existing harnesses
- B4. Design decisions to avoid
- B5. What to build (your custom design)
- B6. Proposed architecture (crate map)
- B7. Suggested order of work (borrow first, build the differentiators second)
- B8. Caveats

---

# Part A: Toolkit (crates, techniques, loop design)

Scope: a lighter, model-agnostic harness in Rust, with FreeLLMAPI / OmniRoute (OpenAI-compatible hosts) as the model backends, plus room to plug in the indexer idea ("Loupe").

Legend: **[searched]** = verified in web results this session. **[known]** = from my own knowledge of the ecosystem, so double-check versions on crates.io before depending on them.

---

## A0. TL;DR recommended stack

| Layer | Pick | Why |
|---|---|---|
| Async runtime | `tokio` | Everything below assumes it |
| LLM client | thin custom client over `reqwest` + SSE, **or** `genai` | Your hosts are OpenAI-compatible; a harness needs raw control over streaming, tool-call deltas, and cache/usage fields |
| Tool schemas | `serde` + `schemars` | JSON Schema generated from Rust structs |
| MCP | `rmcp` (official SDK) | Client for external tools, server for Loupe |
| Editor protocol | `agent-client-protocol` (optional, later) | Lets Zed and other editors drive your harness |
| Search/files | `ignore`, `grep-searcher`/`grep-regex`, `globset` | Ripgrep's own libraries |
| Editing | `similar` (diffs) + your own apply-patch/multi-edit tool | Batch edits, one round trip |
| Code structure | `tree-sitter` + per-language grammars | Outlines, symbols, repo map |
| Shell + sandbox | `tokio::process`, `portable-pty`; Seatbelt (macOS) / Landlock (Linux) | You are on macOS, so Seatbelt matters first |
| TUI | `ratatui` + `crossterm` | De facto standard |
| State | `rusqlite` (sessions) or flat JSONL | Simple, greppable |
| Context management | tool-output truncation to disk + compaction, append-only history | Biggest quality and cost lever (section 10) |
| Tool-call repair | own validate/coerce/retry layer (`jsonschema`) | Free and small models emit malformed calls (section 11) |
| Git | shell out to `git` for writes, `gix` for read-only | `gix` is still missing some write features (section 13) |
| Diagnostics | `cargo check --message-format=json` first, LSP later | Cheap, deterministic edit feedback (section 14) |
| Evals | replay tests + your own task set + Terminal-Bench/SWE-bench | Harness choices change scores a lot (section 17) |

---

## A1. Whole harnesses to read, borrow from, or fork

Read these before writing your own agent loop. Most of the hard design decisions are already visible in them.

- **OpenAI Codex CLI (`codex-rs`)** **[searched]**: the most complete Rust reference. A large Cargo workspace with an agent loop in `core`, a ratatui TUI, headless `exec`, an `app-server` JSON-RPC mode, MCP support, platform sandboxes (Seatbelt on macOS, Landlock/seccomp or Bubblewrap on Linux), and small reusable crates like `codex-apply-patch`, `codex-file-search`, `codex-shell-command`. Best source of truth for sandboxing and patch application. It is big, so cherry-pick ideas instead of forking the whole thing.
- **VT Code** **[searched]**: a Rust terminal coding agent with tree-sitter and ast-grep code intelligence, an execution policy, lifecycle hooks, token-budget tracking, and OS-native sandboxing. It uses `rig` for LLM calls.
- **openbuild** **[searched]**: a model-agnostic agent shell in Rust with a crate split that is close to what you want (tools / config / permissions / sandbox / tui). Early stage, but a useful layout reference.
- **uira** **[searched]**: a standalone Rust agent with multi-provider routing, Seatbelt/Landlock sandboxing, a permission-rule engine, and an MCP client.
- **crab-code** **[searched]**: a 16-crate Rust workspace modeled on Claude Code's workflow. Useful as a tool-system and permission-mode reference (`trait Tool`, registry/executor, enum-dispatched LLM backends).
- **goose (Block)** **[searched, mentioned as a harness]**: Rust agent that the official MCP SDK was originally built for.
- **`agent-harness` crate** **[searched]**: drives Claude Code, Codex, OpenCode and ACP agents behind one `Harness` trait with a normalized event stream. It also ships a built-in agent for OpenAI-compatible open models that runs the tool loop in Rust. This is useful if you want to wrap other agents, not if you want your own loop, but its `RunEvent` vocabulary (text, thinking, tool start/end, plan, usage with cache tokens) is a good event model to copy.

---

## A2. LLM client layer

Your backends (FreeLLMAPI, OmniRoute) are OpenAI-compatible routers, so one chat-completions client with a configurable `base_url` covers everything.

| Option | Notes |
|---|---|
| **`genai`** **[searched]** | Multi-provider client, native protocol per provider, streaming and usage metadata, custom endpoint/auth/model mapping. MIT. Latest activity in 2026. Good default if you want less code. Note: it normalizes chat APIs and is not meant to expose every provider-specific feature. |
| **`rig-core`** **[searched]** | Higher-level: agents, tool calling, streaming, 20+ providers, vector stores. Used by VT Code. More opinionated and more abstraction than a custom harness loop usually wants. |
| **`async-openai`** **[known]** | Typed OpenAI API client, works with any compatible `base_url`. |
| **Raw `reqwest` + SSE** (`eventsource-stream` or `reqwest-eventsource`) **[known]** | About 300 lines for chat completions with streamed tool-call deltas. Gives full control over retries, cancellation, usage accounting, and quirks of free or routed models. |

Recommendation: own a small `Provider` trait with one OpenAI-compatible implementation over `reqwest`. Free/routed models are often inconsistent about tool-call streaming, so you will want to patch behavior yourself, which a heavy abstraction makes harder. Use `genai` if you want multi-provider quickly.

Also: `tiktoken-rs` **[known]** for approximate token counts (a 4-chars-per-token estimate is fine for budgets), `tokio-util` `CancellationToken` for Ctrl-C and timeouts, `backoff` or `tower` retry for flaky free tiers.

---

## A3. Protocols

- **MCP, `rmcp`** **[searched]**: official Rust SDK, tokio-based. Server side has a `#[tool_router]` macro; client side can spawn a child-process MCP server over stdio. Use as a client (load user MCP servers) and as the server for Loupe. An alternative is `rust-mcp-sdk` **[searched]**, with stdio and streamable HTTP transports.
- **ACP, Agent Client Protocol** **[searched]**: standardizes editor-to-agent communication. Official Rust SDK crates: `agent-client-protocol` (roles, builders, protocol types), `-http` transports, `-rmcp` bridge, plus a conductor for proxy chains. Add later if you want Zed/editor integration. Until then, a headless `exec` mode with JSON events is enough.

---

## A4. Tool implementations

**Files and search**
- `ignore` (gitignore-aware walker), `globset`, `grep-searcher` + `grep-regex` (ripgrep internals) **[known]**
- Codex's `codex-file-search` is a reference for fuzzy file search **[searched]**

**Editing**
- `similar` for diffs, `diffy` for applying unified diffs **[known]**
- Study `codex-apply-patch` **[searched]** for a robust patch-format parser
- Design tip from your screenshot: expose a **multi-edit** tool (many replacements in one call) and a **read whole file** default for small files, so the model does not need a dozen small reads

**Shell**
- `tokio::process` for simple commands, `portable-pty` **[known]** for interactive/PTY commands
- Command parsing for permission rules: see `codex-shell-command` **[searched]**

**Sandboxing**
- macOS: wrap commands with `/usr/bin/sandbox-exec` (Seatbelt profiles). Codex does exactly this **[searched]**.
- Linux: Landlock + seccomp, optionally Bubblewrap **[searched]**
- `vetto` **[searched]** is a small daemon-less Landlock/Seatbelt layer for coding agents. Check its maturity before depending on it.

**Permissions**
- A rule engine like `Tool(arg-glob)` allow/ask/deny (seen in openbuild and uira) **[searched]**. Small to write yourself.

---

## A5. Code understanding (where Loupe fits)

Existing indexers already cover a lot, so treat these as dependencies or competitors, not blank space **[searched]**:

- Aider-style repo map: tree-sitter tags + PageRank + budget fitting. Rust ports: `repo-mapper`
- MCP code-map servers: `codeix`, `indxr` (Rust), `stacklit`, `basemind`, `codebase-memory-mcp`, `leindex` (Zoekt-style trigram index + optional SCIP from rust-analyzer), `CodeNib`

Building blocks if you write your own:

| Need | Crate / tool |
|---|---|
| Parsing and outlines | `tree-sitter` + `tree-sitter-rust`, `-python`, `-typescript`, etc.; per-language `tags.scm` queries (Aider's vendored ones are a good start) |
| Exact/regex index | build a trigram index (Zoekt, Cursor's write-up) **[searched]**; or just use `grep-searcher` first and add the index only when you measure it is needed |
| Graph ranking | `petgraph` **[known]** for the symbol graph and PageRank |
| Precise symbols | SCIP via `rust-analyzer scip` **[searched]**, or LSP |
| Storage | `fjall` or `rusqlite` **[known]** |
| Incremental updates | `notify` **[known]** file watcher + content hash per file |
| Fast hashing | `blake3` **[known]** |

---

## A6. Terminal UI

- `ratatui` + `crossterm` **[searched]**: used by Codex, VT Code, uira, crab-code, openbuild
- Alternatives for a lighter, non-fullscreen feel: `reedline` or `rustyline` **[known]** with plain streamed output
- Markdown/syntax rendering: `termimad`, `syntect`, or `tree-sitter-highlight` **[known]**

A plain streaming CLI (no fullscreen TUI) is the fastest way to get a working harness. Add ratatui once the loop is solid.

---

## A7. Config, state, observability

- `serde` + `toml` for config; `directories` for config paths **[known]**
- `clap` for CLI; `anyhow` + `thiserror` for errors **[known]**
- `tracing` + `tracing-subscriber` **[known]**: log every request, tool call, token count, and wall time. You need this to prove "faster and cheaper", which is the point of a lighter harness.
- Session persistence: JSONL transcript per session, or SQLite (Codex uses SQLite for state) **[searched]**

---

## A8. Suggested workspace layout

```
harness/
  crates/
    llm/        Provider trait, OpenAI-compatible client, streaming, usage
    core/       agent loop, context budget, compaction, events
    tools/      read, multi_edit, grep, glob, shell, apply_patch
    perms/      allow/ask/deny rule engine
    sandbox/    seatbelt (+ landlock later)
    mcp/        rmcp client wrapper
    loupe/      indexer + MCP server (can also run in-process)
    tui/        ratatui frontend (later)
    cli/        binary: interactive, `exec`, `--json`
```

Keep `loupe` usable both as an in-process library and as a separate MCP server. The in-process path avoids a protocol round trip and lets the harness inject a repo map into the prompt directly.

---

## A9. Build order

0. Before any code: write 10-20 real tasks from your own repos (with a pass/fail check each). Every later change is measured against them (section 17).
1. `llm` crate: streaming chat with tool-call deltas against FreeLLMAPI/OmniRoute. Test with several free models, since quirks show up here first.
2. Agent loop with three tools: `read_file` (whole file under N tokens, else outline + ranges), `multi_edit`, `shell`.
3. Context budget: token counting, tool-output truncation, simple compaction.
4. Permissions, then Seatbelt sandbox.
5. Loupe v0: tree-sitter outline + `symbol` lookup returning body plus callee signatures in one call.
6. Event log + metrics (tokens, tool calls, wall time per task). Benchmark against your current agent on the same tasks.
7. TUI, MCP client, ACP, trigram index, SCIP: only if the metrics say they matter.

---

## A10. The agent loop and context management (the part most hobby harnesses get wrong)

Sections 1-9 list parts. This section is the design that decides cost, speed, and quality.

**Turn assembly.** Each turn builds: system prompt, task, tool schemas, current plan, managed history. A recent empirical study of coding-harness design (176 matched configurations, 4 models, 2 benchmarks) varied planning, action space, and context management with the loop fixed. It found context management gives its largest gains when context is tight, because it prevents overflow-related termination **[searched]**. Its design keeps the system prompt, task and recent turns verbatim, injects the current plan fresh each turn instead of appending it, and lets older tool outputs be compacted.

**Tool-output handling (do this first).**
- Cap every tool result. Return a truncated preview, write the full output to a file on disk, and tell the model how many bytes/lines are on disk and how to read ranges. Glean's harness does exactly this **[searched]**.
- Optionally add a `recall` tool so the model can fetch discarded output on demand. The same study's T2 tier tests this idea **[searched]**.
- A noisy tool is the usual cause of context blowup, not long user prompts. Pruning inside the runtime (dedupe, strip ANSI, collapse repeated lines, keep errors and the tail of build logs) is cheap and effective **[searched, research direction]**.

**Compaction.** Tiers worth implementing in order:
1. No compaction, fail clearly at the limit (baseline).
2. Truncate old tool outputs, keep the messages.
3. Disk fallback + `recall`.
4. Summarize older history with a separate model call into a running summary.
5. Full context reset with a handoff artifact (goal, decisions, files touched, open TODOs). Anthropic's harness research reports this sometimes beats compaction on long sessions **[searched, secondary source]**.

Compaction loses details, so keep durable facts in project instructions (section 15), which are re-read every turn.

**Prompt caching and prefix stability.** Claude Code's team says it designs its whole harness around prompt caching and treats cache-hit drops as incidents **[searched, secondary source]**. Cache-friendly rules apply even to other providers:
- Keep history **append-only**. Never edit or reorder earlier messages.
- Put static content first (system prompt, tool schemas), dynamic content last (date, cwd, git status).
- Do not reorder or change tool schemas mid-session.
- Run compaction with the same prefix as the parent request, adding the summarize instruction as the final message.
- Check whether your hosts (FreeLLMAPI/OmniRoute and the models behind them) expose cache usage fields. If a router rotates models between turns, you lose the cache, so consider pinning a model per session.

**Loop robustness.**
- Recover from `max_output_tokens` truncation by auto-continuing, and from "prompt too long" (HTTP 413 or provider-specific error) by compacting and retrying **[searched, ClawHarness feature list]**.
- Run read-only tool calls in parallel, writes serially.
- Detect loops (same tool + same args N times) and break out with a nudge or stop.
- Hard limits: max turns, max wall time, max tokens, per task.
- Cancellation via `CancellationToken` threaded through the LLM stream, tools, and child processes.

---

## A11. Tool-call reliability with free, small, or routed models

Your backends are free/routed, so malformed tool calls will happen. Plan for it.

- **Validate, coerce, retry.** Check arguments against the tool's JSON Schema (`jsonschema` crate **[known]**). Coerce common errors deterministically (stringified objects/arrays, numbers wrapped in quotes), fuzzy-match tool names, then on failure return a precise error message and let the model retry. A proxy called `toolrails` documents exactly these failure modes for small models on Ollama **[searched]**; the same repair logic is worth building into your client.
- **Constrained decoding only helps if the backend supports it** (llama.cpp/Ollama grammars, vLLM guided decoding). Through an OpenAI-compatible router you often cannot use it. Also be careful: forcing the whole response through a schema can make models stop calling tools at all **[searched, reports of a measured "constraint tax"]**. Constrain arguments, not the decision to call.
- **Schema shape matters.** Keep tool schemas flat and small. A recent write-up found newer frontier models inventing extra fields in a nested `edits[]` array for one harness's edit tool, plausibly because post-training favored another harness's edit-tool shape **[searched]**. Test your edit tool's schema against every model you plan to use.
- **Per-model capability profile.** Keep a table (tool-call reliability, effective context, supports parallel calls, supports cache, quirks) and give weaker models more scaffolding (simpler schemas, tighter context, extra verification pass). A Go project, `flynn`, describes this as adapting the loop to model strength **[searched]**; the idea transfers.
- **Edit tool design.** Prefer exact-match search/replace or a patch format over line numbers (models miscount). Fail loudly with the closest match shown when the old text is not found, and support multiple edits per call.

---

## A12. Routing, rate limits, and fallback

- Handle `429` and `5xx` with exponential backoff plus jitter, honoring `Retry-After`.
- Fallback chain: if a model fails N times or times out, move to the next model and **say so in the event log**.
- Track per-model usage, latency, and failure rate in the session log so you can see which free model is actually worth using.
- Per-request timeouts and a total stream-idle timeout (free hosts stall silently).
- Normalize usage fields (prompt/completion/cached tokens) into one struct even when a host omits them, falling back to estimates.

---

## A13. Git integration, checkpoints, and undo

- **Checkpoint before edits.** Snapshot affected files (or use `git stash create` / a shadow commit) before each turn that writes, so the user can rewind. Claude Code snapshots files before changes for revert **[searched, secondary source]**.
- **Worktrees** for parallel subagents or risky tasks, so edits do not collide.
- **Which Rust git library.** `gix` (gitoxide) is pure Rust, fast, good for read-only work (status, diff, ignore rules, rev-parse). Its docs list commit creation with hooks, merge, rebase, blame, and full worktree checkout as in development **[searched]**. `git2` is a mature libgit2 binding but needs a C dependency **[searched]**. Practical default: **shell out to the `git` CLI for anything that writes**, and use `gix` for fast read-only queries (or just shell out for everything at first).
- Never auto-commit or push without an explicit user setting. Show the diff first.

---

## A14. Edit feedback: diagnostics, tests, formatters

An agent that only writes and never checks wastes turns.

- **Cheapest loop for Rust projects:** after edits, run `cargo check --message-format=json` and feed back only errors, trimmed. Deterministic and no LSP process to manage.
- **LSP diagnostics** are the general solution. Crates: `lsp-types` for protocol types **[known]**, `async-lsp` **[searched via lingshu-lsp]**, `tokio-lsp` **[searched]**. Examples of agents with LSP clients: `dirge-agent` (a minimal Rust coding agent aimed at low memory use) and `lingshu-lsp` **[searched]**. Gotcha from `dirge-agent`'s client: ignore diagnostics for an older document version, because stale "clean" results hide real errors **[searched]**.
- Auto-format on write (`rustfmt`, `prettier`, `ruff`) so diffs stay small and models stop fighting whitespace.
- Make "run the tests" a first-class, cheap tool that returns a short pass/fail summary plus the first failures, not the whole log.

---

## A15. Project instructions, skills, hooks, subagents

- **Instruction files:** support layered `AGENTS.md`-style files (global → repo → directory), re-read every turn so they survive compaction **[searched, ClawHarness and Claude Code descriptions]**.
- **Hooks:** `PreToolUse` / `PostToolUse` / `UserPromptSubmit` that run a shell command and return allow / deny / modify (crab-code implements these) **[searched]**. Cheap to build, very powerful for policy, formatting, and logging.
- **Skills:** folders with a markdown file loaded on demand, so instructions do not sit in context all the time.
- **Subagents:** isolated context per subagent, returning only a short result to the parent **[searched]**. Add after the single-agent loop is solid.
- **Plan mode / todo tool:** a lightweight task list the model updates helps keep long runs on track. The study above found planning is one of the three levers that matter **[searched]**.

---

## A16. Security and safety

- **Treat tool output and fetched content as data, not instructions.** Prompt injection arrives via files, web pages, issue text, and MCP tool results. Never let it change permissions or trigger network/shell actions without a user-facing check.
- **Path confinement:** resolve symlinks and `..`, deny writes outside the workspace by default.
- **Command policy:** allow/ask/deny lists plus a sandbox (section 4). Block obvious foot-guns (`rm -rf` on broad paths, `git push --force`, `curl | sh`).
- **Network egress:** off by default inside the sandbox; allow-list package registries if needed.
- **Secrets:** scrub env vars passed to child processes, redact keys and tokens in logs and transcripts, never echo API keys. Your own FreeLLMAPI/OmniRoute keys should live in the OS keychain or a locked-down config file, not in session logs.
- **MCP servers are code execution.** Show what a configured server can do and require approval for new ones.

---

## A17. Testing and evaluation

**Test the harness itself (no real model needed):**
- A **mock provider** that returns scripted tool-call sequences, to unit-test the loop, permissions, truncation, compaction and cancellation.
- **Record/replay** of real provider streams (cassette files) for regression tests; `wiremock` for HTTP mocks and `insta` for snapshot tests **[known]**.
- Property tests (`proptest`) for the patch/edit applier and path-confinement logic **[known]**.

**Evaluate the agent:**
- **Your own task set first** (step 0 in section 9). Public benchmarks do not reflect your repos or your free models.
- **Terminal-Bench** (Stanford/Laude; 89 tasks in Terminal-Bench 2.0, containerized terminal tasks, run through the Harbor harness) and **SWE-bench** variants are the public references **[searched]**. Read scores critically: they depend on both the model and the harness, the SWE-bench Verified set has contamination concerns, and a commonly cited rule is to compare only scores produced by the same harness **[searched, secondary sources]**. One source claims harness engineering alone moves Terminal-Bench scores by 20+ points for the same model **[searched, single secondary source, treat as indicative]**.
- **Metrics to log per run:** success, input/output/cached tokens, number of tool calls, number of model turns, wall time, retries, malformed tool calls. These directly measure the "lighter and faster" goal.
- Run each task several times; models are stochastic.

---

## A18. Observability and session storage

- `tracing` spans per turn, per model call, per tool call; optional OpenTelemetry export via `tracing-opentelemetry` **[known]**.
- Store the full transcript (messages, tool calls, raw tool output paths, usage) as JSONL so sessions can be resumed, forked, and replayed.
- A simple `--json` event stream on stdout makes the harness scriptable and easy to wrap in an editor later (compare the `RunEvent` vocabulary in section 1).

---

## A19. Web tools, config, packaging

- **Web fetch/search tools:** `reqwest` + an HTML-to-markdown crate (`htmd`, `html2md`, or `readability`-style extraction **[known]**), with size caps and the same truncation-to-disk rule. Mark fetched content as untrusted (section 16).
- **Config layering:** defaults → global → project → CLI flags/env; `serde` + `toml`. Keep model/host definitions (base URL, key source, capability profile) in config, not code.
- **Packaging:** a single static binary. `cargo-dist` for release builds and installers, `cross` for Linux targets **[known]**. Apple Silicon is a first-class target for you, so build and test there from day one.

---

## A20. Open design decisions to settle early

1. In-process Loupe or MCP server (or both)? Recommended: both, one library.
2. Own LLM client or `genai`? Recommended: own thin client; swap later if it stays boring.
3. Full-screen TUI or streaming CLI first? Recommended: streaming CLI first.
4. How aggressive is truncation, and what is the per-model context budget?
5. Read-whole-file threshold (tokens) before falling back to outline + ranges.
6. What the model sees on edit failure.
7. Which tools are parallel-safe.
8. Pin a model per session (cache) or route per turn (resilience)?

---

## A22. Already published: adopt these, do not reinvent them

Most "clever" harness ideas already have prior art. These are table stakes for a best-in-class harness. Each was checked in web searches this session.

| Technique | What exists | What to take |
|---|---|---|
| **Edit-format choice per model** | A blog benchmark ("The Harness Problem", Feb 2026) found patch formats worst for nearly every one of 16 models, hashline (line-addressed by short content hash) matching or beating plain replace, and weak models gaining most; one cited case went from 6.7% to 68.3% on the same model by changing only the edit tool **[searched, single blog + secondary summaries]**. SWE-Edit (Microsoft) found format should be a per-task decision and that splitting a Viewer and an Editor with clean contexts raised resolve rate 2.1 pp and cut cost 17.9% **[searched]**. A tool-schema-bias paper reports success from total failure to 97% depending only on schema **[searched]**. | Implement 2-3 edit formats (exact replace, hashline, whole-file rewrite) behind one tool interface and test each per model. |
| **Speculative tool execution** | PASTE (pattern-aware, learns tool-call sequences; reported about 1.8x tool throughput), Spex (zero-token transition table for test/lint prefetch), TomasuLLM (out-of-order execution with commit validation), and `zeph`, a Rust agent that dispatches tools while the model is still streaming partial tool-call JSON **[searched]**. | Start with the cheap version: dispatch read-only tools as soon as their streamed JSON is complete, and run independent calls in parallel. |
| **Read dedup** | `lingshu-tools` (Rust) returns a "file unchanged since last read" stub using an mtime cache and warns at 3 identical consecutive reads; Claude Code uses a similar pattern **[searched]**. | Copy it. It is about 100 lines. |
| **Model cascade** | PLANTOD and lean-agent route planning/review to a strong model and bulk edits to a cheap one; an agent-cost-optimizer report on SWE-bench traces claims 62.5% cost cut with a 10.8 pp quality drop for pure cascade, versus 36.9% cut with 3.4 pp drop when escalation is driven by execution feedback **[searched, single report, treat as indicative]**. The same report calls first-step routing mistakes the hardest to recover from. | Escalate on execution feedback (tests fail, repeated edit failure), not on a model's own confidence. |
| **Symbol-level editing and navigation** | Serena (LSP-backed `find_symbol`, `replace_symbol_body`, `insert_after_symbol`) **[searched]**. | Offer symbol-addressed edits from Loupe's tree-sitter spans. Cheaper than LSP and works offline. |
| **Self-improving harness** | AHE, TRIAGE, Self-Harness, PolyHarness; Thoughtworks Radar placed the technique at "Trial" in Nov 2026. Common ingredients: observability, harness parts as separate editable files, and a regression gate before accepting a change **[searched]**. | Do not build the evolver now, but **make it possible**: keep system prompt, tool descriptions, hooks, and memory as separate versioned files, and log full traces. |
| **Structured compaction summaries** | A CooperBench PR measured file+symbol coverage after compaction rising from 40% to 57% by using a template with FILE MAP, KEY SYMBOLS (name → path:line), SEARCH RESULTS including negative results, EDITS ALREADY APPLIED, and BUILD/TEST OUTPUT (small sample: 9 segments, one 9B model) **[searched]**. Deletion-based "verbatim" compaction is also promoted by Morph (vendor claims) **[searched]**. | Use a structured template with negative results and symbol locations. |
| **Pre-built code graph** | CodeGraph reports about 50% fewer tool calls and 54% fewer tokens on one repo, and says agents must be steered to use the graph, otherwise sub-agents read files anyway **[searched, vendor benchmark]**. | Since you own both the harness and the index, you control the tool surface directly: no steering prompt needed. This is an advantage of building them together (my inference). |

---

## A23. Candidate differentiators (no exact prior art found, but not proof of novelty)

I searched for these specific combinations and did not find a direct match. Each has adjacent work, listed so you can read it and judge. Treat them as hypotheses to test, not as established ideas.

**A. Graph-predicted speculation.** Use Loupe's symbol/reference graph (not a learned trajectory table) to predict what the model will ask next. After `symbol(foo)`, pre-stage its callees' signatures and the tests that reference it, and after an edit, pre-run the affected tests. Pre-staged results are cached and only enter context if the model asks.
- Adjacent: PASTE and Spex predict from past trajectories; SpecAgent builds speculative context at index time for code *completion*; `zeph` dispatches from partial streamed JSON. None found that uses a code graph as the predictor inside an agent loop.
- Test: hit rate of prefetched calls, tokens saved, wall time saved, wasted CPU.

**B. Hash-pinned, staleness-aware compaction.** Compaction summaries keep symbol IDs plus content hashes instead of prose about code. After compaction, the model re-reads a symbol through Loupe and gets the **current** body, and is told if it changed since the summary.
- Adjacent: the CooperBench template stores symbol → path:line; the T2 recall tier fetches discarded output from disk. Neither pins hashes or reports staleness.
- Test: re-read rate after compaction, wrong-edit rate from stale memory.

**C. Online edit-format selection from live failure telemetry.** On an edit failure, retry the same intended edit in a different format (for example exact replace → hashline → whole-file), record which format succeeds per model, and let the harness shift its default per model over time (simple bandit).
- Adjacent: static model-to-format maps (Tachikoma), a trained adaptive find-replace/rewrite policy (SWE-Edit), offline harness evolution (AHE). Not found: online per-model adaptation inside the running harness, which matters most for your rotating free models.
- Test: edit-success rate and turns per edit, compared with a fixed format.

If you only build one of these, A is the one that uses what you are already building (Loupe plus the harness), and it directly attacks the round-trip latency you saw in your screenshot.

---

## A24. Scorecard: how you will know it is the best you have built

Track these per task on your own task set (section 17), against your current agent, with several runs each:

| Metric | Why |
|---|---|
| Task success rate | The only metric that really matters |
| Input / output / cached tokens | Context efficiency |
| Model turns and tool calls | Round trips, the cause of slowness in your screenshot |
| Wall time (and time spent waiting on model vs tools) | Speed |
| Malformed tool-call rate, edit-failure rate | Reliability with free models |
| Cache-hit rate (where the host reports it) | Cost and speed |
| Re-read rate (same file/range read twice) | Wasted context |
| Compaction events and post-compaction re-exploration | Memory loss |

---

## A25. Caveats

- Several of these projects are new or pre-1.0 (`openbuild`, `uira`, `crab-code`, `vetto`, `acpx`, `agent-harness` alpha). Read the code, do not assume stability.
- Version numbers drift quickly. Check crates.io for current versions before pinning.
- The **[known]** items were not re-verified in this session.
- I did not verify claims in the projects' own READMEs (performance numbers, token savings). Measure in your own benchmark.
- Section 23's "no exact prior art found" comes from a handful of web searches. It is not a novelty check. Do a proper literature search (arXiv, GitHub) before claiming anything as new.
- Several figures in sections 22-23 come from blog posts, vendor pages, or single reports, and some are small samples. They show direction, not guaranteed gains.

### Sources (Part A)

- [Codex Harness architecture article](https://www.besthub.dev/articles/codex-harness-how-openai-built-an-embeddable-agent-os-in-135-rust-crates-fd42090cccd9)
- [Codex Rust crates overview](https://mintlify.com/openai/codex/architecture/rust-crates)
- [VT Code (docs.rs)](https://docs.rs/crate/vtcode-core/latest)
- [openbuild](https://github.com/rohitg00/openbuild)
- [uira](https://lib.rs/crates/uira-oxc)
- [crab-code](https://github.com/lingcoder/crab-code)
- [agent-harness](https://docs.rs/crate/agent-harness/0.5.1)
- [programmatic-agent-router](https://github.com/KerryRitter/programmatic-agent-router)
- [rust-genai](https://github.com/jeremychone/rust-genai)
- [rig-core](https://docs.rs/crate/rig-core)
- [rmcp (official Rust MCP SDK)](https://github.com/modelcontextprotocol/rust-sdk)
- [rust-mcp-sdk](https://github.com/iflow-mcp/rust-mcp-sdk)
- [ACP Rust SDK](https://github.com/agentclientprotocol/rust-sdk/wiki)
- [vetto](https://docs.rs/crate/vetto/latest/source/Cargo.toml)
- [repo-mapper](https://docs.rs/crate/repo-mapper/0.0.1)
- [Repository map pattern](https://www.agentpatterns.ai/context-engineering/repository-map-pattern/)
- [leindex](https://docs.rs/crate/leindex/2.0.2)
- [CodeNib](https://pypi.org/project/codenib/)
- [Cursor: fast regex search](https://cursor.com/en-US/blog/fast-regex-search)
- [An Empirical Study of Harness Design for Coding Agents (arXiv 2609.20804)](https://www.alphaxiv.org/abs/2609.20804)
- [Glean: building an efficient agent harness](https://www.glean.com/blog/enterprise-agent-harness)
- [SWE-Pruner Pro: tool output pruning](https://www.developersdigest.tech/blog/swe-pruner-pro-tool-output-pruning)
- [Claude Code harness architecture (secondary)](https://wavespeed.ai/blog/posts/claude-code-agent-harness-architecture/)
- [Prompt caching and Claude Code (secondary)](https://www.implicator.ai/anthropic-says-cache-misses-are-production-incidents-reveals-caching-shaped-claude-code/)
- [ClawHarness feature list](https://github.com/agenmod/claw-harness)
- [toolrails: tool-call repair for small models](https://pypi.org/project/toolrails/)
- [Better models, worse tools (edit-tool schema drift)](https://lucumr.pocoo.org/2026/7/4/better-models-worse-tools.md)
- [flynn harness: model-adaptive scaffolding](https://pkg.go.dev/github.com/ionalpha/flynn/harness)
- [gitoxide status](https://x-cmd.com/install/gitoxide/)
- [vcs-core positioning (gix vs git2 vs subprocess)](https://docs.rs/crate/vcs-core/0.7.0/source/docs/positioning.md)
- [dirge-agent LSP client](https://docs.rs/crate/dirge-agent/latest/source/src/lsp/client.rs)
- [lingshu-lsp](https://docs.rs/crate/lingshu-lsp/0.10.0)
- [tokio-lsp](https://docs.rs/crate/tokio-lsp/latest)
- [Terminal-Bench paper (arXiv 2601.11868)](https://arxiv.org/abs/2601.11868?context=cs)
- [SWE-bench vs Terminal-Bench governance guide](https://www.digitalapplied.com/blog/swe-bench-terminal-bench-benchmark-guide-2026)
- [Benchmark literacy guide](https://codex.danielvaughan.com/2026/06/17/benchmark-literacy-guide-reading-coding-agent-benchmarks-critically/)
- [Codesota: Terminal-Bench](https://www.codesota.com/benchmark/terminal-bench)
- [The Harness Problem (edit formats)](https://blog.can.ac/2026/02/12/the-harness-problem)
- [Model Harness: edit format selection (secondary summary)](https://nirvaxstiel.github.io/Tachikoma-Agent-Skills/research/model-harness.html)
- [SWE-Edit (arXiv 2604.26102)](https://arxiv.org/html/2604.26102)
- [Action-Space Shaping / tool-schema bias (arXiv 2609.34971)](https://arxiv.org/pdf/2609.34971)
- [PASTE: speculative tool execution (summary)](https://agentwiki.org/_export/raw/speculative_tool_execution)
- [Spex](https://github.com/cchang3906/spex)
- [TomasuLLM (arXiv 2609.38201)](https://arxiv.org/pdf/2609.38201)
- [zeph speculation engine spec](https://docs.rs/crate/zeph/latest/specs/053-speculation-engine/spec.md)
- [SpecAgent (arXiv 2510.17925)](https://arxiv.org/abs/2510.17925?context=cs)
- [lingshu-tools read tracker](https://docs.rs/crate/lingshu-tools/latest/source/src/read_tracker.rs)
- [Serena](https://github.com/oraios/serena)
- [CodeGraph](https://github.com/scraix/codegraph)
- [Agentic Harness Engineering (AHE)](https://www.alphaxiv.org/audio/2604.25850)
- [TRIAGE / self-evolving harnesses](https://www.emergentmind.com/papers/2608.10178)
- [Self-Harness write-up](https://antoinebuteau.com/agent-harnesses-can-learn-from-their-own-failures)
- [PolyHarness](https://pypi.org/project/polyharness/)
- [Thoughtworks Radar: self-improving harness](https://www.thoughtworks.com/radar/techniques/self-improving-harness)
- [PLANTOD](https://pypi.org/project/plantod/)
- [lean-agent cascade](https://github.com/arnelirobles/lean-agent/wiki/The-Cascade)
- [agent-cost-optimizer report](https://huggingface.co/narcolepticchicken/agent-cost-optimizer/blob/main/docs/final_report.md)
- [CooperBench compaction template PR](https://github.com/cooperbench/CooperBench/pull/54)
- [Morph: deletion-based compaction (vendor)](https://morphllm.com/context-compaction)


---

# Part B: Borrow vs. build, and design decisions from existing harnesses

Part B builds on Part A. Part A lists crates and techniques. Part B answers three questions:

1. What to **borrow** from open-source tools, and in what form (dependency, reimplement, or read only).
2. What to **build** yourself, because it is where your harness can be different.
3. Which **design decisions from existing harnesses** to integrate, adapt, or skip.

Legend: **[searched]** = seen in web results this session. **[known]** = my own knowledge, not re-verified. **[inference]** = my conclusion from the sources, not a published claim.

---

## B1. The rule of thumb

- **Borrow the plumbing and the solved problems**: protocols, parsing, search libraries, sandbox mechanisms, patch formats, TUI widgets.
- **Build what decides quality, cost and speed**: the agent loop, the context manager, the edit tool, the code index (Loupe), per-model adaptation, and the eval harness.
- **Read, don't depend on, big reference harnesses.** Codex, OpenCode, Goose and OpenHands are large and move fast. Take their ideas, not their code.
- **Check every license** before copying code. From the sources: Goose is Apache-2.0, OpenCode and Pi are MIT, Aider is Apache-2.0, repo-mapper is Apache-2.0 **[searched]**. I did not verify the others.

---

## B2. Component-by-component: borrow or build

| Component | Verdict | Borrow from | What you build |
|---|---|---|---|
| Async runtime, serde, clap, tracing | **Depend** | `tokio`, `serde`, `clap`, `tracing` **[known]** | Nothing |
| LLM client (OpenAI-compatible) | **Build thin** | `genai` or `async-openai` as reference **[searched/known]** | Streaming + tool-call delta assembly + usage normalization + retries + per-model quirks. You need control for free/routed hosts |
| Provider/model config | **Borrow the idea** | Goose's declarative providers (a full provider from config, no code) **[searched]** | TOML config: base URL, key source, capability profile |
| MCP client and server | **Depend** | `rmcp` (official Rust SDK) **[searched]** | Wiring, approvals, lazy tool loading |
| Editor protocol (ACP) | **Depend, later** | `agent-client-protocol` crates **[searched]** | Adapter over your event stream |
| File walking, ignore rules, regex search | **Depend** | `ignore`, `globset`, `grep-searcher`/`grep-regex` **[known]** | Thin tool wrappers with output caps |
| Tree-sitter parsing | **Depend** | `tree-sitter` + grammars; Aider's `tags.scm` queries as a starting point **[searched]** | Loupe's outline/symbol model |
| Repo map ranking | **Reimplement** | Aider's repo map (tree-sitter + PageRank + budget fitting); `repo-mapper` is a Rust port **[searched]** | Budget-aware map inside Loupe; use as one query among several |
| Trigram/regex index | **Build later** | Zoekt, Cursor's write-up **[searched]** | Only after measuring that `grep-searcher` is too slow |
| Code index as a service | **Build (Loupe)** | CodeGraph, leindex, codeix, indxr as competitors to study **[searched]** | The differentiator (section 5) |
| Edit tool formats | **Build** | Aider's formats (diff, udiff, whole), `codex-apply-patch` for patch parsing, hashline idea from "The Harness Problem" **[searched]** | 2-3 formats behind one interface, with per-model selection and failure fallback |
| Diff rendering | **Depend** | `similar` **[known]** | Nothing |
| Shell execution | **Depend** | `tokio::process`, `portable-pty` **[known]** | Output capping, timeouts, process-group kill |
| Sandbox | **Borrow design** | Codex: Seatbelt on macOS, Landlock/Bubblewrap on Linux **[searched]** | A small `sandbox` crate that wraps `sandbox-exec` first |
| Permissions | **Build small** | Rule shape `Tool(arg-glob)` from openbuild/uira **[searched]** | Allow/ask/deny engine |
| Agent loop | **Build** | Codex's Thread → Session → Turn structure; OpenHands' event-sourced loop **[searched]** | Core of the product |
| Context manager (truncation, compaction) | **Build** | OpenHands "condenser" strategies; Glean's truncation-to-disk **[searched]** | Token ledger, tiers, hash-pinned compaction (Part A sections A10, A23) |
| Session storage | **Build thin** | Goose's JSONL sessions, OpenCode's branching sessions **[searched]** | JSONL event log with branch/fork/replay |
| Git checkpoints/undo | **Build thin** | OpenCode's automatic git snapshot per step; Aider's auto-commit and `/undo` **[searched]** | Snapshot before write turns; shell out to `git` |
| LSP/diagnostics | **Build thin, optional** | OpenCode feeds only diagnostics to the model **[searched]**; `dirge-agent` LSP client for the stale-version gotcha **[searched]** | `cargo check` JSON first, LSP later |
| Hooks, skills | **Borrow the standard** | Agent Skills standard (used by Pi); Claude-Code-style hooks (crab-code) **[searched]** | Loader plus a hook runner |
| Subagents | **Build later** | OpenCode `task` tool spawns scoped child sessions; OpenHands delegation shares the event stream **[searched]** | Isolated context returning a short result |
| TUI | **Depend** | `ratatui` + `crossterm` **[searched]** | Thin UI over the event stream |
| Evals | **Build** | Terminal-Bench/Harbor and SWE-bench for reference; mini-SWE-agent's trajectory browser idea **[searched]** | Your own task suite and metrics dashboard |
| Release/packaging | **Depend** | `cargo-dist` **[known]** | Nothing |

---

## B3. Design decisions to integrate from existing harnesses

### B3.1 Pi (minimal core, extensible)

What it does **[searched]**:
- Four built-in tools (read, write, edit, bash) and a system prompt under about 1,000 tokens. Everything else is opt-in through an extension system.
- Branching JSONL session trees.
- Its author's reason for leaving MCP out at first was context cost (he measured one browser MCP server at 21 tools and about 13.7k tokens). Later versions added MCP anyway.
- A stable prompt and tool set matters to users: the creator's complaint about other tools was that prompts and tools changed on every release.
- A secondary source reports Terminal-Bench results on par with Cline and OpenCode across four open-weight models; treat as indicative.

Integrate:
- **Small, stable core prompt and tool set.** Version it and do not change it silently between releases (your own benchmark baseline depends on it).
- **Lazy tool loading**: do not put every MCP tool schema in context by default.
- **Branching sessions.**
- **Extension boundary** so new capabilities do not bloat the core.

Adapt: Pi's "bash is enough" stance conflicts with Loupe's "fat structured tools" thesis. Resolve it with data: start bash-first, add each Loupe tool only if an ablation shows it improves success, tokens, or time **[inference]**.

### B3.2 mini-SWE-agent (radical simplicity)

What it does **[searched]**:
- Only bash, no use of the model's tool-calling interface, so it works with any model.
- Linear history: the trajectory is identical to the messages sent to the model, which helps debugging and fine-tuning.
- Each action is an independent subprocess, not a stateful shell session, which makes sandboxing simple.
- A README claims above 74% on SWE-bench Verified with one frontier model; treat as a vendor/author claim.

Integrate:
- **Linear, append-only history** (also good for prompt caching).
- **Stateless commands** (`cd` and env do not persist unless you pass them explicitly).
- A **bash-only fallback mode** that parses a command from plain text, for hosts or models whose tool-calling is unreliable. This is directly relevant to free routed models.
- A trajectory browser early.

### B3.3 OpenCode (client/server, sessions, LSP)

What it does **[searched]**:
- The TUI is one client of a local HTTP + server-sent-events server; a desktop app and an SDK generated from an OpenAPI spec use the same server.
- Sessions can branch, with undo/redo from automatic git snapshots at each agent step.
- Agents defined in config (a full-access "build" agent and a read-only "plan" agent), a `task` tool for child sessions, an `AGENTS.md` generator, and a models catalog.
- LSP is wired in, but only diagnostics are exposed to the model, even though the client supports more.

Integrate:
- **Core as a headless server/library with an event stream; the TUI is just a client.** This also gives you `exec` mode, editor integration, and remote use for free.
- **Git snapshot per write step** for undo.
- **Read-only plan agent** as a config profile, not a separate codebase.
- **Restraint with LSP**: expose diagnostics, not the whole protocol.

### B3.4 Codex CLI (`codex-rs`)

What it does **[searched]**:
- Many-crate workspace: core agent loop, TUI, headless `exec`, JSON-RPC `app-server`, separate sandbox crates, `codex-apply-patch`, `codex-file-search`, `codex-shell-command`.
- Thread → Session → Turn hierarchy; tools built per turn from feature flags; more than one compaction path (local summarization, new window on token budget).
- Sandbox as a standalone executable invoked via argv[0] routing.

Integrate:
- **Protocol-first core** (typed events and requests), with `exec` and a server mode from day one.
- **Per-turn tool assembly**: build the tool list each turn from config and model profile, but keep the order stable for caching.
- **Sandbox as a separate small binary or crate** with a tiny interface.
- **Small single-purpose crates** (patch, file search, shell parsing) that you can test in isolation.

Skip: its 100+ crate scale. Start with about 8-10 crates.

### B3.5 Goose (MCP-native, recipes)

What it does **[searched]**:
- Rust workspace with a core crate, CLI, server (`goosed`), ACP crate, and MCP extension crate.
- Every capability, including built-ins, is an MCP server; built-in ones run in-process over in-memory duplex streams.
- Declarative providers (config only), YAML **recipes** that bundle prompt, extensions and settings, JSONL sessions, and layered config (env over file over defaults).
- A stated principle: trust the type system and avoid defensive code.

Integrate:
- **Declarative providers and models in config**, which suits FreeLLMAPI/OmniRoute.
- **Recipes**: versioned YAML for repeatable workflows (review, fix failing tests, release check).
- **Config precedence** env > file > defaults.

Adapt: for built-in tools, call Rust traits directly instead of going through MCP, to avoid serialization overhead and latency; use MCP for external tools only **[inference]**.

### B3.6 OpenHands (event sourcing, condensers, workspaces)

What it does **[searched]**:
- The V1 SDK is event-sourced: an append-only log of typed events, immutable config, deterministic replay.
- Pluggable **condensers** (LLM summarizing, observation masking, window, no-op) and a signal to compress when the context limit is hit.
- A **workspace** abstraction (local, Docker, remote) so the same agent runs anywhere.
- Auxiliary services (security analyzer, stuck detection, delegation) hang off the event stream.
- Lesson from V0 to V1: a mandatory sandbox made local use and the CLI awkward, so sandboxing became optional and pluggable.

Integrate:
- **Event-sourced core** with typed events; sessions, UI, metrics and replay all consume the same log.
- **`Condenser` trait** with several strategies, selectable per model profile.
- **`Workspace` trait** (local first) so Docker/remote can be added later.
- **Stuck detection** as an observer (repeated identical actions).
- **Opt-in sandbox**, not mandatory.

### B3.7 Aider (repo map, edit formats, architect/editor, git)

What it does **[searched]**:
- Repo map with a token budget that grows when no files are in chat.
- Multiple edit formats chosen per model, plus an architect/editor split where a stronger model plans and a cheaper one writes edits.
- A separate **weak model** for summaries and commit messages.
- Git-native: auto-commits, `/undo`, read-only files, and lint/test output fed back for auto-fix.

Integrate:
- **Weak-model slot** for compaction summaries, commit messages, and titles.
- **Edit format as a per-model setting** (and later an automatic one, see Part A section A23 (C)).
- **Budgeted repo map** as the cold-start context.
- **Read-only files** concept for protected paths.
- **Lint/test feedback loop** after edits.

Optional: architect/editor split, only if your benchmark shows it helps with your models.

### B3.8 Claude Code (from secondary descriptions)

From secondary sources **[searched]**: file snapshots before edits for rewind, truncation with disk fallback, project instruction files re-read every turn, cache-aware design, hooks. Already covered in Part A sections A10, A13, A15. I did not read its code.

---

## B4. Design decisions to avoid

| Avoid | Why |
|---|---|
| A big feature checklist from day one (sub-agents, plan mode, to-dos, background shells) | Pi's whole thesis is that this bloat costs context and stability **[searched]**. Add features only when they win an ablation |
| Mandatory sandboxing | OpenHands V0 had to be redesigned partly because of it **[searched]** |
| Preloading all MCP tool schemas | Large fixed context cost per session |
| Silently changing the system prompt or tool schemas between versions | Breaks cache and invalidates your benchmark baselines |
| Fullscreen TUI before the loop is solid | Time sink; a streaming CLI proves the core faster |
| A stateful shell session | Harder to sandbox, replay, and debug (mini-SWE-agent's design avoids it) |
| Coupling the UI to the agent | Prevents `exec`, editor integration, and testing |
| Trusting model confidence for escalation | Use execution feedback (Part A section A22) |

---

## B5. What to build (your custom design)

These are where your harness can be better, not just equal.

1. **Loupe, in-process.** Because you own both the harness and the index, tools can be designed together: one fat call returns symbol body plus callee signatures plus test locations within a token budget. No steering prompt needed to make the model use it **[inference, based on CodeGraph's note that agents must be steered to use external indexes]**.
2. **Token ledger.** Every tool call declares or estimates its token cost; the loop enforces a per-turn and per-task budget and degrades gracefully (outline instead of full file, preview instead of full output).
3. **Per-model profile.** Capabilities (tool-call reliability, context size, cache support, parallel calls) plus the preferred edit format. Used to choose tool schemas, scaffolding, and condenser.
4. **Edit tool with several formats and a fallback chain**, with telemetry on which format succeeds per model.
5. **Event-sourced core with replay.** Every change can be replayed against a mock provider for regression tests and for benchmarking.
6. **Graph-predicted speculation and hash-pinned compaction** (Part A section A23, hypotheses to test).
7. **Eval harness and metrics dashboard** built in from the start.

---

## B6. Proposed architecture (crate map)

```
cli            binary: interactive, exec, serve
tui            ratatui client (later)
server         local HTTP/SSE or JSON-RPC over the same event types (later)
core           agent loop, turn assembly, token ledger, stuck detection
events         typed event log (append-only JSONL), replay
llm            Provider trait + OpenAI-compatible client, model profiles
context        condenser trait + strategies, truncation-to-disk, recall
tools          read, multi_edit, grep, glob, shell, symbol (Loupe)
edit           edit formats + fallback chain + telemetry
perms          allow/ask/deny rules
sandbox        seatbelt first
workspace      local first, trait for later Docker/remote
loupe          indexer library + optional MCP server binary
mcp            rmcp client wrapper, lazy tool loading
evals          task runner, metrics, trajectory viewer
```

Dependency rule: `core` depends on traits (`Provider`, `Workspace`, `Condenser`, `Tool`) and never on `tui`/`server`/concrete providers. This is the lesson OpenHands learned between V0 and V1 and what OpenCode and Codex both do.

---

## B7. Suggested order of work (borrow first, build the differentiators second)

1. **Borrow/depend**: set up the workspace, `tokio`, `serde`, `ignore`, `grep-*`, `similar`, `tree-sitter`, `rmcp`.
2. **Build the minimum loop** in `core` + `llm` + `tools` (read, multi_edit, shell) with an event log. Pattern from mini-SWE-agent and Pi: linear history, small prompt.
3. **Add evals** (your 10-20 tasks) and record a baseline before adding anything else.
4. **Add safety and undo**: permissions, Seatbelt, git snapshots.
5. **Add Loupe v0** (outline + `symbol`), and run an ablation against bash-only.
6. **Add context management** (truncation to disk, condenser trait, structured summaries).
7. **Add model profiles and edit-format fallback.**
8. **Try the speculation and compaction hypotheses** one at a time, each behind a flag and each judged by the scorecard.
9. Only then: TUI polish, server mode, ACP, subagents, recipes.

---

## B8. Caveats

- Descriptions of other harnesses come from docs, READMEs, and secondary write-ups, not from reading their code. Verify before copying a design detail.
- Benchmark and cost figures quoted in sources are indicative; measure your own.
- Several projects are young or change quickly (Pi, OpenCode, Goose, OpenHands V1). Re-check before depending on specifics.
- I did not verify licenses beyond those noted in section 1.

### Sources (Part B)

- [Pi overview](https://agentic-ai.readthedocs.io/en/latest/AgentHarness/pi-dev/)
- [Pi explained](https://www.eesel.ai/blog/pi-coding-agent)
- [Pi: The Minimal Agent Within OpenClaw](https://lucumr.pocoo.org/2026/1/31/pi.md)
- [Pi architecture deep dive](https://www.developersdigest.tech/blog/pi-deep-dive-agent-toolkit-architecture)
- [Pi write-up (yage.ai)](https://yage.ai/share/pi-coding-agent-locked-out-en-20260518.html)
- [mini-SWE-agent (README mirror)](https://github.com/ofek/mini-swe-agent)
- [mini-SWE-agent (PyPI)](https://pypi.org/project/mini-swe-agent/)
- [OpenCode architecture overview](https://brain-donnyaw-cf-pages-opencode.pages.dev/)
- [Inside OpenCode](https://hn.jakechampion.name/reader/https://cefboud.com/posts/coding-agents-internals-opencode-deepdive)
- [OpenCode server architecture](https://pyshine.com/opencode/)
- [OpenCode wiki](https://www.aiwiki.ai/wiki/opencode/raw)
- [Codex Harness architecture article](https://www.besthub.dev/articles/codex-harness-how-openai-built-an-embeddable-agent-os-in-135-rust-crates-fd42090cccd9)
- [Codex Rust crates overview](https://mintlify.com/openai/codex/architecture/rust-crates)
- [Goose architecture](https://mintlify.com/block/goose/concepts/architecture)
- [Goose development architecture](https://www.mintlify.com/block/goose/development/architecture)
- [Goose MCP-native architecture](https://codex.danielvaughan.com/2026/04/07/goose-mcp-native-architecture/)
- [OpenHands events](https://docs.openhands.dev/sdk/arch/events)
- [OpenHands SDK paper (arXiv 2511.03690)](https://arxiv.org/pdf/2511.03690v1)
- [OpenHands design patterns](https://codedocs-2gm.pages.dev/raw/openhands-2/05-design-patterns)
- [OpenHands deep dive](https://dev.to/truongpx396/openhands-deep-dive-build-your-own-guide-1al0)
- [Aider config reference](https://aider.chat/docs/config/aider_conf.html)
- [Aider overview](https://agentwiki.org/aider)
- [Aider guide](https://deployhq.com/guides/aider)
- [Aider architecture write-up](https://starlog.pages.dev/articles/ai-dev-tools/aider-ai-aider)


---

# Part C: Zephyr Project Tracking & Custom Features

This section tracks project-specific requirements, design decisions, and feature requests for Zephyr.

## C1. Active Feature Requirements

1. **Dual Interface: TUI + GUI Modes**
   - **TUI Frontend:** `ratatui` + `crossterm` for fast, lightweight terminal usage.
   - **GUI Frontend:** `egui` + `eframe` (immediate mode, lightweight single binary) or desktop client powered by a shared core.
   - **Architecture:** Presentation-agnostic `core` emitting strongly typed events (`RunEvent`); both frontends are pure views consuming the stream.

2. **Skills Support (Agent Skills Standard)**
   - Format: Folders with `SKILL.md` (YAML frontmatter + markdown instructions).
   - Discovery: Global (`~/.zephyr/skills/`) and project-local (`.zephyr/skills/`).
   - Lazy Context Loading: Only inject metadata into the initial prompt; load full skill instructions on-demand when triggered to conserve context window.
   - Lifecycle Hooks: Support `PreToolUse` and `PostToolUse` validations within skills.

3. **Custom System Prompts**
   - Layered precedence: Default lean prompt $\to$ Global user prompt $\to$ Project prompt (`.zephyr/system.md` or `AGENTS.md`) $\to$ Directory rules $\to$ CLI `--system-prompt`.
   - Prefix stability: System prompt and tool schemas stay fixed at the top of the prompt to maximize KV cache reuse across turns.

4. **Memory Engineering**
   - **Tier 1 (Working Context):** Tool output truncation to disk (`.zephyr/scratch/`) with range preview to prevent context blowup.
   - **Tier 2 (Structured Compaction):** Pluggable condenser strategies producing structured templates (File Map, Key Symbols, Active Goals, Edits Applied, Negative Results).
   - **Tier 3 (Durable Memory):** SQLite/Markdown storage for long-term project knowledge and user preferences.
   - **Hash-Pinned Anchoring:** Store BLAKE3 hashes of code symbols so memory queries detect when referenced code has changed and report staleness.
