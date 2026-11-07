# Zephyr Specification & Feature Roadmap

> **Status:** Active Specification  
> **Repository:** `rimraf-adi/zephyr`  
> **Last Updated:** 2026-11-07  

---

## 1. Core Vision & Modes

Zephyr is a high-performance, model-agnostic coding agent harness built in Rust. It is designed to be lightweight, token-efficient, and resilient to free and routed model quirks (such as FreeLLMAPI / OmniRoute).

### Interface Modes
1. **TUI Mode (Default):** Built on `ratatui` + `crossterm`. Fast, keyboard-first, terminal-native with split views for streaming tokens, diff inspections, and tool activity.
2. **GUI Mode (`--gui`):** Built on `egui` + `eframe` (or desktop webview via local daemon). Cross-platform, immediate mode, visual file trees, side-by-side graphical diffs, and rich interactive visualizer widgets.
3. **Headless / Exec Mode (`exec` / `--json`):** Non-interactive JSON-RPC / event-stream mode for editor integrations (Zed, VS Code), scripts, and benchmark evals.

---

## 2. Feature Specification

### 2.1 Skills Support
Zephyr adopts an on-demand, lazy-loaded Agent Skills architecture to prevent context window bloat:

- **Format:** Compatible with the open Agent Skills standard.
  - Directory structure: `.zephyr/skills/<skill-name>/SKILL.md` (and `~/.zephyr/skills/` globally).
  - Frontmatter: YAML containing `name`, `description`, `trigger_rules`, and optional required tools or permissions.
  - Body: Markdown instructions, reference documentation, and workflow steps.
- **Lazy Context Loading:** 
  - Instead of dumping all skill documentation into the system prompt at startup, only skill metadata (`name` + 1-line `description`) is registered in a routing index.
  - When the agent loop detects a matching task or intent, it invokes a built-in `activate_skill` tool or dynamically injects the relevant `SKILL.md` into the working context.
- **Skill Hooks:** Skills can define `PreToolUse` and `PostToolUse` scripts or validation checks.

### 2.2 Custom System Prompts
Prompt customization is hierarchical and designed around strict prompt-cache prefix stability:

- **Hierarchy (lowest to highest precedence):**
  1. **Zephyr Default Core Prompt:** Lean core prompt (< 1,000 tokens) defining tool rules and output format.
  2. **Global User Prompt:** `~/.zephyr/system.md` (or configured via `~/.zephyr/config.toml`).
  3. **Project / Workspace Prompt:** `.zephyr/system.md` or repo-level `AGENTS.md`.
  4. **Directory / Scoped Prompt:** Hierarchical subfolder rules.
  5. **Runtime Override:** CLI flags (`--system-prompt <file|string>`) or recipe bundles.
- **Prefix Stability for Caching:**
  - Static system prompts and tool schemas are positioned at the very start of the prompt stream to maximize LLM KV cache hits.
  - Dynamic elements (local time, git branch, cwd, active memory summaries) are placed at the boundary between instructions and conversation history.

### 2.3 Memory Engineering
Zephyr avoids naive context stuffing and compaction loss by implementing a 3-tier memory engine:

```
┌───────────────────────────────────────────────────────────────┐
│                    3-TIER MEMORY ENGINE                       │
├───────────────────────────────┬───────────────────────────────┤
│ Tier 1: Working Memory        │ Append-only turn history      │
│ (Active Context)              │ Tool output preview + disk cap│
├───────────────────────────────┼───────────────────────────────┤
│ Tier 2: Compaction & Condenser│ Structured summary templates  │
│ (Mid-Term History)            │ Hash-pinned symbol anchors    │
├───────────────────────────────┼───────────────────────────────┤
│ Tier 3: Durable Memory        │ Global & project knowledge    │
│ (Long-Term Semantic)          │ SQLite / flat Markdown store  │
└───────────────────────────────┴───────────────────────────────┘
```

1. **Tier 1: Working Memory & Tool Truncation:**
   - Strict output capping: No tool returns thousands of lines into context.
   - Large results are written to disk (`.zephyr/scratch/<call_id>.log`) and the LLM receives an indexed range preview (e.g., first 50 lines, line count, byte size, and instructions on using `read_range`).
2. **Tier 2: Structured Compaction (Condensers):**
   - Pluggable `Condenser` trait with selectable strategies.
   - Summarization follows a structured template preserving:
     - Active Goal & Decisions
     - File Map & Key Symbols (`name -> path:line:blake3_hash`)
     - Applied Edits & Outstanding TODOs
     - Negative Results (what was tried and failed, preventing repetitive loops)
   - **Hash-Pinned Staleness Detection:** Symbol references retain content hashes. If a symbol was modified since compaction, the memory layer flags it as stale upon lookup.
3. **Tier 3: Durable / Long-Term Knowledge:**
   - Stored in `.zephyr/memory.db` (SQLite) or `.zephyr/memory/*.md`.
   - Captures user preferences, architectural rules, environment quirks, and successful test commands across sessions.

---

## 3. Workspace Architecture (Crate Layout)

```
zephyr/
├── Cargo.toml                  # Workspace definition
├── crates/
│   ├── cli/                    # Main binary entrypoint (clap: --tui, --gui, exec)
│   ├── core/                   # Agent loop, turn assembly, memory engine, token ledger
│   ├── llm/                    # OpenAI-compatible streaming client, router fallback, retries
│   ├── tools/                  # Ripgrep, multi_edit, shell executor, diagnostics
│   ├── loupe/                  # Tree-sitter code indexer & symbol graph
│   ├── skills/                 # Skill discovery, loader, and lazy injector
│   ├── prompt/                 # Hierarchical system prompt engine & templating
│   ├── memory/                 # 3-tier memory management, truncation & condensers
│   ├── events/                 # Event-sourced JSONL session logging and replay
│   ├── perms/                  # Allow/Ask/Deny security rule engine
│   ├── sandbox/                # macOS Seatbelt profiles (sandbox-exec) & Linux Landlock
│   ├── tui/                    # Ratatui + crossterm terminal interface
│   └── gui/                    # egui + eframe desktop graphical interface
└── docs/                       # Specifications, research, and guides
```

---

## 4. Tracked Feature Requests & Decisions Log

| Date | Feature / Decision | Description | Status |
|---|---|---|---|
| 2026-11-07 | **TUI + GUI Dual Mode** | Support both `ratatui` (TUI) and `egui` (GUI) backed by presentation-agnostic core | Defined |
| 2026-11-07 | **Skills Support** | On-demand Agent Skills standard with lazy prompt loading | Defined |
| 2026-11-07 | **Custom System Prompts** | Layered prompt resolution (global $\to$ project $\to$ CLI) with cache-friendly layout | Defined |
| 2026-11-07 | **Memory Engineering** | 3-tier memory: working output capping, structured compaction, hash-pinned durable memory | Defined |
