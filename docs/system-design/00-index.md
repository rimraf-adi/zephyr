# System Design: AI Coding Agent Harness

> **Interview Prep & Engineering Architecture Reference**  
> Extracted and modularized from the comprehensive Rust coding-agent harness guide.

This series breaks down the system design of an enterprise-grade, high-performance coding agent harness (like Claude Code, Codex-rs, OpenCode, or Zephyr) into 12 targeted system design modules.

---

## Curriculum & Module Index

| # | Document | Primary System Design Topics Covered |
|---|---|---|
| **01** | [`01-architecture-overview-and-stack.md`](file:///Users/adityakinjawadekar/Documents/100xcode/zephyr/docs/system-design/01-architecture-overview-and-stack.md) | High-level architecture, Rust crate map, decoupled presentation layers (TUI/GUI/Headless), technology stack selection. |
| **02** | [`02-agent-loop-and-context-management.md`](file:///Users/adityakinjawadekar/Documents/100xcode/zephyr/docs/system-design/02-agent-loop-and-context-management.md) | Core agent loop, turn assembly, token budget enforcement, tool-output truncation to disk, compaction tiers, prompt caching & prefix stability. |
| **03** | [`03-llm-client-protocols-and-routing.md`](file:///Users/adityakinjawadekar/Documents/100xcode/zephyr/docs/system-design/03-llm-client-protocols-and-routing.md) | LLM client layer, streaming SSE, protocol standards (MCP, ACP), rate limiting, exponential backoff, multi-model fallback chains. |
| **04** | [`04-tool-design-reliability-and-repair.md`](file:///Users/adityakinjawadekar/Documents/100xcode/zephyr/docs/system-design/04-tool-design-reliability-and-repair.md) | Tool schemas, multi-edit tools, dealing with malformed tool calls from small/routed models, deterministic coercion and repair, constrained decoding trade-offs. |
| **05** | [`05-code-intelligence-and-repo-indexing.md`](file:///Users/adityakinjawadekar/Documents/100xcode/zephyr/docs/system-design/05-code-intelligence-and-repo-indexing.md) | Codebase understanding, Tree-sitter AST parsing, Aider-style PageRank repo maps, symbol graphs, trigram regex indexes, SCIP integration. |
| **06** | [`06-git-checkpoints-and-edit-feedback.md`](file:///Users/adityakinjawadekar/Documents/100xcode/zephyr/docs/system-design/06-git-checkpoints-and-edit-feedback.md) | Deterministic edit feedback (`cargo check` JSON / LSP diagnostics), git-based snapshotting, shadow commits, `/undo`, worktree isolation. |
| **07** | [`07-security-sandboxing-and-permissions.md`](file:///Users/adityakinjawadekar/Documents/100xcode/zephyr/docs/system-design/07-security-sandboxing-and-permissions.md) | Threat modeling, prompt injection as data, OS-level sandboxing (macOS Seatbelt, Linux Landlock), path confinement, command permission policies. |
| **08** | [`08-skills-prompts-hooks-and-subagents.md`](file:///Users/adityakinjawadekar/Documents/100xcode/zephyr/docs/system-design/08-skills-prompts-hooks-and-subagents.md) | Hierarchical system prompts, Agent Skills standard with lazy loading, lifecycle hooks (`PreToolUse`/`PostToolUse`), isolated subagents. |
| **09** | [`09-observability-state-and-evals.md`](file:///Users/adityakinjawadekar/Documents/100xcode/zephyr/docs/system-design/09-observability-state-and-evals.md) | Observability (`tracing`), event sourcing (JSONL transcripts), deterministic replay, evaluation methodologies (Terminal-Bench, SWE-bench), scorecard metrics. |
| **10** | [`10-case-studies-existing-agent-harnesses.md`](file:///Users/adityakinjawadekar/Documents/100xcode/zephyr/docs/system-design/10-case-studies-existing-agent-harnesses.md) | Deep-dive comparative case studies: OpenAI Codex, Pi, mini-SWE-agent, OpenCode, Goose, OpenHands, Aider, and anti-patterns to avoid. |
| **11** | [`11-borrow-vs-build-matrix.md`](file:///Users/adityakinjawadekar/Documents/100xcode/zephyr/docs/system-design/11-borrow-vs-build-matrix.md) | Engineering strategy: component-by-component borrow vs build analysis, identifying high-leverage differentiators. |
| **12** | [`12-novel-differentiators-and-research.md`](file:///Users/adityakinjawadekar/Documents/100xcode/zephyr/docs/system-design/12-novel-differentiators-and-research.md) | Advanced frontiers: graph-predicted speculative execution, hash-pinned staleness compaction, online edit-format adaptation, cited literature. |

---

## Common Interview Questions Mapped to Modules

1. **"How do you prevent context window exhaustion in an autonomous coding agent?"**  
   $\to$ See [Module 02](file:///Users/adityakinjawadekar/Documents/100xcode/zephyr/docs/system-design/02-agent-loop-and-context-management.md) (Tool Truncation to Disk, Condenser Tiers, KV Cache Prefix Stability).
2. **"How do you design a secure execution environment when an AI agent runs arbitrary shell commands?"**  
   $\to$ See [Module 07](file:///Users/adityakinjawadekar/Documents/100xcode/zephyr/docs/system-design/07-security-sandboxing-and-permissions.md) (Seatbelt/Landlock Sandboxes, Path Confinement, Allow/Ask/Deny Rules).
3. **"How would you architect a coding agent to support both a Terminal UI (TUI) and a Desktop GUI?"**  
   $\to$ See [Module 01](file:///Users/adityakinjawadekar/Documents/100xcode/zephyr/docs/system-design/01-architecture-overview-and-stack.md) & [Module 10](file:///Users/adityakinjawadekar/Documents/100xcode/zephyr/docs/system-design/10-case-studies-existing-agent-harnesses.md) (Event-Sourced Headless Core with Presentation-Agnostic Views).
4. **"Why do smaller or open-weight models frequently fail at tool calling, and how do you mitigate this in the harness?"**  
   $\to$ See [Module 04](file:///Users/adityakinjawadekar/Documents/100xcode/zephyr/docs/system-design/04-tool-design-reliability-and-repair.md) (Deterministic Coercion, Schema Flattening, Edit Format Fallback Chains).
5. **"How do you provide repository-wide context without blowing the prompt budget?"**  
   $\to$ See [Module 05](file:///Users/adityakinjawadekar/Documents/100xcode/zephyr/docs/system-design/05-code-intelligence-and-repo-indexing.md) (Tree-sitter Outlines, PageRank Repo Maps, Trigram Indexing).
6. **"How do you benchmark and evaluate if an AI agent is actually getting better?"**  
   $\to$ See [Module 09](file:///Users/adityakinjawadekar/Documents/100xcode/zephyr/docs/system-design/09-observability-state-and-evals.md) (Trajectory Replay, Token Efficiency Ratios, Terminal-Bench / SWE-bench Pitfalls).
