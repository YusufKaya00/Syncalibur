# Changelog

## [v2.10.1] - 2026-10-22

### 📚 Tooling, Tests & Fleet
- **syncalibur**: update syncalibur.sh (+11/-25) ([`6c650e8`](https://github.com/YusufKaya00/Syncalibur/commit/6c650e8))


## [v2.2.0] - 2026-10-20

### 🚀 Features
- **main**: update runRelease() (+10/-1) ([`86dfbcc`](https://github.com/YusufKaya00/Syncalibur/commit/86dfbcc))

### 🐛 Bug Fixes & Stability
- **main**: fix runRelease() (+4/-4) ([`f336c50`](https://github.com/YusufKaya00/Syncalibur/commit/f336c50))

### 📚 Tooling, Tests & Fleet
- **release**: update release.yml (+3/-0) ([`dadc4bd`](https://github.com/YusufKaya00/Syncalibur/commit/dadc4bd))


## [v2.1.0] - 2026-10-20

### 🚀 Features
- **main**: update runRelease() (+43/-1) ([`4cd5bfa`](https://github.com/YusufKaya00/Syncalibur/commit/4cd5bfa))
- **main**: update printVersion() (+8/-1) ([`c2a0a8a`](https://github.com/YusufKaya00/Syncalibur/commit/c2a0a8a))
- **syncalibur**: update syncalibur.sh (+2/-2) ([`0055b16`](https://github.com/YusufKaya00/Syncalibur/commit/0055b16))
- **syncalibur**: update syncalibur.sh (+2/-0) ([`26fe9d1`](https://github.com/YusufKaya00/Syncalibur/commit/26fe9d1))
- update logo.ansi, main.go (+5 more) (+429/-39) ([`3a84104`](https://github.com/YusufKaya00/Syncalibur/commit/3a84104))
- **20**: add 20 (+21/-0) ([`3f1db06`](https://github.com/YusufKaya00/Syncalibur/commit/3f1db06))
- **main**: update main.go (+24/-0) ([`fe3569d`](https://github.com/YusufKaya00/Syncalibur/commit/fe3569d))
- **main**: update runInteractiveTUI() (+1/-0) ([`fd92f9c`](https://github.com/YusufKaya00/Syncalibur/commit/fd92f9c))
- **main**: update runInteractiveTUI() (+23/-3) ([`0835365`](https://github.com/YusufKaya00/Syncalibur/commit/0835365))
- **main**: update renderLineANSI() (+8/-2) ([`b4b541d`](https://github.com/YusufKaya00/Syncalibur/commit/b4b541d))
- **logo_data**: add logo_data.go (+26/-0) ([`2af6911`](https://github.com/YusufKaya00/Syncalibur/commit/2af6911))
- add main() (+309/-5) ([`11ec779`](https://github.com/YusufKaya00/Syncalibur/commit/11ec779))

### 🐛 Bug Fixes & Stability
- **engine**: support cross-platform Windows terminal size and process check ([`1640872`](https://github.com/YusufKaya00/Syncalibur/commit/1640872))

### 📚 Tooling, Tests & Fleet
- **release**: add release.yml (+45/-0) ([`78e69bd`](https://github.com/YusufKaya00/Syncalibur/commit/78e69bd))
- **main_test**: update main_test.go (+1/-0) ([`6999445`](https://github.com/YusufKaya00/Syncalibur/commit/6999445))
- **logo**: remove logo.ansi (+0/-21) ([`b966615`](https://github.com/YusufKaya00/Syncalibur/commit/b966615))
- **main_test**: update TestPlanSmartStaging (+47/-0) ([`196e884`](https://github.com/YusufKaya00/Syncalibur/commit/196e884))
- **main_test**: update TestLoadLogoLines (+3/-2) ([`b089bd1`](https://github.com/YusufKaya00/Syncalibur/commit/b089bd1))
- **main_test**: update main_test.go (+0/-1) ([`6795ff7`](https://github.com/YusufKaya00/Syncalibur/commit/6795ff7))
- **main**: update validateFileSyntax() (+0/-14) ([`4109772`](https://github.com/YusufKaya00/Syncalibur/commit/4109772))
- **main**: update main.go (+20/-26) ([`3b131ee`](https://github.com/YusufKaya00/Syncalibur/commit/3b131ee))
- **ci**: update ci.yml (+1/-1) ([`89a8553`](https://github.com/YusufKaya00/Syncalibur/commit/89a8553))
- **20**: remove 20 (+0/-21) ([`3ff15f2`](https://github.com/YusufKaya00/Syncalibur/commit/3ff15f2))
- **CONTRIBUTING**: update CONTRIBUTING.md (+1/-1) ([`e9fb79f`](https://github.com/YusufKaya00/Syncalibur/commit/e9fb79f))
- **ci**: update ci.yml (+1/-1) ([`bb9e44d`](https://github.com/YusufKaya00/Syncalibur/commit/bb9e44d))
- **logo_data**: remove logo_data.go (+0/-26) ([`6ac5a48`](https://github.com/YusufKaya00/Syncalibur/commit/6ac5a48))
- **main_test**: update TestLoadLogoLines (+17/-1) ([`2ee33a8`](https://github.com/YusufKaya00/Syncalibur/commit/2ee33a8))
- **main**: update loadLogoLines() (+2/-25) ([`33eb136`](https://github.com/YusufKaya00/Syncalibur/commit/33eb136))


All notable changes to **Syncalibur** are documented in this file.
The project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html) and follows [Conventional Commits](https://www.conventionalcommits.org/).

---

## [v2.0.0] - 2026-10-20

### 🚀 Major Architectural Breakthroughs
- **Compiled Go Native Core Engine (`bin/syncalibur`):** Migrated entire synchronization, AST parsing, and terminal rendering pipeline from legacy shell scripts to a compiled Go standard library binary (`cmd/syncalibur-engine/main.go`). Runs under its own compiled process name with zero external dependencies, ~0.0% idle CPU, and ~7 MB RAM.
- **Universal Multi-Language Syntax Hold:** Intelligent pre-flight syntax holding that prevents commits mid-typing if brackets, braces, quotes, or language statements are unclosed. Combines native compilers (`go/parser`, `node --check`, `py_compile`, `tsc`, `javac`, `bash -n`) with a custom Universal Lexical Balance Engine supporting 11+ languages (Go, Python, React JSX/TSX, TypeScript, Node.js, Java, Bash, Rust, C/C++, YAML, PHP).
- **Shannon Entropy Secret Shield:** Real-time mathematical entropy calculation ($H \ge 4.4$ bit/char) paired with high-confidence pattern scanning (AWS keys, GitHub PATs, OpenAI tokens, Slack webhooks, Private RSA/ECDSA keys). Automatically quarantines sensitive files out of Git staging without breaking uninterrupted workflow.
- **Concurrent Multi-Repository Fleet Supervisor:** Parallel discovery and supervision of multiple repositories (`syncalibur fleet scan`, `repos.list`) with status synchronization across all local projects.
- **Interactive Two-Pane Live TUI Cockpit (`syncalibur ui`):** Full-screen terminal dashboard built with raw terminal mode and alternate screen buffering. Features 24-bit TrueColor ANSI logo rendering, live commit sparkline, real-time daemon state, scrollable log viewport with Vim keys (`j`/`k`, `g`/`G`), and single-key actions (`p`, `s`, `b`, `t`, `d`, `q`).
- **Bidirectional Rebase Auto-Sync:** Automated remote tracking with `--rebase --autostash` safety. Instantly detects remote divergence, pulls cleanly on clean trees, and automatically aborts (`git rebase --abort`) if merge conflicts arise, alerting the user via terminal and webhook.
- **Automated SemVer & Release Engine (`syncalibur release`):** Native commit history analyzer parsing Conventional Commits, calculating SemVer bumps, generating formatted changelogs, creating annotated Git tags, and broadcasting to remotes.

---

### 🛡️ Shields & Safety Pre-flight (12 Commits)
- **Mathematical Entropy Analysis:** Implemented `calculateShannonEntropy()` to flag cryptographically random high-entropy strings and quarantine credentials (`calculateShannonEntropy()`, `detectSecretsInFile()`).
- **Lexical Balance Engine:** Implemented `bracketToken` stack parser that ignores string literals, regexes, and comments while validating nesting integrity (`checkLexicalBalance()`).
- **YAML & Helm Block Scalar Support:** Implemented dedicated `validateYAMLSyntax()` supporting block scalars (`|`, `>`), Jinja (`{% ... %}`), and Helm templates (`{{ ... }}`) without falsely treating CSS hex colors as comments.
- **Pre-flight Pipeline Integration:** Bound secrets scanning, syntax hold verification, and large-file filters into a non-blocking commit gate (`runPreflight()`, `validateFileSyntax()`).
- **AST Enclosing Symbol Detection:** Enclosing function and component detection for precise Semantic Conventional Commits (`detectEnclosingSymbol()`).

---

### ✨ Features & Enhancements (53 Commits)
- **Live TUI Cockpit & Emblem:** Side-by-side TrueColor sword emblem and live telemetry cockpit with dynamic viewport scrolling, terminal resize auto-detection (`TIOCGWINSZ`), and instant toggle (`[t]`).
- **TrueColor ANSI Graphics:** 24-bit TrueColor terminal graphic rendering of Syncalibur emblem directly in terminal output (`assets/syncalibur-logo.ansi`).
- **Fleet Orchestration:** Added `syncalibur fleet scan`, interactive fleet tables, and multi-repo process oversight (`cmd_fleet`).
- **Notifications & Webhooks:** Rich alerts dispatched to Telegram, Discord, Slack, and generic webhooks with test commands (`syncalibur notify`).
- **11-Point System Doctor:** Diagnostic command verifying Git identity, compiler health, watcher daemons, SSH connectivity, and push privileges (`syncalibur doctor`).
- **One-Liner Global Installer:** Self-contained global installer script linking into `~/.local/bin/syncalibur` (`install.sh`).

---

### 🐛 Bug Fixes & Hardening (5 Commits)
- **Installer & Symlink Resolution:** Fixed cross-platform symlinking and permission issues in global install script (`install.sh`).
- **Engine Process Signaling:** Corrected PID lockfile handling and clean subprocess termination on exit (`syncalibur.sh`).
- **Parser Robustness:** Fixed edge-case nil pointer panic in Go AST parser when scanning empty or malformed files (`main.go`).
- **Rebase Recovery:** Ensured clean working-tree recovery when aborting remote rebase conflicts (`syncalibur.sh`).
- **YAML Lexing:** Fixed CSS hex colors in Helm ConfigMap templates triggering false unclosed bracket syntax errors.

---

### 🛠️ Tooling, Testing & Documentation (33 Commits)
- **Unit Test Coverage:** Comprehensive test harness in `main_test.go` covering entropy formulas, lexical balance, bracket tokens, SemVer conventional commits, and syntax validations.
- **CI/CD Pipeline:** Added GitHub Actions automation for `shellcheck`, `bash -n`, `go vet`, and security hygiene verification (`.github/workflows/ci.yml`).
- **PR Governance:** Enforced branch protection, CODEOWNERS, structured PR templates, and auto-labelers (`.github/workflows/pr-governance.yml`).
- **Documentation:** Authored comprehensive architectural guides, Mermaid pipeline diagrams, Lucide glyph legends, and developer guidelines (`CONTRIBUTING.md`, `README.md`).

---

### 📦 Installation & Upgrade

```bash
# One-liner quick upgrade / install
curl -fsSL https://raw.githubusercontent.com/YusufKaya00/Syncalibur/main/install.sh | bash

# Verify installation & system doctor
syncalibur doctor

# Launch Live TUI Cockpit
syncalibur ui
```