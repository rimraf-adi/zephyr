# Zephyr Phased Engineering Plan & Build Checklist

This document lays out the exact build order for Zephyr, structuring each phase around the required cycle:
1. **Build** the specific crate/module.
2. **Test** with unit tests, mocks, and real command runs.
3. **Take Feedback** from the user at each milestone.
4. **Change / Refine** based on feedback.
5. **Commit** with clean, human commit messages.

---

## High-Level Phase Overview

```
Phase 0: Workspace Foundation & Core Event Types
   ▼
Phase 1: LLM Client & SSE Streaming Layer
   ▼
Phase 2: Fundamental Tools & Tool-Call Repair Engine
   ▼
Phase 3: Core Agent Loop, Token Ledger & Condenser
   ▼
Phase 4: Skills System & Hierarchical System Prompts
   ▼
Phase 5: Code Intelligence (Loupe AST & Repo Map)
   ▼
Phase 6: Security, Sandboxing (Seatbelt) & Git Checkpoints
   ▼
Phase 7: Terminal UI (TUI) via Ratatui
   ▼
Phase 8: Desktop GUI via egui
   ▼
Phase 9: Evaluation Suite, Replay & Benchmarks
```

---

## Detailed Checklists

### Phase 0: Workspace Foundation & Core Event Types
- [ ] Initialize root Cargo workspace (`Cargo.toml`)
- [ ] Setup initial crates:
  - `crates/events`: Strongly typed `RunEvent`, `Action`, and JSONL serializer
  - `crates/core`: Agent loop interfaces and presentation-agnostic traits
  - `crates/cli`: Binary entrypoint with `clap` arguments (`--tui`, `--gui`, `exec`)
- [ ] Add mock event stream test
- [ ] **Checkpoint:** Validate workspace compiles and tests pass; review architecture with user.

---

### Phase 1: LLM Client & SSE Streaming Layer (`crates/llm`)
- [ ] Build thin OpenAI-compatible HTTP client over `reqwest`
- [ ] Implement Server-Sent Events (SSE) streaming parser
- [ ] Build delta accumulator for fragmented streaming tool-call JSON
- [ ] Implement exponential backoff with jitter and `Retry-After` handling
- [ ] Standardize `TokenUsage` accounting (prompt, completion, cached tokens)
- [ ] Add idle-stream watchdog timer (handles silent TCP disconnects)
- [ ] **Checkpoint:** Test against mock streaming endpoint and real OpenAI-compatible endpoint.

---

### Phase 2: Fundamental Tools & Tool-Call Reliability (`crates/tools`)
- [ ] Implement `read_file` (whole file preview + range reader)
- [ ] Implement `multi_edit` (exact string replacement across multiple blocks)
- [ ] Implement `shell_exec` (asynchronous command execution with process group kill)
- [ ] Implement tool-output disk truncation:
  - Max in-memory preview (50–100 lines)
  - Full output persisted to `.zephyr/scratch/<call_id>.log`
- [ ] Build tool-call repair pipeline (`jsonschema`):
  - Deterministic coercion (strip quotes around numbers, parse nested JSON strings)
  - Schema error feedback retry loop
- [ ] **Checkpoint:** Run tool tests (handling truncated outputs and malformed JSON recovery).

---

### Phase 3: Core Agent Loop & Context Condensers (`crates/core`)
- [ ] Implement turn assembler with strict prefix stability for KV caching
- [ ] Implement token budget ledger (tracking per-turn and total session limits)
- [ ] Build circuit breakers:
  - Stuck detection (interrupting identical repeated tool calls)
  - Auto-continuation on `finish_reason: length`
  - Max turn ceiling and `CancellationToken` support
- [ ] Implement `Condenser` trait:
  - Output eviction tier
  - Structured compaction template (Goals, File Map, Symbols, Negative Results, TODOs)
- [ ] **Checkpoint:** Run multi-turn synthetic agent tasks with memory compaction.

---

### Phase 4: Skills System & Hierarchical Prompts (`crates/skills`, `crates/prompt`)
- [ ] Build hierarchical prompt resolver:
  - Default lean base rules (< 1,000 tokens)
  - Global user prompt (`~/.zephyr/system.md`)
  - Project prompt (`.zephyr/system.md` or `AGENTS.md`)
  - CLI overrides (`--system-prompt`)
- [ ] Implement Agent Skills standard parser (`SKILL.md` with YAML frontmatter)
- [ ] Implement lazy prompt loading (catalog registration at boot, full body injection on demand)
- [ ] Build `PreToolUse` and `PostToolUse` lifecycle hook runner
- [ ] **Checkpoint:** Test skill activation and prompt layering.

---

### Phase 5: Code Intelligence & Repo Indexing (`crates/loupe`)
- [ ] Integrate `tree-sitter` and language grammars (Rust, Python, TypeScript)
- [ ] Implement structural file outline generator (signatures without bodies)
- [ ] Build symbol reference graph using `petgraph`
- [ ] Implement Aider-style PageRank repository map within a tight token budget
- [ ] Add fast lexical search over ripgrep internals (`grep-searcher`)
- [ ] **Checkpoint:** Test repo map generation and symbol search on real repositories.

---

### Phase 6: Security, Sandboxing & Git Checkpoints (`crates/sandbox`, `crates/perms`)
- [ ] Implement path confinement (symlink resolution and workspace boundary checks)
- [ ] Implement macOS Seatbelt sandbox wrapper (`sandbox-exec` profiles)
- [ ] Build command permission engine (Allow / Ask / Deny matching)
- [ ] Implement git shadow checkpoints before write turns (`refs/zephyr/checkpoints/`)
- [ ] Implement `/undo` non-destructive rollback
- [ ] Subprocess environment variable sanitization and secrets redaction
- [ ] **Checkpoint:** Verify sandboxing prevents unauthorized disk/network operations and verify `/undo`.

---

### Phase 7: Terminal UI (TUI) Mode (`crates/tui`)
- [ ] Setup `ratatui` + `crossterm` presentation layer
- [ ] Build responsive layout:
  - Conversational streaming panel with markdown rendering
  - Interactive file diff inspector (`similar`)
  - Tool execution spinner and active status bar
  - Multi-line user input prompt with command history
- [ ] Wire TUI event loop directly to core `RunEvent` stream
- [ ] **Checkpoint:** Interactive test session of TUI workflow.

---

### Phase 8: Desktop GUI Mode (`crates/gui`)
- [ ] Setup `egui` + `eframe` immediate-mode graphical desktop application
- [ ] Connect GUI to the exact same core `RunEvent` stream
- [ ] Implement graphical features:
  - Project file tree with modified file status
  - Side-by-side visual diff viewer
  - Expandable tool execution logs and scratch inspector
  - Dark/Light developer themes
- [ ] Configure `zephyr --gui` CLI dispatch
- [ ] **Checkpoint:** Interactive test run of GUI mode.

---

### Phase 9: Evaluation Suite, Replay & Benchmarks
- [ ] Implement JSONL session recorder and deterministic trajectory replayer
- [ ] Build mock provider with recorded HTTP cassettes (`wiremock`)
- [ ] Establish initial 10-task regression test suite
- [ ] Generate scorecard metrics (resolve rate, turns, token efficiency, cache hits)
- [ ] **Checkpoint:** Run full automated regression evaluation.
