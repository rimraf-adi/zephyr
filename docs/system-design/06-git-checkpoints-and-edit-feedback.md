# 06: Git Checkpoints, Undo & Edit Feedback Loops

> **Focus Area:** Automatic Git Snapshotting, Shadow Commits, Non-Destructive Undo, Cargo Check JSON Diagnostics, Test Feedback.

---

## 1. Edit Feedback Loops: The Secret to High Resolve Rates

An agent that writes code and immediately concludes the turn without verification has a low success rate. Benchmark research shows that **deterministic feedback loops** dramatically increase task completion while reducing total turns.

```
Agent Proposes Edit
        │
        ▼
Apply Diff to Workspace
        │
        ▼
Run Fast Feedback Checks
  - Rust: `cargo check --message-format=json`
  - Python: `ruff check` / `pyright`
  - TypeScript: `tsc --noEmit`
        │
        ├── Errors Detected ──► Trim errors & feed back in same turn
        │                       (Model fixes mistakes immediately)
        │
        └── Clean ────────────► Run unit tests / Complete turn
```

---

## 2. Low-Cost Diagnostics: `cargo check --message-format=json`

Managing an interactive LSP server (e.g. `rust-analyzer`) in background processes is prone to synchronization races and crashes. For Rust codebases:
- Run `cargo check --message-format=json` directly after an edit.
- Filter and parse JSON diagnostics for `"level": "error"`.
- Extract file path, line number, primary message, and source code snippet.
- Feed this back as tool output. The model resolves compiler errors before the user even inspects the changes.

---

## 3. Git Checkpoints & Rollback Architecture

### The Problem
Agents can make destructive edits or head down a dead-end reasoning path. The user must be able to roll back instantly with `/undo`.

### The Solution: Snapshotting Before Write Turns
Before executing any tool that modifies the file system:
1. **Dirty Check:** Check if the working tree has uncommitted modifications.
2. **Shadow Commit / Stash:** 
   - Execute `git stash create` or commit a shadow ref (`refs/zephyr/checkpoints/<timestamp>`).
   - Store the ref SHA alongside the turn ID in the session log.
3. **Rollback (`/undo`):**
   - The user or agent can revert to any prior checkpoint cleanly without losing history.

```
Turn 1 (Read)        ──► No modification
Turn 2 (Edit lib.rs) ──► Checkpoint ref: `refs/zephyr/ckpt_01` created ──► Edit applied
Turn 3 (Edit cli.rs) ──► Checkpoint ref: `refs/zephyr/ckpt_02` created ──► Edit applied
User types `/undo`   ──► Hard reset to `ckpt_01` (Turn 3 changes reverted)
```

---

## 4. Worktree Isolation for Subagents

When multiple subagents operate concurrently, sharing the same working directory leads to file race conditions and compiler lock contention.

- **Solution:** Provision an isolated `git worktree` per subagent (`git worktree add ../agent-task-branch`).
- Each subagent operates on its own branch in an isolated workspace.
- The parent agent merges the child branch upon successful completion.

---

## 5. Git Library Trade-offs in Rust

| Approach | Read Performance | Write Support | Verdict |
|---|---|---|---|
| **`gix` (Gitoxide)** | Extremely Fast (Pure Rust) | Incomplete (writes/merges still in development) | Best for fast status, diffs, and rev-parsing. |
| **`git2` (libgit2)** | Fast | Full | Requires C toolchain / OpenSSL linking issues. |
| **Subprocess `git` CLI** | Baseline | 100% Complete | **Recommended default for writes.** Shell out to `git` CLI for commits, checkouts, and stashes. |
