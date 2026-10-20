# Contributing to Syncalibur

Thank you for contributing to **Syncalibur**! To keep the `main` branch stable for all users running the autonomous self-updater, direct pushes and force pushes to `main` are disabled. All changes must go through a **Branch + Pull Request** workflow with **1 required review** and passing **CI checks**.

## Pull Request Workflow

1. **Fork & Clone** the repository (or create a feature branch if you have write access):
   ```bash
   git checkout -b feat/your-feature-name
   ```
2. **Make Your Changes**:
   - **Go Core Engine:** Edit `cmd/syncalibur-engine/main.go` (pure Go standard library, zero external dependencies).
   - **Daemon & Service Manager:** Edit `syncalibur.sh`.
3. **Run Local Verification Before Pushing**:
   ```bash
   # 1. Check Bash syntax
   bash -n syncalibur.sh && bash -n install.sh

   # 2. Build Go engine
   go vet ./cmd/syncalibur-engine
   go build -o bin/syncalibur-engine ./cmd/syncalibur-engine

   # 3. Run 10-point diagnostic suite
   ./syncalibur.sh doctor
   ```
4. **Push Your Branch & Open a Pull Request**:
   ```bash
   git push -u origin feat/your-feature-name
   gh pr create --fill
   ```
5. **Automated CI & Code Review**:
   - GitHub Actions (`Syncalibur CI`) will automatically run `bash -n`, `shellcheck`, `go vet`, `go build`, and secret hygiene checks.
   - `@YusufKaya00` (`CODEOWNERS`) is automatically requested for review.
   - Once **1 approval** is granted and **CI checks pass**, the PR can be merged cleanly.

## Security & Privacy Rules

- **Never commit** your personal `syncalibur.env`, `repos.list`, or compiled `bin/` binaries.
- Only update `syncalibur.env.example` and `repos.list.example` when adding new configuration variables.
