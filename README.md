<div align="center">

<img src="assets/syncalibur-logo.png" alt="Syncalibur Emblem" width="220" />

# Syncalibur

**Zero-Data-Loss Git Autopilot & Workstation Sentinel • Multi-Remote Sync Engine, Developer Flight Recorder & Live TUI Cockpit**

[![CI Status](https://img.shields.io/github/actions/workflow/status/YusufKaya00/Syncalibur/ci.yml?branch=main&style=for-the-badge&label=CI%20Verify)](https://github.com/YusufKaya00/Syncalibur/actions)
[![Platform](https://img.shields.io/badge/Platform-Linux%20%7C%20macOS%20%7C%20Windows-0ea5e9?style=for-the-badge)](#cli-command-reference)
[![Engine](https://img.shields.io/badge/Core-Go%201.23%20%2B%20Bash-00ADD8?style=for-the-badge&logo=go&logoColor=white)](#architecture--sync-pipeline)
[![Git Remotes](https://img.shields.io/badge/Multi--Remote-GitHub%20%2B%20GitLab%20%2B%20Self--Hosted-10b981?style=for-the-badge)](#what-is-syncalibur)
[![PRs Welcome](https://img.shields.io/badge/PRs-Protected%20Main%20%7C%201%20Review%20Required-a855f7?style=for-the-badge)](CONTRIBUTING.md)

*Forge-grade **Git auto-commit**, **auto-pusher**, and **Developer Flight Recorder** for your repositories — powered by a compiled Go core (`syncalibur`), AST-aware Conventional Commit generation, concurrent Multi-Repo Fleet orchestration, Telegram & Multi-Channel Webhooks, and an Interactive Live TUI Cockpit.*

</div>

---

## Terminal Preview (`syncalibur status`)

When you run `syncalibur status` (or `syncalibur logs`), the native Go engine renders [`assets/syncalibur-logo.png`](assets/syncalibur-logo.png) directly inside your terminal using **24-bit TrueColor ANSI escape sequences** alongside live repository telemetry, verified accounts, active operating mode, and a 7-day commit velocity sparkline:

<div align="center">
  <img src="assets/syncalibur-terminal.png" alt="Syncalibur Terminal Preview" width="960" />
</div>

---

## One-Liner Quick Install

Anyone can install Syncalibur globally in seconds:

```bash
curl -fsSL https://raw.githubusercontent.com/YusufKaya00/Syncalibur/main/install.sh | bash
```

Or clone manually and watch any local repository:

```bash
git clone https://github.com/YusufKaya00/Syncalibur.git ~/Syncalibur
cd ~/Syncalibur
./syncalibur.sh install /path/to/your/repo
```

Once installed, the global `syncalibur` command is automatically linked into `~/.local/bin/syncalibur` so you can run `syncalibur status`, `syncalibur mode`, `syncalibur fleet`, `syncalibur notify`, or `syncalibur doctor` from any directory.

---

## What is Syncalibur?

### 1. AST Git Intelligence, Conventional Commits & Multi-Language Support
- **Compiled Go Core (`cmd/syncalibur-engine/main.go`):** Zero external Go dependencies (100% Go standard library). Automatically compiled into `bin/syncalibur` on first run or after any self-update.
- **AST & Function-Aware Conventional Commit Engine:** Goes far beyond file-level messages (`auto: update file.js`). The Go engine inspects Go AST nodes (`ast.FuncDecl`, `ast.TypeSpec`) and Unified Diff Hunks across React (`export const Header = () => ...`), TypeScript, JavaScript, Java, Python, Rust, and C++ to generate pinpoint Conventional Commits:
  - `feat(Header): update Header() (+15/-2)`
  - `feat(main): update detectSecretsInFile() (+1/-1)`
  - `feat(Payment): update processTransaction() (+20/-4)`
  - `docs(README): update README.md (+85/-12)`
  - Optional local **Ollama AI Commit Hook** (`SYNCALIBUR_AI_COMMIT="true"`) with a strict 2-second non-blocking timeout and automatic heuristic fallback.
- **Zero-Friction Commit Protection (Syntax Hold & Secret Shield):**
  - **Syntax Hold:** Pre-flight gate that holds back commits if code has broken syntax (open braces, unclosed string/template literals, half-written statements) until you finish typing. Combines native compilers (`go/parser`, `py_compile`, `node --check`, `bash -n`, `tsc`, `javac`) with a native **Universal Lexical Balance Engine** supporting Go, Python, React (JSX/TSX), TypeScript, Node.js, Java, Bash, YAML, C/C++, Rust, and PHP.
  - **Secret Shield & Shannon Entropy Quarantine:** Detects credentials via signatures (AWS, GitHub PAT, OpenAI `sk-...`, Slack, Private Keys) and mathematical Shannon Entropy ($H \ge 4.4$ bit/char). Quarantines sensitive files into `git add` exclude patterns without blocking other safe files.
  - **DevOps & GitOps Guard:** Specialized preflight inspection for Kubernetes, Helm, Dockerfile, Terraform, and Git merge hygiene:
    - **Kubernetes Port Verification:** Cross-checks Service `targetPort` (both numeric and named ports) against Pod/Deployment `containerPort` to eliminate silent blackhole traffic drops.
    - **ConfigMap Plaintext Secret Prevention:** Blocks raw tokens, private keys, and passwords from slipping into ConfigMaps or Terraform `.tf` files.
    - **Container Health Probe Checks:** Validates `livenessProbe` and `readinessProbe` `httpGet.port` alignment with declared container ports.
    - **Dockerfile Hygiene Linter:** Warns on unpinned base images (`:latest`) and missing tags.
    - **Git Conflict Marker Blocker:** Rejects accidental commits containing unresolved merge conflict markers (`<<<<<<<`, `=======`, `>>>>>>>`).
- **Resilient Git Submodules & Stale Lock Auto-Healing:**
  - Automatic recursive submodule synchronization (`git push --recurse-submodules=on-demand`, `git submodule update --init --recursive`) and native support for Git worktrees and submodules where `.git` is a pointer file.
  - Non-blocking lock detection with procfs (`/proc/*/fd/*`) verification that automatically heals stale Git locks (>60s) without risking live index corruption.
- **Safe Bidirectional Sync (`--rebase --autostash`):** Periodically fetches from your primary remote (`origin`), detects when remote commits are ahead of local `HEAD`, and safely runs `git pull --rebase --autostash` when the working tree is clean. If a merge conflict is detected, Syncalibur **immediately aborts the rebase** (`git rebase --abort`), leaves your working tree untouched, and pops open a diagnostic terminal window.
- **Protected Branch Auto-PR with Rich Template:** If a target branch (like `main`) is protected with required Pull Request reviews or force-push restrictions, Syncalibur automatically pushes a `sync/auto-*` branch and opens a Pull Request populated with a structured Markdown PR template (diff stats, component checkboxes, and verification checklist).
- **Developer Flight Recorder & History Continuity:** Naturally captures your real day-to-day engineering progress into steady, well-scoped semantic checkpoints and conventional commits — preserving continuous history without risking broken builds or leaked secrets.
- **Native Filesystem Event Watching (`inotifywait` / `fswatch`):** Zero-CPU event-driven watching when `inotifywait` (Linux) or `fswatch` (macOS) is installed, with automatic fallback to lightweight polling.
- **Multi-Repository Fleet Supervisor (`repos.list`):** Watch 5, 10, or 50 repositories simultaneously! Run `syncalibur fleet scan ~/projects` to auto-discover all Git repositories using concurrent Go goroutines and supervise them in parallel.

### 2. Operating Modes & Adaptive History Hygiene
Syncalibur provides four specialized automation personas designed for different development workflows, configurable **per-repository** (via `.syncalibur`) or globally:
- **`instant-push` (Continuous Stream Mode):** Every save (even a 1-line edit) passes Syntax Hold & Secret Shield, settles after a 10s cooldown quiescence, is staged into an atomic Conventional Commit, and pushed directly to `main` without squashing. Perfect for solo study/notes repos (`k8sfundamentals`), learning journals, and personal sandboxes (backward-compatible alias: `commit-generator`).
- **`developer` (Adaptive Burst & Squash Mode) [Default]:** Balances continuous local micro-commits with clean remote history. 1-2 small commits push directly; if you make 3+ rapid changes (active burst sprint), Syncalibur waits for a 20s cooldown quiescence, squashes them into 1 clean Conventional Commit, and pushes. If the remote branch is protected, it automatically opens a Pull Request instead.
- **`devops` (DevOps & GitOps Guard Mode):** Engineered for Kubernetes manifests, Helm charts, Dockerfiles, and Terraform repos. Enforces a 30s cooldown quiescence, automatic squash hygiene (`tidy ≥3`), and executes the deep **DevOps & GitOps Guard** preflight gate (verifying K8s ports, probe ports, ConfigMap secrets, and Dockerfile tags).
- **`pr` (Strict Pull Request Mode):** Never pushes directly to `main`; always creates an isolated sync branch (`sync/<user>/<branch>-<timestamp>`) and opens a Pull Request with a structured Markdown PR template.
- **Per-Repository Tailoring:** Different repositories can run different modes! For example:
  - `~/k8sfundamentals` in `instant-push` mode
  - `~/Syncalibur` in `developer` mode
  - `~/k8s-infra` in `devops` mode
  - `~/team-project` in `pr` mode
- **Switch Modes Anytime:**
  - For current repository: `syncalibur mode <instant-push|developer|devops|pr>` (saves to `<repo>/.syncalibur`)
  - Globally for all repositories: `syncalibur mode <mode> --global`
  - Interactive Cockpit: Press `[m]` in `syncalibur status`

### 3. Offline Mode & Network Disconnection Shield
- **100% Autonomous Offline Protection:** Working on a train, plane, or when your WiFi drops? Syncalibur continues monitoring files, running AST syntax validations, scanning Shannon entropy secrets, and creating clean local atomic commits on your local branch without interruption.
- **Zero Disruptive Alerts:** Network timeouts or connectivity drops are treated as normal working conditions—Syncalibur never spawns diagnostic alert terminal popups for offline drops.
- **Local Queuing & Non-Blocking TCP Dial:** Uses a fast non-blocking TCP probe ($\le 1.0\text{s}$) to detect network status. Commits queue safely on local `HEAD`.
- **Automatic Reconnection Flush:** As soon as connectivity returns, Syncalibur automatically performs a safe fetch and rebase (`git pull --rebase --autostash`), flushes the queued commits in an atomic push, and logs the restored state.

### 4. Two-Pane Live Cockpit, Lucide Typography & Diagnostics
- **Native `syncalibur` Process Identity (`top` / `htop` / Activity Monitor / Task Manager):** Runs under its own compiled binary process name **`syncalibur`** across Linux, macOS, and Windows so you can inspect exact CPU (~0.0%) and RAM (~7 MB) usage directly in `top` or `syncalibur status`.
- **Interactive Two-Pane Live Cockpit (`syncalibur status`):** Full-screen, flicker-free terminal control center built in Go using raw terminal mode and alternate screen buffering. Running `syncalibur status` in an interactive terminal launches this cockpit automatically (or use `syncalibur status --static` for a plain text snapshot). Split into two dedicated panes:
  - **Top Pane (Live Telemetry Cockpit):** Displays real-time daemon state, active operating mode, today's commit count, 7-day velocity sparkline (` ▂█▅··▅▃`), Syntax & Secret Shield status, and hotkey cheatsheet (updated live every second).
  - **Bottom Pane (Scrollable Live Log Viewport):** Dynamically sized viewport that reads `syncalibur.log` with full keyboard scrolling (`↑/↓` or Vim `j/k`, `PageUp/PageDown`, `g` Top, `G` Bottom) and Follow mode toggle (`f`).
  - **Single-Keystroke Hotkeys:**
    - `[m]` Toggle Operating Mode (`instant-push` ➔ `developer` ➔ `devops` ➔ `pr`)
    - `[p]` Pause / Resume Auto-Sync
    - `[s]` Force Immediate Commit & Push
    - `[b]` Force Bidirectional Pull (`--rebase`)
    - `[t]` Toggle TrueColor Sword Emblem View
    - `[f]` Toggle Live Log Follow Mode
    - `[↑/↓]` or `[j/k]` Scroll Log Viewport
    - `[d]` Run 12-Point System Doctor
    - `[q]` Quit Cockpit cleanly
- **Lucide Minimalist Typography:** All terminal outputs, emblems, banners, and diagnostic tables use sleek, standardized Lucide minimalist glyphs (`✓`, `✕`, `▲`, `⛨`, `✦`, `↗`, `↙`, `●`, `○`, `⇄`, `✚`, `◆`) instead of distracting colored emojis.
- **Telegram Bot & Multi-Channel Webhooks (`Telegram`, `Discord`, `Slack`, `Generic JSON`):** Configure `SYNCALIBUR_TELEGRAM_BOT_TOKEN` & `SYNCALIBUR_TELEGRAM_CHAT_ID` (or `SYNCALIBUR_WEBHOOK_URL`) in `syncalibur.env` to receive instant rich notifications on push failures, rebase conflicts, self-updates, or every sync, and test anytime with `syncalibur notify` (`syncalibur telegram-test`).
- **12-Point `syncalibur doctor` Diagnostic Suite:** Verifies Git identity, Go engine binary health, filesystem watcher availability, Git locks & worktree health, submodule integrity, SSH connectivity (`github.com` / `gitlab.com`), `gh` / `glab` CLI authentication, OS background service registration, remote push permissions, Syntax & Secret Shield, and DevOps & GitOps Guard.

---

## Architecture & Sync Pipeline

```mermaid
flowchart TD
    A["Filesystem Change or Timer Tick"] --> B{"Event Watcher"}
    B -->|inotifywait or fswatch| C["Debounce Window"]
    B -->|Polling Fallback| C
    C --> D["Bidirectional Sync Check"]
    D -->|Remote Ahead| E["git pull --rebase --autostash"]
    E -->|Conflict Detected| F["Abort Rebase + Alert Terminal + Webhook"]
    E -->|Clean Rebase| G["Safety Guard (<50 MiB & Secret Filter)"]
    D -->|Already In Sync| G
    G --> H["Go Engine: Semantic Commit Generator"]
    H --> I["git commit -m 'feat(scope): update ... (+X/-Y)'"]
    I --> J{"Multi-Remote Push"}
    J -->|origin| K["GitHub (or Auto-PR if Protected)"]
    J -->|gitlab| L["GitLab / Self-Hosted"]
    K --> M["Telegram / Discord / Slack Webhook"]
    L --> M
```

```text
  ┌────────────────────────────────────────────────────────────────────────────────────────┐
  │                       S Y N C A L I B U R   E X E C U T I O N   F L O W                │
  └────────────────────────────────────────────────────────────────────────────────────────┘

   [1] File Change Detected        ──► inotifywait / fswatch (0% CPU) or 3s Polling Fallback
            │
            ▼
   [2] Bidirectional Rebase Check  ──► Fetches remote; runs 'git pull --rebase --autostash'
            │                          (Auto-aborts & pops open Alert Terminal on conflict)
            ▼
   [3] Pre-Staging Safety Shield   ──► Universal Syntax Pre-flight & Shannon Entropy Secret Shield
            │                          (Go, Python, React JSX/TSX, TypeScript, Node, Java, Bash, YAML, C/C++, Rust)
            │                          (Holds broken syntax ticks & quarantines high-entropy secrets)
            ▼
   [4] Go Engine AST Commit        ──► Inspects AST/hunks -> "feat(Component): update myMethod() (+42/-7)"
            │
   [5] Multi-Remote Push & Alert   ──► Pushes to GitHub + GitLab -> Dispatches Telegram / Webhook
```

---

## CLI Command Reference

### Developer Push Ergonomics & Universal Safety

| Command | Description |
| :--- | :--- |
| `syncalibur push [MSG] [--tidy]` | **One-Shot Push:** Runs preflight syntax & secret check, generates AST conventional commit (or uses custom msg), pulls safely, and pushes immediately to all remotes (<1.5s) |
| `syncalibur commit [MSG]` | **Local Safe Commit:** Preflight checks & commit without network push (perfect for offline coding) |
| `syncalibur check` / `preflight` | **Pre-Push Inspection:** Inspects working tree, tests Syntax Hold & Secret Shield, shows proposed AST commit message, and prints sync matrix |
| `syncalibur tidy` | **History Hygiene (Smart Squash):** Squashes unpushed local micro-commits into a single clean AST-driven Conventional Commit with automatic rollback checkpoint (`refs/syncalibur/backup-HEAD-...`) |
| `syncalibur undo` | **Time Machine Rollback:** Safely reverts working branch to the previous checkpoint via non-destructive soft reset (working tree edits are preserved) |
| `syncalibur checkpoint [save]` | **Safety Checkpoints:** Inspects or creates safety rollback checkpoints under `refs/syncalibur/*` |
| `syncalibur hook install` | **Universal Git Hook:** Installs `.git/hooks/pre-commit` (Secret Shield & Syntax Hold) and `.git/hooks/prepare-commit-msg` (AST suggestions) so standard `git commit` commands outside Syncalibur are guarded |
| `syncalibur hook status` / `uninstall` | Check active Git hook status or uninstall Syncalibur Git hooks cleanly |

### Autonomous Daemon & Fleet Cockpit

| Command | Description |
| :--- | :--- |
| `syncalibur status [--static]` | Launch the **Interactive Two-Pane Live Cockpit** in TTY (TrueColor sword emblem, live telemetry, scrollable log viewport, hotkeys `[m]` mode, `[p]` pause, `[s]` sync, `[b]` pull, `[t]` logo, `[d]` doctor, `[q]` quit; or `--static` for plain text snapshot) |
| `syncalibur mode [name]` | Inspect or switch active operating mode (`commit-generator` \| `developer` \| `devops` \| `pr`) |
| `syncalibur install [PATH]` | Auto-detect repo, create missing remote repo if needed, link `~/.local/bin/syncalibur`, and start the OS service |
| `syncalibur logs` | Render the emblem header and stream live color-coded log events (`tail -f`) |
| `syncalibur sync` | Trigger an immediate one-shot bidirectional pull, semantic commit, and multi-remote push |
| `syncalibur pull` | Trigger an immediate safe bidirectional pull (`git pull --rebase --autostash`) |
| `syncalibur pause` / `resume` | Temporarily pause or resume background auto-commits without uninstalling the service |
| `syncalibur notify [MSG]` | Send an instant test notification to your configured **Telegram Bot** (`syncalibur telegram-test`) and Webhook |
| `syncalibur doctor` | Run the 12-point diagnostic health check (Git, Go Engine, inotify, SSH, `gh`/`glab`, remotes, Git locks, submodules, Syntax, Secret Shield, and DevOps & GitOps Guard) |
| `syncalibur fleet status` | Display a concurrent status table of all repositories in `repos.list` |
| `syncalibur fleet scan [DIR]` | Recursively scan a directory for Git repositories using Go goroutines and append them to `repos.list` |
| `syncalibur fleet start` | Spawn background Syncalibur daemons for every active repository in `repos.list` |
| `syncalibur fleet stop` | Stop all background fleet daemons |
| `syncalibur alert` | Test the graphical pop-up diagnostic terminal window on your desktop environment |
| `syncalibur uninstall` | Stop the background daemon and remove OS startup registration |

---

## Configuration (`syncalibur.env.example` → `syncalibur.env`)

Your personal `syncalibur.env` and `repos.list` files are **gitignored** so your local paths and webhook tokens are never pushed to GitHub. To customize settings, copy the template:

```bash
cp syncalibur.env.example syncalibur.env
```

```dotenv
# Target Git repository path to watch and auto-sync
SYNCALIBUR_REPO_DIR="/path/to/your/git/repository"

# Push target strategy:
#   "all"    -> Push to ALL configured git remotes (e.g., GitHub + GitLab simultaneously)
#   "origin" -> Push only to origin
SYNCALIBUR_REMOTE="all"

# Commit message style:
#   "semantic" -> Go Engine generates Conventional Commits: feat(scope): update X (+12/-3)
#   "simple"   -> Classic timestamped auto: file1, file2 (YYYY-MM-DD HH:MM:SS)
SYNCALIBUR_COMMIT_STYLE="semantic"

# Optional Local AI Commit Message Generation via Ollama (true / false)
SYNCALIBUR_AI_COMMIT="false"
SYNCALIBUR_OLLAMA_MODEL="qwen2.5-coder:1.5b"

# Safe Bidirectional Sync (auto-pull --rebase --autostash before pushing)
SYNCALIBUR_BIDIRECTIONAL="true"
SYNCALIBUR_PULL_INTERVAL="60"

# Native Filesystem Event Watching (inotifywait / fswatch with polling fallback)
SYNCALIBUR_USE_FS_EVENTS="true"

# Multi-Channel Webhooks (Discord, Slack, Telegram, or Generic JSON)
SYNCALIBUR_WEBHOOK_URL=""
SYNCALIBUR_WEBHOOK_ON_SUCCESS="false"
SYNCALIBUR_TELEGRAM_BOT_TOKEN=""
SYNCALIBUR_TELEGRAM_CHAT_ID=""
```

---

## Telegram Bot & Multi-Channel Webhook Alerts

Syncalibur's Go engine includes a native multi-channel notification dispatcher supporting **Telegram Bot API**, **Discord Embeds**, **Slack Blocks**, and **Generic JSON Webhooks**.

### 1. Configure Telegram Bot Alerts
1. Message [@BotFather](https://t.me/BotFather) on Telegram and run `/newbot` to get your bot token.
2. Obtain your chat or channel ID (for example via `@userinfobot` or `https://api.telegram.org/bot<TOKEN>/getUpdates`).
3. Add your credentials to `syncalibur.env`:

```dotenv
SYNCALIBUR_TELEGRAM_BOT_TOKEN="123456789:ABCdefGHIjklMNOpqrSTUvwxYZ"
SYNCALIBUR_TELEGRAM_CHAT_ID="987654321"
# Set to "true" if you also want notifications on every successful push (default: failures & conflicts only)
SYNCALIBUR_WEBHOOK_ON_SUCCESS="true"
```

### 2. Configure Discord or Slack Webhooks (Optional)
```dotenv
SYNCALIBUR_WEBHOOK_URL="https://discord.com/api/webhooks/..."
```

### 3. Test Your Notification Channel
Verify your Telegram Bot or Webhook configuration at any time from your terminal:

```bash
syncalibur notify "Hello from Syncalibur!"
# or
syncalibur telegram-test
```

---

## Multi-Repository Fleet Mode (`repos.list`)

Want to watch multiple projects at once? Use **Fleet Mode**:

```bash
# Automatically scan a workspace directory and add all Git repositories to repos.list
syncalibur fleet scan ~/workspace

# View real-time parallel status across all repositories in your fleet
syncalibur fleet status

# Start background sync daemons for every repository in repos.list
syncalibur fleet start
```

---

## Open-Source Governance & Pull Request Workflow

To protect all users running Syncalibur's autonomous self-updater, the `main` branch is protected with strict repository governance:
- **Force Pushes & Branch Deletion Disabled:** History on `main` cannot be overwritten.
- **Pull Request & 1 Approval Required:** Direct pushes to `main` from external contributors are blocked; all changes require a Pull Request and **at least 1 approving review** from `@YusufKaya00` (`CODEOWNERS`).
- **Automated CI Verification (`Syncalibur CI`):** Every PR automatically runs Bash syntax verification, `ShellCheck`, `go vet`, native `go build`, engine smoke tests, and secret hygiene checks before merging is allowed.
- **Automated PR Triage (`PR Governance & Auto-Labeler`):** Pull requests are automatically labeled by component (`go-engine`, `shell-daemon`, `documentation`, `ci-cd`) and diff size (`size/S`, `size/M`, `size/L`).

See [CONTRIBUTING.md](CONTRIBUTING.md) for step-by-step instructions on opening a Pull Request.

---

## Repository Structure

```text
Syncalibur/
├── .github/
│   ├── CODEOWNERS                 # Auto-assigns @YusufKaya00 for mandatory PR review
│   ├── pull_request_template.md   # Standardized PR checklist
│   └── workflows/
│       ├── ci.yml                 # Automated Go build, ShellCheck & smoke test CI
│       └── pr-governance.yml      # Automatic PR component & size labeler
├── assets/
│   ├── syncalibur-logo.png        # Official Syncalibur Emblem (rendered in TrueColor ANSI)
│   └── syncalibur-terminal.png    # High-DPI screenshot of `syncalibur status`
├── cmd/
│   └── syncalibur-engine/
│       ├── main.go                # Go Core Engine (PNG Renderer, AST Commits, Fleet, Webhooks, Live TUI)
│       └── main_test.go           # Go Unit Tests (Shannon Entropy, Secret Shield, Syntax & AST)
├── CONTRIBUTING.md                # Contributor & Pull Request workflow guide
├── go.mod                         # Go module definition (zero external dependencies)
├── install.sh                     # 1-Liner public curl|bash installer
├── syncalibur.sh                  # Cross-platform Service Manager & Daemon Supervisor
├── syncalibur.env.example         # Public configuration template (syncalibur.env is gitignored)
├── repos.list.example             # Multi-repository fleet configuration template
└── README.md                      # Documentation
```

---

## Search Keywords & Topics

`git auto commit` • `auto commit pusher` • `commit gainer` • `github auto commit` • `git auto push` • `git auto sync daemon` • `organic commit gainer` • `github contribution graph` • `semantic commit generator` • `multi-remote git sync` • `github gitlab mirror`

---

## License

MIT License — Free for personal and commercial use.
