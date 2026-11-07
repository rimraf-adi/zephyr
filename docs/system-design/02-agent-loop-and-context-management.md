# 02: The Agent Loop & Context Management

> **Focus Area:** The Turn Loop, Token Budgets, Output Truncation to Disk, Compaction Condensers, Prompt Caching & Prefix Stability.

---

## 1. Context Management: The High-Leverage Failure Point

In system design interviews for agentic systems, a critical insight is:
> *Most hobby harnesses fail not because their models are weak, but because unmanaged tool outputs blow the context window, triggering unexpected termination, amnesia, or prohibitive costs.*

An empirical study across 176 configurations demonstrated that **context management provides its largest gains when context is constrained**, directly preventing context overflow failures.

---

## 2. Turn Assembly Anatomy

Every turn sent to the LLM must be assembled with deterministic ordering:

```
┌────────────────────────────────────────────────────────┐
│ 1. Static System Prompt (Instructions, Guidelines)     │ ◄── CACHED PREFIX
│ 2. Tool Schemas (JSON Schema of available tools)       │ ◄── CACHED PREFIX
├────────────────────────────────────────────────────────┤
│ 3. Working Memory / Active Goal / Current Plan         │ ◄── DYNAMIC INJECTION
│ 4. Environment Metadata (Timestamp, cwd, git branch)   │ ◄── DYNAMIC INJECTION
├────────────────────────────────────────────────────────┤
│ 5. Managed Conversation History                        │
│    - Compacted Summary (Older turns)                   │
│    - Verbatim History (Recent turns & Tool results)    │
├────────────────────────────────────────────────────────┤
│ 6. Latest User Turn / Tool Output Feedback             │
└────────────────────────────────────────────────────────┘
```

---

## 3. Tool-Output Handling (Disk Capping & Range Previews)

**Rule: Never allow unbounded tool outputs to enter the conversation context.**

When a tool produces large output (e.g., a 4,000-line test log or huge file):
1. **Truncate in Memory:** Cap the output string returned to the model (e.g., maximum 50–100 lines or ~2,000 tokens).
2. **Spill to Disk:** Write the complete raw output to a dedicated scratch directory (`.zephyr/scratch/<call_id>.log`).
3. **Structured Notice:** Inform the model:
   ```
   [Output truncated: showing lines 1-100 of 4,210. 
    Full output stored at: .zephyr/scratch/call_98f1a.log (184 KB).
    To inspect specific lines, call read_file_range(path, start, end).]
   ```
4. **Output Pruning:** Strip ANSI escape codes, deduplicate repeated stack traces, and preserve the head and tail of build logs.

---

## 4. Multi-Tier Compaction Strategies

When the total context usage approaches the safety limit (e.g., 75% of context window), the harness invokes a pluggable `Condenser`:

```
┌─────────────────────────────────────────────────────────────────┐
│                    TIERS OF COMPACTION                          │
├─────────────────────────────────────────────────────────────────┤
│ Tier 1: Baseline                                                │
│ Fail gracefully with an explicit budget warning.                │
├─────────────────────────────────────────────────────────────────┤
│ Tier 2: Output Eviction                                         │
│ Replace historical tool outputs with 1-line execution summaries.│
├─────────────────────────────────────────────────────────────────┤
│ Tier 3: Disk Fallback + Recall                                  │
│ Evict historical content to disk; provide a `recall` tool.      │
├─────────────────────────────────────────────────────────────────┤
│ Tier 4: Structured LLM Compaction                               │
│ Use a weak/fast model to distill history into a structured spec.│
├─────────────────────────────────────────────────────────────────┤
│ Tier 5: Context Reset with Handoff Artifact                     │
│ Complete state reset passing only active TODOs and touched files│
└─────────────────────────────────────────────────────────────────┘
```

### The Structured Compaction Template
Unstructured prose summaries lose crucial facts. Compaction must generate structured markdown:
- **ACTIVE TASK & GOAL:** Primary objective and explicit user constraints.
- **FILE MAP & KEY SYMBOLS:** Modified files, key types, and functions (`Symbol -> path:line:hash`).
- **APPLIED EDITS:** Diffs already finalized.
- **NEGATIVE RESULTS:** Hypotheses tested and discarded (prevents infinite repetition).
- **OPEN TODOS:** Immediate next steps.

---

## 5. Prompt Caching & Prefix Stability

Frontier models (Anthropic, Gemini, OpenAI) heavily discount prompt tokens that match the KV cache prefix. To maintain cache-hit rates:
1. **Append-Only History:** Never edit or reorder earlier messages in the transcript.
2. **Static-First Ordering:** Static system instructions and tool definitions come first; dynamic timestamps and cwd come last.
3. **Fixed Tool Ordering:** Never shuffle tool schemas between turns.
4. **Pinned Model Routing:** Keep the session pinned to a specific model so router rotations do not wipe the server-side KV cache.

---

## 6. Loop Robustness & Circuit Breakers

To avoid runaway costs or infinite agent loops:
- **Stuck Detection:** Track `(tool_name, hash(arguments))`. If called identically 3 consecutive times, interrupt with an execution error prompt.
- **Auto-Continuation:** If the model terminates with `finish_reason: length` (`max_output_tokens`), automatically inject a `"continue"` prompt and stitch outputs.
- **Hard Resource Limits:** Hard turn ceiling (e.g., 50 turns), max total token budget, and wall-clock timeout enforced via `tokio_util::sync::CancellationToken`.
