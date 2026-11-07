# 10: Deep-Dive Case Studies: Modern Agent Harnesses

> **Focus Area:** Architectural analysis of Codex-rs, Pi, mini-SWE-agent, OpenCode, Goose, OpenHands, and Aider.

---

## 1. OpenAI Codex CLI (`codex-rs`)

### Architecture
- **Scale:** Large multi-crate Rust workspace (~135 crates).
- **Core Abstractions:** Hierarchical `Thread` $\to$ `Session` $\to$ `Turn` model.
- **Sandboxing:** Best-in-class OS isolation using macOS Seatbelt (`sandbox-exec`) and Linux Landlock/Bubblewrap.
- **Reusable Crates:** `codex-apply-patch`, `codex-file-search`, `codex-shell-command`.

### Key Lessons to Borrow
- Protocol-first core: Headless JSON-RPC server (`app-server`) cleanly decoupled from the Ratatui TUI.
- Isolate OS sandboxing and patch application into independent, testable crates.

### Anti-Patterns to Avoid
- Extreme crate explosion: 100+ crates introduce compile-time overhead and maintenance complexity.

---

## 2. Pi (Minimal Core, Extensible)

### Architecture
- **Ultra-Lean Core:** 4 built-in tools (`read`, `write`, `edit`, `bash`) with a system prompt under 1,000 tokens.
- **Extension Boundary:** All non-essential features are pushed to an external extension system.
- **Branching Sessions:** JSONL-backed session tree with undo and fork capabilities.

### Key Lessons to Borrow
- Keep the core prompt tiny and rock-solid. Prompt churn invalidates benchmark baselines and destroys caching.
- Lazy tool loading: Never dump large MCP toolsets into context up front.

---

## 3. mini-SWE-agent (Radical Simplicity)

### Architecture
- **Pure Bash Actions:** Operates entirely over shell execution, without using the model's native tool-calling API.
- **Stateless Subprocesses:** Every command is an isolated subprocess (no persistent shell session).
- **Strict Linear History:** Messages sent to the model match the raw execution trajectory 1:1.

### Key Lessons to Borrow
- A bash-only fallback mode is an incredible safety net for weak models whose native JSON tool-calling fails.
- Stateless execution makes sandboxing, replay, and debugging effortless.

---

## 4. OpenCode (Client-Server Architecture)

### Architecture
- **Decoupled Engine:** The agent engine runs as an HTTP + Server-Sent Events (SSE) server.
- **Multi-Client:** The TUI, desktop app, and CLI all consume the same SSE server.
- **Step Snapshots:** Automatically snapshots git state before each turn for instant undo.
- **Restrained LSP:** Connects to language servers, but feeds only concise diagnostics to the model.

### Key Lessons to Borrow
- Decoupling the engine via an SSE event stream allows seamless support for **both TUI and GUI modes**.
- Expose diagnostics rather than drowning the LLM in complex LSP JSON responses.

---

## 5. Goose (Block / MCP-Native)

### Architecture
- **Pure MCP-Native:** Every capability (including internal tools) is structured as an MCP server.
- **In-Memory Streams:** Built-in MCP servers communicate in-process over in-memory duplex streams.
- **Declarative YAML Recipes:** Bundles prompts, extensions, and settings into shareable automation workflows.

### Key Lessons to Borrow
- Declarative configuration: Providers and models defined entirely via TOML/YAML.
- Recipes for repeatable developer workflows.

---

## 6. OpenHands (Event-Sourced Condensers)

### Architecture
- **V1 SDK Event Sourcing:** Complete state represented as an immutable stream of typed events.
- **Pluggable Condensers:** Modular context compression strategies (masking, summarization, windowing).
- **Stuck Detection:** Observer pattern tracking repeated actions.

### Key Lessons to Borrow
- Separation of `Condenser` strategies from the agent loop.
- Opt-in sandboxing (mandatory sandboxing severely degraded local CLI developer experience in OpenHands V0).

---

## 7. Aider (Repo Maps & Edit Format Selection)

### Architecture
- **PageRank Repo Map:** Uses Tree-sitter tags and PageRank to fit global repo context into token budgets.
- **Adaptive Edit Formats:** Chooses between unified diffs, whole-file rewrites, or search/replace based on model profile.
- **Weak Model Delegation:** Uses a cheap/fast model for commit messages, titles, and context summaries.

### Key Lessons to Borrow
- Weak model slot: Offload compaction and metadata extraction to cheap models.
- Automatic lint/test loop following code writes.
