# 01: High-Level Architecture, Crate Map & Technology Stack

> **Focus Area:** System decomposition, modular workspace design, multi-frontend decoupling (TUI + GUI + Headless).

---

## 1. System Design Goals

An enterprise-grade coding agent harness must satisfy four foundational architectural requirements:
1. **Model-Agnostic Resilience:** Must operate reliably with frontier models, free tiers, and routed proxies (OpenAI-compatible endpoints).
2. **Deterministic Context Control:** Prevent context blowup through strict token budgeting and disk truncation.
3. **Pluggable Presentation Layers:** Support a Terminal UI (TUI), a native Desktop GUI, and headless automated execution (`exec` mode) without duplicating business logic.
4. **Local-First Safety:** Confinement to project workspace, sandboxed shell execution, and non-destructive git undo.

---

## 2. Multi-Frontend Decoupled Architecture

The biggest architectural mistake in coding agent harnesses is tightly coupling the agent loop with the terminal interface. To support **both TUI and GUI modes**, Zephyr adopts an **event-sourced, presentation-agnostic core**.

```
                              ┌────────────────────────┐
                              │     CLI Dispatcher     │
                              │  zephyr [--tui|--gui]  │
                              └───────────┬────────────┘
                                          │
                    ┌─────────────────────┴─────────────────────┐
                    ▼                                           ▼
         ┌─────────────────────┐                     ┌─────────────────────┐
         │     crates/tui      │                     │     crates/gui      │
         │ (ratatui+crossterm) │                     │    (egui+eframe)    │
         └──────────┬──────────┘                     └──────────┬──────────┘
                    │                                           │
                    │         Typed Action / Event Bus          │
                    │        (SubmitPrompt, Cancel, etc.)       │
                    └─────────────────────┬─────────────────────┘
                                          ▼
                               ┌─────────────────────┐
                               │     crates/core     │
                               │  (Agent Loop, Turn  │
                               │  Assembly, Budget)  │
                               └──────────┬──────────┘
                                          │
       ┌──────────────────┬───────────────┴───────────────┬──────────────────┐
       ▼                  ▼                               ▼                  ▼
┌──────────────┐   ┌──────────────┐                ┌──────────────┐   ┌──────────────┐
│  crates/llm  │   │ crates/tools │                │ crates/loupe │   │crates/sandbox│
│(SSE/OpenAI/  │   │(Read/Edit/   │                │ (Tree-sitter │   │  (Seatbelt/  │
│  Routers)    │   │ Grep/Shell)  │                │ Symbol Graph)│   │  Landlock)   │
└──────────────┘   └──────────────┘                └──────────────┘   └──────────────┘
```

### The Inversion of Control Rule
The core agent loop (`crates/core`) **never depends on UI crates**. It depends strictly on traits:
- `trait Provider`: Abstract LLM completions and streaming deltas.
- `trait Tool`: Abstract executable action schemas and runners.
- `trait Workspace`: Abstract file system and subprocess execution environment.
- `trait Condenser`: Abstract context compaction strategies.

Frontends subscribe to a broadcast stream of typed `RunEvent`s:
```rust
pub enum RunEvent {
    TextDelta(String),
    ThinkingDelta(String),
    ToolCallStarted { id: String, name: String, args: Value },
    ToolCallFinished { id: String, output_preview: String, disk_path: Option<PathBuf> },
    PlanUpdated(Vec<TaskItem>),
    UsageReported(TokenUsage),
    ErrorOccurred(AgentError),
    Done,
}
```

---

## 3. Recommended Technology Stack

| Layer | Recommended Choice | Rationale |
|---|---|---|
| **Async Runtime** | `tokio` (multi-threaded) | Industry standard, required by network and process libraries. |
| **LLM Client** | Custom thin client over `reqwest` + SSE | Direct control over streaming deltas, usage token fields, backoff, and model quirks. |
| **Tool Schemas** | `serde` + `schemars` | Automatically derives OpenAPI/JSON Schema from Rust structs. |
| **Extensible Protocols** | `rmcp` (official Rust MCP SDK) | Client for community MCP servers; server for local code indexer. |
| **Search & Traversal** | `ignore`, `grep-searcher`, `globset` | Ripgrep's core libraries; fastest gitignore-aware search in Rust. |
| **Diffs & Patching** | `similar` + custom patch parser | Robust unified diff calculation and display. |
| **AST & Symbols** | `tree-sitter` + grammars | Fast, language-agnostic syntax tree and outline extraction. |
| **Sandboxing (macOS)** | macOS `sandbox-exec` (Seatbelt) | Kernel-level file system and network confinement. |
| **TUI Frontend** | `ratatui` + `crossterm` | De facto standard for interactive terminal applications. |
| **GUI Frontend** | `egui` + `eframe` | Pure Rust, lightweight immediate-mode GUI with zero webview overhead. |
| **State & Persistence** | `rusqlite` + flat JSONL | Simple, crash-safe, and greppable transcripts. |

---

## 4. Workspace Crate Map

```
zephyr/
  crates/
    core/       # Agent loop, turn assembly, token ledger, event dispatch
    llm/        # Provider trait, OpenAI client, streaming SSE, usage normalization
    tools/      # File read/write, multi_edit, ripgrep, shell execution
    perms/      # Allow/Ask/Deny security rule engine
    sandbox/    # macOS Seatbelt profiles and Linux Landlock wrappers
    loupe/      # Tree-sitter code outline, symbol graph, and repo map
    skills/     # Agent Skills loader, triggers, and lazy prompt injection
    memory/     # 3-tier memory engine, output disk truncation, condensers
    events/     # Event-sourced JSONL persistence and session replay
    tui/        # Ratatui terminal UI
    gui/        # egui/eframe desktop UI
    cli/        # Main binary entrypoint: `zephyr`, `zephyr --gui`, `zephyr exec`
```

---

## 5. Architectural Rule of Thumb: Borrow vs. Build

1. **Borrow the plumbing and solved problems:** Protocols (`rmcp`), regex search (`grep-searcher`), syntax parsing (`tree-sitter`), diffing (`similar`), and UI widgets (`ratatui`, `egui`).
2. **Build what decides quality, cost, and speed:** The agent loop, context window condenser, tool-call repair layer, code index (Loupe), and evaluation harness.
3. **Study reference architectures:** Read Codex-rs, OpenCode, Goose, and OpenHands to borrow design patterns rather than taking heavy dependencies.
