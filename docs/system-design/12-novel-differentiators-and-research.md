# 12: Advanced Differentiators, Speculation & Research Frontiers

> **Focus Area:** Novel Architectural Hypotheses, Speculative Execution, Hash-Pinned Memory, Online Bandit Adaptation, Research Literature Citations.

---

## 1. Candidate Differentiators (Cutting-Edge Frontiers)

### A. Graph-Predicted Speculative Execution
- **The Problem:** Tool calls are synchronous round trips. Waiting for an LLM to generate a tool call, executing it, and waiting for the next turn introduces massive latency.
- **The Technique:** Use the Tree-sitter symbol graph (Loupe) to **predict what the model will ask for next**:
  - When the model reads `symbol("process_payment")`, proactively pre-stage the AST signatures of its callees and relevant test files.
  - When the model edits a function, speculatively trigger the corresponding test suite in the background.
  - Pre-staged results are stored in an in-memory cache and only enter context if requested, achieving zero context waste.

### B. Hash-Pinned, Staleness-Aware Memory Compaction
- **The Problem:** Historical compaction summaries write prose descriptions of code (e.g., *"The user fixed the auth token validation in `auth.rs`"*). As the codebase evolves, this memory becomes silently incorrect, causing hallucinations.
- **The Technique:** Compaction templates store **Symbol IDs + BLAKE3 Content Hashes** instead of prose:
  - Format: `Symbol(AuthService::validate) -> src/auth.rs:45 [hash: a7f8c1]`.
  - When the agent later queries this memory, the harness checks the current file hash.
  - If changed, it warns the agent: `[Notice: Symbol AuthService::validate was modified since this summary was created. Re-reading current definition.]`

### C. Online Edit-Format Bandit Adaptation
- **The Problem:** Fixed edit schemas fail when routing across diverse open-weight models.
- **The Technique:** Implement an online multi-armed bandit inside the harness:
  - If a model fails an edit format twice in a row, automatically retry using hashline or whole-file rewrite.
  - Maintain a persistent local SQLite table mapping `(model_name, edit_format) -> success_rate`.
  - The harness automatically converges on the optimal edit format for any free or routed model.

---

## 2. Table-Stakes Techniques (Adopt, Do Not Reinvent)

| Technique | Prior Art | System Design Insight |
|---|---|---|
| **Speculative Tool Streaming** | `zeph` (Rust), PASTE, Spex | Dispatch read-only tools while the model is still streaming partial tool-call JSON chunks. |
| **Read Deduplication** | `lingshu-tools` (Rust), Claude Code | Return a `"File unchanged since last read"` stub using file mtime caches; warn after 3 identical reads. |
| **Model Cascade** | PLANTOD, lean-agent | Route high-level planning to a strong model, and bulk mechanical edits to a cheap/fast model. |
| **Structured Compaction** | CooperBench, Morph | Use structured markdown tables (Goals, Symbols, Negative Results) instead of freeform prose summaries. |

---

## 3. Comprehensive Academic & Industry Sources

1. **Harness & Loop Design:**
   - *An Empirical Study of Harness Design for Coding Agents* (arXiv:2609.20804)
   - *Glean: Building an Efficient Enterprise Agent Harness* (Glean Engineering)
   - *Claude Code Agent Harness Architecture* (Wavespeed AI Deep Dive)
   - *Pi: The Minimal Agent Within OpenClaw* (Armin Ronacher)
2. **Edit Formats & Tool Reliability:**
   - *The Harness Problem: Benchmarking Edit Tool Formats* (Can.ac Research)
   - *SWE-Edit: Action-Space Shaping for Coding Agents* (arXiv:2604.26102)
   - *Better Models, Worse Tools: Tool Schema Drift* (Lucumr)
   - *Toolrails: Deterministic Tool-Call Repair* (Toolrails Project)
3. **Speculation & Indexing:**
   - *PASTE: Pattern-Aware Speculative Tool Execution*
   - *Spex: Zero-Token Pre-execution Engine* (Spex Lab)
   - *TomasuLLM: Out-of-Order Execution in LLM Loops* (arXiv:2609.38201)
   - *CodeGraph: Pre-built Code Graphs for Agents* (Scraix)
4. **Benchmarks & Governance:**
   - *Terminal-Bench: A Benchmark for AI Agents in Terminal Environments* (arXiv:2601.11868)
   - *SWE-bench: Can Language Models Resolve Real-world GitHub Issues?* (Princeton NLP)
   - *Agentic Harness Engineering (AHE)* (AlphaXiv Audio)
