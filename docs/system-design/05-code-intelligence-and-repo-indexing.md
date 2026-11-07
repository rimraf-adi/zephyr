# 05: Codebase Understanding & Indexing (Loupe Engine)

> **Focus Area:** Tree-Sitter AST, Aider-Style Repo Maps, PageRank Graph Ranking, Trigram Search, SCIP Symbols.

---

## 1. The Code Understanding Problem

An AI coding agent facing a repository with 500,000 lines of code cannot stuff the entire codebase into its context window. It needs:
1. **High-Level Orientation:** A compact summary of files, classes, structs, and function signatures.
2. **Deep Semantic Search:** Pinpointing definitions, call sites, and type implementations.
3. **Exact Token Matching:** Blazing-fast regex and symbol queries.

---

## 2. Architecture of the Indexer (Loupe)

```
                            Source Files on Disk
                                     │
                     ┌───────────────┴───────────────┐
                     ▼                               ▼
            Tree-Sitter Grammars             Trigram / Regex Index
           (Rust, Python, TS, Go)             (Ripgrep / Zoekt)
                     │                               │
                     ▼                               ▼
           Symbol & Tag Extractor             Exact Match Hits
             (`tags.scm` queries)                    │
                     │                               │
                     ▼                               │
             Symbol Reference Graph                  │
             (petgraph PageRank)                     │
                     │                               │
                     ▼                               │
            Budgeted Repo Map                        │
                     │                               │
                     └───────────────┬───────────────┘
                                     ▼
                        Unified Agent Tools:
                    - repo_map(token_budget)
                    - symbol_lookup(name)
                    - find_references(symbol)
```

---

## 3. Tree-Sitter Tag Extraction & Outlines

Rather than running full language servers (LSP), which are resource-heavy and slow to boot, Zephyr uses `tree-sitter` grammars:
- Extracts definitions: `struct`, `enum`, `fn`, `class`, `interface`, `trait`.
- Extracts export visibility (`pub`).
- Produces lightweight file outlines (e.g., function signatures without internal implementation bodies).

---

## 4. Aider-Style PageRank Repository Map

To generate a global overview of the repository that fits within a 1,000–2,000 token budget:
1. **Graph Construction:** Build a directed graph where nodes are source files and edges represent symbol references/imports (using `petgraph`).
2. **PageRank Scoring:** Run PageRank over the graph. Core utility files and shared types receive the highest ranking.
3. **Greedy Budget Fitting:** Starting from the highest-ranked files, add file paths and top symbol signatures to the prompt until the allocated token budget is exhausted.
4. **Context Adaptation:** As the user and agent discuss specific files, boost those files' scores in the graph so the repo map dynamically focuses on the active subsystem.

---

## 5. SCIP vs. Tree-Sitter vs. Trigrams

| Indexing Method | Precision | Build Cost | Memory Overhead | Use Case |
|---|---|---|---|---|
| **Trigram Search (Zoekt)** | Low (Lexical) | Fast (~seconds) | Moderate | Exact string & regex queries across millions of lines. |
| **Tree-Sitter Tags** | Medium (Syntactic) | Very Fast (ms) | Low | Structural outlines, file skeletons, signature summaries. |
| **SCIP / LSP** | High (Semantic) | Slow (minutes) | High | True cross-file "Go to Definition" and type resolution. |

**System Design Recommendation:**  
Start with Tree-sitter outlines and ripgrep internals (`grep-searcher`). Add PageRank ranking next. Only introduce SCIP or full LSP when deep cross-crate semantic resolution is strictly required.
