<div align="center">

<img src="assets/syncalibur-logo.png" alt="Syncalibur Emblem" width="220" />

# Syncalibur

**The Zero-Data-Loss Git Sentinel & Autopilot for Developers**  
*Autonomous Background Sync • AST Semantic Commits • Shannon Entropy Secret Shield • Instant Time Machine Rollback*

[![CI Status](https://img.shields.io/github/actions/workflow/status/YusufKaya00/Syncalibur/ci.yml?branch=main&style=for-the-badge&label=CI%20Verify)](https://github.com/YusufKaya00/Syncalibur/actions)
[![Platform](https://img.shields.io/badge/Platform-Linux%20%7C%20macOS%20%7C%20Windows-0ea5e9?style=for-the-badge)](#cli-command-reference)
[![Engine](https://img.shields.io/badge/Core-Go%20Standard%20Library%20(0%20Deps)-00ADD8?style=for-the-badge&logo=go&logoColor=white)](#architecture)
[![Git Remotes](https://img.shields.io/badge/Multi--Remote-GitHub%20%2B%20GitLab%20%2B%20Self--Hosted-10b981?style=for-the-badge)](#multi-remote-push)
[![License](https://img.shields.io/badge/License-MIT-amber?style=for-the-badge)](LICENSE)

*Syncalibur is a developer workstation flight recorder and autonomous Git autopilot. It runs silently in the background, detects cognitive typing pauses, validates code syntax, quarantines credentials with Shannon entropy, creates AST-aware Conventional Commits, and pushes to all your Git remotes in real-time. If you ever make a mistake, its zero-cost Time Machine rewinds commits in 1 keystroke without losing a single uncommitted character.*

</div>

---

## ⚡ Quick Install (30 Seconds)

Install globally in one command:

```bash
curl -fsSL https://raw.githubusercontent.com/YusufKaya00/Syncalibur/main/install.sh | bash
```

Or install via standard Go tooling:

```bash
go install github.com/YusufKaya00/Syncalibur/cmd/syncalibur-engine@latest
```

Or clone and watch your repository:

```bash
git clone https://github.com/YusufKaya00/Syncalibur.git ~/Syncalibur
cd ~/Syncalibur
./syncalibur.sh install /path/to/your/project
```

Once installed, the global `syncalibur` CLI is accessible anywhere in your terminal.

---

## 🎯 Why Syncalibur? (Problems It Solves)

| The Problem | How Traditional Tools Fail | The Syncalibur Solution |
| :--- | :--- | :--- |
| **Lost Uncommitted Code** (OS crash, battery dies, IDE freezes) | Manual `git commit` is forgotten during deep work; simple auto-sync scripts commit broken, half-written syntax. | **Zero-Data-Loss Autopilot**: 5-second cognitive settle debounce, syntax balance verification, AST conventional commit, and instant multi-remote push. |
| **Credential Leakage** (`.env`, OpenAI keys, AWS secrets) | Blind `git add -A` scripts push API keys to public repositories within seconds. | **Shannon Entropy Secret Shield**: Mathematical entropy scanning ($H \ge 4.4\text{ bit/char}$) + signature detection quarantines secrets *before* staging. |
| **Fear of Losing Control** ("What did it just commit?") | Hard to undo without git reflog gymnastics, often destroying working tree state. | **Time Machine Checkpoints**: Type `syncalibur undo [N]` or press `[z]` in the live cockpit to safely rewind HEAD with soft reset (preserves all working tree edits). |
| **Commit History Pollution** | Naive scripts fill Git history with 500 spam commits like `update file.js (+1/-1)`. | **AST & Function-Aware Engine**: Inspects diffs to generate meaningful Conventional Commits (`feat(auth): update ValidateToken() (+12/-3)`). |
| **Background Service Friction** | Some developers prefer working on-demand rather than running a background daemon. | **One-Shot Quicksave**: `syncalibur save [msg]` runs the full security scan, AST commit, and push in `<150ms`. |

---

## 🖥 Terminal Preview (`syncalibur ui`)

Launch the calm, high-contrast cockpit anytime by running `syncalibur ui` (or `syncalibur status`):

<div align="center">
  <img src="assets/syncalibur-terminal.png" alt="Syncalibur Cockpit Preview" width="960" />
</div>

*Features embedded 24-bit TrueColor ANSI logo, live commit velocity sparkline (`▄··▄▆█▂`), telemetry metrics, and instant interactive hotkeys (`[q]` Quit, `[p]` Pause, `[z]` Undo, `[s]` Sync, `[b]` Pull, `[d]` Doctor).*

---

## 🛠 Core Capabilities

### 1. Zero-Data-Loss Autopilot
- Runs silently with near-zero CPU usage (`inotifywait` / `fswatch` native hybrid with polling fallback).
- Detects bursts of coding activity. Once you pause typing for 5 seconds (configurable settle window), it triggers a safe commit cycle.
- Automatically pushes to multiple configured Git remotes in parallel (e.g. GitHub + GitLab + Private Server).

### 2. Time Machine Local Checkpoints (`syncalibur undo`)
- Every auto-commit and manual save records a zero-cost timeline checkpoint ref (`refs/syncalibur/timeline/<epoch>`).
- Did you accidentally commit something you didn't want?
  ```bash
  syncalibur undo     # Reverts 1 commit safely
  syncalibur undo 3   # Reverts 3 commits
  ```
- **Guaranteed non-destructive:** Uses `git reset --soft` so your working tree, open editor files, and uncommitted changes are **100% preserved**. A redo reference is always saved if you want to jump forward again.
- In the TUI cockpit, just press **`[z]`** for instant rollback.

### 3. One-Shot Quicksave (`syncalibur save`)
For developers who prefer on-demand saves without a continuous daemon:
```bash
syncalibur save                        # AST auto-generates commit message & pushes
syncalibur save "feat: add user login" # Custom message + secret scan + instant push
syncalibur save --dry-run              # Inspect what would be committed without pushing
```
Runs in `<150ms` using the compiled Go engine.

### 4. Shannon Entropy Secret Shield
Pre-flight security gate that intercepts secrets before `git add`:
- **Mathematical Entropy Analysis:** Computes Shannon entropy per string token. High-randomness strings ($H \ge 4.4$) like private keys, hex tokens, and base64 hashes are flagged immediately.
- **Pattern Signatures:** Detects AWS keys (`AKIA...`), GitHub PATs (`ghp_...`), OpenAI keys (`sk-...`), Slack tokens, private keys (`-----BEGIN PRIVATE KEY-----`), and `.env` files.
- **Selective Quarantine:** Quarantines sensitive files from staging using path exclusion rules while allowing clean source files to be committed.

### 5. Universal Syntax & Lexical Gate
Ensures broken syntax is never committed:
- Detects unclosed braces, brackets, quotes, template literals, and unresolved merge conflict markers (`<<<<<<<`, `=======`).
- Multi-language verification: Go, TypeScript, React JSX/TSX, Python, JavaScript, Java, Bash, Rust, C/C++, and YAML.
- Holds commits back until you finish typing the closing brace or statement, keeping Git history buildable.

### 6. Multi-Repository Fleet Management
Supervise all your local repositories from a single terminal:
```bash
syncalibur fleet scan ~/projects       # Auto-discovers all Git repositories
syncalibur fleet add ~/projects/my-api # Enrolls repository into fleet
syncalibur fleet status                # Tabular dashboard of all repos, unpushed counts & dirty files
syncalibur fleet start                 # Starts background autostart supervisor for all enrolled repos
```

### 7. 15ms Native Git Hook (`syncalibur hook install`)
Want protection even when running standard `git commit` in your shell?
```bash
syncalibur hook install
```
Installs a compiled sub-15ms `pre-commit` hook that executes the Go engine's Secret Shield and Syntax Gate on standard Git operations.

---

## 📖 CLI Command Reference

```text
Developer Ergonomics & Flight Recorder:
  syncalibur save [msg]             One-shot quicksave: secret scan + syntax check + AST commit + push
  syncalibur undo [N]               Time Machine: safely revert N commits via soft reset (preserves working tree)
  syncalibur push [msg]             Alias to 'save' (one-shot quicksave & multi-remote push)
  syncalibur commit [msg]           Local preflight check & commit (skips network push)
  syncalibur check | preflight      Run pre-push inspection (syntax hold, secret shield, diff, proposed msg)
  syncalibur checkpoint [list|save] List or create safety rollback checkpoints
  syncalibur release [mode]         Automated SemVer bump, CHANGELOG generation & Git release tag
  syncalibur version | -v           Print official release version & platform milestone
  syncalibur tidy                   Squash unpushed local micro-commits into 1 clean conventional commit
  syncalibur hook <install|status>  Install universal pre-commit & prepare-commit-msg Git hooks

Autonomous Daemon & Fleet Control:
  syncalibur install [path]         Configure OS auto-start, compile Go engine, link global CLI & start now
  syncalibur status [--static]      Launch interactive cockpit in TTY (or static dashboard with --static)
  syncalibur ui                     Launch interactive TrueColor cockpit & live log viewer
  syncalibur mode [name]            Inspect or toggle operating architecture: autopilot | manual
  syncalibur logs                   Stream live colorized telemetry
  syncalibur fleet <subcmd>         Manage multi-repo fleet: status | add <path> | remove <path> | scan | start | stop
  syncalibur doctor                 Run 12-point system diagnostic health check
  syncalibur notify [msg]           Send an instant test notification to Telegram / Discord / Slack
  syncalibur pr [base|--preview]    Open a structured Pull Request using Syncalibur's PR Template
  syncalibur pull                   Trigger immediate bidirectional remote fetch & safe rebase
  syncalibur pause | resume         Pause or resume background auto-commits without stopping the service
  syncalibur alert [--test]         Display full status + Sync Failed diagnostic report
  syncalibur restart | stop | start Control background daemon lifecycle
  syncalibur uninstall              Stop Syncalibur and remove OS auto-start service
```

---

## ⚙ Configuration (`syncalibur.env`)

Configure Syncalibur globally or per-repository (via `.syncalibur` in any repo root):

```bash
# Path to watch:
SYNCALIBUR_REPO_PATH="."

# Operating Mode: "autopilot" (continuous background stream) or "manual" (on-demand quicksave)
SYNCALIBUR_MODE="autopilot"

# Quiescent settle window (seconds) after typing stops before auto-committing:
SYNCALIBUR_BURST_COOLDOWN=5

# Bidirectional sync (fetches & safely rebases clean remote changes):
SYNCALIBUR_BIDIRECTIONAL="true"
SYNCALIBUR_PULL_INTERVAL=45

# AST Semantic Conventional Commits:
SYNCALIBUR_SEMANTIC_COMMITS="true"
SYNCALIBUR_COMMIT_STYLE="conventional"

# Optional local Ollama AI commit messages:
SYNCALIBUR_AI_COMMIT="false"
SYNCALIBUR_OLLAMA_MODEL="qwen2.5-coder:1.5b"

# Multi-Channel Alert Notifications (Telegram / Discord / Slack):
SYNCALIBUR_TELEGRAM_BOT_TOKEN=""
SYNCALIBUR_TELEGRAM_CHAT_ID=""
SYNCALIBUR_WEBHOOK_URL=""
```

---

## ✚ Diagnostic Doctor (`syncalibur doctor`)

Syncalibur includes a built-in 12-point diagnostic engine that audits your workstation environment:

```text
  ╭────────────────────────────────────────────────────────────────────────────────────────╮
  │  ✚  S Y N C A L I B U R   D O C T O R   —   1 2 - P O I N T   D I A G N O S T I C S  │
  ╰────────────────────────────────────────────────────────────────────────────────────────╯

    [✓] Go Native Engine           : Compiled & active (syncalibur)
    [✓] Filesystem Event Watcher   : Polling/inotify mode active
    [✓] Git Repository State       : git version 2.43.0 • Working tree unlocked
    [✓] Daemon Process             : Running (syncalibur)
    [✓] OS Auto-Start Service      : systemd / launchd registered
    [✓] Provider Authentication    : GitHub (gh+SSH) • GitLab (SSH)
    [✓] Push Destinations          : Verified reachable
    [✓] Self-Watching Updater      : Active (every 120s)
    [✓] Global CLI Command         : ~/.local/bin/syncalibur linked
    [✓] Telegram & Webhooks        : Diagnostic notification channels verified
    [✓] Syntax & Secret Shield     : Active (Shannon entropy & lexical balance verified clean)
    [✓] DevOps & GitOps Guard      : Active (Kubernetes ports, Dockerfile, Terraform verified)

  ────────────────────────────────────────────────────────────────────────────────────────
  Summary: 12 Passed  •  0 Warnings  •  0 Failures
```

---

## 🏛 Architecture

Syncalibur is engineered for maximum performance, security, and developer ergonomics:

```
Workstation File Changes
         │
         ▼
[Filesystem Event Watcher] (inotifywait / fswatch / stat polling)
         │
         ▼ (5s cognitive settle debounce)
[Preflight Security & Syntax Gate] (Compiled Go Core - Pure Standard Library)
   ├── 1. Shannon Entropy Secret Shield (H >= 4.4 bit/char quarantine)
   ├── 2. Universal Lexical Balance Gate (unclosed braces, conflict markers)
   └── 3. Static DevOps Manifest Inspection (ports, Dockerfile, ConfigMaps)
         │ (Clean)
         ▼
[AST & Function-Aware Conventional Commit]
   └── Inspects AST / diff hunks: "feat(auth): update ValidateToken() (+12/-3)"
         │
         ├──► Record Time Machine Ref (refs/syncalibur/timeline/<epoch>)
         │
         ▼
[Safe Multi-Remote Push Engine]
   ├── Push to GitHub (main)
   ├── Push to GitLab (main)
   └── Push to Self-Hosted Remote (main)
         │ (If Offline)
         └──► Offline Shield: Queue commits locally, flush on TCP reconnect
```

- **Zero External Dependencies:** Built with 100% Go standard library (`go/parser`, `go/ast`, `crypto/sha256`, `math`, `net`, `os/exec`).
- **Archive Notice:** The legacy 4-mode experimentation matrix (`developer`, `devops`, `pr`, `instant-push`) is preserved in the [`archive/multi-mode-system`](https://github.com/YusufKaya00/Syncalibur/tree/archive/multi-mode-system) branch for historical reference.

---

## 🤝 Contributing

Contributions, issues, and feature requests are welcome!
Please check the [Contributing Guidelines](CONTRIBUTING.md) and open a pull request.

---

## 📜 License

Licensed under the [MIT License](LICENSE). Copyright © 2026 Yusuf Kaya.
