# 11: Borrow vs. Build Strategic Matrix

> **Focus Area:** Strategic Engineering Decisions, Core Competencies, High-Leverage Custom Components.

---

## 1. The Core Engineering Rule

> **Borrow the solved plumbing; build what decides cost, speed, and intelligence.**

Large reference harnesses (Codex, OpenHands, Goose) contain tens of thousands of lines of code. Forking or depending directly on them creates massive maintenance debt. The correct strategy is to depend on specialized, low-level Rust crates and write the core loop yourself.

---

## 2. Component-by-Component Decision Matrix

| Subsystem | Strategy | Recommended Dependencies / Reference | What to Build Custom |
|---|---|---|---|
| **Async Runtime & Utilities** | **Depend** | `tokio`, `serde`, `clap`, `tracing` | Standard application configuration. |
| **LLM Client Layer** | **Build Thin** | Study `async-openai` or `genai` | SSE parser, streaming delta accumulator, router retry logic, usage normalization. |
| **MCP Integration** | **Depend** | `rmcp` (Official SDK) | Lazy tool loading, context-budgeted schema injection. |
| **Editor Protocol (ACP)** | **Depend (Later)** | `agent-client-protocol` crates | Bridge mapping Zephyr's internal event stream to ACP. |
| **File Walking & Search** | **Depend** | `ignore`, `globset`, `grep-searcher` | Thin tool wrappers with token caps and disk spilling. |
| **AST & Code Navigation** | **Depend** | `tree-sitter` + official grammars | Symbol extraction, file outlines, structural skeletons. |
| **Repo Map Ranking** | **Reimplement** | Aider algorithm / `repo-mapper` | PageRank graph scoring fitted to dynamic token budgets. |
| **Edit Tool Engine** | **Build Custom** | `similar`, study `codex-apply-patch` | Multi-format edit executor (replace, hashline, rewrite) with fallback logic. |
| **Process Execution & PTY** | **Depend** | `tokio::process`, `portable-pty` | Subprocess timeouts, process-group killing, output truncation. |
| **Kernel Sandboxing** | **Borrow Design** | Study Codex Seatbelt / Landlock | Small `sandbox` crate wrapping macOS `sandbox-exec`. |
| **Security Permissions** | **Build Small** | Rule structure from `uira`/`openbuild` | Allow/Ask/Deny pattern-matching engine. |
| **Agent Loop Engine** | **Build Custom** | Study Codex `Turn` & OpenHands loop | Turn assembly, circuit breakers, auto-continuation, stuck detection. |
| **Context Management** | **Build Custom** | OpenHands condenser pattern | Disk truncation, structured compaction templates, hash-pinned staleness detection. |
| **Session Persistence** | **Build Thin** | Goose / OpenCode JSONL sessions | Append-only event-sourced logging and deterministic replay. |
| **Git Checkpoints & Undo** | **Build Thin** | Subprocess `git` CLI | Shadow commit / stash snapshots before write turns; `/undo` handler. |
| **Edit Feedback Loop** | **Build Thin** | `cargo check --message-format=json` | JSON diagnostic parser feeding error spans back to model. |
| **Terminal UI (TUI)** | **Depend** | `ratatui` + `crossterm` | View client subscribing to core event stream. |
| **Graphical UI (GUI)** | **Depend** | `egui` + `eframe` | Immediate-mode desktop client subscribing to core event stream. |

---

## 3. High-Leverage Custom Differentiators

If you build everything listed above, your agent will be equal to standard commercial harnesses. To be **better, faster, and cheaper**, focus engineering effort on:

1. **In-Process Loupe Indexer:** Direct in-process query API between the agent loop and Tree-sitter symbol graphs (bypassing slow IPC serialization).
2. **Deterministic Tool Repair:** Eliminating wasted API turns on small/routed models by fixing malformed tool arguments before calling the tool.
3. **Structured Compaction with Hash Pinning:** Storing BLAKE3 hashes of code symbols in summaries so the agent detects when memory is stale.
4. **Adaptive Multi-Format Edits:** Dynamically shifting edit formats when a model repeatedly struggles with exact search/replace.
