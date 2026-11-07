# 04: Tool Design, Schema Shaping & Tool-Call Reliability

> **Focus Area:** Tool Schema Design, Multi-Edit Tools, Repairing Malformed Tool Calls, Constrained Decoding Trade-offs, Model-Adaptive Scaffolding.

---

## 1. Tool-Call Reliability with Small & Routed Models

While frontier models (Claude 3.5 Sonnet, GPT-4o) follow tool schemas accurately, open-weight models (Llama 3 8B/70B, Qwen 2.5 Coder, DeepSeek) running on free or routed tiers frequently emit malformed tool calls:
- Numbers wrapped in quotation marks (`"line": "42"`).
- JSON objects encoded as escaped strings (`"args": "{\"path\": \"src/lib.rs\"}"`).
- Invented extra fields or hallucinations in nested schemas.
- Minor typos in tool names (`readfile` instead of `read_file`).

A robust agent harness **never fails immediately on a malformed call**. It implements a deterministic **Validate $\to$ Coerce $\to$ Retry** pipeline.

---

## 2. The Validate $\to$ Coerce $\to$ Retry Pipeline

```
Raw Model Tool Call
        │
        ▼
┌─────────────────────────────────┐
│ 1. Schema Validation            │
│    (jsonschema crate)           │
└───────┬─────────────────────────┘
        │
        ├── Pass ──► Execute Tool
        │
        └── Fail
                │
                ▼
        ┌─────────────────────────────────┐
        │ 2. Deterministic Coercion       │
        │    - Unquote numbers/booleans   │
        │    - Parse nested JSON strings  │
        │    - Fuzzy-match tool names     │
        └───────┬─────────────────────────┘
                │
                ├── Pass ──► Execute Tool
                │
                └── Fail
                        │
                        ▼
                ┌─────────────────────────────────┐
                │ 3. Targeted Feedback Retry      │
                │    Inject precise schema error  │
                │    and request immediate retry  │
                └─────────────────────────────────┘
```

---

## 3. Tool Schema Shaping: Flat vs. Nested

A critical finding in recent agent research is that **schema structure directly dictates model success**:
- **Avoid deep nesting:** Newer models often hallucinate additional attributes inside complex nested arrays (e.g., `edits: [{ action: "replace", options: { case_sensitive: true } }]`).
- **Keep parameters flat:** Flat, explicit schemas minimize parse errors and token overhead.
- **Example: Multi-Edit Tool Schema:**
  ```json
  {
    "type": "object",
    "properties": {
      "path": { "type": "string", "description": "Absolute or relative file path" },
      "replacements": {
        "type": "array",
        "items": {
          "type": "object",
          "properties": {
            "old_str": { "type": "string", "description": "Exact text to find and replace" },
            "new_str": { "type": "string", "description": "Replacement content" }
          },
          "required": ["old_str", "new_str"]
        }
      }
    },
    "required": ["path", "replacements"]
  }
  ```

---

## 4. Edit Tool Formats & Fallback Strategy

Models exhibit different affinities for edit formats based on their post-training:

| Edit Format | How it Works | Pros / Cons |
|---|---|---|
| **Exact Match Search & Replace** | Model supplies exact `old_str` and `new_str`. | Safe, deterministic, whitespace-sensitive. |
| **Hashline** | Target lines are referenced by a short content hash. | Resilient to line miscounts; strong on smaller models. |
| **Unified Diff / Patch** | Standard diff syntax (`@@ -1,5 +1,6 @@`). | Models miscalculate chunk line counts; high failure rate. |
| **Whole-File Rewrite** | Model returns the entire updated file. | High token cost; 100% reliable on small files. |

### The Edit Fallback Chain
If an exact match search/replace fails because the model omitted whitespace or indentation:
1. Fuzzy-match the `old_str` with whitespace normalization.
2. If unresolved, return the closest matching lines to the model with context.
3. If failure persists, fall back to whole-file replacement.

---

## 5. Constrained Decoding vs. Free Generation

- **Grammar-Based Constrained Decoding (e.g., Outlines, vLLM guided decoding):** Guarantees 100% valid JSON syntax.
- **The "Constraint Tax":** Forcing an LLM through a rigid grammar can degrade reasoning quality or cause models to stop calling tools altogether.
- **Best Practice:** Constrain arguments once the tool call is initiated, but never constrain the freeform thinking/reasoning before the call.
