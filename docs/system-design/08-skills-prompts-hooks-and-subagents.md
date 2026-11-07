# 08: Skills, Hierarchical Prompts, Hooks & Subagents

> **Focus Area:** Hierarchical System Prompts, Agent Skills Standard, Lazy Context Loading, Lifecycle Hooks, Isolated Subagents.

---

## 1. System Prompt Architecture & Layered Precedence

Hardcoding system prompts into the binary creates rigid, uncustomizable agents. Zephyr organizes system prompts into a **layered inheritance hierarchy**:

```
                       ┌─────────────────────────┐
                       │  Zephyr Built-in Core   │
                       │    (Lean Base Rules)    │
                       └────────────┬────────────┘
                                    │
                                    ▼
                       ┌─────────────────────────┐
                       │ Global User Preference  │
                       │  (~/.zephyr/system.md)  │
                       └────────────┬────────────┘
                                    │
                                    ▼
                       ┌─────────────────────────┐
                       │ Project/Repo Directive  │
                       │   (.zephyr/system.md    │
                       │     or AGENTS.md)       │
                       └────────────┬────────────┘
                                    │
                                    ▼
                       ┌─────────────────────────┐
                       │ Task/CLI Flag Override  │
                       │  (--system-prompt <..>) │
                       └─────────────────────────┘
```

### Prompt-Cache Prefix Stability
To maximize KV cache hits across turns, the layered prompt engine ensures that **static directives remain identical at the start of the prompt**. Dynamic state (git branch, active time) is strictly injected at the boundary between instructions and conversation history.

---

## 2. Agent Skills: On-Demand Capability Expansion

Skills are self-contained folders that teach the agent specialized workflows (e.g. database migrations, Rust unsafe auditing, cloud deployment):

### Structure of a Skill
```
.zephyr/skills/
  └── rust-auditor/
      ├── SKILL.md          # YAML frontmatter + markdown instructions
      ├── scripts/          # Helper scripts or linters
      └── references/       # API cheat-sheets or docs
```

### The Lazy Context Loading Pattern
Preloading dozens of skill manuals into the system prompt exhausts tokens before the agent does any work.
- **Registration:** At boot, Zephyr registers only skill headers (`name` and a 1-sentence `description`) into a lightweight catalog.
- **Dynamic Activation:** When the model needs a skill, it calls `activate_skill("rust-auditor")`, which injects the full `SKILL.md` instructions into context on-demand.

---

## 3. Lifecycle Hooks

Hooks enable deterministic policies, formatting, and security checks without model overhead:

```
User Prompt ──► [Hook: UserPromptSubmit] ──► Inject Context / Validate
                                                  │
                                                  ▼
                                            Agent Loop
                                                  │
                                                  ▼
Model Proposes Tool ──► [Hook: PreToolUse] ──► Allow / Deny / Rewrite Args
                                                  │
                                                  ▼
                                             Execute Tool
                                                  │
                                                  ▼
Tool Result ──────────► [Hook: PostToolUse] ─► Format / Trim / Auto-test
```

- **Example `PostToolUse` for Rust:** Automatically run `rustfmt` after any write to a `.rs` file so diffs stay clean.

---

## 4. Subagent Architecture: Isolated Contexts

When a task requires deep exploration or parallel research, delegating to subagents preserves the parent's context window:

```
┌────────────────────────────────────────────────────────┐
│                      Parent Agent                      │
│               (Coordinates High-Level Goal)            │
└───────────────┬────────────────────────┬───────────────┘
                │                        │
       Spawns Subagent 1        Spawns Subagent 2
       (Search Codebase)        (Run Benchmark)
                │                        │
                ▼                        ▼
     ┌──────────────────────┐ ┌──────────────────────┐
     │ Subagent 1 Context   │ │ Subagent 2 Context   │
     │ - 40 turns of search │ │ - 20 turns of logs   │
     └──────────┬───────────┘ └──────────┬───────────┘
                │                        │
                └───────────┬────────────┘
                            │ Returns only 2-line summary!
                            ▼
              Parent Context Stays Tiny & Clean!
```

- Subagents run in isolated workspaces (or git worktrees).
- Subagents do not stream their internal tool transcripts to the parent—only the final synthesized answer.
