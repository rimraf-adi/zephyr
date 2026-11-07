# Zephyr

Next-generation AI coding agent harness in Rust, supporting both Terminal UI (TUI) and Desktop Graphical (GUI) workflows.

---

## Documentation & System Design

The architecture and engineering design of Zephyr are documented under [`docs/`](file:///Users/adityakinjawadekar/Documents/100xcode/zephyr/docs):

- **[Zephyr Living Specification](file:///Users/adityakinjawadekar/Documents/100xcode/zephyr/docs/zephyr-spec.md):** Active features, roadmap, skills standard, layered prompts, and 3-tier memory engine.
- **[System Design Interview & Architecture Guide](file:///Users/adityakinjawadekar/Documents/100xcode/zephyr/docs/system-design/00-index.md):** 12-part system design curriculum covering real-world coding-agent architecture:
  1. [`01-architecture-overview-and-stack.md`](file:///Users/adityakinjawadekar/Documents/100xcode/zephyr/docs/system-design/01-architecture-overview-and-stack.md) - System goals, crate layout, decoupled presentation layers (TUI/GUI/Headless).
  2. [`02-agent-loop-and-context-management.md`](file:///Users/adityakinjawadekar/Documents/100xcode/zephyr/docs/system-design/02-agent-loop-and-context-management.md) - The agent loop, turn assembly, token budget enforcement, disk truncation, KV cache prefix stability.
  3. [`03-llm-client-protocols-and-routing.md`](file:///Users/adityakinjawadekar/Documents/100xcode/zephyr/docs/system-design/03-llm-client-protocols-and-routing.md) - SSE streaming, MCP/ACP protocols, backoff, multi-model fallback chains.
  4. [`04-tool-design-reliability-and-repair.md`](file:///Users/adityakinjawadekar/Documents/100xcode/zephyr/docs/system-design/04-tool-design-reliability-and-repair.md) - Tool schemas, multi-edit tools, tool-call repair, constrained decoding trade-offs.
  5. [`05-code-intelligence-and-repo-indexing.md`](file:///Users/adityakinjawadekar/Documents/100xcode/zephyr/docs/system-design/05-code-intelligence-and-repo-indexing.md) - Codebase understanding, Tree-sitter outlines, PageRank repo maps, trigram search.
  6. [`06-git-checkpoints-and-edit-feedback.md`](file:///Users/adityakinjawadekar/Documents/100xcode/zephyr/docs/system-design/06-git-checkpoints-and-edit-feedback.md) - Edit feedback loops (`cargo check` JSON), git checkpoints, `/undo`, worktree isolation.
  7. [`07-security-sandboxing-and-permissions.md`](file:///Users/adityakinjawadekar/Documents/100xcode/zephyr/docs/system-design/07-security-sandboxing-and-permissions.md) - OS sandboxing (macOS Seatbelt, Linux Landlock), path confinement, command permissions.
  8. [`08-skills-prompts-hooks-and-subagents.md`](file:///Users/adityakinjawadekar/Documents/100xcode/zephyr/docs/system-design/08-skills-prompts-hooks-and-subagents.md) - Hierarchical system prompts, Agent Skills standard with lazy loading, lifecycle hooks, subagents.
  9. [`09-observability-state-and-evals.md`](file:///Users/adityakinjawadekar/Documents/100xcode/zephyr/docs/system-design/09-observability-state-and-evals.md) - Tracing, JSONL event-sourcing, replay, benchmarks (Terminal-Bench, SWE-bench).
  10. [`10-case-studies-existing-agent-harnesses.md`](file:///Users/adityakinjawadekar/Documents/100xcode/zephyr/docs/system-design/10-case-studies-existing-agent-harnesses.md) - Deep-dive case studies: Codex-rs, Pi, mini-SWE-agent, OpenCode, Goose, OpenHands, Aider.
  11. [`11-borrow-vs-build-matrix.md`](file:///Users/adityakinjawadekar/Documents/100xcode/zephyr/docs/system-design/11-borrow-vs-build-matrix.md) - Engineering borrow vs build decision matrix, high-leverage differentiators.
  12. [`12-novel-differentiators-and-research.md`](file:///Users/adityakinjawadekar/Documents/100xcode/zephyr/docs/system-design/12-novel-differentiators-and-research.md) - Advanced research: graph-predicted speculation, hash-pinned compaction, bandit edit adaptation.
