# Syncalibur — Comprehensive Project Memory & Living System Blueprint

> **Status:** Active Platform Specification & Living Memory Ledger  
> **Platform Version:** `v2.3.0` (Milestone: *"DevOps Intelligence & GitOps Guard"*)  
> **Architecture:** Hybrid Native Go Core (`cmd/syncalibur-engine`) + Fault-Tolerant Bash Supervisor (`syncalibur.sh`)  
> **Living Memory Rule:** *Every functional modification, architecture change, or configuration addition MUST update this document.*

---

## 1. Executive Summary & Mission Statement

**Syncalibur** is an autonomous, production-grade Git synchronization platform and developer code guard. It operates as both a zero-friction background daemon and an ergonomic, one-shot CLI tool (`syncalibur push`, `syncalibur commit`, `syncalibur status`, `syncalibur mode`, `syncalibur tidy`, `syncalibur doctor`).
 
### Core Design Philosophy
1. **Zero Data Loss Guarantee:** File changes are snapshotted and committed locally immediately upon debounce quiescence. Even if the system loses power or network drops, uncommitted edits are never lost.
2. **Clean Conventional Commit History:** Unlike crude auto-commit scripts that flood Git history with dozens of spam commits (`auto: update file.js`), Syncalibur enforces semantic Conventional Commits (`feat(scope): ... (+X/-Y)`), Smart Staging (partitioning multi-concern edits), and Burst Cadence Auto-Tidy (squashing micro-commits into one clean milestone commit before pushing).
3. **Zero External Dependencies:** Built with pure Go standard library (`go/parser`, `go/ast`, `image`, `syscall`, `net/http`) and native POSIX/Bash commands. Compiles in $<1.5\text{s}$ into a standalone stripped binary (`bin/syncalibur`).
4. **Fail-Safe Secret & Syntax Shield:** High-entropy credentials, private keys, and broken syntax are quarantined before hitting the Git index or remote servers.
5. **True Cross-Platform Operation:** Linux (`systemd`, `inotify`, `ioctl`), macOS (`launchd`, `fswatch`, `ioctl`), and Windows (`windows-startup`, PowerShell RawUI, Win32 `OpenProcess`).

---

## 2. System Architecture & Dual-Engine Strategy

Syncalibur employs a **symbiotic dual-engine architecture**:

```
 ┌────────────────────────────────────────────────────────────────────────┐
 │                        USER INTERFACES & CLI                           │
 │  syncalibur push | commit | tidy | mode | pr | status | doctor | fleet │
 └───────────────────────────────────┬────────────────────────────────────┘
                                     │
         ┌───────────────────────────┴───────────────────────────┐
         ▼                                                       ▼
┌─────────────────────────────────┐     ┌─────────────────────────────────┐
│     NATIVE GO CORE ENGINE       │     │     BASH SUPERVISOR & DAEMON    │
│    `cmd/syncalibur-engine/`     │     │         `syncalibur.sh`         │
├─────────────────────────────────┤     ├─────────────────────────────────┤
│ • AST Symbol & Diff Parsing     │◄───►│ • OS Service Supervision        │
│ • Shannon Entropy Secret Shield │     │   (systemd, launchd, startup)   │
│ • Polyglot Syntax Hold Engine   │     │ • Filesystem Watcher Hierarchy  │
│ • Smart Staging Planner         │     │   (inotifywait, fswatch, poll)  │
│ • Interactive Two-Pane TUI      │     │ • Debounce & Snapshot Hashing   │
│ • TrueColor 24-bit ANSI Logo    │     │ • Burst Cadence & Auto-Tidy     │
│ • SemVer & Release Calculator   │     │ • Bidirectional Sync & Rebase   │
│ • Fleet Scanner & Prober        │     │ • Multi-Remote Routing & Auth   │
│ • Webhooks & Telegram Bot       │     │ • Self-Update & Hot Reloading   │
└─────────────────────────────────┘     └─────────────────────────────────┘
```

### Why Hybrid?
- **Bash Supervisor (`syncalibur.sh`):** Excels at operating-system integration, background process lifecycle, signal trapping (`trap on_exit EXIT`), managing system service managers (`systemd --user`, macOS `launchd`, Windows VBS startup), and wrapping arbitrary git CLI flows without brittle CGO bindings.
- **Go Engine (`cmd/syncalibur-engine`):** Excels at CPU-intensive tasks requiring microsecond performance: walking AST trees, parsing lexical balances across 22 programming languages, computing Shannon entropy frequencies, downsampling TrueColor ANSI graphics, rendering interactive fullscreen raw TUIs, and executing concurrent goroutine fleet probes.

---

## 3. Subsystem Deep-Dive

### 3.1 Go Native Engine (`cmd/syncalibur-engine/`)

The compiled binary resides in `bin/syncalibur` (symlinked as `bin/syncalibur-engine`).

#### Subcommands Supported
- `daemon-exec <scriptPath> [args...]`: Spawns and supervises bash background daemon. Forwards `SIGINT`, `SIGTERM`, `SIGHUP`. Injects `SYNCALIBUR_SUPERVISOR_PID` and re-executes so `top`, `ps`, and Activity Monitor display `syncalibur`.
- `commit-msg [repoRoot]`: Inspects staged git diff, performs AST symbol lookup, and outputs Conventional Commit message (`<type>(<scope>): <subject> (+add/-del)`).
- `smart-stage [repoRoot] [--json]`: Partitions multi-file working tree changes into atomic semantic groups (`GROUP|<scope>|<type>|<files>`).
- `preflight [repoRoot] [--staged-only] [--json]`: Scans modified files for secrets (Shannon entropy), syntax errors, and DevOps policies.
  - **Exit Code 0:** Clean (pass).
  - **Exit Code 2:** Syntax Hold (code contains unclosed brackets/syntax errors or conflict markers; commit is held until fixed).
  - **Exit Code 3:** Secret Shield (high-entropy credential detected; specific files are quarantined from git add).
  - **Exit Code 4:** DevOps Policy Hold (Kubernetes port mismatch, ConfigMap secret, Dockerfile :latest / root USER, or Terraform hygiene failure; commit is held until fixed).
- `status / ui [scriptPath] [repoRoot]`: Launches fullscreen interactive two-pane cockpit with live telemetry, active operating mode, and log viewport.
- `render-logo <src.png> <dst.ansi>`: Downsamples PNG to 44x21 character TrueColor ANSI using 2x2 quadrant subpixel clustering.
- `sparkline [repoRoot]`: Computes 7-day commit velocity bar chart (`  ▂▃▄▅▆▇█`).
- `fleet-scan [dir]`: Recursively discovers Git repositories up to depth 3 using parallel goroutines.
- `fleet-status <fleetFilePath>`: Concurrently probes (`sync.WaitGroup`) all repos in `repos.list` and outputs ANSI status table.
- `webhook <url> <event> <repo> <branch> <title> <msg> <details>`: Dispatches JSON payloads to Discord, Slack, or generic webhooks with 8s timeout.
- `telegram <botToken> <chatId> <event> <repo> <branch> <title> <msg> <details>`: Dispatches formatted HTML alert cards to Telegram Bot.
- `release [mode] [--dry-run] [repoRoot]`: Automated SemVer bumper, changelog generator, git tagger, and GitHub release creator.
- `version`: Outputs platform version, milestone, commit hash, OS/Arch.

#### Shannon Entropy Secret Shield (`calculateShannonEntropy`)
- Implements Shannon entropy formula:
  $$H(X) = -\sum_{i=1}^{n} P(x_i) \log_2 P(x_i)$$
- Words $\ge 24$ chars with entropy $\ge 4.4\text{ bits/char}$ trigger credential quarantine.
- Hardcoded & regex signatures:
  - AWS Access Key IDs (`AKIA[0-9A-Z]{16}`) and Secret Access Keys (`(?i)aws_secret_access_key\s*=\s*[A-Za-z0-9/+=]{40}`)
  - GitHub Classic PATs (`ghp_[A-Za-z0-9]{36}`) and Fine-Grained PATs (`github_pat_[A-Za-z0-9_]{82}`)
  - GitLab Personal Access Tokens (`glpat-[A-Za-z0-9\-_]{20,}`)
  - Slack tokens (`xox[baprs]-[0-9a-zA-Z]{10,48}`)
  - OpenAI API keys (`sk-[A-Za-z0-9_-]{32,}`) and Anthropic API keys (`sk-ant-[A-Za-z0-9_-]{32,}`)
  - Stripe Secret & Restricted API keys (`sk_live_[0-9a-zA-Z]{24,}`, `rk_live_[0-9a-zA-Z]{24,}`)
  - RSA/EC/OpenSSH private keys (`-----BEGIN (RSA|EC|DSA|OPENSSH) PRIVATE KEY-----`)
  - Generic token/password assignment statements with high-entropy values.
- Expanded file blacklists in `DEFAULT_SENSITIVE_PATTERNS`: `.env`, `.env.*`, `*.tfstate*`, `*.tfvars`, `*.tfvars.json`, `credentials.json`, `client_secret*.json`, `service-account*.json`, `azureProfile.json`, `*.pem`, `*.key`, `*.p12`, `*.pfx`, `*.pkcs12`, `*.jks`, `*.keystore`, `*.kdbx`, `*.ovpn`, `*.tblk`, `auth.json`, `.netrc`, `.npmrc`, `.pypirc`, `*.dump`, `*.rdb`, `*.bak`, `*.backup`.
- **Precision Filter & Template Whitelist:** Automatically whitelists benign template files (`*.example`, `*.sample`, `*.template`, `*.dist`, `*.default`, `*.skel`), UUIDs, SVG paths, Subresource Integrity hashes (`sha512-`), base64 data URIs (`data:image/...`), pure hex hashes, URLs, file paths, and shell expansions (`${...}`).

#### Single-Pass Commit Velocity Engine (`getCommitVelocity`)
- Derives 7-day commit counts, today's commit count, and calibrated sparkline glyphs (`  ▂▃▄▅▆▇█`) from a single atomic `git log --format=%ct` invocation.
- Eliminates 7 subprocess forks per second in live TUI.
- Computes exact rolling 7-day average (`%.1f/d`) and peak daily velocity with zero timezone boundary double-counting.

#### Non-Blocking Network Liveness Probe (`probeRemoteLiveness`)
- Probes remote Git target host and port via fast TCP dial (`net.DialTimeout`) with $\le 1.0\text{s}$ timeout.
- Pure Go standard library (zero external packages). Supports SSH (`git@host:port`), HTTPS (`https://host:port`), and custom ports.

#### Polyglot Syntax Hold Engine (`checkLexicalBalance` & Native Parsers)
- **Go:** Full AST validation via native `go/parser.ParseFile`.
- **JSON / JSONC:** Custom `stripCommentsFromJSON()` strips `//` and `/* */` comments before parsing.
- **YAML / Helm:** Rejects forbidden tab indents (YAML specification), handles literal block scalars (`|`, `|-`, `>`, `>-`), allows embedded HTML/CSS hex codes (`#0f172a`), validates Jinja/Helm delimiters (`{{ ... }}`).
- **CSS / SCSS / Less:** Validates bracket balance and color declarations.
- **Terraform / HCL:** Heredoc parsing (`<<EOF ... EOF`), variable blocks, and bracket hierarchies.
- **SQL:** Comment parsing (`--`, `#`), escaped quotes (`''`), and parentheses.
- **Dockerfile:** Validates instruction verbs (`FROM`, `RUN`, `CMD`, `COPY`, etc.) and continuation slashes (`\`).
- **External CLI Fallbacks:** If installed, verifies with `python3 -m py_compile`, `node --check`, `bash -n`, `javac`, `ruby -c`, `php -l`.

#### DevOps & GitOps Guard Subsystem (`cmd/syncalibur-engine`)
Syncalibur provides an automated infrastructure policy engine for Kubernetes, Helm, Docker, and Terraform:
1. **Kubernetes Port & Service Cross-Validation (`validateServicePortTargetPort`):**
   - Correlates Kubernetes `Service` manifests with matching `Deployment`, `StatefulSet`, `DaemonSet`, and `Pod` resources by resolving `spec.selector` label sets against `spec.template.metadata.labels`.
   - Validates that `service.spec.ports[].targetPort` (numeric or named) matches at least one declared `containerPort` in the matching workload.
   - Prevents silent blackhole routing failures (where traffic to a Service is dropped with `Connection refused` because of a typo between port and targetPort).
2. **ConfigMap Plaintext Credential Shield (`validateConfigMapSecrets`):**
   - Analyzes keys and values within `ConfigMap` data blocks.
   - Flags sensitive keys (`password`, `api_key`, `access_key`, `auth_token`, `secret`, `private_key`) containing high-entropy or raw credentials, prompting developers to migrate to Kubernetes `Secret`, `SealedSecret`, or external vault providers.
3. **Container Health Probe Alignment (`validateContainerProbes`):**
   - Validates that `livenessProbe` and `readinessProbe` HTTP/TCP target ports point to declared container ports, preventing deployment CrashLoopBackOff loops.
4. **Image Tag Pinning (`validateImageTags`):**
   - Flags unpinned container images or `:latest` tags, enforcing immutable tags or SHA256 digests.
5. **Dockerfile Security & Best Practices (`lintDockerfile`):**
   - Flags `FROM image:latest` mutable base images.
   - Enforces non-root execution (`USER <non-root>`).
   - Flags broad `COPY . .` commands when `.dockerignore` is missing (preventing accidental leakage of `.git`, secrets, and build caches into container layers).
6. **Terraform Hygiene & State Secret Shield (`lintTerraformHygiene`):**
   - Enforces `.gitignore` coverage for `.terraform/` provider caches and `*.tfstate*` remote state files.
   - Flags plaintext credentials and secrets in `terraform.tfvars`.
7. **Git Conflict Marker Gate (`detectConflictMarkers`):**
   - Sniffs files for unresolved Git conflict markers (`<<<<<<<`, `=======`, `>>>>>>>`).
8. **Null-Byte Binary Sniffer (`isBinaryFile`):**
   - Sniffs the first 8000 bytes for null bytes (`0x00`) with a 10MB analysis ceiling, preventing accidental commits of compiled binaries or VM disk images.
9. **Symlink Path-Traversal Boundary Guard (`isSafePath`):**
   - Inspects `os.Lstat` and ensures symlinks never resolve outside the repository boundary.

#### Smart Staging Planner (`planSmartStaging`)
- Groups files by concern:
  1. `docs`: Markdown, text, licenses, `docs/` folder.
  2. `ci`: Workflows (`.github/`, `.gitlab/`), build files (`Makefile`, `package.json`, `go.mod`).
  3. `k8s` / `helm`: Kubernetes manifests and Helm charts (`*.yaml`, `*.yml` in templates/ or k8s/).
  4. `docker`: Dockerfiles, Containerfiles, `docker-compose.yml`.
  5. `tf`: Terraform configuration (`*.tf`, `*.tfvars`).
  6. `style`: CSS, SCSS, stylesheets.
  7. `code:<scope>`: Application code partitioned by module/directory.
  8. `test`: Unit tests (`*_test.go`, `*.spec.ts`, `tests/`).
- Allows Syncalibur to stage and commit each group independently as an atomic commit (`feat(scope): ...`, `docs: ...`, `infra(k8s): ...`, `helm: ...`, `docker: ...`, `ci: ...`).

#### High-Definition TrueColor ANSI Emblem & Interactive TUI
- **Renderer:** Maps 2x2 subpixels to Unicode quadrant blocks (`▀`, `▌`, `▐`, `▄`, `▘`, `▝`, `▖`, `▗`, `▚`, `▞`, `▛`, `▜`, `▙`, `▟`, `█`).
- **Embedded Fallback (`defaultEmbeddedLogoLines`):** 21 lines of TrueColor ANSI embedded directly in Go binary. If `assets/syncalibur-logo.ansi` is missing, the emblem renders flawlessly.
- **Raw Mode Terminal:** Non-blocking async keyboard reader (`keyCh := make(chan []byte, 32)`), alternate screen buffer (`\033[?1049h`), 1-second ticker refresh.
- **Keybindings:**
  - `q` / `Ctrl+C` / `Esc`: Clean exit with terminal state restoration.
  - `m`: Cycle operating mode (`commit-generator` ➔ `developer` ➔ `devops` ➔ `pr` ➔ `commit-generator`) and persist to `.syncalibur` or `syncalibur.env`.
  - `p`: Toggle pause/resume (`syncalibur.pause`).
  - `s`: Trigger immediate sync cycle (`syncalibur.sync`).
  - `b`: Trigger bidirectional remote pull (`syncalibur pull`).
  - `u`: Trigger self-update check (`syncalibur.upd`).
  - `d`: Flash status for 11-point diagnostic check.
  - `t` / `e`: Toggle side-by-side Emblem Cockpit vs Compact Tail view.
  - `f`: Lock follow/tail mode.
  - `↑` / `k` / `PgUp`: Scroll viewport up (disables follow mode).
  - `↓` / `j` / `PgDn`: Scroll viewport down (re-enables follow mode at bottom).
  - `←` / `h` & `→` / `l`: Horizontal pan (10 columns). `0`: Reset pan.

#### Cross-Platform Hardware Abstraction
- **`terminal_unix.go` (`!windows`):** Direct kernel `ioctl(TIOCGWINSZ)` syscall without spawning subprocesses. Signal 0 process liveness probing.
- **`terminal_windows.go` (`windows`):** PowerShell `RawUI.WindowSize` dimensions query. Win32 `OpenProcess` and `GetExitCodeProcess` liveness probing.

---

### 3.2 Bash Daemon Core (`syncalibur.sh` & `install.sh`)

#### Daemon Lifecycle & Single-Instance Lock
- File descriptor advisory lock via `flock -w 5` on `${GIT_DIR}/syncalibur.lock`. Fallback to PID probing via `kill -0` if flock is unavailable.
- Stored PID in `${GIT_DIR}/syncalibur.pid`.
- Traps: `trap on_exit EXIT`, `trap 'exit 0' INT TERM`.

#### Filesystem Event Watching & Debounce Quiescence
- **Hierarchy:** `inotifywait` (Linux kernel inotify, <10ms trigger, 0.00% CPU) $\rightarrow$ `fswatch` (macOS FSEvents) $\rightarrow$ stat-based polling fallback.
- **Snapshot Hashing:** Deterministic SHA-1 working tree digest computed via `file_signature` (size + epoch mtime) piped to `git hash-object --stdin`.
- **Debounce:** `wait_for_quiet` polls working tree until consecutive snapshots match over `DEBOUNCE_WAIT` (default 5s). Capped at `MAX_DEBOUNCE` (default 60s).

#### Operating Modes & Adaptive History Hygiene (`SYNCALIBUR_MODE`)
Syncalibur provides four specialized automation personas configured per-repository via `.syncalibur` or globally via `syncalibur.env` (or via `syncalibur mode <name>` / `[m]` hotkey):
1. **`commit-generator` (Continuous Stream Mode):**
   - Immediate push on every save (even a 1-line edit) after passing Syntax Hold & Secret Shield.
   - Quiescence settle cooldown defaults to 10s (`BURST_COOLDOWN=10`), allowing rapid consecutive saves to settle cleanly without spamming micro-commits.
   - Squashing is bypassed (`TIDY_THRESHOLD=999`, `AUTO_TIDY="false"`), and `AUTO_PR="false"`.
   - Directly pushes to `main` without delay. Ideal for personal study/notes repos (`k8sfundamentals`), learning journals, and solo sandbox projects.
2. **`developer` (Adaptive Burst & Squash Mode) [Default]:**
   - 1-2 small commits: Pushed directly to `main` without squashing.
   - 3+ rapid commits (burst sprint): Enters `BURST_COOLDOWN=20s`. When coding pauses for 20s, auto-squashes (`cmd_tidy`) all unpushed commits into 1 clean Conventional Commit with a rollback backup ref (`refs/syncalibur/backup-HEAD-<timestamp>`).
   - If user has push permission: Pushes the consolidated commit directly to `main`.
   - If user lacks permission or branch is protected (`GH006`): Automatically branches to `sync/auto-*` and opens an auto-PR.
3. **`devops` (Kubernetes, Helm & Infrastructure Guard Mode):**
   - Built specifically for cloud-native infrastructure, Kubernetes clusters, Helm charts, Dockerfiles, and Terraform repos.
   - Quiescence settle cooldown is 30s (`BURST_COOLDOWN=30`).
   - Deep Preflight Gate: Enforces Service port $\leftrightarrow$ targetPort cross-validation, ConfigMap secret protection, Dockerfile non-root execution, Terraform state/credential hygiene, and conflict marker detection.
   - Multi-group smart staging automatically scopes commits using conventional tags: `infra(k8s):`, `helm:`, `docker:`, `tf:`, `gitops:`.
   - Halts commit/push with exit code `4` (`DEV_OPS_HOLD`) if any infrastructure policy error is present.
4. **`pr` (Strict Pull Request Mode):**
   - Never pushes directly to `main`. Commits are always pushed to an isolated sync branch (`sync/<user>/<branch>-<timestamp>`) and opened as a rich Pull Request with AI summary.
5. **Configuration Precedence Cascade:**
   `CLI Flags (--mode=...)` > `Environment (SYNCALIBUR_MODE)` > `Fleet Inline (repos.list [mode=...])` > `Repo-local (<repo>/.syncalibur)` > `Global User Config (syncalibur.env)` > `Default (developer)`.

#### Offline Mode & Network Disconnection Shield
- **Zero Data Loss Guarantee:** When disconnected from the internet or VPN drops, local file watching, AST syntax verification, Shannon entropy secret scanning, and atomic git commits continue operating at 100% capacity.
- **Non-Blocking Fast Liveness Probe:** Uses `check_network_liveness` (calling Go engine's TCP dial or `/dev/tcp`) before executing `git fetch` or `git push`.
- **Silent Local Queuing:** Network connection errors do NOT trigger disruptive diagnostic alert windows. Instead, `.git/syncalibur.offline` is flagged, commits accumulate safely on local HEAD, and the TUI Cockpit displays `🌐 OFFLINE (N queued)`.
- **Automatic Reconnection Flush:** Once network connectivity restores, the daemon detects the connection, performs safe fetch & rebase (`safe_bidirectional_sync`), flushes queued commits via atomic push, removes the offline flag, and logs the restored state.

#### Safe Bidirectional Synchronization
- Runs `safe_bidirectional_sync` before local commit cycles and periodically (`SYNCALIBUR_PULL_INTERVAL=45`).
- Executes `git pull --rebase --autostash`.
- **Conflict Self-Healing:** If an auto-rebase encounters a conflict, Syncalibur immediately aborts (`git rebase --abort`) to prevent working tree corruption, records diagnostic metadata in `.git/syncalibur.failure`, and opens a diagnostic terminal alert window.

#### Git Submodules & Lock Contention Resilience
- **Recursive Submodule Push:** All `git push` operations include `--recurse-submodules=on-demand`, guaranteeing that submodules are pushed in dependency order before the parent repository.
- **Submodule Directory Resolution (`isGitRepoDir` & `resolveGitDir`):** Recognizes `.git` files (gitdir pointers in submodules and worktrees) and locates `syncalibur.pid` and `syncalibur.log` accurately via `git rev-parse --absolute-git-dir`.
- **Process-Aware Stale Lock Healing:** Inspects active processes via `fuser`, `lsof`, and `/proc/*/fd/*` rather than flaky process name regexes. Stale locks (>60s) without live open file descriptors are automatically pruned.
- **Exponential Backoff:** If Git is transiently busy, Syncalibur retries operations with exponential backoff (1s, 2s, 4s, 8s, 16s) instead of aborting.
- **Detached HEAD Guard (`assert_git_operable`):** Halts commit and push commands if HEAD is detached (`git symbolic-ref -q HEAD`), preventing orphaned commits and inadvertent overwrites.
- **Non-Interactive Batch Mode:** Exports `GIT_TERMINAL_PROMPT=0` and non-interactive `GIT_SSH_COMMAND` (`-o BatchMode=yes -o ConnectTimeout=10`), guaranteeing background daemons never hang waiting for user keyboard input.

#### Multi-Remote Routing & Zero-Touch Repository Provisioning
- Queries push URLs across `remote.<name>.pushurl` and `remote.<name>.url` to target multi-remote configurations (e.g. GitHub + GitLab mirrors).
- **Auto-Provisioning:** If remote repository does not exist on GitHub/GitLab (404 / Repository not found), uses `gh repo create` or GitLab API with user token to provision the private/public repository on the fly and bind remote.

#### Self-Update & Hot Reload Engine
- Daemon periodically checks Syncalibur's own repository (`SELF_REPO_ROOT`).
- If upstream changes exist: pulls updates (`git pull --rebase --autostash`), recompiles Go engine (`go build -ldflags="-s -w"`), and executes **in-place process replacement**:
  ```bash
  trap - EXIT
  exec bin/syncalibur daemon-exec syncalibur.sh run
  ```
  The running daemon image is replaced in memory without modifying PID, keeping systemd/launchd supervisors intact.

#### 12-Point Diagnostic Doctor (`cmd_doctor`)
1. Go Native Engine compiled & executable.
2. Filesystem event watcher active (`inotifywait` / `fswatch` / `polling`).
3. Git repository state (unlocked, no stale `index.lock`, process-aware lock detection).
4. Daemon process status (PID + CPU%/RAM consumption).
5. OS Auto-Start service registration (`systemd`, `launchd`, `vbs`, `cron`).
6. Provider authentication (GitHub/GitLab CLI or token).
7. Push destinations configured.
8. Self-watching updater enabled.
9. Global CLI command symlinked in `$PATH`.
10. Telegram Bot & Webhooks reachable.
11. Preflight syntax & secret shield verification.
12. DevOps & GitOps Guard (Kubernetes manifest validator, Dockerfile linter, Terraform hygiene, Git conflict markers).

---

## 4. Multi-Repository Fleet Mode (`repos.list`)

Syncalibur supervises an arbitrary fleet of repositories simultaneously, with each repository running in its own tailored operating mode.
- **Config File:** `repos.list` (gitignored, template: `repos.list.example`).
- **Extended Fleet Syntax:**
  ```text
  # Standard path (uses repo's .syncalibur or global default)
  /home/ysf/Syncalibur

  # Bracketed attributes
  /home/ysf/k8sfundamentals [mode=commit-generator] [remote=origin]

  # Inline key-value attributes
  /home/ysf/Erasmus-Connect mode=pr remote=origin
  ```
- **Commands:**
  - `syncalibur fleet scan [dir]`: Recursively crawls directory to depth 3 and auto-enrolls all Git repositories.
  - `syncalibur fleet add <path> [--mode=commit-generator|developer|pr]`: Enrolls repository with specified mode.
  - `syncalibur fleet status`: Concurrent Go probe displaying tabular status of all enrolled repositories, including active `MODE` (`⚡ commit-gen`, `🛠 developer`, `🔀 pr`), daemon PID, unpushed counts, and dirty file status.
  - `syncalibur fleet start`: Starts background sync daemons for all repositories according to their configured per-repo modes.
  - `syncalibur fleet stop`: Terminates all fleet background daemons.

---

## 5. CI/CD & Automated Release Pipelines

### 5.1 Continuous Integration (`.github/workflows/ci.yml`)
- Triggers on `push` and `pull_request` to `main`/`master`.
- **Gates:**
  1. `bash -n syncalibur.sh && bash -n install.sh` (Shell syntax validation).
  2. `shellcheck -S error syncalibur.sh install.sh` (Static analysis).
  3. `go vet ./cmd/syncalibur-engine` (Go compiler static analysis).
  4. `go build -trimpath -ldflags="-s -w" -o bin/syncalibur-engine ./cmd/syncalibur-engine` (Engine build).
  5. Smoke tests (`render-logo`, `sparkline`).
  6. Git index hygiene assertion (verifies `syncalibur.env` and `repos.list` are never tracked).

### 5.2 Automated SemVer & Release Pipeline (`.github/workflows/release.yml`)
- Triggers on `push` to `main` (ignoring changes to `CHANGELOG.md`).
- Infinite Loop Guard: `if: "!startsWith(github.event.head_commit.message, 'chore(release):')"`.
- **Mechanics:**
  1. Evaluates all commits since last Git tag.
  2. Bumps SemVer:
     - Breaking change $\rightarrow$ Major (`vX+1.0.0`)
     - `feat:` $\rightarrow$ Minor (`vX.Y+1.0`)
     - `fix:`, `perf:`, `chore:`, etc. $\rightarrow$ Patch (`vX.Y.Z+1`)
  3. Prepend categorized release notes to `CHANGELOG.md`.
  4. Syncs version constants across `cmd/syncalibur-engine/main.go` and `syncalibur.sh`.
  5. Commits `chore(release): bump version to <tag>`, creates annotated tag, and pushes to remote.
  6. Creates GitHub Release via `gh release create` and attaches compiled binary `bin/syncalibur`.

---

## 6. Complete Configuration Reference (`syncalibur.env`)

| Variable Name | Default Value | Description |
| :--- | :--- | :--- |
| `SYNCALIBUR_REPO_PATH` | `"."` | Local folder or repository to watch and sync. |
| `SYNCALIBUR_BIDIRECTIONAL` | `"true"` | Enables automatic remote fetch and rebase (`auto-pull`). |
| `SYNCALIBUR_PULL_INTERVAL` | `45` | Interval in seconds between background remote fetch cycles. |
| `SYNCALIBUR_SEMANTIC_COMMITS` | `"true"` | Formats commits according to Conventional Commits. |
| `SYNCALIBUR_COMMIT_STYLE` | `"conventional"` | Commit style (`conventional` or `simple`). |
| `SYNCALIBUR_SMART_STAGING` | `"true"` | Partitions multi-concern changes into atomic semantic commits. |
| `SYNCALIBUR_AI_COMMIT` | `"false"` | Local Ollama LLM commit message hook. |
| `SYNCALIBUR_OLLAMA_MODEL` | `"qwen2.5-coder:1.5b"`| Local Ollama model identifier. |
| `SYNCALIBUR_SELF_AUTO_UPDATE` | `"true"` | Watches Syncalibur's own repo, auto-pulls, rebuilds, hot-reloads. |
| `SYNCALIBUR_SELF_CHECK_INTERVAL`| `120` | Interval in seconds for checking upstream Syncalibur updates. |
| `SYNCALIBUR_AUTO_CREATE_REMOTE` | `"true"` | Auto-creates remote repository on GitHub/GitLab if absent. |
| `SYNCALIBUR_REPO_VISIBILITY` | `"private"` | Visibility for auto-created repositories (`private`/`public`). |
| `SYNCALIBUR_ALERT_TERMINAL` | `"true"` | Pops open a diagnostic terminal window upon sync error. |
| `SYNCALIBUR_ALERT_COOLDOWN` | `600` | Cooldown period (in seconds) between popup alert windows. |
| `SYNCALIBUR_TELEGRAM_BOT_TOKEN` | `""` | Telegram Bot token for instant mobile/desktop HTML cards. |
| `SYNCALIBUR_TELEGRAM_CHAT_ID` | `""` | Telegram destination chat ID. |
| `SYNCALIBUR_WEBHOOK_URL` | `""` | Discord, Slack, Telegram, or custom webhook URL. |
| `SYNCALIBUR_WEBHOOK_EVENTS` | `"failure,recovery,summary"` | Comma-delimited list of webhook events to dispatch. |
| `SYNCALIBUR_WEBHOOK_ON_SUCCESS` | `"false"` | Dispatches webhook on every successful push when true. |
| `SYNCALIBUR_REMOTE` | `""` | Target remote name (empty auto-selects `"all"` or `"origin"`). |
| `SYNCALIBUR_USE_FS_EVENTS` | `"true"` | Uses kernel filesystem events (`inotifywait` / `fswatch`). |
| `SYNCALIBUR_POLL_INTERVAL` | `3` | Idle watcher timeout / polling loop interval in seconds. |
| `SYNCALIBUR_DEBOUNCE_WAIT` | `5` | Quiescence delay (seconds) after last file write before committing. |
| `SYNCALIBUR_MAX_DEBOUNCE` | `60` | Maximum debounce wait before forcing commit. |
| `SYNCALIBUR_BURST_COOLDOWN` | `20` | Cooldown window (seconds) to batch rapid edits before pushing. |
| `SYNCALIBUR_AUTO_TIDY` | `"true"` | Automatically squashes burst micro-commits (`unpushed > 1`) into 1 clean commit. |
| `SYNCALIBUR_AUTO_PR` | `"false"` | Pushes to feature branch and auto-creates Pull Request. |
| `SYNCALIBUR_MAX_FILE_SIZE_MB` | `50` | Maximum file size in MiB before quarantine. |
| `SYNCALIBUR_SENSITIVE_PATTERNS` | `*.tfstate ...` | Space-delimited patterns for credential quarantine. |
| `SYNCALIBUR_PROTECT_SYNTAX` | `"true"` | Multi-language syntax hold guard. |
| `SYNCALIBUR_SECRET_SHIELD` | `"true"` | Shannon entropy & pattern credential detector. |
| `SYNCALIBUR_RETRY_BASE` | `30` | Initial exponential backoff retry interval (seconds). |
| `SYNCALIBUR_RETRY_MAX` | `900` | Maximum exponential backoff retry cap (15 minutes). |
| `SYNCALIBUR_PUSH_TIMEOUT` | `120` | Network timeout (seconds) for git push commands. |
| `SYNCALIBUR_LOG_MAX_KB` | `1024` | Maximum log size before automatic rotation. |

---

## 7. Living Memory Protocol & Maintenance Rules

Whenever changes are made to this codebase in future sessions:
1. **Never Break Zero-Dependency Standard:** The Go engine must remain 100% standard library. Do not add external third-party Go modules.
2. **Never Break Test Suite:** Any new Go feature must have matching unit test cases in `cmd/syncalibur-engine/main_test.go` (`go test -v ./...` must pass).
3. **Never Break POSIX/Bash Syntax:** All shell script modifications must pass `bash -n syncalibur.sh` and `bash -n install.sh`.
4. **Update This Memory Ledger:** When introducing new configuration flags, altering lifecycle behavior, or expanding command capabilities, update:
   - `PROJECT_MEMORY.md` (this file)
   - `syncalibur.env.example`
   - `syncalibur.sh` (including `CONFIG_KEYS` array)
