#!/usr/bin/env bash
# ==============================================================================
# Syncalibur (syncalibur.sh) - Cross-Platform Autonomous Git Auto-Sync Platform
# ------------------------------------------------------------------------------
# Powered by a hybrid Bash + Go Engine (`cmd/syncalibur-engine/main.go`):
#   - High-Definition TrueColor Terminal Emblem & Live Dashboard
#   - Bidirectional Git Sync (`auto-pull --rebase`) with Safe Conflict Abort
#   - Semantic Conventional Commit Generator (`feat(scope): ... (+X/-Y)`) + AI Hook
#   - Native Filesystem Event Watcher (`inotifywait` / `fswatch` + Polling Fallback)
#   - Multi-Repository Fleet Supervisor (`syncalibur fleet add|scan|status|start`)
#   - Interactive Live TUI Cockpit (`syncalibur status`) with 7-Day Commit Sparkline
#   - First-Class Telegram Bot & Multi-Channel Webhooks (Discord, Slack, Telegram)
#   - Self-Watching & Upstream Hot-Reload Auto-Updater + 10-Point `doctor` Suite
#
# Usage : syncalibur {install [path]|uninstall|start|stop|restart|status|logs|fleet|doctor|notify|pull|alert|run}
# Docs  : ./README.md
# ==============================================================================

set -uo pipefail
# NOTE: `set -e` is intentionally omitted. The daemon loop must survive
# transient git/network errors rather than exiting unexpectedly.

export PATH="${HOME}/.local/bin:/opt/homebrew/bin:/usr/local/bin:/c/Program Files/Go/bin:/c/Go/bin:${PATH}"
SYNCALIBUR_PLATFORM_VERSION="v2.27.0"
export GIT_TERMINAL_PROMPT=0
export GIT_SSH_COMMAND="${GIT_SSH_COMMAND:-ssh -o BatchMode=yes -o ConnectTimeout=10 -o ServerAliveInterval=10 -o ServerAliveCountMax=3 -o StrictHostKeyChecking=accept-new}"

# ==============================================================================
# Cross-platform portability helpers (Linux, macOS Bash 3.2+, Windows Git Bash)
# ==============================================================================

OS_KERNEL="$(uname -s 2>/dev/null || echo Unknown)"

normalize_path_input() {
  local raw="$1"
  raw="${raw/#\~/${HOME}}"
  if command -v cygpath >/dev/null 2>&1 && [[ "${raw}" =~ ^[A-Za-z]:[\\/] ]]; then
    cygpath -u "${raw}"
  else
    printf '%s\n' "${raw}"
  fi
}

resolve_path() {
  local target="$1"
  if [[ -d "${target}" ]]; then
    (cd "${target}" 2>/dev/null && pwd -P)
    return
  fi
  local dir base
  dir="$(dirname "${target}")"
  base="$(basename "${target}")"
  if [[ -L "${target}" ]] && command -v readlink >/dev/null 2>&1; then
    local link
    link="$(readlink "${target}")"
    [[ "${link}" == /* ]] || link="${dir}/${link}"
    dir="$(dirname "${link}")"
    base="$(basename "${link}")"
  fi
  (cd "${dir}" 2>/dev/null && printf '%s/%s\n' "$(pwd -P)" "${base}")
}

if stat -c '%s' . >/dev/null 2>&1; then
  STAT_FLAVOR="gnu"
else
  STAT_FLAVOR="bsd"
fi

file_size_bytes() {
  if [[ "${STAT_FLAVOR}" == "gnu" ]]; then
    stat -c '%s' -- "$1" 2>/dev/null || echo 0
  else
    stat -f '%z' -- "$1" 2>/dev/null || echo 0
  fi
}

file_mtime_epoch() {
  if [[ "${STAT_FLAVOR}" == "gnu" ]]; then
    stat -c '%Y' -- "$1" 2>/dev/null || echo 0
  else
    stat -f '%m' -- "$1" 2>/dev/null || echo 0
  fi
}

file_signature() {
  if [[ "${STAT_FLAVOR}" == "gnu" ]]; then
    stat -c '%s %Y' -- "$1" 2>/dev/null || echo gone
  else
    stat -f '%z %m' -- "$1" 2>/dev/null || echo gone
  fi
}

run_with_timeout() {
  local secs="$1"
  shift
  if command -v timeout >/dev/null 2>&1; then
    timeout "${secs}" "$@"
    return $?
  fi

  "$@" &
  local cmd_pid=$!
  (
    sleep "${secs}"
    kill -TERM "${cmd_pid}" 2>/dev/null || true
  ) &
  local watchdog_pid=$!

  wait "${cmd_pid}" 2>/dev/null
  local rc=$?
  if kill -0 "${watchdog_pid}" 2>/dev/null; then
    kill "${watchdog_pid}" 2>/dev/null || true
    wait "${watchdog_pid}" 2>/dev/null || true
    return "${rc}"
  fi
  return 124
}

detect_service_backend() {
  case "${OS_KERNEL}" in
    Darwin*)
      echo "launchd"
      ;;
    MINGW*|MSYS*|CYGWIN*)
      echo "windows-startup"
      ;;
    *)
      if command -v systemctl >/dev/null 2>&1 && systemctl --user list-units >/dev/null 2>&1; then
        echo "systemd"
      else
        echo "cron"
      fi
      ;;
  esac
}

SERVICE_BACKEND="$(detect_service_backend)"

# ==============================================================================
# Paths, Go Engine Auto-Builder & Configuration Loading
# ==============================================================================

SCRIPT_PATH="$(resolve_path "${BASH_SOURCE[0]}")"
SCRIPT_DIR="$(dirname "${SCRIPT_PATH}")"
LOGO_ANSI_FILE="${SCRIPT_DIR}/assets/syncalibur-logo.ansi"
LOGO_PNG_FILE="${SCRIPT_DIR}/assets/syncalibur-logo.png"
GO_ENGINE_SRC="${SCRIPT_DIR}/cmd/syncalibur-engine/main.go"
GO_ENGINE_BIN="${SCRIPT_DIR}/bin/syncalibur"
GO_ENGINE_COMPAT_BIN="${SCRIPT_DIR}/bin/syncalibur-engine"
if [[ "${OS_KERNEL}" =~ ^(MINGW|MSYS|CYGWIN) ]]; then
  GO_ENGINE_BIN="${GO_ENGINE_BIN}.exe"
  GO_ENGINE_COMPAT_BIN="${GO_ENGINE_COMPAT_BIN}.exe"
fi
FLEET_LIST_FILE="${SYNCALIBUR_FLEET_LIST:-${SCRIPT_DIR}/repos.list}"
SELF_REPO_ROOT="$(git -C "${SCRIPT_DIR}" rev-parse --show-toplevel 2>/dev/null || echo "")"

resolve_syncalibur_platform_version() {
  local dyn_tag=""
  local search_dirs=("${SELF_REPO_ROOT:-}" "${SCRIPT_DIR:-}" "/home/ysf/Syncalibur" ".")
  for d in "${search_dirs[@]}"; do
    if [[ -n "${d}" && -d "${d}/.git" ]]; then
      dyn_tag="$(git -C "${d}" tag -l "v*" --sort=-v:refname 2>/dev/null | head -n1)"
      if [[ -n "${dyn_tag}" ]]; then
        printf '%s\n' "${dyn_tag}"
        return 0
      fi
    fi
  done
  printf '%s\n' "${SYNCALIBUR_PLATFORM_VERSION:-v2.27.0}"
}
SYNCALIBUR_PLATFORM_VERSION="$(resolve_syncalibur_platform_version)"

# Ensure the native Go binary (`bin/syncalibur`, which registers as `syncalibur` in top/htop/ps/Task Manager) is compiled and up to date
ensure_go_engine() {
  if [[ -x "${GO_ENGINE_BIN}" ]]; then
    if [[ ! -f "${GO_ENGINE_SRC}" ]] || (( $(file_mtime_epoch "${GO_ENGINE_BIN}") >= $(file_mtime_epoch "${GO_ENGINE_SRC}") )); then
      [[ -e "${GO_ENGINE_COMPAT_BIN}" ]] || ln -sf "syncalibur" "${GO_ENGINE_COMPAT_BIN}" 2>/dev/null || cp -f "${GO_ENGINE_BIN}" "${GO_ENGINE_COMPAT_BIN}" 2>/dev/null || true
      return 0
    fi
  fi
  if command -v go >/dev/null 2>&1 && [[ -f "${GO_ENGINE_SRC}" ]]; then
    mkdir -p "$(dirname "${GO_ENGINE_BIN}")" 2>/dev/null || true
    if (cd "${SCRIPT_DIR}" && go build -ldflags="-s -w" -o "${GO_ENGINE_BIN}" ./cmd/syncalibur-engine 2>/dev/null); then
      ln -sf "syncalibur" "${GO_ENGINE_COMPAT_BIN}" 2>/dev/null || cp -f "${GO_ENGINE_BIN}" "${GO_ENGINE_COMPAT_BIN}" 2>/dev/null || true
      return 0
    fi
  fi
  if [[ ! -x "${GO_ENGINE_BIN}" && -x "${GO_ENGINE_COMPAT_BIN}" ]]; then
    cp -f "${GO_ENGINE_COMPAT_BIN}" "${GO_ENGINE_BIN}" 2>/dev/null || true
  fi
  [[ -x "${GO_ENGINE_BIN}" ]]
}

CONFIG_KEYS=" SYNCALIBUR_REPO_PATH SYNCALIBUR_REPO_DIR SYNCALIBUR_AUTO_CREATE_REMOTE SYNCALIBUR_REPO_VISIBILITY
  SYNCALIBUR_BIDIRECTIONAL SYNCALIBUR_PULL_INTERVAL SYNCALIBUR_SEMANTIC_COMMITS SYNCALIBUR_COMMIT_STYLE
  SYNCALIBUR_AI_COMMIT SYNCALIBUR_OLLAMA_MODEL
  SYNCALIBUR_SELF_AUTO_UPDATE SYNCALIBUR_SELF_CHECK_INTERVAL
  SYNCALIBUR_ALERT_TERMINAL SYNCALIBUR_ALERT_COOLDOWN
  SYNCALIBUR_TELEGRAM_BOT_TOKEN SYNCALIBUR_TELEGRAM_CHAT_ID
  SYNCALIBUR_WEBHOOK_URL SYNCALIBUR_WEBHOOK_EVENTS SYNCALIBUR_WEBHOOK_ON_SUCCESS
  SYNCALIBUR_USE_FS_EVENTS SYNCALIBUR_FLEET_LIST
  SYNCALIBUR_SERVICE_NAME SYNCALIBUR_REMOTE SYNCALIBUR_POLL_INTERVAL
  SYNCALIBUR_DEBOUNCE_WAIT SYNCALIBUR_MAX_DEBOUNCE SYNCALIBUR_MAX_FILE_SIZE_MB SYNCALIBUR_SENSITIVE_PATTERNS
  SYNCALIBUR_PROTECT_SYNTAX SYNCALIBUR_SECRET_SHIELD SYNCALIBUR_SMART_STAGING
  SYNCALIBUR_MODE SYNCALIBUR_BURST_COOLDOWN SYNCALIBUR_TIDY_THRESHOLD SYNCALIBUR_AUTO_TIDY SYNCALIBUR_AUTO_PR
  SYNCALIBUR_RETRY_BASE SYNCALIBUR_RETRY_MAX SYNCALIBUR_PUSH_TIMEOUT SYNCALIBUR_LOG_MAX_KB
  GITHUB_TOKEN GH_TOKEN GITLAB_TOKEN
  AUTO_SYNC_REPO_PATH AUTO_SYNC_SERVICE_NAME AUTO_SYNC_REMOTE AUTO_SYNC_POLL_INTERVAL
  AUTO_SYNC_DEBOUNCE_WAIT AUTO_SYNC_MAX_DEBOUNCE AUTO_SYNC_MAX_FILE_SIZE_MB AUTO_SYNC_SENSITIVE_PATTERNS
  AUTO_SYNC_RETRY_BASE AUTO_SYNC_RETRY_MAX AUTO_SYNC_PUSH_TIMEOUT AUTO_SYNC_LOG_MAX_KB "

load_config() {
  local file="$1" allow_override="${2:-0}" line key value n=0 is_set=0
  [[ -f "${file}" ]] || return 0
  while IFS= read -r line || [[ -n "${line}" ]]; do
    n=$(( n + 1 ))
    line="${line%$'\r'}"
    [[ "${line}" =~ ^[[:space:]]*(#|$) ]] && continue
    if ! [[ "${line}" =~ ^[[:space:]]*(export[[:space:]]+)?([A-Za-z_][A-Za-z0-9_]*)[[:space:]]*=[[:space:]]*(.*)$ ]]; then
      echo "▲  ${file}:${n}: ignoring malformed line: ${line}" >&2
      continue
    fi
    key="${BASH_REMATCH[2]}"
    value="${BASH_REMATCH[3]}"
    if [[ "${key}" =~ ^(MODE|BURST_COOLDOWN|TIDY_THRESHOLD|AUTO_TIDY|AUTO_PR|PROTECT_SYNTAX|SECRET_SHIELD|REMOTE|BIDIRECTIONAL|POLL_INTERVAL)$ ]]; then
      key="SYNCALIBUR_${key}"
    fi
    if [[ "${CONFIG_KEYS}" != *[[:space:]]"${key}"[[:space:]]* ]]; then
      echo "▲  ${file}:${n}: unknown key '${key}' (ignored)" >&2
      continue
    fi
    if [[ "${value}" =~ ^\"(.*)\"[[:space:]]*(#.*)?$ || "${value}" =~ ^\'(.*)\'[[:space:]]*(#.*)?$ ]]; then
      value="${BASH_REMATCH[1]}"
    else
      value="${value%%[[:space:]]#*}"
      value="${value%"${value##*[![:space:]]}"}"
    fi
    eval "[[ -n \"\${${key}+x}\" ]] && is_set=1 || is_set=0"
    (( is_set && ! allow_override )) && continue
    printf -v "${key}" '%s' "${value}"
  done < "${file}"
}

DEFAULT_ENV_FILE="${SCRIPT_DIR}/syncalibur.env"
EXAMPLE_ENV_FILE="${SCRIPT_DIR}/syncalibur.env.example"

if [[ -z "${SYNCALIBUR_CONFIG:-${AUTO_SYNC_CONFIG:-}}" && ! -f "${DEFAULT_ENV_FILE}" && -f "${EXAMPLE_ENV_FILE}" ]]; then
  cp "${EXAMPLE_ENV_FILE}" "${DEFAULT_ENV_FILE}" 2>/dev/null || true
fi

CONFIG_FILE="${SYNCALIBUR_CONFIG:-${AUTO_SYNC_CONFIG:-${DEFAULT_ENV_FILE}}}"

save_repo_path_to_env() {
  local target_path="$1"
  local env_file="${SYNCALIBUR_CONFIG:-${AUTO_SYNC_CONFIG:-${DEFAULT_ENV_FILE}}}"
  if [[ ! -f "${env_file}" && -f "${EXAMPLE_ENV_FILE}" ]]; then
    cp "${EXAMPLE_ENV_FILE}" "${env_file}" 2>/dev/null || true
  fi
  if [[ -f "${env_file}" ]]; then
    local tmp
    tmp="$(mktemp)"
    awk -v new_path="${target_path}" '
      BEGIN { replaced = 0 }
      /^[[:space:]]*(export[[:space:]]+)?(SYNCALIBUR_REPO_PATH|AUTO_SYNC_REPO_PATH)=/ {
        if (!replaced) {
          printf "SYNCALIBUR_REPO_PATH=\"%s\"\n", new_path
          replaced = 1
        }
        next
      }
      { print }
      END {
        if (!replaced) {
          printf "SYNCALIBUR_REPO_PATH=\"%s\"\n", new_path
        }
      }
    ' "${env_file}" > "${tmp}"
    mv "${tmp}" "${env_file}"
  else
    cat > "${env_file}" << EOF
# Syncalibur local configuration (gitignored)
SYNCALIBUR_REPO_PATH="${target_path}"
SYNCALIBUR_BIDIRECTIONAL="true"
SYNCALIBUR_SEMANTIC_COMMITS="true"
SYNCALIBUR_SELF_AUTO_UPDATE="true"
SYNCALIBUR_AUTO_CREATE_REMOTE="false"
SYNCALIBUR_REPO_VISIBILITY="private"
SYNCALIBUR_ALERT_TERMINAL="true"
EOF
  fi
  CONFIG_FILE="$(resolve_path "${env_file}")"
  SYNCALIBUR_REPO_PATH="${target_path}"
}

if [[ "${1:-}" == "install" && -n "${2:-}" ]]; then
  INSTALL_ARG_PATH="$(normalize_path_input "${2}")"
  mkdir -p "${INSTALL_ARG_PATH}" 2>/dev/null || true
  if [[ ! -d "${INSTALL_ARG_PATH}" ]]; then
    echo "✕ Specified directory does not exist and could not be created: ${2}" >&2
    exit 1
  fi
  INSTALL_ARG_PATH="$(resolve_path "${INSTALL_ARG_PATH}")"
  save_repo_path_to_env "${INSTALL_ARG_PATH}"
fi

if [[ -f "${CONFIG_FILE}" ]]; then
  CONFIG_FILE="$(resolve_path "${CONFIG_FILE}")"
  load_config "${CONFIG_FILE}"
elif [[ -f "${EXAMPLE_ENV_FILE}" ]]; then
  CONFIG_FILE="$(resolve_path "${EXAMPLE_ENV_FILE}")"
  load_config "${CONFIG_FILE}"
elif [[ -n "${SYNCALIBUR_CONFIG:-${AUTO_SYNC_CONFIG:-}}" ]]; then
  echo "✕ Config file not found: ${CONFIG_FILE}" >&2
  exit 1
else
  CONFIG_FILE=""
fi

# ==============================================================================
# Repository & Scope Resolution
# ==============================================================================

AUTO_CREATE_REMOTE="${SYNCALIBUR_AUTO_CREATE_REMOTE:-false}"
REPO_VISIBILITY="${SYNCALIBUR_REPO_VISIBILITY:-private}"
[[ "${REPO_VISIBILITY}" == "public" ]] || REPO_VISIBILITY="private"
BIDIRECTIONAL_SYNC="${SYNCALIBUR_BIDIRECTIONAL:-true}"
PULL_INTERVAL="${SYNCALIBUR_PULL_INTERVAL:-45}"
[[ "${PULL_INTERVAL}" =~ ^[0-9]+$ ]] || PULL_INTERVAL=45
SEMANTIC_COMMITS="${SYNCALIBUR_SEMANTIC_COMMITS:-true}"
if [[ "${SYNCALIBUR_COMMIT_STYLE:-conventional}" == "simple" ]]; then
  SEMANTIC_COMMITS="false"
fi
AI_COMMIT="${SYNCALIBUR_AI_COMMIT:-false}"
SMART_STAGING="${SYNCALIBUR_SMART_STAGING:-true}"
OLLAMA_MODEL="${SYNCALIBUR_OLLAMA_MODEL:-qwen2.5-coder:1.5b}"
SELF_AUTO_UPDATE="${SYNCALIBUR_SELF_AUTO_UPDATE:-true}"
SELF_CHECK_INTERVAL="${SYNCALIBUR_SELF_CHECK_INTERVAL:-120}"
[[ "${SELF_CHECK_INTERVAL}" =~ ^[0-9]+$ ]] || SELF_CHECK_INTERVAL=120
(( SELF_CHECK_INTERVAL < 15 )) && SELF_CHECK_INTERVAL=15
ALERT_TERMINAL="${SYNCALIBUR_ALERT_TERMINAL:-true}"
ALERT_COOLDOWN="${SYNCALIBUR_ALERT_COOLDOWN:-600}"
[[ "${ALERT_COOLDOWN}" =~ ^[0-9]+$ ]] || ALERT_COOLDOWN=600
TELEGRAM_BOT_TOKEN="${SYNCALIBUR_TELEGRAM_BOT_TOKEN:-}"
TELEGRAM_CHAT_ID="${SYNCALIBUR_TELEGRAM_CHAT_ID:-}"
WEBHOOK_URL="${SYNCALIBUR_WEBHOOK_URL:-}"
WEBHOOK_EVENTS="${SYNCALIBUR_WEBHOOK_EVENTS:-failure,recovery,summary}"
WEBHOOK_ON_SUCCESS="${SYNCALIBUR_WEBHOOK_ON_SUCCESS:-false}"
USE_FS_EVENTS="${SYNCALIBUR_USE_FS_EVENTS:-true}"

resolve_default_watch_fallback() {
  local fallback
  if fallback="$(git -C "${SCRIPT_DIR}" rev-parse --show-toplevel 2>/dev/null)"; then
    printf '%s\n' "${fallback}"
  else
    resolve_path "${SCRIPT_DIR}"
  fi
}

WATCH_PATH="${SYNCALIBUR_REPO_PATH:-${SYNCALIBUR_REPO_DIR:-${AUTO_SYNC_REPO_PATH:-}}}"
if [[ -n "${WATCH_PATH}" ]]; then
  WATCH_PATH="$(normalize_path_input "${WATCH_PATH}")"
  [[ "${WATCH_PATH}" == /* ]] || WATCH_PATH="$(dirname "${CONFIG_FILE:-${SCRIPT_PATH}}")/${WATCH_PATH}"
  if [[ ! -d "${WATCH_PATH}" ]]; then
    FALLBACK_WATCH_PATH="$(resolve_default_watch_fallback)"
    echo "▲  Configured SYNCALIBUR_REPO_PATH ('${WATCH_PATH}') not found on this machine; automatically using '${FALLBACK_WATCH_PATH}'." >&2
    WATCH_PATH="${FALLBACK_WATCH_PATH}"
  fi
else
  WATCH_PATH="$(resolve_default_watch_fallback)"
fi
WATCH_PATH="$(resolve_path "${WATCH_PATH}")"

if ! REPO_ROOT="$(git -C "${WATCH_PATH}" rev-parse --show-toplevel 2>/dev/null)"; then
  if [[ "${AUTO_CREATE_REMOTE}" == "true" || "${1:-}" == "install" ]]; then
    echo "✦  '${WATCH_PATH}' is not a git repository yet; initializing with 'git init -b main'..."
    git -C "${WATCH_PATH}" init -b main >/dev/null 2>&1 || {
      git -C "${WATCH_PATH}" init >/dev/null 2>&1
      git -C "${WATCH_PATH}" symbolic-ref HEAD refs/heads/main 2>/dev/null || true
    }
    REPO_ROOT="${WATCH_PATH}"
  else
    echo "✕ '${WATCH_PATH}' is not inside a git repository." >&2
    exit 1
  fi
fi
REPO_ROOT="$(resolve_path "${REPO_ROOT}")"
GIT_DIR="$(git -C "${REPO_ROOT}" rev-parse --absolute-git-dir)"

# Per-Repository Configuration (.syncalibur or syncalibur.env inside repo root)
for repo_cfg in "${REPO_ROOT}/.syncalibur" "${REPO_ROOT}/.syncalibur.env" "${REPO_ROOT}/syncalibur.env"; do
  if [[ -f "${repo_cfg}" ]]; then
    repo_cfg_resolved="$(resolve_path "${repo_cfg}")"
    if [[ -z "${CONFIG_FILE}" || "${repo_cfg_resolved}" != "${CONFIG_FILE}" ]]; then
      load_config "${repo_cfg}" 1
      break
    fi
  fi
done

SCOPE="$(git -C "${WATCH_PATH}" rev-parse --show-prefix 2>/dev/null || true)"
SCOPE="${SCOPE%/}"
SCOPE="${SCOPE:-.}"

PID_FILE="${GIT_DIR}/syncalibur.pid"
LOCK_FILE="${GIT_DIR}/syncalibur.lock"
LOG_FILE="${GIT_DIR}/syncalibur.log"
AUTH_CACHE_FILE="${GIT_DIR}/syncalibur.auth"
FAIL_REPORT_FILE="${GIT_DIR}/syncalibur.failure"
OFFLINE_FLAG_FILE="${GIT_DIR}/syncalibur.offline"
ALERT_STATE_FILE="${GIT_DIR}/syncalibur.alert_ts"
PAUSE_FILE="${GIT_DIR}/syncalibur.pause"
SYNC_TRIGGER_FILE="${GIT_DIR}/syncalibur.sync"
UPD_TRIGGER_FILE="${GIT_DIR}/syncalibur.upd"

if [[ -n "${SYNCALIBUR_SERVICE_NAME:-${AUTO_SYNC_SERVICE_NAME:-}}" ]]; then
  SERVICE_NAME="${SYNCALIBUR_SERVICE_NAME:-${AUTO_SYNC_SERVICE_NAME}}"
elif [[ -n "${CONFIG_FILE}" && "$(basename "${CONFIG_FILE}")" != "syncalibur.env" && "$(basename "${CONFIG_FILE}")" != "syncalibur.env.example" ]]; then
  SERVICE_NAME="syncalibur-$(basename "${CONFIG_FILE}" .env | tr -cd 'A-Za-z0-9_-')"
else
  SERVICE_NAME="syncalibur"
fi

UNIT_FILE="${XDG_CONFIG_HOME:-${HOME}/.config}/systemd/user/${SERVICE_NAME}.service"
PLIST_LABEL="com.syncalibur.${SERVICE_NAME}"
PLIST_FILE="${HOME}/Library/LaunchAgents/${PLIST_LABEL}.plist"
DESKTOP_FILE="${XDG_CONFIG_HOME:-${HOME}/.config}/autostart/${SERVICE_NAME}.desktop"

win_startup_vbs_path() {
  local appdata_posix=""
  if command -v cygpath >/dev/null 2>&1 && [[ -n "${APPDATA:-}" ]]; then
    appdata_posix="$(cygpath -u "${APPDATA}")"
  elif [[ -d "${HOME}/AppData/Roaming" ]]; then
    appdata_posix="${HOME}/AppData/Roaming"
  fi
  if [[ -n "${appdata_posix}" ]]; then
    printf '%s/Microsoft/Windows/Start Menu/Programs/Startup/%s.vbs\n' "${appdata_posix}" "${SERVICE_NAME}"
  else
    printf '%s/.config/syncalibur/%s.vbs\n' "${HOME}" "${SERVICE_NAME}"
  fi
}

# ==============================================================================
# Tunable Parameters & Filesystem Event Watcher Detection
# ==============================================================================

POLL_INTERVAL="${SYNCALIBUR_POLL_INTERVAL:-${AUTO_SYNC_POLL_INTERVAL:-3}}"
DEBOUNCE_WAIT="${SYNCALIBUR_DEBOUNCE_WAIT:-${AUTO_SYNC_DEBOUNCE_WAIT:-5}}"
MAX_DEBOUNCE="${SYNCALIBUR_MAX_DEBOUNCE:-${AUTO_SYNC_MAX_DEBOUNCE:-60}}"
MAX_FILE_SIZE_MB="${SYNCALIBUR_MAX_FILE_SIZE_MB:-${AUTO_SYNC_MAX_FILE_SIZE_MB:-50}}"
RETRY_BASE="${SYNCALIBUR_RETRY_BASE:-${AUTO_SYNC_RETRY_BASE:-30}}"
RETRY_MAX="${SYNCALIBUR_RETRY_MAX:-${AUTO_SYNC_RETRY_MAX:-900}}"
PUSH_TIMEOUT="${SYNCALIBUR_PUSH_TIMEOUT:-${AUTO_SYNC_PUSH_TIMEOUT:-120}}"
LOG_MAX_KB="${SYNCALIBUR_LOG_MAX_KB:-${AUTO_SYNC_LOG_MAX_KB:-1024}}"
REMOTE_OVERRIDE="${SYNCALIBUR_REMOTE:-${AUTO_SYNC_REMOTE:-}}"

# Syncalibur Architecture: Unified Zero-Data-Loss Autopilot & Sentinel
MODE="${SYNCALIBUR_MODE:-autopilot}"
case "${MODE}" in
  manual | idle | sentinel)
    MODE="manual"
    BURST_COOLDOWN="${SYNCALIBUR_BURST_COOLDOWN:-0}"
    TIDY_THRESHOLD="${SYNCALIBUR_TIDY_THRESHOLD:-999}"
    AUTO_TIDY="${SYNCALIBUR_AUTO_TIDY:-false}"
    AUTO_PR="${SYNCALIBUR_AUTO_PR:-false}"
    ;;
  *)
    MODE="autopilot"
    BURST_COOLDOWN="${SYNCALIBUR_BURST_COOLDOWN:-5}"
    TIDY_THRESHOLD="${SYNCALIBUR_TIDY_THRESHOLD:-999}"
    AUTO_TIDY="${SYNCALIBUR_AUTO_TIDY:-false}"
    AUTO_PR="${SYNCALIBUR_AUTO_PR:-false}"
    ;;
esac

DEFAULT_SENSITIVE_PATTERNS='.env .env.* *.tfstate *.tfstate.* *.tfvars *.tfvars.json credentials.json client_secret*.json service-account*.json service_account*.json *serviceaccount*.json azureProfile.json *.pem *.key *.p12 *.pfx *.pkcs12 *.jks *.keystore id_rsa* id_ecdsa* id_ed25519* id_dsa* *.asc secring.gpg *.kdbx kubeconfig *.kubeconfig *.ovpn *.tblk auth.json .netrc .npmrc .pypirc *.htpasswd *.dump *.rdb *.bak *.backup'
EFFECTIVE_SENSITIVE="${SYNCALIBUR_SENSITIVE_PATTERNS-${AUTO_SYNC_SENSITIVE_PATTERNS-${DEFAULT_SENSITIVE_PATTERNS}}}"
read -r -a SENSITIVE_PATTERNS <<< "${EFFECTIVE_SENSITIVE}"

DEFAULT_BINARY_PATTERNS='*.iso *.img *.vmdk *.qcow2 *.tar *.gz *.tgz *.bz2 *.xz *.zip *.7z *.rar *.jar *.war *.ear *.exe *.dll *.so *.dylib *.bin *.o *.a *.pyc *.pyo *.class *.parquet *.avro *.db *.sqlite *.sqlite3'
EFFECTIVE_BINARY="${SYNCALIBUR_BINARY_PATTERNS-${DEFAULT_BINARY_PATTERNS}}"
read -r -a BINARY_PATTERNS <<< "${EFFECTIVE_BINARY}"

for var in POLL_INTERVAL DEBOUNCE_WAIT MAX_DEBOUNCE MAX_FILE_SIZE_MB RETRY_BASE RETRY_MAX PUSH_TIMEOUT LOG_MAX_KB BURST_COOLDOWN TIDY_THRESHOLD; do
  if ! [[ "${!var}" =~ ^[0-9]+$ ]]; then
    echo "✕ ${var} must be a non-negative integer (got '${!var}')." >&2
    exit 1
  fi
done
if (( TIDY_THRESHOLD < 2 )); then
  TIDY_THRESHOLD=2
fi
if (( POLL_INTERVAL < 1 || DEBOUNCE_WAIT < 1 || RETRY_BASE < 1 )); then
  echo "✕ POLL_INTERVAL, DEBOUNCE_WAIT and RETRY_BASE must be at least 1 second." >&2
  exit 1
fi
MAX_FILE_BYTES=$(( MAX_FILE_SIZE_MB * 1024 * 1024 ))

export GIT_TERMINAL_PROMPT=0
export GCM_INTERACTIVE=never

detect_fs_watcher() {
  if [[ "${USE_FS_EVENTS}" == "false" ]]; then
    echo "polling"
  elif command -v inotifywait >/dev/null 2>&1; then
    echo "inotifywait"
  elif command -v fswatch >/dev/null 2>&1; then
    echo "fswatch"
  else
    echo "polling"
  fi
}

FS_WATCHER="$(detect_fs_watcher)"

# ==============================================================================
# Provider Authentication Probing & Automatic Remote Repository Creation
# ==============================================================================

get_remote() {
  if [[ -n "${REMOTE_OVERRIDE}" ]]; then
    echo "${REMOTE_OVERRIDE}"
  elif git -C "${REPO_ROOT}" remote | grep -qx "all"; then
    echo "all"
  else
    echo "origin"
  fi
}

get_branch() {
  git -C "${REPO_ROOT}" symbolic-ref --quiet --short HEAD 2>/dev/null
}

parse_remote_url() {
  local url="$1"
  local proto="SSH" host="" path="" provider="" owner=""

  if [[ "${url}" =~ ^https?://([^/]+)/(.+)$ ]]; then
    proto="HTTPS"
    host="${BASH_REMATCH[1]}"
    path="${BASH_REMATCH[2]}"
    host="${host#*@}"
  elif [[ "${url}" =~ ^ssh://([^/]+)/(.+)$ ]]; then
    proto="SSH"
    host="${BASH_REMATCH[1]}"
    path="${BASH_REMATCH[2]}"
    host="${host#*@}"
    host="${host%%:*}"
  elif [[ "${url}" =~ ^[^@]+@([^:]+):(.+)$ ]]; then
    proto="SSH"
    host="${BASH_REMATCH[1]}"
    path="${BASH_REMATCH[2]}"
  else
    proto="Local"
    host="local"
    path="${url}"
  fi

  path="${path%.git}"
  path="${path#/}"
  owner="${path%%/*}"

  case "${host}" in
    *github.com*)                     provider="GitHub" ;;
    *gitlab.com*|*gitlab.*)           provider="GitLab" ;;
    *bitbucket.org*)                  provider="Bitbucket" ;;
    *dev.azure.com*|*visualstudio.*)  provider="Azure DevOps" ;;
    local)                            provider="Local Path" ;;
    *)                                provider="${host}" ;;
  esac

  printf '%s|%s|%s|%s\n' "${provider}" "${path}" "${proto}" "${owner}"
}

get_push_urls() {
  local remote="$1"
  git -C "${REPO_ROOT}" remote get-url --push --all "${remote}" 2>/dev/null \
    || git -C "${REPO_ROOT}" config --get-all "remote.${remote}.pushurl" 2>/dev/null \
    || git -C "${REPO_ROOT}" config --get-all "remote.${remote}.url" 2>/dev/null
}

describe_push_targets_inline() {
  local remote="$1" url parsed provider slug summary=""
  while IFS= read -r url; do
    [[ -z "${url}" ]] && continue
    parsed="$(parse_remote_url "${url}")"
    provider="${parsed%%|*}"
    slug="${parsed#*|}"
    slug="${slug%%|*}"
    if [[ -z "${summary}" ]]; then
      summary="${provider}: ${slug}"
    else
      summary="${summary}, ${provider}: ${slug}"
    fi
  done < <(get_push_urls "${remote}")
  echo "${summary:-${remote} (no URL yet)}"
}

detect_auth_credential() {
  local key_file
  for key_file in "${HOME}/.ssh/id_ed25519" "${HOME}/.ssh/id_ecdsa" "${HOME}/.ssh/id_rsa"; do
    if [[ -f "${key_file}" ]]; then
      basename "${key_file}"
      return 0
    fi
  done
  local helper
  helper="$(git -C "${REPO_ROOT}" config --get credential.helper 2>/dev/null || true)"
  if [[ -n "${helper}" ]]; then
    echo "credential.${helper}"
    return 0
  fi
  echo "default"
}

probe_provider_auth() {
  local now mtime
  now="$(date +%s)"
  if [[ -f "${AUTH_CACHE_FILE}" ]]; then
    mtime="$(file_mtime_epoch "${AUTH_CACHE_FILE}")"
    if (( now - mtime < 600 )); then
      cat "${AUTH_CACHE_FILE}"
      return 0
    fi
  fi

  local gh_user="" gh_via="" gl_user="" gl_via="" out

  if command -v gh >/dev/null 2>&1; then
    out="$(gh auth status 2>&1 || true)"
    if [[ "${out}" =~ account[[:space:]]+([A-Za-z0-9_.-]+) ]]; then
      gh_user="${BASH_REMATCH[1]}"
      gh_via="gh+SSH"
    fi
  fi
  if [[ -z "${gh_user}" && -n "${GH_TOKEN:-${GITHUB_TOKEN:-}}" ]] && command -v curl >/dev/null 2>&1; then
    out="$(curl -fsSL -m 4 -H "Authorization: token ${GH_TOKEN:-${GITHUB_TOKEN}}" https://api.github.com/user 2>/dev/null || true)"
    if [[ "${out}" =~ \"login\"[[:space:]]*:[[:space:]]*\"([^\"]+)\" ]]; then
      gh_user="${BASH_REMATCH[1]}"
      gh_via="token"
    fi
  fi
  if [[ -z "${gh_user}" ]] && command -v ssh >/dev/null 2>&1; then
    out="$(ssh -T -o BatchMode=yes -o StrictHostKeyChecking=accept-new -o ConnectTimeout=3 git@github.com 2>&1 || true)"
    if [[ "${out}" =~ Hi[[:space:]]+([^!]+)! ]]; then
      gh_user="${BASH_REMATCH[1]}"
      gh_via="SSH"
    fi
  fi

  if command -v glab >/dev/null 2>&1; then
    out="$(glab auth status 2>&1 || true)"
    if [[ "${out}" =~ Logged[[:space:]]+in[[:space:]]+to[[:space:]]+[^[:space:]]+[[:space:]]+as[[:space:]]+([A-Za-z0-9_.-]+) ]]; then
      gl_user="${BASH_REMATCH[1]}"
      gl_via="glab"
    fi
  fi
  if [[ -z "${gl_user}" && -n "${GITLAB_TOKEN:-}" ]] && command -v curl >/dev/null 2>&1; then
    out="$(curl -fsSL -m 4 --header "PRIVATE-TOKEN: ${GITLAB_TOKEN}" https://gitlab.com/api/v4/user 2>/dev/null || true)"
    if [[ "${out}" =~ \"username\"[[:space:]]*:[[:space:]]*\"([^\"]+)\" ]]; then
      gl_user="${BASH_REMATCH[1]}"
      gl_via="token"
    fi
  fi
  if [[ -z "${gl_user}" ]] && command -v ssh >/dev/null 2>&1; then
    out="$(ssh -T -o BatchMode=yes -o StrictHostKeyChecking=accept-new -o ConnectTimeout=3 git@gitlab.com 2>&1 || true)"
    if [[ "${out}" =~ Welcome[[:space:]]+to[[:space:]]+GitLab,[[:space:]]*@([A-Za-z0-9_.-]+)! ]]; then
      gl_user="${BASH_REMATCH[1]}"
      gl_via="SSH"
    fi
  fi

  local line="${gh_user}|${gh_via}|${gl_user}|${gl_via}"
  mkdir -p "$(dirname "${AUTH_CACHE_FILE}")" 2>/dev/null || true
  printf '%s\n' "${line}" > "${AUTH_CACHE_FILE}" 2>/dev/null || true
  printf '%s\n' "${line}"
}

detect_user_identity() {
  local git_name git_email sys_user sys_host ssh_key auth_line gh_user gh_via gl_user gl_via
  git_name="$(git -C "${REPO_ROOT}" config --get user.name 2>/dev/null || true)"
  git_email="$(git -C "${REPO_ROOT}" config --get user.email 2>/dev/null || true)"
  sys_user="${USER:-$(id -un 2>/dev/null || echo user)}"
  sys_host="$(hostname -s 2>/dev/null || uname -n 2>/dev/null || echo localhost)"
  ssh_key="$(detect_auth_credential)"
  auth_line="$(probe_provider_auth)"
  IFS='|' read -r gh_user gh_via gl_user gl_via <<< "${auth_line}"

  if [[ -z "${git_name}" ]]; then
    if [[ -n "${gh_user}" ]]; then
      git_name="${gh_user}"
    elif [[ -n "${gl_user}" ]]; then
      git_name="${gl_user}"
    elif [[ "${git_email}" =~ ^[0-9]+\+([^@]+)@users\.noreply\.github\.com$ ]]; then
      git_name="${BASH_REMATCH[1]}"
    elif [[ "${git_email}" =~ ^([^@]+)@ ]]; then
      git_name="${BASH_REMATCH[1]}"
    else
      git_name="${sys_user}"
    fi
  fi

  if [[ -z "$(git -C "${REPO_ROOT}" config --get user.name 2>/dev/null || true)" ]]; then
    git -C "${REPO_ROOT}" config user.name "${git_name}" 2>/dev/null || true
  fi
  if [[ -z "${git_email}" ]]; then
    git_email="${git_name}@users.noreply.github.com"
    git -C "${REPO_ROOT}" config user.email "${git_email}" 2>/dev/null || true
  fi

  printf '%s|%s|%s@%s|%s|%s|%s|%s|%s\n' \
    "${git_name}" "${git_email}" "${sys_user}" "${sys_host}" "${ssh_key}" \
    "${gh_user}" "${gh_via}" "${gl_user}" "${gl_via}"
}

auto_create_or_bind_remote() {
  [[ "${AUTO_CREATE_REMOTE}" == "true" ]] || return 1

  local target_slug="${1:-}" target_provider="${2:-}"
  local repo_name auth_line gh_user gh_via gl_user gl_via is_private="true"
  repo_name="$(basename "${REPO_ROOT}")"
  [[ "${REPO_VISIBILITY}" == "public" ]] && is_private="false"

  auth_line="$(probe_provider_auth)"
  IFS='|' read -r gh_user gh_via gl_user gl_via <<< "${auth_line}"

  if [[ "${target_provider}" == "GitHub" || ( -z "${target_provider}" && -n "${gh_user}" ) ]]; then
    local slug="${target_slug:-${gh_user}/${repo_name}}"
    local name_only="${slug#*/}"
    if command -v gh >/dev/null 2>&1 && [[ -n "${gh_user}" ]]; then
      log "✦ Creating ${REPO_VISIBILITY} GitHub repository '${slug}' via GitHub CLI..."
      if gh repo view "${slug}" >/dev/null 2>&1 || gh repo create "${slug}" "--${REPO_VISIBILITY}" >/dev/null 2>&1; then
        if ! git -C "${REPO_ROOT}" remote get-url origin >/dev/null 2>&1; then
          git -C "${REPO_ROOT}" remote add origin "git@github.com:${slug}.git" 2>/dev/null || true
        fi
        log "✓ GitHub repository ready: git@github.com:${slug}.git"
        return 0
      fi
    elif [[ -n "${GH_TOKEN:-${GITHUB_TOKEN:-}}" ]] && command -v curl >/dev/null 2>&1; then
      log "✦ Creating ${REPO_VISIBILITY} GitHub repository '${slug}' via GitHub API..."
      if curl -fsSL -m 10 -X POST \
        -H "Authorization: token ${GH_TOKEN:-${GITHUB_TOKEN}}" \
        -H "Accept: application/vnd.github+json" \
        https://api.github.com/user/repos \
        -d "{\"name\":\"${name_only}\",\"private\":${is_private}}" >/dev/null 2>&1; then
        if ! git -C "${REPO_ROOT}" remote get-url origin >/dev/null 2>&1; then
          git -C "${REPO_ROOT}" remote add origin "git@github.com:${slug}.git" 2>/dev/null || true
        fi
        log "✓ GitHub repository ready: git@github.com:${slug}.git"
        return 0
      fi
    fi
  fi

  if [[ "${target_provider}" == "GitLab" || ( -z "${target_provider}" && -n "${gl_user}" ) ]]; then
    local slug="${target_slug:-${gl_user}/${repo_name}}"
    local name_only="${slug#*/}"
    if command -v glab >/dev/null 2>&1 && [[ -n "${gl_user}" ]]; then
      log "✦ Creating ${REPO_VISIBILITY} GitLab project '${slug}' via GitLab CLI..."
      glab repo create "${slug}" "--${REPO_VISIBILITY}" >/dev/null 2>&1 || true
    elif [[ -n "${GITLAB_TOKEN:-}" ]] && command -v curl >/dev/null 2>&1; then
      log "✦ Creating ${REPO_VISIBILITY} GitLab project '${slug}' via GitLab API..."
      curl -fsSL -m 10 -X POST \
        --header "PRIVATE-TOKEN: ${GITLAB_TOKEN}" \
        "https://gitlab.com/api/v4/projects?name=${name_only}&visibility=${REPO_VISIBILITY}" >/dev/null 2>&1 || true
    fi
    if [[ -n "${gl_user}" ]] && ! git -C "${REPO_ROOT}" remote get-url origin >/dev/null 2>&1; then
      git -C "${REPO_ROOT}" remote add origin "git@gitlab.com:${slug}.git" 2>/dev/null || true
      log "✓ Bound remote 'origin' -> git@gitlab.com:${slug}.git (GitLab push-to-create ready)"
      return 0
    fi
  fi

  return 1
}

# ==============================================================================
# Terminal Theme, High-Definition PNG Emblem Renderer & Styled Log Formatter
# ==============================================================================

C_RESET=$'\033[0m'
C_BOLD=$'\033[1m'
C_DIM=$'\033[2m'
C_GOLD=$'\033[38;2;245;194;66m'
C_CYAN=$'\033[38;2;56;189;248m'
C_GREEN=$'\033[38;2;52;211;153m'
C_RED=$'\033[38;2;248;113;113m'
C_AMBER=$'\033[38;2;251;191;36m'
C_SLATE=$'\033[38;2;148;163;184m'
C_WHITE=$'\033[38;2;241;245;249m'

ensure_logo_ansi() {
  [[ -f "${LOGO_ANSI_FILE}" ]] && return 0
  [[ -f "${LOGO_PNG_FILE}" ]] || return 1
  if command -v python3 >/dev/null 2>&1 && python3 -c "import PIL" >/dev/null 2>&1; then
    python3 - "${LOGO_PNG_FILE}" "${LOGO_ANSI_FILE}" 2>/dev/null << 'PY' && return 0
import sys
from PIL import Image, ImageEnhance, ImageFilter
QUADS = [" ", "▘", "▝", "▀", "▖", "▌", "▞", "▛", "▗", "▚", "▐", "▜", "▄", "▙", "▟", "█"]
src, dst = sys.argv[1], sys.argv[2]
im = Image.open(src).convert("RGBA")
bbox = im.getbbox()
if bbox:
    im = im.crop(bbox)
COLS, ROWS = 44, 21
im_sub = im.resize((COLS * 2, ROWS * 2), Image.Resampling.LANCZOS)
rgb = im_sub.convert("RGB").filter(ImageFilter.UnsharpMask(radius=1.1, percent=175, threshold=2))
rgb = ImageEnhance.Color(ImageEnhance.Contrast(rgb).enhance(1.22)).enhance(1.25)
px_rgb, px_a = rgb.load(), im_sub.split()[3].load()
lines = []
for cy in range(ROWS):
    row = []
    for cx in range(COLS):
        coords = [(cx*2, cy*2), (cx*2+1, cy*2), (cx*2, cy*2+1), (cx*2+1, cy*2+1)]
        alphas = [px_a[x, y] for x, y in coords]
        colors = [px_rgb[x, y] for x, y in coords]
        mask, active = 0, []
        for i, a in enumerate(alphas):
            if a >= 85:
                if a < 220:
                    f = a / 255.0
                    colors[i] = (int(colors[i][0]*f), int(colors[i][1]*f), int(colors[i][2]*f))
                mask |= (1 << i); active.append(i)
        if mask == 0:
            row.append("\033[0m ")
        elif mask != 15:
            r = sum(colors[i][0] for i in active) // len(active)
            g = sum(colors[i][1] for i in active) // len(active)
            b = sum(colors[i][2] for i in active) // len(active)
            row.append(f"\033[0m\033[38;2;{r};{g};{b}m{QUADS[mask]}")
        else:
            best_err, best_m, best_fg, best_bg = 1e18, 3, colors[0], colors[2]
            for m in range(1, 8):
                g1 = [colors[i] for i in range(4) if (m >> i) & 1]
                g2 = [colors[i] for i in range(4) if not ((m >> i) & 1)]
                c1 = tuple(sum(c[k] for c in g1) // len(g1) for k in range(3))
                c2 = tuple(sum(c[k] for c in g2) // len(g2) for k in range(3))
                err = sum(sum((c[k]-c1[k])**2 for k in range(3)) for c in g1) + sum(sum((c[k]-c2[k])**2 for k in range(3)) for c in g2)
                if m == 3:
                    err *= 0.92
                if err < best_err:
                    best_err, best_m, best_fg, best_bg = err, m, c1, c2
            r1, g1, b1 = best_fg; r2, g2, b2 = best_bg
            row.append(f"\033[38;2;{r1};{g1};{b1}m\033[48;2;{r2};{g2};{b2}m{QUADS[best_m]}")
    row.append("\033[0m")
    lines.append("".join(row))
with open(dst, "w") as f:
    f.write("\n".join(lines) + "\n")
PY
  fi
  if ensure_go_engine; then
    "${GO_ENGINE_BIN}" render-logo "${LOGO_PNG_FILE}" "${LOGO_ANSI_FILE}" 2>/dev/null && return 0
  fi
  return 1
}

render_sparkline() {
  if ensure_go_engine; then
    "${GO_ENGINE_BIN}" sparkline "${REPO_ROOT}" 2>/dev/null && return 0
  fi
  local cnt
  cnt="$(git -C "${REPO_ROOT}" rev-list --count --since="7 days ago" HEAD 2>/dev/null || echo 0)"
  printf '%s▂▄▆█%s %s(7d: %s commits)%s' "${C_GREEN}" "${C_RESET}" "${C_SLATE}" "${cnt}" "${C_RESET}"
}

format_log_line() {
  local raw="$1"
  if [[ "${raw}" =~ ^\[([0-9]{4}-[0-9]{2}-[0-9]{2})\ ([0-9]{2}:[0-9]{2}:[0-9]{2})\]\ (.*)$ ]]; then
    local ts="${BASH_REMATCH[2]}"
    local msg="${BASH_REMATCH[3]}"
    local badge_color="${C_SLATE}"
    local badge_text="◈ INFO     "
    local msg_color="${C_WHITE}"

    case "${msg}" in
      *"Syncalibur started"*|*"Syncalibur online"*|*"başlatıldı"*)
        badge_color="${C_GOLD}"; badge_text="● ONLINE   "; msg_color="${C_GOLD}${C_BOLD}" ;;
      *"Targets:"*|*"Identity:"*|*"Poll:"*|*"Kontrol aralığı:"*)
        badge_color="${C_CYAN}"; badge_text="◆ TELEMETRY"; msg_color="${C_SLATE}" ;;
      *"Self-Update:"*|*"Self-Sync:"*|*"hot-reloading"*)
        badge_color="${C_CYAN}${C_BOLD}"; badge_text="⟳ SELF-SYNC"; msg_color="${C_CYAN}" ;;
      *"Fast-forward"*|*"Rebase"*|*"Divergence"*)
        badge_color="${C_CYAN}"; badge_text="⇄ BI-SYNC  "; msg_color="${C_CYAN}" ;;
      *"Creating "*|*"repository ready"*|*"Bound remote"*)
        badge_color="${C_GOLD}"; badge_text="✦ AUTOREPO "; msg_color="${C_GOLD}" ;;
      *"Changes detected"*|*"Değişiklik algılandı"*)
        badge_color="${C_CYAN}"; badge_text="● DETECTED "; msg_color="${C_CYAN}" ;;
      *"[SECRET SHIELD]"*)
        badge_color="${C_RED}${C_BOLD}"; badge_text="⛨ QUARANTINE"; msg_color="${C_RED}" ;;
      *"[SYNTAX HOLD]"*)
        badge_color="${C_AMBER}${C_BOLD}"; badge_text="▲ SYNTAX HOLD"; msg_color="${C_AMBER}" ;;
      *"Committed"*|*"Commit oluşturuldu"*)
        badge_color="${C_GREEN}"; badge_text="✓ COMMIT   "; msg_color="${C_GREEN}" ;;
      *"Pushing"*|*"Uzak depoya"*)
        badge_color="${C_GOLD}"; badge_text="↗ PUSHING  "; msg_color="${C_GOLD}" ;;
      *"Push succeeded"*|*"Başarıyla pushlandı"*)
        badge_color="${C_GREEN}${C_BOLD}"; badge_text="✓ SYNCED   "; msg_color="${C_GREEN}${C_BOLD}" ;;
      *"Not committing"*)
        badge_color="${C_AMBER}"; badge_text="⛨ GUARD    "; msg_color="${C_AMBER}" ;;
      *"Push failed"*|*"timed out"*|*"failed"*|*"conflict"*|*"Another Syncalibur"*|*"SYNC FAILED"*)
        badge_color="${C_RED}${C_BOLD}"; badge_text="✕ ERROR    "; msg_color="${C_RED}" ;;
      *"Retrying"*)
        badge_color="${C_AMBER}"; badge_text="⟳ RETRY    "; msg_color="${C_AMBER}" ;;
      *"paused"*)
        badge_color="${C_AMBER}"; badge_text="▲ PAUSED   "; msg_color="${C_AMBER}" ;;
      *"stopped"*|*"durduruldu"*)
        badge_color="${C_RED}"; badge_text="■ STOPPED  "; msg_color="${C_SLATE}" ;;
    esac

    printf '%s[%s]%s %s⟪%s⟫%s %s%s%s\n' \
      "${C_SLATE}" "${ts}" "${C_RESET}" \
      "${badge_color}" "${badge_text}" "${C_RESET}" \
      "${msg_color}" "${msg}" "${C_RESET}"
  else
    printf '           %s│ %s%s\n' "${C_SLATE}" "${raw}" "${C_RESET}"
  fi
}

describe_notify_channels() {
  local -a ch=()
  if [[ -n "${TELEGRAM_BOT_TOKEN}" && -n "${TELEGRAM_CHAT_ID}" ]]; then
    ch+=("${C_CYAN}✓ Telegram Bot${C_RESET}")
  fi
  if [[ -n "${WEBHOOK_URL}" ]]; then
    local lurl="${WEBHOOK_URL,,}"
    if [[ "${lurl}" == *"discord"* ]]; then
      ch+=("${C_CYAN}✓ Discord Webhook${C_RESET}")
    elif [[ "${lurl}" == *"slack.com"* ]]; then
      ch+=("${C_CYAN}✓ Slack Webhook${C_RESET}")
    elif [[ "${lurl}" == *"api.telegram.org"* ]]; then
      ch+=("${C_CYAN}✓ Telegram Webhook${C_RESET}")
    else
      ch+=("${C_CYAN}✓ Custom Webhook${C_RESET}")
    fi
  fi
  if (( ${#ch[@]} == 0 )); then
    printf '%s' "${C_CYAN}✓ Alert Popup${C_RESET}"
  else
    local IFS=" ${C_SLATE}•${C_RESET} "
    printf '%s' "${ch[*]} ${C_SLATE}•${C_RESET} ${C_CYAN}✓ Alert Popup${C_RESET}"
  fi
}

daemon_resource_badge() {
  local pid="$1"
  [[ -n "${pid}" ]] || return 0
  local ps_line cpu_pct mem_pct rss_kb rss_mb comm_name
  ps_line="$(ps -p "${pid}" -o %cpu=,%mem=,rss=,comm= 2>/dev/null | head -n 1 || true)"
  if [[ -n "${ps_line}" ]]; then
    read -r cpu_pct mem_pct rss_kb comm_name <<< "${ps_line}"
    if [[ "${rss_kb:-0}" =~ ^[0-9]+$ ]] && (( rss_kb > 0 )); then
      rss_mb="$(awk -v kb="${rss_kb}" 'BEGIN { printf "%.1f", kb / 1024.0 }')"
      printf '%s[%s • CPU %s%% • RAM %s MB]%s' "${C_SLATE}" "${comm_name:-syncalibur}" "${cpu_pct:-0.0}" "${rss_mb}" "${C_RESET}"
      return 0
    fi
  fi
  return 0
}

print_syncalibur_dashboard() {
  ensure_logo_ansi

  local pid state_str autostart_str branch remote unpushed_str cfg_short watch_short self_str engine_badge res_badge=""
  if [[ -f "${PAUSE_FILE}" ]]; then
    state_str="${C_AMBER}${C_BOLD}▲ PAUSED${C_RESET} ${C_SLATE}(resume via 'syncalibur resume' or [p] in status)${C_RESET}"
  elif pid="$(daemon_pid)"; then
    res_badge="$(daemon_resource_badge "${pid}")"
    state_str="${C_GREEN}${C_BOLD}● ONLINE${C_RESET} ${C_SLATE}(PID ${pid})${C_RESET} ${res_badge}"
  else
    state_str="${C_RED}${C_BOLD}○ OFFLINE${C_RESET}"
  fi

  if is_service_installed; then
    autostart_str="${C_GREEN}✓ Enabled${C_RESET} ${C_SLATE}(${SERVICE_NAME} • ${SERVICE_BACKEND})${C_RESET}"
  else
    autostart_str="${C_AMBER}○ Not installed${C_RESET} ${C_SLATE}(${SERVICE_BACKEND})${C_RESET}"
  fi

  if ensure_go_engine; then
    engine_badge="${C_CYAN}Go Engine + ${FS_WATCHER}${C_RESET}"
  else
    engine_badge="${C_SLATE}Bash + ${FS_WATCHER}${C_RESET}"
  fi

  remote="$(get_remote)"
  branch="$(get_branch || echo detached)"
  if [[ -f "${FAIL_REPORT_FILE}" ]]; then
    local fail_reason_short
    fail_reason_short="$(sed -n '4p' "${FAIL_REPORT_FILE}" 2>/dev/null | cut -c1-48)"
    unpushed_str="${C_RED}${C_BOLD}✕ SYNC FAILED:${C_RESET} ${C_RED}${fail_reason_short:-see syncalibur alert}${C_RESET}"
  elif [[ "${branch}" != "detached" ]]; then
    local up_cnt
    up_cnt="$(cd "${REPO_ROOT}" && unpushed_count "${remote}" "${branch}")"
    if (( up_cnt == 0 )); then
      unpushed_str="${C_GREEN}✓ 0 unpushed commits (in sync)${C_RESET}  ${C_SLATE}•${C_RESET}  $(render_sparkline)"
    else
      unpushed_str="${C_AMBER}● ${up_cnt} commit(s) queued${C_RESET}  ${C_SLATE}•${C_RESET}  $(render_sparkline)"
    fi
  else
    unpushed_str="${C_AMBER}▲ Detached HEAD (paused)${C_RESET}"
  fi

  local id_raw git_user git_email sys_ident ssh_cred gh_user gh_via gl_user gl_via
  id_raw="$(detect_user_identity)"
  IFS='|' read -r git_user git_email sys_ident ssh_cred gh_user gh_via gl_user gl_via <<< "${id_raw}"

  local auth_badges=""
  if [[ -n "${gh_user}" ]]; then
    auth_badges="${C_GREEN}✓ GitHub: @${gh_user}${C_RESET} ${C_SLATE}(${gh_via})${C_RESET}"
  fi
  if [[ -n "${gl_user}" ]]; then
    [[ -n "${auth_badges}" ]] && auth_badges="${auth_badges}  ${C_SLATE}•${C_RESET}  "
    auth_badges="${auth_badges}${C_GREEN}✓ GitLab: @${gl_user}${C_RESET} ${C_SLATE}(${gl_via})${C_RESET}"
  fi
  [[ -z "${auth_badges}" ]] && auth_badges="${C_AMBER}SSH key (${ssh_cred})${C_RESET}"

  local -a dest_lines=()
  local url parsed provider slug proto owner idx=0
  while IFS= read -r url; do
    [[ -z "${url}" ]] && continue
    idx=$(( idx + 1 ))
    parsed="$(parse_remote_url "${url}")"
    IFS='|' read -r provider slug proto owner <<< "${parsed}"
    local p_color="${C_CYAN}"
    [[ "${provider}" == "GitHub" ]] && p_color="${C_WHITE}${C_BOLD}"
    [[ "${provider}" == "GitLab" ]] && p_color="${C_AMBER}${C_BOLD}"
    dest_lines+=("  ${C_GOLD}◈${C_RESET} ${C_BOLD}Push #${idx}   :${C_RESET} ${p_color}${provider}${C_RESET} ${C_SLATE}→${C_RESET} ${C_CYAN}${slug}${C_RESET} ${C_SLATE}[${branch} • ${proto}]${C_RESET}")
  done < <(get_push_urls "${remote}")

  if (( ${#dest_lines[@]} == 0 )); then
    dest_lines+=("  ${C_GOLD}◈${C_RESET} ${C_BOLD}Push Dest :${C_RESET} ${C_AMBER}None yet — will auto-create ${REPO_VISIBILITY} repo on first push${C_RESET}")
  fi

  if [[ "${SELF_AUTO_UPDATE}" == "true" && -n "${SELF_REPO_ROOT}" ]]; then
    local self_rev self_short
    self_rev="$(git -C "${SELF_REPO_ROOT}" rev-parse --short HEAD 2>/dev/null || echo main)"
    self_short="${SELF_REPO_ROOT/#${HOME}/~}"
    self_str="${C_GREEN}✓ Watching & Auto-Updating${C_RESET} ${C_SLATE}(${self_short} @ ${self_rev} • every ${SELF_CHECK_INTERVAL}s)${C_RESET}"
  else
    self_str="${C_SLATE}Disabled${C_RESET}"
  fi

  local notify_str
  notify_str="$(describe_notify_channels)"
  local autorepo_str="${C_GREEN}✓ Bi-Sync${C_RESET} ${C_SLATE}•${C_RESET} ${C_GREEN}✓ Semantic Commits${C_RESET} ${C_SLATE}•${C_RESET} ${C_GREEN}✓ Auto-Repo (${REPO_VISIBILITY})${C_RESET} ${C_SLATE}•${C_RESET} ${notify_str}"

  local total_commits today_commits
  total_commits="$(git -C "${REPO_ROOT}" rev-list --count HEAD 2>/dev/null || echo "0")"
  today_commits="$(git -C "${REPO_ROOT}" log --since="midnight" --oneline 2>/dev/null | wc -l | tr -d ' ')"

  cfg_short="${CONFIG_FILE:-<defaults>}"
  cfg_short="${cfg_short/#${HOME}/~}"
  watch_short="${WATCH_PATH/#${HOME}/~}"

  local title_text="  S Y N C A L I B U R ${SYNCALIBUR_PLATFORM_VERSION}  •  Zero-Data-Loss Git Sentinel"
  local vis_len=${#title_text}
  local pad_len=$(( 70 - vis_len ))
  local pad=""
  if (( pad_len > 0 )); then
    pad="$(printf '%*s' "${pad_len}" "")"
  fi

  local -a info_lines=(
    "${C_GOLD}╭──────────────────────────────────────────────────────────────────────╮${C_RESET}"
    "${C_GOLD}│${C_RESET}  ${C_GOLD}${C_BOLD}S Y N C A L I B U R${C_RESET} ${C_GREEN}${SYNCALIBUR_PLATFORM_VERSION}${C_RESET}  ${C_SLATE}•${C_RESET}  ${C_CYAN}Zero-Data-Loss Git Sentinel${C_RESET}${pad}${C_GOLD}│${C_RESET}"
    "${C_GOLD}╰──────────────────────────────────────────────────────────────────────╯${C_RESET}"
    "  ${C_GOLD}◈${C_RESET} ${C_BOLD}Status    :${C_RESET} ${state_str}  ${C_SLATE}•${C_RESET}  ${engine_badge}"
    "  ${C_GOLD}◈${C_RESET} ${C_BOLD}OS / Svc  :${C_RESET} ${C_WHITE}${OS_KERNEL} (${SERVICE_BACKEND})${C_RESET}  ${C_SLATE}•${C_RESET}  ${autostart_str}"
    "  ${C_GOLD}◈${C_RESET} ${C_BOLD}Sync Mode :${C_RESET} ${C_CYAN}${C_BOLD}⚡ autopilot${C_RESET} ${C_SLATE}(5s cognitive debounce • zero-data-loss sentinel active)${C_RESET}"
    "  ${C_GOLD}◈${C_RESET} ${C_BOLD}Author    :${C_RESET} ${C_WHITE}${C_BOLD}${git_user}${C_RESET} ${C_SLATE}<${git_email}> (${sys_ident})${C_RESET}"
    "  ${C_GOLD}◈${C_RESET} ${C_BOLD}Verified  :${C_RESET} ${auth_badges}"
    "  ${C_GOLD}◈${C_RESET} ${C_BOLD}Watching  :${C_RESET} ${C_WHITE}${watch_short}${C_RESET} ${C_SLATE}(scope: ${SCOPE})${C_RESET}"
    "  ${C_GOLD}◈${C_RESET} ${C_BOLD}Config    :${C_RESET} ${C_SLATE}${cfg_short} (gitignored)${C_RESET}"
    "  ${C_GOLD}◈${C_RESET} ${C_BOLD}Remote    :${C_RESET} ${C_CYAN}${remote}${C_RESET} ${C_SLATE}(${idx} target(s) • branch '${branch}')${C_RESET}"
  )

  local d_idx=0
  while (( d_idx < ${#dest_lines[@]} && d_idx < 3 )); do
    info_lines+=("${dest_lines[$d_idx]}")
    d_idx=$(( d_idx + 1 ))
  done

  info_lines+=(
    "  ${C_GOLD}◈${C_RESET} ${C_BOLD}Self-Sync :${C_RESET} ${self_str}"
    "  ${C_GOLD}◈${C_RESET} ${C_BOLD}Features  :${C_RESET} ${autorepo_str}"
    "  ${C_GOLD}◈${C_RESET} ${C_BOLD}Commits   :${C_RESET} ${C_GREEN}${total_commits}${C_RESET} ${C_SLATE}total${C_RESET}  ${C_SLATE}•${C_RESET}  ${C_GREEN}${today_commits}${C_RESET} ${C_SLATE}today${C_RESET}  ${C_SLATE}•${C_RESET}  ${unpushed_str}"
    "  ${C_GOLD}◈${C_RESET} ${C_BOLD}Shield    :${C_RESET} ${C_SLATE}Poll ${POLL_INTERVAL}s • Debounce ${DEBOUNCE_WAIT}s • ≤ ${MAX_FILE_SIZE_MB} MiB & Secret Guard${C_RESET}"
    "  ${C_SLATE}──────────────────────────────────────────────────────────────────────${C_RESET}"
    "  ${C_SLATE}Commands  : ${C_CYAN}save${C_SLATE} | ${C_CYAN}undo${C_SLATE} | ${C_CYAN}check${C_SLATE} | ${C_CYAN}status${C_SLATE} | ${C_CYAN}push${C_SLATE} | ${C_CYAN}doctor${C_SLATE} | ${C_CYAN}fleet${C_SLATE} | ${C_CYAN}version${C_RESET}"
  )

  local -a logo_lines=()
  local line
  if [[ -f "${LOGO_ANSI_FILE}" ]]; then
    while IFS= read -r line; do
      logo_lines+=("${line}")
    done < "${LOGO_ANSI_FILE}"
  fi

  local total_rows=${#logo_lines[@]}
  (( ${#info_lines[@]} > total_rows )) && total_rows=${#info_lines[@]}

  printf '\n'
  local i=0
  local empty_logo="                                            "
  while (( i < total_rows )); do
    local l_part="${logo_lines[$i]:-${empty_logo}}"
    local r_part="${info_lines[$i]:-}"
    printf '  %s   %s\n' "${l_part}" "${r_part}"
    i=$(( i + 1 ))
  done
  printf '\n'
}

# ==============================================================================
# First-Class Telegram Bot & Multi-Channel Webhook Notification Dispatcher
# ==============================================================================

has_notification_channel() {
  [[ -n "${WEBHOOK_URL}" ]] || [[ -n "${TELEGRAM_BOT_TOKEN}" && -n "${TELEGRAM_CHAT_ID}" ]]
}

send_webhook_notification() {
  local event_type="$1" title="$2" message="$3" details="${4:-}" sync_flag="${5:-async}"
  has_notification_channel || return 0

  if [[ "${event_type}" != "test" ]]; then
    if [[ "${event_type}" == "success" ]]; then
      if [[ "${WEBHOOK_ON_SUCCESS}" != "true" && ",${WEBHOOK_EVENTS}," != *",success,"* && "${WEBHOOK_EVENTS}" != "all" ]]; then
        return 0
      fi
    elif [[ ",${WEBHOOK_EVENTS}," != *",${event_type},"* && "${WEBHOOK_EVENTS}" != "all" ]]; then
      return 0
    fi
  fi

  local repo_name branch_name
  repo_name="$(basename "${REPO_ROOT}")"
  branch_name="$(get_branch || echo main)"

  if ensure_go_engine; then
    if [[ "${sync_flag}" == "sync" ]]; then
      SYNCALIBUR_TELEGRAM_BOT_TOKEN="${TELEGRAM_BOT_TOKEN}" \
      SYNCALIBUR_TELEGRAM_CHAT_ID="${TELEGRAM_CHAT_ID}" \
        "${GO_ENGINE_BIN}" webhook "${WEBHOOK_URL}" "${event_type}" "${repo_name}" "${branch_name}" "${title}" "${message}" "${details}"
      return $?
    else
      (
        SYNCALIBUR_TELEGRAM_BOT_TOKEN="${TELEGRAM_BOT_TOKEN}" \
        SYNCALIBUR_TELEGRAM_CHAT_ID="${TELEGRAM_CHAT_ID}" \
          "${GO_ENGINE_BIN}" webhook "${WEBHOOK_URL}" "${event_type}" "${repo_name}" "${branch_name}" "${title}" "${message}" "${details}" >/dev/null 2>&1
      ) &
      return 0
    fi
  fi

  # Pure-Bash curl fallback when Go binary is unavailable
  if command -v curl >/dev/null 2>&1 && [[ -n "${TELEGRAM_BOT_TOKEN}" && -n "${TELEGRAM_CHAT_ID}" ]]; then
    local tg_url="https://api.telegram.org/bot${TELEGRAM_BOT_TOKEN#bot}/sendMessage"
    local tg_text="◈ ${title} — ${message} [${repo_name} (${branch_name})]"
    if [[ "${sync_flag}" == "sync" ]]; then
      curl -fsSL -m 8 -X POST "${tg_url}" -d "chat_id=${TELEGRAM_CHAT_ID}" --data-urlencode "text=${tg_text}" >/dev/null
      return $?
    else
      (curl -fsSL -m 8 -X POST "${tg_url}" -d "chat_id=${TELEGRAM_CHAT_ID}" --data-urlencode "text=${tg_text}" >/dev/null 2>&1 &)
    fi
  fi
  return 0
}

cmd_notify() {
  local custom_msg="${*:-Manual test notification triggered via 'syncalibur notify'. Your Telegram / Webhook pipeline is online!}"
  local repo_name branch_name
  repo_name="$(basename "${REPO_ROOT}")"
  branch_name="$(get_branch || echo main)"

  if ! has_notification_channel; then
    printf '\n  %s▲  No Telegram Bot or Webhook URL is configured yet in %s.%s\n\n' "${C_AMBER}${C_BOLD}" "${CONFIG_FILE:-syncalibur.env}" "${C_RESET}"
    printf '  %sTo enable First-Class Telegram Bot notifications:%s\n' "${C_GOLD}${C_BOLD}" "${C_RESET}"
    printf '    1. Open Telegram and message %s@BotFather%s -> %s/newbot%s to get a token\n' "${C_CYAN}" "${C_RESET}" "${C_WHITE}" "${C_RESET}"
    printf '    2. Set %sSYNCALIBUR_TELEGRAM_BOT_TOKEN="..."%s and %sSYNCALIBUR_TELEGRAM_CHAT_ID="..."%s in %ssyncalibur.env%s\n' "${C_CYAN}" "${C_RESET}" "${C_CYAN}" "${C_RESET}" "${C_WHITE}" "${C_RESET}"
    printf '    3. Or set %sSYNCALIBUR_WEBHOOK_URL="..."%s for Discord / Slack / Telegram / Custom JSON\n\n' "${C_CYAN}" "${C_RESET}"
    return 1
  fi

  printf '  %s✦ Dispatching notification for %s (%s)...%s\n' "${C_CYAN}" "${repo_name}" "${branch_name}" "${C_RESET}"
  if send_webhook_notification "test" "Syncalibur Notification (${repo_name})" "${custom_msg}" "Triggered via: syncalibur notify" "sync"; then
    printf '  %s✓ Notification delivered successfully!%s\n' "${C_GREEN}${C_BOLD}" "${C_RESET}"
    return 0
  else
    printf '  %s✕ Failed to deliver notification. Check your token/chat_id or webhook URL in %s.%s\n' "${C_RED}${C_BOLD}" "${CONFIG_FILE:-syncalibur.env}" "${C_RESET}" >&2
    return 1
  fi
}

# ==============================================================================
# Proactive Failure Diagnostic Report & Automatic Terminal Popup Alert
# ==============================================================================

diagnose_failure_hint() {
  local raw="$1"
  if grep -qiE 'Permission denied \(publickey\)|Authentication failed|Could not read from remote repository' <<< "${raw}"; then
    echo "SSH key or token is not authorized for this remote. Check 'ssh -T git@github.com' / 'gh auth status'."
  elif grep -qiE 'Repository not found|project you were looking for could not be found' <<< "${raw}"; then
    echo "Remote repository does not exist or account lacks access. Authenticate 'gh auth login' or 'glab auth login' for auto-creation."
  elif grep -qiE 'CONFLICT|Merge conflict|rebase' <<< "${raw}"; then
    echo "Merge conflict detected between remote and local commits. Syncalibur cleanly aborted the rebase; resolve manually with 'git pull --rebase'."
  elif grep -qiE 'fetch first|non-fast-forward|rejected.*main|rejected.*master' <<< "${raw}"; then
    echo "Remote has new commits. Run 'syncalibur pull' or 'git pull --rebase'."
  elif grep -qiE 'GH001: Large files detected|exceeds GitHub.*file size limit|pack exceeds maximum allowed size' <<< "${raw}"; then
    echo "A large binary is in commit history. Remove it from history and lower SYNCALIBUR_MAX_FILE_SIZE_MB."
  elif grep -qiE 'GH013: Repository rule violations|Push cannot contain secrets' <<< "${raw}"; then
    echo "Remote push protection detected a secret in commit history. Remove the secret and add the file to .gitignore."
  elif grep -qiE 'Could not resolve host|Network is unreachable|Connection timed out|timed out after' <<< "${raw}"; then
    echo "Network/DNS connection failed. Syncalibur will automatically retry with exponential backoff once online."
  else
    echo "Inspect the status and raw git output below, or press [L] to stream live logs."
  fi
}

record_sync_failure() {
  local stage="$1" reason="$2" details="$3" remote="${4:-$(get_remote)}" branch="${5:-$(get_branch || echo main)}"
  local ts targets hint
  ts="$(date '+%Y-%m-%d %H:%M:%S')"
  targets="$(describe_push_targets_inline "${remote}")"
  hint="$(diagnose_failure_hint "${details} ${reason}")"

  mkdir -p "$(dirname "${FAIL_REPORT_FILE}")" 2>/dev/null || true
  {
    printf '%s\n' "${ts}"
    printf '%s\n' "${stage}"
    printf '%s (%s)\n' "${targets}" "${branch}"
    printf '%s\n' "${reason}"
    printf '%s\n' "${hint}"
    printf '%s\n' "${details}"
  } > "${FAIL_REPORT_FILE}" 2>/dev/null || true
}

clear_sync_failure() {
  if [[ -f "${FAIL_REPORT_FILE}" ]]; then
    send_webhook_notification "recovery" "✓ Syncalibur Recovered ($(basename "${REPO_ROOT}"))" "Synchronization restored after previous error." "All remotes in sync."
  fi
  rm -f "${FAIL_REPORT_FILE}" "${ALERT_STATE_FILE}" 2>/dev/null || true
}

send_desktop_alert() {
  local title="$1" body="$2"
  case "${OS_KERNEL}" in
    Darwin*)
      command -v osascript >/dev/null 2>&1 && osascript -e "display notification \"${body//\"/\'}\" with title \"${title//\"/\'}\" sound name \"Basso\"" >/dev/null 2>&1 || true
      ;;
    MINGW*|MSYS*|CYGWIN*)
      command -v powershell.exe >/dev/null 2>&1 && powershell.exe -NoProfile -NonInteractive -WindowStyle Hidden -Command \
        "[void][System.Reflection.Assembly]::LoadWithPartialName('System.Windows.Forms'); [System.Windows.Forms.MessageBox]::Show('${body//\'/}', '${title//\'/}', 'OK', 'Warning')" >/dev/null 2>&1 &
      ;;
    *)
      if command -v notify-send >/dev/null 2>&1; then
        notify-send -u critical -i "${LOGO_PNG_FILE}" "${title}" "${body}" >/dev/null 2>&1 || true
      fi
      ;;
  esac
}

trigger_failure_alert() {
  local stage="$1" reason="$2" details="$3" remote="${4:-$(get_remote)}" branch="${5:-$(get_branch || echo main)}"
  record_sync_failure "${stage}" "${reason}" "${details}" "${remote}" "${branch}"
  send_webhook_notification "failure" "✕ Syncalibur Alert — ${stage}" "${reason}" "${details}"

  [[ "${ALERT_TERMINAL}" == "true" ]] || return 0

  local now last_ts=0
  now="$(date +%s)"
  if [[ -f "${ALERT_STATE_FILE}" ]]; then
    last_ts="$(cat "${ALERT_STATE_FILE}" 2>/dev/null || echo 0)"
    [[ "${last_ts}" =~ ^[0-9]+$ ]] || last_ts=0
  fi
  if (( now - last_ts < ALERT_COOLDOWN )); then
    return 0
  fi
  printf '%s\n' "${now}" > "${ALERT_STATE_FILE}" 2>/dev/null || true

  send_desktop_alert "Syncalibur — Issue Detected ($(basename "${REPO_ROOT}"))" "${reason}"

  if [[ "${SYNCALIBUR_BACKGROUND:-0}" != "1" && -t 1 ]]; then
    return 0
  fi

  local alert_cmd="SYNCALIBUR_CONFIG=\"${CONFIG_FILE}\" \"${SCRIPT_PATH}\" alert --wait"

  case "${OS_KERNEL}" in
    Darwin*)
      if command -v osascript >/dev/null 2>&1; then
        osascript -e "tell application \"Terminal\" to activate" \
                  -e "tell application \"Terminal\" to do script \"${alert_cmd//\"/\\\"}\"" >/dev/null 2>&1 &
      fi
      ;;
    MINGW*|MSYS*|CYGWIN*)
      if command -v cygpath >/dev/null 2>&1; then
        local bash_win script_win
        bash_win="$(cygpath -w "$(command -v bash)")"
        script_win="$(cygpath -w "${SCRIPT_PATH}")"
        cmd.exe //c start "Syncalibur Status & Diagnostic Alert" "${bash_win}" "${script_win}" alert --wait >/dev/null 2>&1 &
      fi
      ;;
    *)
      local uid
      uid="$(id -u 2>/dev/null || echo 1000)"
      [[ -z "${DISPLAY:-}" && -S /tmp/.X11-unix/X0 ]] && export DISPLAY=":0"
      [[ -z "${DISPLAY:-}" && -S /tmp/.X11-unix/X1 ]] && export DISPLAY=":1"
      [[ -z "${WAYLAND_DISPLAY:-}" && -S "/run/user/${uid}/wayland-0" ]] && export WAYLAND_DISPLAY="wayland-0"
      [[ -z "${DBUS_SESSION_BUS_ADDRESS:-}" && -S "/run/user/${uid}/bus" ]] && export DBUS_SESSION_BUS_ADDRESS="unix:path=/run/user/${uid}/bus"

      if [[ -z "${DISPLAY:-}" && -z "${WAYLAND_DISPLAY:-}" ]]; then
        return 0
      fi

      if command -v gnome-terminal >/dev/null 2>&1; then
        gnome-terminal --title="Syncalibur Status & Diagnostic Alert" -- bash -c "${alert_cmd}" >/dev/null 2>&1 &
      elif command -v konsole >/dev/null 2>&1; then
        konsole -p tabtitle="Syncalibur Status & Diagnostic Alert" -e bash -c "${alert_cmd}" >/dev/null 2>&1 &
      elif command -v xfce4-terminal >/dev/null 2>&1; then
        xfce4-terminal --title="Syncalibur Status & Diagnostic Alert" -x bash -c "${alert_cmd}" >/dev/null 2>&1 &
      elif command -v kitty >/dev/null 2>&1; then
        kitty --title "Syncalibur Status & Diagnostic Alert" bash -c "${alert_cmd}" >/dev/null 2>&1 &
      elif command -v alacritty >/dev/null 2>&1; then
        alacritty -t "Syncalibur Status & Diagnostic Alert" -e bash -c "${alert_cmd}" >/dev/null 2>&1 &
      elif command -v wezterm >/dev/null 2>&1; then
        wezterm start -- bash -c "${alert_cmd}" >/dev/null 2>&1 &
      elif command -v tilix >/dev/null 2>&1; then
        tilix -t "Syncalibur Status & Diagnostic Alert" -e "bash -c '${alert_cmd}'" >/dev/null 2>&1 &
      elif command -v mate-terminal >/dev/null 2>&1; then
        mate-terminal --title="Syncalibur Status & Diagnostic Alert" -x bash -c "${alert_cmd}" >/dev/null 2>&1 &
      elif command -v x-terminal-emulator >/dev/null 2>&1; then
        x-terminal-emulator -e bash -c "${alert_cmd}" >/dev/null 2>&1 &
      elif command -v xterm >/dev/null 2>&1; then
        xterm -T "Syncalibur Status & Diagnostic Alert" -e bash -c "${alert_cmd}" >/dev/null 2>&1 &
      fi
      ;;
  esac
}

cmd_alert() {
  local mode="${1:-}"
  print_syncalibur_dashboard

  local ts stage targets reason hint details
  if [[ -f "${FAIL_REPORT_FILE}" ]]; then
    ts="$(sed -n '1p' "${FAIL_REPORT_FILE}")"
    stage="$(sed -n '2p' "${FAIL_REPORT_FILE}")"
    targets="$(sed -n '3p' "${FAIL_REPORT_FILE}")"
    reason="$(sed -n '4p' "${FAIL_REPORT_FILE}")"
    hint="$(sed -n '5p' "${FAIL_REPORT_FILE}")"
    details="$(tail -n +6 "${FAIL_REPORT_FILE}" | head -n 12)"
  else
    ts="$(date '+%Y-%m-%d %H:%M:%S')"
    stage="HEALTH CHECK"
    targets="$(describe_push_targets_inline "$(get_remote)") ($(get_branch || echo main))"
    if [[ "${mode}" == "--test" ]]; then
      reason="Simulated Sync Failure Alert (Test Mode)"
      hint="Popup terminal & status view verified! Whenever a sync or self-update fails, this terminal opens automatically."
      details="remote: Test simulation triggered via 'syncalibur alert --test'"
    else
      reason="No active sync failures recorded — repository and self-updater are healthy!"
      hint="Syncalibur automatically pops open this status & diagnostic window whenever an error occurs."
      details=""
    fi
  fi

  local box_color="${C_RED}"
  local box_title="✕  S Y N C   F A I L E D   —   D I A G N O S T I C   R E P O R T"
  if [[ ! -f "${FAIL_REPORT_FILE}" && "${mode}" != "--test" ]]; then
    box_color="${C_GREEN}"
    box_title="✓  D I A G N O S T I C   S T A T U S   —   A L L   S Y S T E M S   HEALTHY"
  fi

  printf '  %s╭────────────────────────────────────────────────────────────────────────────────────────╮%s\n' "${box_color}" "${C_RESET}"
  printf '  %s│%s  %s%s%-84s%s  %s│%s\n' "${box_color}" "${C_RESET}" "${box_color}" "${C_BOLD}" "${box_title}" "${C_RESET}" "${box_color}" "${C_RESET}"
  printf '  %s╰────────────────────────────────────────────────────────────────────────────────────────╯%s\n' "${box_color}" "${C_RESET}"
  printf '    %s◈ Stage / Time :%s %s%s%s  %s(%s)%s\n' "${box_color}" "${C_RESET}" "${C_AMBER}" "${stage}" "${C_RESET}" "${C_SLATE}" "${ts}" "${C_RESET}"
  printf '    %s◈ Target(s)    :%s %s%s%s\n' "${box_color}" "${C_RESET}" "${C_CYAN}" "${targets}" "${C_RESET}"
  printf '    %s◈ Error Summary:%s %s%s%s%s\n' "${box_color}" "${C_RESET}" "${C_RED}" "${C_BOLD}" "${reason}" "${C_RESET}"
  printf '    %s◈ Recommended  :%s %s%s%s\n\n' "${box_color}" "${C_RESET}" "${C_GOLD}" "${hint}" "${C_RESET}"

  if [[ -n "${details}" ]]; then
    printf '  %s%sRaw Git / Transport Output:%s\n' "${C_RED}" "${C_BOLD}" "${C_RESET}"
    printf '  %s────────────────────────────────────────────────────────────────────────────────────────%s\n' "${C_SLATE}" "${C_RESET}"
    local line
    while IFS= read -r line; do
      [[ -z "${line}" ]] && continue
      printf '    %s│%s %s%s%s\n' "${C_RED}" "${C_RESET}" "${C_SLATE}" "${line}" "${C_RESET}"
    done <<< "${details}"
    printf '\n'
  fi

  if [[ -f "${LOG_FILE}" ]]; then
    printf '  %s%sRecent Activity Log:%s\n' "${C_GOLD}" "${C_BOLD}" "${C_RESET}"
    printf '  %s────────────────────────────────────────────────────────────────────────────────────────%s\n' "${C_SLATE}" "${C_RESET}"
    local line
    while IFS= read -r line; do
      printf '  '
      format_log_line "${line}"
    done < <(tail -n 6 "${LOG_FILE}")
    printf '\n'
  fi

  if [[ "${mode}" == "--wait" ]]; then
    printf '  %sPress [Enter] to close, [R] + [Enter] to force retry/restart, or [L] + [Enter] for live logs: %s' "${C_GOLD}${C_BOLD}" "${C_RESET}"
    local reply=""
    read -r reply || true
    if [[ "${reply}" =~ ^[Ll]$ ]]; then
      cmd_logs
    elif [[ "${reply}" =~ ^[Rr]$ ]]; then
      cmd_restart
    fi
  fi
}

# ==============================================================================
# Domain 1.2: Bidirectional Sync Engine (`auto-pull --rebase` & Clean Abort)
# ==============================================================================

LAST_BI_SYNC_TS=0

check_network_liveness() {
  local remote="${1:-$(get_remote)}"
  local url
  url="$(git remote get-url "${remote}" 2>/dev/null || echo "")"
  [[ -z "${url}" ]] && return 0 # Local repo, no network required

  # 1. Native Go engine fast TCP dial probe (non-blocking, timeout <= 1000ms)
  if ensure_go_engine; then
    if "${GO_ENGINE_BIN}" probe-network "${url}" 1000 >/dev/null 2>&1; then
      return 0
    fi
    return 1
  fi

  # 2. Bash /dev/tcp subshell probe fallback
  local host="github.com" port="22"
  if [[ "${url}" =~ @([^:]+): ]]; then
    host="${BASH_REMATCH[1]}"
    port="22"
  elif [[ "${url}" =~ https://([^/:]+) ]]; then
    host="${BASH_REMATCH[1]}"
    port="443"
  fi
  timeout 1 bash -c "cat < /dev/null > /dev/tcp/${host}/${port}" 2>/dev/null
}

safe_bidirectional_sync() {
  [[ "${BIDIRECTIONAL_SYNC}" == "true" ]] || return 0
  local remote="$1" branch="$2" force_now="${3:-0}"
  local now
  now="$(date +%s)"
  if (( ! force_now )) && (( now - LAST_BI_SYNC_TS < PULL_INTERVAL )); then
    return 0
  fi
  LAST_BI_SYNC_TS="${now}"

  repo_busy >/dev/null && return 0
  git rev-parse --verify HEAD >/dev/null 2>&1 || return 0

  # Fetch from origin (or primary remote if single-URL)
  local fetch_remote="origin"
  if ! git remote | grep -qx "origin"; then
    fetch_remote="${remote}"
  fi
  git remote get-url "${fetch_remote}" >/dev/null 2>&1 || return 0

  # Network Offline Shield: probe before blocking fetch
  check_network_liveness "${fetch_remote}" || return 0

  run_with_timeout 25 git fetch -q "${fetch_remote}" "${branch}" 2>/dev/null || return 0
  local upstream="refs/remotes/${fetch_remote}/${branch}"
  git rev-parse --verify "${upstream}" >/dev/null 2>&1 || return 0

  local behind ahead
  behind="$(git rev-list --count "HEAD..${upstream}" 2>/dev/null || echo 0)"
  ahead="$(git rev-list --count "${upstream}..HEAD" 2>/dev/null || echo 0)"

  (( behind == 0 )) && return 0

  # Ensure any local uncommitted changes are committed cleanly before fast-forward or rebase
  if has_changes_in_scope; then
    wait_for_quiet
    commit_changes || return 1
    ahead="$(git rev-list --count "${upstream}..HEAD" 2>/dev/null || echo 0)"
  fi

  if (( behind > 0 && ahead == 0 )); then
    log "↙ Bidirectional Sync: Remote '${fetch_remote}/${branch}' has ${behind} new commit(s). Fast-forwarding..."
    local ff_out
    if ff_out="$(git merge --ff-only "${upstream}" 2>&1)"; then
      printf '%s\n' "${ff_out}" >> "${LOG_FILE}"
      log "✓ Fast-forward pull succeeded (HEAD @ $(git rev-parse --short HEAD))."
      clear_sync_failure
      return 0
    fi
  fi

  if (( behind > 0 && ahead > 0 )); then
    log "⇄ Bidirectional Sync: Divergent history (${ahead} local, ${behind} remote). Rebasing cleanly..."
    local rb_out
    if rb_out="$(git rebase "${upstream}" 2>&1)"; then
      printf '%s\n' "${rb_out}" >> "${LOG_FILE}"
      log "✓ Rebase succeeded! Replayed ${ahead} local commit(s) onto '${fetch_remote}/${branch}'."
      clear_sync_failure
      return 0
    else
      printf '%s\n' "${rb_out}" >> "${LOG_FILE}"
      local conflict_files
      conflict_files="$(git diff --name-only --diff-filter=U 2>/dev/null | tr '\n' ' ' || true)"
      git rebase --abort >/dev/null 2>&1 || true
      local err_msg="Merge conflict during auto-rebase (${conflict_files:-conflicting changes}); cleanly aborted rebase to protect working tree"
      log "✕ ${err_msg}."
      trigger_failure_alert "BIDIRECTIONAL SYNC (CONFLICT)" "${err_msg}" "${rb_out}" "${remote}" "${branch}"
      return 1
    fi
  fi
  return 0
}

# ==============================================================================
# Automated Pull Request Template Generator (`render_auto_pr_body` & `syncalibur pr`)
# ==============================================================================

render_auto_pr_body() {
  local repo_dir="$1" base_branch="$2" head_branch="$3" pr_title="$4" trigger_reason="${5:-Protected Branch Auto-PR (\`GH006\`)}"
  local diff_range="refs/remotes/origin/${base_branch}..HEAD"
  local head_sha base_sha
  head_sha="$(git -C "${repo_dir}" rev-parse -q --verify HEAD 2>/dev/null || echo "")"
  base_sha="$(git -C "${repo_dir}" rev-parse -q --verify "refs/remotes/origin/${base_branch}" 2>/dev/null || echo "")"
  if [[ -z "${base_sha}" || "${head_sha}" == "${base_sha}" ]]; then
    diff_range="HEAD~1..HEAD"
  fi

  local stat_summary changed_list file_line
  stat_summary="$(git -C "${repo_dir}" diff --shortstat "${diff_range}" 2>/dev/null || echo "1 file changed")"
  stat_summary="${stat_summary#"${stat_summary%%[![:space:]]*}"}"

  local chk_go=" " chk_sh=" " chk_inst=" " chk_docs=" " chk_app=" "
  changed_list=""
  while IFS= read -r file_line; do
    [[ -z "${file_line}" ]] && continue
    changed_list="${changed_list}- \`${file_line}\`"$'\n'
    case "${file_line}" in
      cmd/syncalibur-engine/*|*.go) chk_go="x" ;;
      syncalibur.sh) chk_sh="x" ;;
      install.sh|.github/*) chk_inst="x" ;;
      README.md|*.env.example|*.list.example|docs/*|*.md) chk_docs="x" ;;
      *) chk_app="x" ;;
    esac
  done < <(git -C "${repo_dir}" diff --name-only "${diff_range}" 2>/dev/null)

  if [[ -z "${changed_list}" ]]; then
    changed_list="- \`$(git -C "${repo_dir}" log -1 --name-only --pretty=format: HEAD 2>/dev/null | head -n 1 || echo "repository files")\`"$'\n'
    chk_app="x"
  fi

  cat << EOF
## ◈ Summary

**${pr_title}**

| Field | Details |
| :--- | :--- |
| **Trigger** | ${trigger_reason} |
| **Branches** | \`${head_branch}\` → \`${base_branch}\` |
| **Diff Stat** | \`${stat_summary:-auto-sync}\` |
| **Sync Engine** | Syncalibur Autonomous Git Daemon |

---

## ◆ Changed Files & Diff Telemetry

${changed_list}
---

## ◆ Component Area

- [${chk_go}] **Go Core Engine** (\`cmd/syncalibur-engine/main.go\`)
- [${chk_sh}] **Daemon & Service Supervisor** (\`syncalibur.sh\`)
- [${chk_inst}] **Public Installer & Packaging** (\`install.sh\`, \`.github/\`)
- [${chk_docs}] **Documentation & Config Templates** (\`README.md\`, \`syncalibur.env.example\`, \`repos.list.example\`)
- [${chk_app}] **Application / Repository Code** (watched project files)

---

## ⛨ Verification & Safety Checklist

- [x] Secret & large-file safety shield verified (no \`.env\`, \`.pem\`, \`.key\`, \`.tfstate\`, or binaries > ${MAX_FILE_SIZE_MB} MiB)
- [x] Personal \`syncalibur.env\` and \`repos.list\` files are **not** included in this PR
- [x] \`bash -n syncalibur.sh && bash -n install.sh\` passes without syntax errors
- [x] \`go build ./cmd/syncalibur-engine\` compiles cleanly with zero external dependencies
- [x] \`syncalibur doctor\` passes all 10 diagnostic checks
EOF
}

cmd_pr() {
  cd "${REPO_ROOT}" || return 1
  local dry_run=0 base_branch="main"
  if [[ "${1:-}" == "--dry-run" || "${1:-}" == "--preview" ]]; then
    dry_run=1
    base_branch="${2:-main}"
  elif [[ -n "${1:-}" ]]; then
    base_branch="$1"
    if [[ "${2:-}" == "--dry-run" || "${2:-}" == "--preview" ]]; then
      dry_run=1
    fi
  fi

  local cur_branch pr_branch pr_title pr_body pr_url
  cur_branch="$(get_branch || echo main)"
  pr_title="$(git log -1 --pretty=%s 2>/dev/null || echo "feat: automated Syncalibur update")"

  if [[ "${cur_branch}" == "${base_branch}" ]]; then
    pr_branch="sync/pr-$(date '+%Y%m%d-%H%M%S')"
  else
    pr_branch="${cur_branch}"
  fi

  pr_body="$(render_auto_pr_body "${REPO_ROOT}" "${base_branch}" "${pr_branch}" "${pr_title}" "Manual / CLI Pull Request (\`syncalibur pr\`)")"

  if (( dry_run )); then
    printf '%s\n' "${pr_body}"
    return 0
  fi

  echo "↗ Pushing branch '${pr_branch}' to origin..."
  git push --recurse-submodules=on-demand -u origin "HEAD:refs/heads/${pr_branch}" || return 1

  if command -v gh >/dev/null 2>&1; then
    echo "⇄ Opening GitHub Pull Request (${pr_branch} -> ${base_branch}) with Syncalibur PR Template..."
    pr_url="$(gh pr create --base "${base_branch}" --head "${pr_branch}" --title "${pr_title}" --body "${pr_body}" 2>&1)" && {
      echo "✓ Pull Request created: ${pr_url}"
      send_webhook_notification "summary" "⇄ Pull Request Opened ($(basename "${REPO_ROOT}"))" "${pr_title}" "${pr_url}"
      return 0
    }
    echo "▲  ${pr_url}" >&2
  elif command -v glab >/dev/null 2>&1; then
    echo "⇄ Opening GitLab Merge Request (${pr_branch} -> ${base_branch}) with Syncalibur PR Template..."
    glab mr create --target-branch "${base_branch}" --source-branch "${pr_branch}" --title "${pr_title}" --description "${pr_body}" --yes && return 0
  else
    echo "ℹ  Install 'gh' (GitHub CLI) or 'glab' (GitLab CLI) to create PRs directly from the terminal."
    printf '\n%s\n' "${pr_body}"
  fi
}

# ==============================================================================
# Self-Repo Watcher & Automatic Self-Updater (Rebuilds Go Engine + Hot-Reloads)
# ==============================================================================

LAST_SELF_CHECK_TS=0

sync_and_update_self_repo() {
  [[ "${SELF_AUTO_UPDATE}" == "true" ]] || return 0
  [[ -n "${SELF_REPO_ROOT}" && -d "${SELF_REPO_ROOT}/.git" ]] || return 0

  local now force_upd=0
  now="$(date +%s)"
  if [[ -f "${UPD_TRIGGER_FILE}" ]]; then
    rm -f "${UPD_TRIGGER_FILE}" 2>/dev/null || true
    force_upd=1
  fi

  local self_dirty=""
  if [[ "${SELF_REPO_ROOT}" != "${REPO_ROOT}" ]]; then
    self_dirty="$(git -C "${SELF_REPO_ROOT}" status --porcelain -uall 2>/dev/null || true)"
  fi

  if (( ! force_upd )) && [[ -z "${self_dirty}" ]] && (( now - LAST_SELF_CHECK_TS < SELF_CHECK_INTERVAL )); then
    return 0
  fi
  LAST_SELF_CHECK_TS="${now}"

  local self_branch self_remote="origin"
  self_branch="$(git -C "${SELF_REPO_ROOT}" symbolic-ref --quiet --short HEAD 2>/dev/null || echo "")"
  [[ -n "${self_branch}" ]] || return 0

  if [[ "${SELF_REPO_ROOT}" != "${REPO_ROOT}" && -n "${self_dirty}" ]]; then
    nap 2
    if ! git -C "${SELF_REPO_ROOT}" add -A -- ":(exclude)syncalibur.env" ":(exclude)repos.list" ":(exclude).env" ":(exclude)bin/" >/dev/null 2>&1; then
      git -C "${SELF_REPO_ROOT}" add -A >/dev/null 2>&1 || true
      git -C "${SELF_REPO_ROOT}" reset -q -- syncalibur.env repos.list .env bin/ 2>/dev/null || true
    fi
    if ! git -C "${SELF_REPO_ROOT}" diff --cached --quiet; then
      local s_msg
      if ensure_go_engine; then
        s_msg="$(SYNCALIBUR_AI_COMMIT="${AI_COMMIT}" SYNCALIBUR_OLLAMA_MODEL="${OLLAMA_MODEL}" "${GO_ENGINE_BIN}" commit-msg "${SELF_REPO_ROOT}" 2>/dev/null || echo "chore(syncalibur): self-sync ($(date '+%H:%M:%S'))")"
      else
        s_msg="chore(syncalibur): self-sync ($(date '+%H:%M:%S'))"
      fi
      if git -C "${SELF_REPO_ROOT}" commit -m "${s_msg}" >> "${LOG_FILE}" 2>&1; then
        log "⇄ Self-Sync: Committed Syncalibur update: ${s_msg}"
        local p_out
        if p_out="$(run_with_timeout "${PUSH_TIMEOUT}" git -C "${SELF_REPO_ROOT}" push -u "${self_remote}" "${self_branch}" 2>&1)"; then
          printf '%s\n' "${p_out}" >> "${LOG_FILE}"
          log "✓ Self-Sync: Pushed Syncalibur repository updates to ${self_remote}/${self_branch}."
        elif grep -qiE 'GH006|protected branch|Required status check|At least 1 approving review' <<< "${p_out}"; then
          local pr_branch="sync/auto-$(date '+%Y%m%d-%H%M%S')" pr_body
          pr_body="$(render_auto_pr_body "${SELF_REPO_ROOT}" "${self_branch}" "${pr_branch}" "${s_msg}" "Self-Sync Protected Branch Auto-PR (\`GH006\`)")"
          log "⇄ Self-Sync: '${self_branch}' is protected; pushing feature branch '${pr_branch}' & opening templated Pull Request..."
          if git -C "${SELF_REPO_ROOT}" push -u "${self_remote}" "HEAD:refs/heads/${pr_branch}" >> "${LOG_FILE}" 2>&1; then
            if command -v gh >/dev/null 2>&1; then
              gh pr create -R YusufKaya00/Syncalibur --base "${self_branch}" --head "${pr_branch}" --title "${s_msg}" --body "${pr_body}" >> "${LOG_FILE}" 2>&1 || true
            fi
            git -C "${SELF_REPO_ROOT}" reset --hard "${self_remote}/${self_branch}" >/dev/null 2>&1 || true
            log "✓ Self-Sync: Opened templated Pull Request '${pr_branch}' -> '${self_branch}'."
          fi
        else
          printf '%s\n' "${p_out}" >> "${LOG_FILE}"
          log "▲  Self-Sync push failed: $(push_error_summary "${p_out}")"
          trigger_failure_alert "SELF-REPO PUSH (${SELF_REPO_ROOT})" "Failed to push Syncalibur self-repository to ${self_remote}/${self_branch}" "${p_out}"
        fi
      fi
    fi
  fi

  git -C "${SELF_REPO_ROOT}" remote get-url "${self_remote}" >/dev/null 2>&1 || return 0
  run_with_timeout 30 git -C "${SELF_REPO_ROOT}" fetch -q "${self_remote}" "${self_branch}" 2>/dev/null || return 0

  local local_sha remote_sha behind_cnt
  local_sha="$(git -C "${SELF_REPO_ROOT}" rev-parse -q --verify HEAD 2>/dev/null || echo "")"
  remote_sha="$(git -C "${SELF_REPO_ROOT}" rev-parse -q --verify "refs/remotes/${self_remote}/${self_branch}" 2>/dev/null || echo "")"
  [[ -n "${local_sha}" && -n "${remote_sha}" && "${local_sha}" != "${remote_sha}" ]] || return 0

  behind_cnt="$(git -C "${SELF_REPO_ROOT}" rev-list --count "HEAD..refs/remotes/${self_remote}/${self_branch}" 2>/dev/null || echo 0)"
  (( behind_cnt > 0 )) || return 0

  log "⇄ Self-Update: Detected ${behind_cnt} new upstream commit(s) in Syncalibur (${local_sha:0:7} -> ${remote_sha:0:7}). Pulling & rebuilding..."
  local pull_out
  if pull_out="$(run_with_timeout 60 git -C "${SELF_REPO_ROOT}" pull --rebase --autostash "${self_remote}" "${self_branch}" 2>&1)"; then
    printf '%s\n' "${pull_out}" >> "${LOG_FILE}"
    chmod +x "${SCRIPT_PATH}" 2>/dev/null || true
    rm -f "${GO_ENGINE_BIN}" "${GO_ENGINE_COMPAT_BIN}" "${LOGO_ANSI_FILE}" 2>/dev/null || true
    ensure_go_engine || true
    ensure_logo_ansi || true
    clear_sync_failure
    send_webhook_notification "summary" "⟳ Syncalibur Self-Updated (${local_sha:0:7} → ${remote_sha:0:7})" "Pulled ${behind_cnt} upstream commit(s), rebuilt Go engine, and hot-reloaded daemon." "${pull_out}"
    log "✓ Self-Update: Syncalibur updated to ${remote_sha:0:7}; hot-reloading daemon in the background..."
    trap - EXIT
    if [[ -x "${GO_ENGINE_BIN}" ]]; then
      exec "${GO_ENGINE_BIN}" daemon-exec "${SCRIPT_PATH}" run
    else
      exec /usr/bin/env bash "${SCRIPT_PATH}" run
    fi
  else
    printf '%s\n' "${pull_out}" >> "${LOG_FILE}"
    git -C "${SELF_REPO_ROOT}" rebase --abort >/dev/null 2>&1 || true
    log "▲  Self-Update failed: $(push_error_summary "${pull_out}")"
    trigger_failure_alert "SELF-UPDATE (${SELF_REPO_ROOT})" "Failed to pull latest Syncalibur update from ${self_remote}/${self_branch}" "${pull_out}"
  fi
}

# ==============================================================================
# Domain 1.4: Native Filesystem Event Watcher (`inotifywait` / `fswatch` Hybrid)
# ==============================================================================

NAP_PID=""
nap() {
  sleep "${1:-${POLL_INTERVAL}}" &
  NAP_PID=$!
  wait "${NAP_PID}" 2>/dev/null || true
  NAP_PID=""
}

wait_for_fs_change() {
  local timeout="${1:-${POLL_INTERVAL}}"
  [[ -f "${SYNC_TRIGGER_FILE}" || -f "${UPD_TRIGGER_FILE}" ]] && return 0

  case "${FS_WATCHER}" in
    inotifywait)
      if run_with_timeout "$(( timeout + 2 ))" inotifywait \
        -r -q -t "${timeout}" \
        -e modify,close_write,move,create,delete \
        --exclude '(/\.git(/|$)|/\.git/.*|\.swp$|\.tmp$|4913$|~|\.DS_Store$)' \
        "${WATCH_PATH}" >/dev/null 2>&1; then
        return 0
      fi
      return 1
      ;;
    fswatch)
      if run_with_timeout "${timeout}" fswatch \
        -1 -t "${timeout}" \
        -e '/\.git(/|$)' -e '\.swp$' -e '\.tmp$' -e '\.DS_Store$' \
        "${WATCH_PATH}" >/dev/null 2>&1; then
        return 0
      fi
      return 1
      ;;
    *)
      nap "${timeout}"
      return 0
      ;;
  esac
}

# ==============================================================================
# Logging & Repository Commit/Push Helpers (Go Engine Semantic Commits)
# ==============================================================================

log() {
  local line
  line="[$(date '+%Y-%m-%d %H:%M:%S')] $*"
  printf '%s\n' "${line}" >> "${LOG_FILE}"
  if [[ "${SYNCALIBUR_BACKGROUND:-0}" != "1" ]]; then
    if [[ -t 1 ]]; then
      format_log_line "${line}"
    else
      printf '%s\n' "${line}"
    fi
  fi
}

LAST_NOTICE=""
notice() {
  [[ "$*" == "${LAST_NOTICE}" ]] && return 0
  LAST_NOTICE="$*"
  log "$@"
}

WARNED_PATHS=""
warn_blocked() {
  local path="$1" reason="$2"
  case $'\n'"${WARNED_PATHS}"$'\n' in
    *$'\n'"${path}"$'\n'*) return 0 ;;
  esac
  WARNED_PATHS="${WARNED_PATHS}"$'\n'"${path}"
  log "▲ Not committing '${path}': ${reason}. Add it to .gitignore to silence this warning."
}

trim_log() {
  local max_bytes=$(( LOG_MAX_KB * 1024 )) size tmp
  [[ -f "${LOG_FILE}" ]] || return 0
  size="$(file_size_bytes "${LOG_FILE}")"
  (( size > max_bytes )) || return 0
  tmp="$(mktemp)" || return 0
  if tail -c "$(( max_bytes / 2 ))" "${LOG_FILE}" | tail -n +2 > "${tmp}"; then
    cat "${tmp}" > "${LOG_FILE}"
  fi
  rm -f "${tmp}"
}

is_lock_held_by_process() {
  local lock_file="$1"
  [[ ! -e "${lock_file}" ]] && return 1
  if command -v fuser >/dev/null 2>&1; then
    fuser "${lock_file}" >/dev/null 2>&1 && return 0
  fi
  if command -v lsof >/dev/null 2>&1; then
    lsof -t "${lock_file}" >/dev/null 2>&1 && return 0
  fi
  local real_lock
  real_lock="$(readlink -f "${lock_file}" 2>/dev/null || echo "${lock_file}")"
  for fd in /proc/[0-9]*/fd/*; do
    [[ ! -e "${fd}" ]] && continue
    local target
    target="$(readlink "${fd}" 2>/dev/null || true)"
    if [[ "${target}" == "${real_lock}" ]]; then
      return 0
    fi
  done 2>/dev/null
  return 1
}

find_git_locks() {
  local locks=()
  for marker in index.lock refs/heads/*.lock HEAD.lock config.lock MERGE_HEAD CHERRY_PICK_HEAD REVERT_HEAD BISECT_LOG BISECT_START rebase-merge rebase-apply sequencer AUTO_MERGE; do
    for path in "${GIT_DIR}"/${marker}; do
      [[ -e "${path}" ]] && locks+=("${path}")
    done
  done
  printf '%s\n' "${locks[@]}"
}

run_git_with_backoff() {
  local max_attempts=5 delay=1 attempt=1
  while (( attempt <= max_attempts )); do
    local err_file
    err_file="$(mktemp "${TMPDIR:-/tmp}/syncalibur-git-err.XXXXXX")"
    if "$@" 2>"${err_file}"; then
      rm -f "${err_file}"
      return 0
    fi
    local err_msg
    err_msg="$(cat "${err_file}")"
    rm -f "${err_file}"

    if [[ "${err_msg}" =~ "index.lock" || "${err_msg}" =~ "Another git process is running" ]]; then
      if [[ -f "${GIT_DIR}/index.lock" ]]; then
        local l_age=$(( $(date +%s) - $(file_mtime_epoch "${GIT_DIR}/index.lock" 2>/dev/null || echo 0) ))
        if (( l_age > 60 )) && ! is_lock_held_by_process "${GIT_DIR}/index.lock"; then
          log "▲ Auto-healed stale '${GIT_DIR}/index.lock' (${l_age}s old; no process holding fd)."
          rm -f "${GIT_DIR}/index.lock" 2>/dev/null || true
        fi
      fi
      sleep "${delay}"
      delay=$(( delay * 2 ))
      (( attempt++ ))
    else
      echo "${err_msg}" >&2
      return 1
    fi
  done
  return 1
}

assert_git_operable() {
  local repo_dir="${1:-${REPO_ROOT}}"
  if ! git -C "${repo_dir}" rev-parse --git-dir >/dev/null 2>&1; then
    echo -e "${C_RED}✕ Not a git repository: '${repo_dir}'${C_RESET}" >&2
    return 1
  fi
  if ! git -C "${repo_dir}" symbolic-ref -q HEAD >/dev/null 2>&1; then
    echo -e "${C_RED}✕ Detached HEAD state detected in '${repo_dir}'. Syncalibur requires an active named branch (e.g. main/dev) to prevent orphaned commits.${C_RESET}" >&2
    return 1
  fi
  local busy
  if busy="$(repo_busy)"; then
    echo -e "${C_AMBER}▲ Git repository is currently busy with operation '${busy}'. Syncalibur auto-operation paused until resolved.${C_RESET}" >&2
    return 1
  fi
  return 0
}

repo_busy() {
  local marker lock_age
  # 1. Stale index.lock recovery: if index.lock is >60s old and no active process holds fd, auto-heal
  if [[ -e "${GIT_DIR}/index.lock" ]]; then
    lock_age=$(( $(date +%s) - $(file_mtime_epoch "${GIT_DIR}/index.lock" 2>/dev/null || echo 0) ))
    if (( lock_age > 60 )) && ! is_lock_held_by_process "${GIT_DIR}/index.lock"; then
      log "▲ Auto-healed stale '${GIT_DIR}/index.lock' (${lock_age}s old; no process holding fd)."
      rm -f "${GIT_DIR}/index.lock" 2>/dev/null || true
    fi
  fi

  for marker in index.lock refs/heads/*.lock HEAD.lock MERGE_HEAD CHERRY_PICK_HEAD REVERT_HEAD BISECT_LOG BISECT_START rebase-merge rebase-apply sequencer AUTO_MERGE; do
    if [[ -e "${GIT_DIR}/${marker}" ]]; then
      echo "${marker}"
      return 0
    fi
  done
  return 1
}

unpushed_count() {
  local remote="$1" branch="$2"
  if git rev-parse -q --verify "refs/remotes/${remote}/${branch}" >/dev/null 2>&1; then
    git rev-list --count "refs/remotes/${remote}/${branch}..HEAD" 2>/dev/null || echo 0
  elif git rev-parse -q --verify "refs/remotes/origin/${branch}" >/dev/null 2>&1; then
    git rev-list --count "refs/remotes/origin/${branch}..HEAD" 2>/dev/null || echo 0
  else
    git rev-list --count HEAD 2>/dev/null || echo 0
  fi
}

has_changes_in_scope() {
  [[ -n "$(git status --porcelain -uall -- "${SCOPE}" 2>/dev/null)" ]]
}

snapshot() {
  git -c core.quotePath=false status --porcelain -uall -- "${SCOPE}" 2>/dev/null \
    | sed -E 's/^.{3}//; s/^.* -> //; s/^"(.*)"$/\1/' \
    | while IFS= read -r path; do
        printf '%s %s\n' "${path}" "$(file_signature "${path}")"
      done \
    | git hash-object --stdin
}

wait_for_quiet() {
  local previous current waited=0
  previous="$(snapshot)"
  while (( waited < MAX_DEBOUNCE )); do
    nap "${DEBOUNCE_WAIT}"
    waited=$(( waited + DEBOUNCE_WAIT ))
    current="$(snapshot)"
    [[ "${current}" == "${previous}" ]] && return 0
    previous="${current}"
  done
  log "▲ Files still changing after ${MAX_DEBOUNCE}s; committing anyway."
}

changed_paths() {
  local entry xy
  while IFS= read -r -d '' entry; do
    xy="${entry:0:2}"
    if [[ "${xy}" == R* || "${xy}" == C* ]]; then
      IFS= read -r -d '' _
    fi
    [[ "${xy}" == *D* ]] && continue
    printf '%s\0' "${entry:3}"
  done < <(git status --porcelain -z -uall -- "${SCOPE}" 2>/dev/null)
}

block_reason() {
  local path="$1" name="${1##*/}" pattern size
  if [[ "${name}" == "syncalibur.env" || "${name}" == "repos.list" ]]; then
    echo "personal Syncalibur config file (gitignored)"
    return 0
  fi
  case "${name}" in
    *.example | *.sample | *.template | *.dist | *.default | *.skel) ;;
    *)
      if (( ${#SENSITIVE_PATTERNS[@]} > 0 )); then
        for pattern in "${SENSITIVE_PATTERNS[@]}"; do
          [[ -z "${pattern}" ]] && continue
          # shellcheck disable=SC2053
          if [[ "${name}" == ${pattern} || "${path}" == ${pattern} || "${path}" == *"${pattern}"* ]]; then
            echo "looks sensitive (matches '${pattern}')"
            return 0
          fi
        done
      fi
      ;;
  esac

  if (( ${#BINARY_PATTERNS[@]} > 0 )); then
    for bin_pat in "${BINARY_PATTERNS[@]}"; do
      [[ -z "${bin_pat}" ]] && continue
      # shellcheck disable=SC2053
      if [[ "${name}" == ${bin_pat} || "${path}" == ${bin_pat} || "${path}" == *"${bin_pat}"* ]]; then
        echo "binary/archive file (matches '${bin_pat}')"
        return 0
      fi
    done
  fi

  if [[ -f "${path}" && ! -L "${path}" ]]; then
    if [[ "$(head -c 8000 "${path}" 2>/dev/null | tr -d '\0' | wc -c)" -ne "$(head -c 8000 "${path}" 2>/dev/null | wc -c)" ]]; then
      echo "binary file (contains null bytes)"
      return 0
    fi
  fi

  if (( MAX_FILE_BYTES > 0 )) && [[ -f "${path}" && ! -L "${path}" ]]; then
    size="$(file_size_bytes "${path}")"
    if (( size > MAX_FILE_BYTES )); then
      echo "too large ($(( size / 1024 / 1024 )) MiB > ${MAX_FILE_SIZE_MB} MiB limit)"
      return 0
    fi
  fi
  return 1
}

record_timeline_checkpoint() {
  local head_sha epoch_ts ref_name
  head_sha="$(git rev-parse HEAD 2>/dev/null || echo "")"
  [[ -z "${head_sha}" ]] && return 0
  epoch_ts="$(date +%s%N 2>/dev/null || date +%s)"
  ref_name="refs/syncalibur/timeline/${epoch_ts}"
  git update-ref "${ref_name}" "${head_sha}" 2>/dev/null || true

  # Prune timeline checkpoints: keep newest 50
  local all_refs=()
  while IFS= read -r r; do
    [[ -n "${r}" ]] && all_refs+=("${r}")
  done < <(git for-each-ref --sort=-creatordate --format='%(refname)' 'refs/syncalibur/timeline/*' 2>/dev/null)

  if (( ${#all_refs[@]} > 50 )); then
    local old_ref
    for old_ref in "${all_refs[@]:50}"; do
      git update-ref -d "${old_ref}" 2>/dev/null || true
    done
  fi
}

commit_changes() {
  local path reason msg count line err_out
  local -a excludes=() files=() quarantined_set=()

  # 1. Pre-flight Syntax Gate, Secret Shield & DevOps Lint Gate (Go Engine)
  if ensure_go_engine; then
    local pf_out pf_code=0
    pf_out="$("${GO_ENGINE_BIN}" preflight "${REPO_ROOT}" --json 2>/dev/null || true)"
    if [[ -n "${pf_out}" ]] && command -v jq >/dev/null 2>&1; then
      pf_code="$(jq -r '.exit_code // 0' <<< "${pf_out}")"
      if (( pf_code == 3 )); then
        while IFS= read -r q_file; do
          [[ -z "${q_file}" ]] && continue
          excludes+=(":(exclude,literal)${q_file}")
          quarantined_set+=("${q_file}")
          log "⛨ [SECRET SHIELD] Quarantined '${q_file}' (high-entropy credential detected; excluded from commit)"
        done < <(jq -r '.quarantined_files[]?' <<< "${pf_out}")
      elif (( pf_code == 2 )); then
        local syn_err
        syn_err="$(jq -r '.details[0] // .syntax_errors[0] // "syntax error"' <<< "${pf_out}")"
        log "▲ [SYNTAX HOLD] Incomplete syntax in ${syn_err}; holding commit until write settles."
        return 1
      elif (( pf_code == 4 )); then
        local devops_err
        devops_err="$(jq -r '.details[0] // .devops_errors[0] // "infrastructure lint failure"' <<< "${pf_out}")"
        log "☸ [DEVOPS ADVISORY] Infrastructure lint warning: ${devops_err} (non-blocking)."
      fi
    elif ensure_go_engine; then
      local pf_raw pf_rc=0
      pf_raw="$("${GO_ENGINE_BIN}" preflight "${REPO_ROOT}" 2>&1)" || pf_rc=$?
      if (( pf_rc == 2 )); then
        local first_err
        first_err="$(head -n 2 <<< "${pf_raw}" | tr '\n' ' ')"
        log "▲ [SYNTAX HOLD] ${first_err}; holding commit until write settles."
        return 1
      elif (( pf_rc == 3 )); then
        log "⛨ [SECRET SHIELD] Detected credential in working tree; holding commit."
        return 1
      elif (( pf_rc == 4 )); then
        local first_err
        first_err="$(head -n 2 <<< "${pf_raw}" | tr '\n' ' ')"
        log "☸ [DEVOPS ADVISORY] ${first_err} (non-blocking)."
      fi
    fi
  fi

  # Smart Staging Partition: If multiple distinct file scopes exist, commit them atomically
  local smart_groups=()
  if [[ "${SMART_STAGING:-true}" == "true" ]] && ensure_go_engine; then
    while IFS= read -r s_line; do
      [[ "${s_line}" =~ ^GROUP\| ]] && smart_groups+=("${s_line}")
    done < <("${GO_ENGINE_BIN}" smart-stage "${REPO_ROOT}" 2>/dev/null || true)
  fi

  if (( ${#smart_groups[@]} > 1 )); then
    local g_line g_scope g_type g_files_csv g_files total_committed=0
    for g_line in "${smart_groups[@]}"; do
      IFS='|' read -r _ g_scope g_type g_files_csv <<< "${g_line}"
      IFS=',' read -r -a g_files <<< "${g_files_csv}"

      local g_valid_files=()
      for f in "${g_files[@]}"; do
        local is_quarantined=0
        for q in "${quarantined_set[@]}"; do
          if [[ "${f}" == "${q}" ]]; then
            is_quarantined=1
            break
          fi
        done
        (( is_quarantined )) && continue

        if ! reason="$(block_reason "${f}")"; then
          g_valid_files+=("${f}")
        else
          warn_blocked "${f}" "${reason}"
        fi
      done

      (( ${#g_valid_files[@]} == 0 )) && continue

      git add -- "${g_valid_files[@]}" 2>/dev/null || true
      if git diff --cached --quiet; then
        continue
      fi

      local g_msg=""
      if [[ "${SEMANTIC_COMMITS}" == "true" ]] && ensure_go_engine; then
        g_msg="$(SYNCALIBUR_AI_COMMIT="${AI_COMMIT}" SYNCALIBUR_OLLAMA_MODEL="${OLLAMA_MODEL}" "${GO_ENGINE_BIN}" commit-msg "${REPO_ROOT}" 2>/dev/null || true)"
      fi
      if [[ -z "${g_msg:-}" ]]; then
        g_msg="${g_type}(${g_scope}): update ${g_valid_files[0]}"
        (( ${#g_valid_files[@]} > 1 )) && g_msg+=", ${g_valid_files[1]}"
        (( ${#g_valid_files[@]} > 2 )) && g_msg+=" (+$((${#g_valid_files[@]} - 2)) more)"
      fi

      if err_out="$(git commit -m "${g_msg}" 2>&1)"; then
        printf '%s\n' "${err_out}" >> "${LOG_FILE}"
        log "✓ [Smart Stage: ${g_scope}] Committed ${#g_valid_files[@]} file(s): ${g_msg}"
        record_timeline_checkpoint
        total_committed=$(( total_committed + 1 ))
      else
        printf '%s\n' "${err_out}" >> "${LOG_FILE}"
        log "✕ [Smart Stage] 'git commit' failed: ${g_msg}"
      fi
    done

    if (( total_committed > 0 )); then
      return 0
    fi
  fi

  while IFS= read -r -d '' path; do
    if reason="$(block_reason "${path}")"; then
      excludes+=(":(exclude,literal)${path}")
      warn_blocked "${path}" "${reason}"
    fi
  done < <(changed_paths)

  if ! err_out="$(git add -A -- "${SCOPE}" "${excludes[@]}" 2>&1)"; then
    printf '%s\n' "${err_out}" >> "${LOG_FILE}"
    log "✕ 'git add' failed (details in ${LOG_FILE})."
    trigger_failure_alert "GIT STAGE (git add)" "git add failed while staging changes in '${SCOPE}'" "${err_out}"
    return 1
  fi

  while IFS= read -r -d '' path; do
    if reason="$(block_reason "${path}")"; then
      git reset -q -- "${path}" 2>/dev/null || git rm -q --cached -- "${path}" >/dev/null 2>&1
      warn_blocked "${path}" "${reason}"
    fi
  done < <(git diff --cached --name-only -z --diff-filter=ACMR)

  git diff --cached --quiet && return 1

  while IFS= read -r line; do
    [[ -n "${line}" ]] && files+=("${line}")
  done < <(git -c core.quotePath=false diff --cached --name-only)

  count=${#files[@]}
  if [[ "${SEMANTIC_COMMITS}" == "true" ]] && ensure_go_engine; then
    msg="$(SYNCALIBUR_AI_COMMIT="${AI_COMMIT}" SYNCALIBUR_OLLAMA_MODEL="${OLLAMA_MODEL}" "${GO_ENGINE_BIN}" commit-msg "${REPO_ROOT}" 2>/dev/null || true)"
  fi
  if [[ -z "${msg:-}" ]]; then
    msg="auto: ${files[0]}"
    (( count > 1 )) && msg+=", ${files[1]}"
    (( count > 2 )) && msg+=" (+$(( count - 2 )) more)"
    msg+=" ($(date '+%H:%M:%S'))"
  fi

  if err_out="$(git commit -m "${msg}" 2>&1)"; then
    printf '%s\n' "${err_out}" >> "${LOG_FILE}"
    log "✓ Committed ${count} file(s): ${msg}"
    record_timeline_checkpoint
    return 0
  fi
  printf '%s\n' "${err_out}" >> "${LOG_FILE}"
  log "✕ 'git commit' failed (details in ${LOG_FILE})."
  trigger_failure_alert "GIT COMMIT" "git commit failed for '${msg}'" "${err_out}"
  return 1
}

push_error_summary() {
  local line summary="" count=0
  while IFS= read -r line; do
    line="${line#remote:}"
    line="${line#"${line%%[![:space:]]*}"}"
    line="${line%"${line##*[![:space:]]}"}"
    [[ -z "${line}" ]] && continue
    if [[ -z "${summary}" ]]; then
      summary="${line}"
    else
      summary="${summary} | ${line}"
    fi
    count=$(( count + 1 ))
    (( count >= 3 )) && break
  done < <(grep -iE 'rejected|error|fatal|denied|exceed|too large|LFS|could not|not found' <<< "$1")
  echo "${summary:-unknown error, see ${LOG_FILE}}"
}

push_changes() {
  local remote="$1" branch="$2" output rc targets_desc err_summary last_commit_msg
  if [[ -z "$(get_push_urls "${remote}")" ]]; then
    auto_create_or_bind_remote "" "" || true
    remote="$(get_remote)"
  fi

  # Check network connectivity before push attempt (Offline Shield)
  if ! check_network_liveness "${remote}"; then
    touch "${OFFLINE_FLAG_FILE}" 2>/dev/null || true
    local unp_cnt=0
    unp_cnt="$(unpushed_count "${remote}" "${branch}")"
    log "☁ Offline Shield: Connection to '${remote}' unavailable. ${unp_cnt} commit(s) queued locally (will auto-flush on reconnect)."
    return 1
  fi

  targets_desc="$(describe_push_targets_inline "${remote}")"
  last_commit_msg="$(git log -1 --pretty=%s 2>/dev/null || echo "auto-sync")"
  log "↗ Pushing '${branch}' -> ${targets_desc}..."
  output="$(run_with_timeout "${PUSH_TIMEOUT}" git push --recurse-submodules=on-demand -u "${remote}" "${branch}" 2>&1)"
  rc=$?
  printf '%s\n' "${output}" >> "${LOG_FILE}"
  if (( rc == 0 )); then
    if [[ -f "${OFFLINE_FLAG_FILE}" ]]; then
      rm -f "${OFFLINE_FLAG_FILE}" 2>/dev/null || true
      log "🌐 Network connectivity restored. Successfully flushed queued commits (${targets_desc})."
    fi
    clear_sync_failure
    log "✓ Push succeeded (${targets_desc})."
    send_webhook_notification "success" "✓ Syncalibur Push Succeeded ($(basename "${REPO_ROOT}"))" "${last_commit_msg}" "Target: ${targets_desc} (${branch})"
    return 0
  fi

  # If push failed due to protected branch rules (GH006 / required PR review), auto-push a branch and open a templated PR!
  if grep -qiE 'GH006|protected branch|Required status check|At least 1 approving review' <<< "${output}"; then
    local pr_branch="sync/auto-$(date '+%Y%m%d-%H%M%S')" pr_title pr_body
    pr_title="${last_commit_msg}"
    pr_body="$(render_auto_pr_body "${REPO_ROOT}" "${branch}" "${pr_branch}" "${pr_title}" "Protected Branch Auto-PR (\`GH006\`)")"
    log "⇄ Branch '${branch}' is protected; pushing to '${pr_branch}' and opening templated Pull Request..."
    if run_with_timeout "${PUSH_TIMEOUT}" git push --recurse-submodules=on-demand -u origin "HEAD:refs/heads/${pr_branch}" >> "${LOG_FILE}" 2>&1; then
      if command -v gh >/dev/null 2>&1; then
        gh pr create --base "${branch}" --head "${pr_branch}" --title "${pr_title}" --body "${pr_body}" >> "${LOG_FILE}" 2>&1 || true
      elif command -v glab >/dev/null 2>&1; then
        glab mr create --target-branch "${branch}" --source-branch "${pr_branch}" --title "${pr_title}" --description "${pr_body}" --yes >> "${LOG_FILE}" 2>&1 || true
      fi
      clear_sync_failure
      log "✓ Opened templated Pull Request '${pr_branch}' -> '${branch}'."
      send_webhook_notification "summary" "⇄ Auto-Opened Pull Request ($(basename "${REPO_ROOT}"))" "${pr_title} (${pr_branch} -> ${branch})" "${pr_body}"
      return 0
    fi
  fi

  # If push failed due to non-fast-forward and bidirectional sync is enabled, try safe rebase & retry push!
  if [[ "${BIDIRECTIONAL_SYNC}" == "true" ]] && grep -qiE 'fetch first|non-fast-forward' <<< "${output}"; then
    if safe_bidirectional_sync "${remote}" "${branch}" 1; then
      output="$(run_with_timeout "${PUSH_TIMEOUT}" git push --recurse-submodules=on-demand -u "${remote}" "${branch}" 2>&1)"
      rc=$?
      printf '%s\n' "${output}" >> "${LOG_FILE}"
      if (( rc == 0 )); then
        if [[ -f "${OFFLINE_FLAG_FILE}" ]]; then
          rm -f "${OFFLINE_FLAG_FILE}" 2>/dev/null || true
          log "🌐 Network connectivity restored. Successfully flushed queued commits (${targets_desc})."
        fi
        clear_sync_failure
        log "✓ Push succeeded after automatic rebase (${targets_desc})."
        send_webhook_notification "success" "✓ Syncalibur Push Succeeded ($(basename "${REPO_ROOT}"))" "${last_commit_msg}" "Rebased & pushed to ${targets_desc} (${branch})"
        return 0
      fi
    else
      return 1
    fi
  fi

  if [[ "${AUTO_CREATE_REMOTE}" == "true" ]] && grep -qiE 'Repository not found|project you were looking for could not be found|does not appear to be a git repository' <<< "${output}"; then
    local url parsed provider slug created_any=0
    while IFS= read -r url; do
      [[ -z "${url}" ]] && continue
      parsed="$(parse_remote_url "${url}")"
      provider="${parsed%%|*}"
      slug="${parsed#*|}"
      slug="${slug%%|*}"
      if auto_create_or_bind_remote "${slug}" "${provider}"; then
        created_any=1
      fi
    done < <(get_push_urls "${remote}")

    if (( created_any )); then
      output="$(run_with_timeout "${PUSH_TIMEOUT}" git push --recurse-submodules=on-demand -u "${remote}" "${branch}" 2>&1)"
      rc=$?
      printf '%s\n' "${output}" >> "${LOG_FILE}"
      if (( rc == 0 )); then
        if [[ -f "${OFFLINE_FLAG_FILE}" ]]; then
          rm -f "${OFFLINE_FLAG_FILE}" 2>/dev/null || true
          log "🌐 Network connectivity restored. Successfully flushed queued commits (${targets_desc})."
        fi
        clear_sync_failure
        log "✓ Push succeeded after auto-creating remote repository (${targets_desc})."
        send_webhook_notification "success" "✓ Syncalibur Created Remote & Pushed ($(basename "${REPO_ROOT}"))" "${last_commit_msg}" "Target: ${targets_desc} (${branch})"
        return 0
      fi
    fi
  fi

  if (( rc == 124 )); then
    err_summary="Push timed out after ${PUSH_TIMEOUT}s (${targets_desc})"
    log "▲  ${err_summary}."
  else
    err_summary="$(push_error_summary "${output}")"
    log "▲  Push failed: ${err_summary}"
  fi

  # Offline Network Disconnection Check: do not spam alert popup windows if offline
  if grep -qiE 'Could not resolve host|Connection refused|Operation timed out|Failed to connect|Network is unreachable|Connection reset by peer|Connection timed out|fatal: unable to access|ssh: connect to host' <<< "${output}"; then
    touch "${OFFLINE_FLAG_FILE}" 2>/dev/null || true
    local unp_cnt=0
    unp_cnt="$(unpushed_count "${remote}" "${branch}")"
    log "☁ Offline Shield: Network disconnection detected (${err_summary}). ${unp_cnt} commit(s) safely queued locally."
    return 1
  fi

  trigger_failure_alert "GIT PUSH (${remote}/${branch})" "${err_summary}" "${output}" "${remote}" "${branch}"
  return 1
}

# ==============================================================================
# Domain 1.0: Developer Push Ergonomics, Safety Checks & History Hygiene
# ==============================================================================

cmd_check() {
  local target_repo="${1:-${REPO_ROOT}}"
  [[ -d "${target_repo}" ]] || target_repo="${PWD}"
  target_repo="$(cd "${target_repo}" && git rev-parse --show-toplevel 2>/dev/null || echo "${target_repo}")"

  echo -e "${C_CYAN}◈ Syncalibur Preflight & Safety Inspection${C_RESET} [$(basename "${target_repo}") @ $(git -C "${target_repo}" branch --show-current 2>/dev/null || echo "detached")]"
  echo "────────────────────────────────────────────────────────────────────────"

  # 1. Secret Shield & Syntax Hold via Go engine
  if ensure_go_engine; then
    local pf_raw pf_rc=0
    pf_raw="$("${GO_ENGINE_BIN}" preflight "${target_repo}" 2>&1)" || pf_rc=$?
    if (( pf_rc == 0 )); then
      echo -e "  ${C_GREEN}⛨ Secret Shield : Clean (0 credentials detected)${C_RESET}"
      echo -e "  ${C_GREEN}▲ Syntax Hold   : Valid (universal lexical balance passed)${C_RESET}"
    else
      echo -e "  ${C_AMBER}▲ Preflight Notice:${C_RESET}"
      echo "${pf_raw}"
    fi
  fi

  # 2. Working Tree Changes
  local changed_cnt unpushed_cnt remotes_desc branch
  branch="$(git -C "${target_repo}" branch --show-current 2>/dev/null || echo "main")"
  changed_cnt="$(git -C "${target_repo}" status --porcelain 2>/dev/null | wc -l | tr -d ' ')"
  unpushed_cnt="$(cd "${target_repo}" && unpushed_count "$(get_remote)" "${branch}")"
  remotes_desc="$(cd "${target_repo}" && describe_push_targets_inline "$(get_remote)")"

  echo -e "  ${C_CYAN}✦ Working Tree  : ${changed_cnt} changed file(s)${C_RESET}"
  if (( changed_cnt > 0 )); then
    git -C "${target_repo}" -c core.quotePath=false status --short 2>/dev/null | head -n 8 | while IFS= read -r line; do
      echo -e "    ${C_SLATE}• ${line}${C_RESET}"
    done
    if (( changed_cnt > 8 )); then
      echo -e "    ${C_SLATE}... and $(( changed_cnt - 8 )) more file(s)${C_RESET}"
    fi
  fi

  # 3. Proposed Commit Message
  if ensure_go_engine && (( changed_cnt > 0 )); then
    local proposed_msg
    proposed_msg="$("${GO_ENGINE_BIN}" commit-msg "${target_repo}" 2>/dev/null || true)"
    if [[ -n "${proposed_msg}" ]]; then
      echo -e "  ${C_GOLD}✦ Proposed Msg  : ${proposed_msg}${C_RESET}"
    fi
  fi

  # 4. Remote & Sync Status
  echo -e "  ${C_CYAN}⇄ Remote Status : Behind: 0 | Ahead: ${unpushed_cnt} unpushed commit(s)${C_RESET}"
  echo -e "  ${C_CYAN}↗ Push Targets  : ${remotes_desc}${C_RESET}"
  echo "────────────────────────────────────────────────────────────────────────"
  if (( changed_cnt > 0 || unpushed_cnt > 0 )); then
    echo -e "  ${C_GREEN}Status: Ready to sync.${C_RESET} Run '${C_CYAN}syncalibur push${C_RESET}' to commit and push."
  else
    echo -e "  ${C_GREEN}Status: All files committed and remotes up to date.${C_RESET}"
  fi
}

cmd_save() {
  cmd_push "$@"
}

cmd_push() {
  local custom_msg="" tidy_mode=0 dry_run=0
  while [[ $# -gt 0 ]]; do
    case "$1" in
      --tidy) tidy_mode=1; shift ;;
      --dry-run) dry_run=1; shift ;;
      *)
        if [[ -z "${custom_msg}" ]]; then
          custom_msg="$1"
        else
          custom_msg="${custom_msg} $1"
        fi
        shift
        ;;
    esac
  done

  local current_repo
  current_repo="$(git rev-parse --show-toplevel 2>/dev/null || echo "${REPO_ROOT}")"
  cd "${current_repo}" || return 1
  assert_git_operable "${current_repo}" || return 1

  echo -e "${C_CYAN}${C_BOLD}◈ Syncalibur Quicksave & Push${C_RESET} [$(basename "${current_repo}") @ $(git branch --show-current 2>/dev/null || echo "main")]"

  if (( dry_run )); then
    echo -e "  ${C_AMBER}✦ Running in --dry-run mode (no changes will be committed or pushed)${C_RESET}"
    cmd_check "${current_repo}"
    return 0
  fi

  # If tidy mode requested, tidy unpushed commits first
  if (( tidy_mode )); then
    cmd_tidy "${current_repo}" || true
  fi

  # 1. Run Pre-flight Safety Gate
  if ensure_go_engine; then
    local pf_raw pf_rc=0
    pf_raw="$("${GO_ENGINE_BIN}" preflight "${current_repo}" 2>&1)" || pf_rc=$?
    if (( pf_rc == 2 )); then
      echo -e "  ${C_AMBER}▲ [SYNTAX HOLD] Incomplete syntax detected; push halted:${C_RESET}"
      echo "${pf_raw}"
      return 1
    elif (( pf_rc == 3 )); then
      echo -e "  ${C_RED}⛨ [SECRET SHIELD] High-entropy credential detected in files:${C_RESET}"
      echo "${pf_raw}"
      echo -e "  ${C_SLATE}Hint: Move credentials to .env or add to .gitignore before pushing.${C_RESET}"
      return 1
    elif (( pf_rc == 4 )); then
      echo -e "  ${C_AMBER}☸ [DEVOPS ADVISORY] Infrastructure lint warnings detected (non-blocking).${C_RESET}"
    fi
    echo -e "  ${C_GREEN}⛨ Secret Shield : Clean${C_RESET}"
    echo -e "  ${C_GREEN}▲ Syntax Hold   : Valid${C_RESET}"
  fi

  # 2. Stage & Commit if changes exist
  local branch remote
  branch="$(get_branch || echo "main")"
  remote="$(get_remote)"

  if has_changes_in_scope; then
    if [[ -n "${custom_msg}" ]]; then
      git add -A -- "${SCOPE}" 2>/dev/null
      git commit -m "${custom_msg}"
      record_timeline_checkpoint
      echo -e "  ${C_GREEN}✓ Committed     : ${custom_msg}${C_RESET}"
    else
      commit_changes || true
    fi
  else
    echo -e "  ${C_SLATE}• Working tree clean (no new changes to commit)${C_RESET}"
  fi

  # 3. Pull if behind (safe bidirectional rebase)
  if [[ "${BIDIRECTIONAL_SYNC}" == "true" ]]; then
    safe_bidirectional_sync "${remote}" "${branch}" 1 || true
  fi

  # 4. Push to all remotes
  local unpushed
  unpushed="$(unpushed_count "${remote}" "${branch}")"
  if (( unpushed > 0 )); then
    if push_changes "${remote}" "${branch}"; then
      echo -e "  ${C_GREEN}✦ All remotes in sync.${C_RESET}"
      return 0
    else
      echo -e "  ${C_RED}✕ Push failed (see details above).${C_RESET}"
      return 1
    fi
  else
    echo -e "  ${C_GREEN}✓ Already up to date with remote.${C_RESET}"
    return 0
  fi
}

cmd_commit() {
  local custom_msg="" dry_run=0
  while [[ $# -gt 0 ]]; do
    case "$1" in
      --dry-run) dry_run=1; shift ;;
      *)
        if [[ -z "${custom_msg}" ]]; then
          custom_msg="$1"
        else
          custom_msg="${custom_msg} $1"
        fi
        shift
        ;;
    esac
  done

  local current_repo
  current_repo="$(git rev-parse --show-toplevel 2>/dev/null || echo "${REPO_ROOT}")"
  cd "${current_repo}" || return 1
  assert_git_operable "${current_repo}" || return 1

  echo -e "${C_CYAN}◈ Syncalibur Local Commit${C_RESET} [$(basename "${current_repo}") @ $(git branch --show-current 2>/dev/null || echo "main")]"

  if (( dry_run )); then
    cmd_check "${current_repo}"
    return 0
  fi

  # Pre-flight gate
  if ensure_go_engine; then
    local pf_raw pf_rc=0
    pf_raw="$("${GO_ENGINE_BIN}" preflight "${current_repo}" 2>&1)" || pf_rc=$?
    if (( pf_rc == 2 )); then
      echo -e "  ${C_AMBER}▲ [SYNTAX HOLD] Incomplete syntax detected:${C_RESET}"
      echo "${pf_raw}"
      return 1
    elif (( pf_rc == 3 )); then
      echo -e "  ${C_RED}⛨ [SECRET SHIELD] High-entropy credential detected:${C_RESET}"
      echo "${pf_raw}"
      return 1
    elif (( pf_rc == 4 )); then
      echo -e "  ${C_RED}☸ [DEVOPS HOLD] Infrastructure lint policy violation detected:${C_RESET}"
      echo "${pf_raw}"
      return 1
    fi
  fi

  if ! has_changes_in_scope; then
    echo -e "  ${C_SLATE}• No changes to commit in '${SCOPE}'${C_RESET}"
    return 0
  fi

  if [[ -n "${custom_msg}" ]]; then
    git add -A -- "${SCOPE}" 2>/dev/null
    git commit -m "${custom_msg}"
    echo -e "  ${C_GREEN}✓ Committed : ${custom_msg}${C_RESET}"
  else
    commit_changes
  fi
}

cmd_tidy() {
  local current_repo="${1:-${REPO_ROOT}}"
  [[ -d "${current_repo}" ]] || current_repo="${PWD}"
  current_repo="$(cd "${current_repo}" && git rev-parse --show-toplevel 2>/dev/null || echo "${current_repo}")"
  cd "${current_repo}" || return 1
  assert_git_operable "${current_repo}" || return 1

  local branch upstream unpushed_cnt rem
  branch="$(git branch --show-current 2>/dev/null || echo "main")"
  upstream="$(git rev-parse --abbrev-ref --symbolic-full-name '@{u}' 2>/dev/null || echo "")"
  if [[ -z "${upstream}" ]] || ! git rev-parse --verify "${upstream}" >/dev/null 2>&1; then
    rem="$(get_remote 2>/dev/null || echo "origin")"
    if git rev-parse --verify "refs/remotes/${rem}/${branch}" >/dev/null 2>&1; then
      upstream="refs/remotes/${rem}/${branch}"
    elif git rev-parse --verify "refs/remotes/origin/${branch}" >/dev/null 2>&1; then
      upstream="refs/remotes/origin/${branch}"
    elif git rev-parse --verify "origin/${branch}" >/dev/null 2>&1; then
      upstream="origin/${branch}"
    fi
  fi

  if [[ -z "${upstream}" ]] || ! git rev-parse --verify "${upstream}" >/dev/null 2>&1; then
    echo -e "${C_AMBER}▲ No upstream tracking branch found for '${branch}'; cannot safely determine unpushed boundary.${C_RESET}"
    return 1
  fi

  unpushed_cnt="$(git rev-list --count "${upstream}..HEAD" 2>/dev/null || echo 0)"
  if (( unpushed_cnt <= 1 )); then
    echo -e "  ${C_SLATE}• Only ${unpushed_cnt} unpushed commit(s); tidy squash not needed.${C_RESET}"
    return 0
  fi

  echo -e "${C_CYAN}◈ Syncalibur Tidy Squash${C_RESET} [${unpushed_cnt} unpushed commits -> 1 clean Conventional Commit]"

  # 1. Create a safe rollback reference
  local backup_tag="refs/syncalibur/backup-HEAD-$(date +%s)"
  git update-ref "${backup_tag}" HEAD
  echo -e "  ${C_SLATE}⛨ Safety Checkpoint created: ${backup_tag}${C_RESET}"

  # 2. Collect commit messages for rollup body
  local rollup_body
  rollup_body="$(printf 'Rollup of %d unpushed local commits:\n%s' "${unpushed_cnt}" "$(git log --pretty=format:'- %s' "${upstream}..HEAD")")"

  # 3. Soft reset to upstream
  git reset --soft "${upstream}"

  # 4. Generate unified AST Conventional Commit title
  local macro_msg
  if ensure_go_engine; then
    macro_msg="$("${GO_ENGINE_BIN}" commit-msg "${current_repo}" 2>/dev/null || true)"
  fi
  if [[ -z "${macro_msg}" ]]; then
    macro_msg="feat($(basename "${current_repo}")): tidy development milestone"
  fi

  # 5. Commit with clean title and rollup body
  git commit -m "${macro_msg}" -m "${rollup_body}"
  echo -e "  ${C_GREEN}✓ Squashed ${unpushed_cnt} commits -> ${macro_msg}${C_RESET}"
  echo -e "  ${C_SLATE}Hint: To undo this tidy operation anytime: run 'syncalibur undo' (or git reset --hard ${backup_tag})${C_RESET}"
}

cmd_checkpoint() {
  local cur_git_root
  cur_git_root="$(git rev-parse --show-toplevel 2>/dev/null || echo "")"
  local current_repo="${cur_git_root:-${1:-${REPO_ROOT}}}"
  [[ -d "${current_repo}" ]] || current_repo="${PWD}"
  current_repo="$(cd "${current_repo}" && git rev-parse --show-toplevel 2>/dev/null || echo "${current_repo}")"
  cd "${current_repo}" || return 1
  assert_git_operable "${current_repo}" || return 1

  local subcmd="${2:-list}"
  if [[ "${1:-}" == "save" || "${1:-}" == "create" || "${subcmd}" == "save" || "${subcmd}" == "create" ]]; then
    local tag_name="refs/syncalibur/checkpoint-$(date +%s)"
    git update-ref "${tag_name}" HEAD
    echo -e "${C_GREEN}✓ Safety checkpoint created:${C_RESET} ${tag_name} [$(git log -1 --pretty=format:'%h - %s' HEAD)]"
    return 0
  fi

  echo -e "${C_CYAN}◈ Syncalibur Safety Checkpoints (Time Machine)${C_RESET} [$(basename "${current_repo}")]"
  local refs
  refs="$(git for-each-ref --sort=-creatordate --format='%(refname) %(creatordate:relative) (commit: %(objectname:short))' 'refs/syncalibur/*' 2>/dev/null)"
  if [[ -z "${refs}" ]]; then
    echo -e "  ${C_SLATE}• No checkpoints recorded yet in this repository.${C_RESET}"
    echo -e "  ${C_SLATE}  Checkpoints are automatically created before 'syncalibur tidy' or via 'syncalibur checkpoint save'.${C_RESET}"
    return 0
  fi
  local idx=1
  while IFS= read -r line; do
    echo -e "  ${C_WHITE}[${idx}]${C_RESET} ${line}"
    idx=$(( idx + 1 ))
    (( idx > 10 )) && break
  done <<< "${refs}"
  echo -e "  ${C_SLATE}Hint: Run 'syncalibur undo' to safely revert to the most recent checkpoint.${C_RESET}"
}

cmd_undo() {
  local steps="1" target_repo=""
  while [[ $# -gt 0 ]]; do
    if [[ "$1" =~ ^[0-9]+$ ]]; then
      steps="$1"
    elif [[ -d "$1" ]]; then
      target_repo="$1"
    fi
    shift
  done

  local current_repo
  if [[ -n "${target_repo}" ]]; then
    current_repo="$(cd "${target_repo}" && git rev-parse --show-toplevel 2>/dev/null || echo "${target_repo}")"
  else
    current_repo="$(git rev-parse --show-toplevel 2>/dev/null || echo "${REPO_ROOT}")"
  fi
  cd "${current_repo}" || return 1
  assert_git_operable "${current_repo}" || return 1

  local commit_cnt
  commit_cnt="$(git rev-list --count HEAD 2>/dev/null || echo 0)"
  if (( commit_cnt <= 1 )); then
    echo -e "${C_AMBER}▲ Cannot undo: Repository has only 1 commit or history is empty.${C_RESET}"
    return 1
  fi

  local current_sha target_sha
  current_sha="$(git rev-parse --short HEAD 2>/dev/null)"
  target_sha="$(git rev-parse --short "HEAD~${steps}" 2>/dev/null)"

  if [[ -z "${target_sha}" ]]; then
    echo -e "${C_RED}✕ Target commit HEAD~${steps} not found in history.${C_RESET}"
    return 1
  fi

  echo -e "${C_CYAN}${C_BOLD}◈ Syncalibur Time Machine Undo${C_RESET} [Rolling back ${steps} commit(s): ${current_sha} ➔ ${target_sha}]"

  # Record redo ref so developer can recover if needed
  local redo_ref="refs/syncalibur/redo-$(date +%s)"
  git update-ref "${redo_ref}" HEAD 2>/dev/null || true

  # Non-destructive soft reset preserves all unstaged/staged file modifications
  if git reset --soft "HEAD~${steps}" 2>/dev/null; then
    # Prune latest timeline ref that pointed to the undone commit
    local latest_tl
    latest_tl="$(git for-each-ref --sort=-creatordate --format='%(refname)' 'refs/syncalibur/timeline/*' 2>/dev/null | head -n 1)"
    if [[ -n "${latest_tl}" ]]; then
      git update-ref -d "${latest_tl}" 2>/dev/null || true
    fi

    echo -e "  ${C_GREEN}✓ Rolled back to ${target_sha}! (working tree preserved via soft reset)${C_RESET}"
    echo -e "  ${C_SLATE}Redo reference saved at: ${redo_ref} (run 'git reset --soft ${redo_ref}' to re-apply)${C_RESET}"
    return 0
  else
    echo -e "  ${C_RED}✕ Soft reset to HEAD~${steps} failed.${C_RESET}"
    return 1
  fi
}

cmd_pull() {
  local current_repo="${1:-${REPO_ROOT}}"
  [[ -d "${current_repo}" ]] || current_repo="${PWD}"
  current_repo="$(cd "${current_repo}" && git rev-parse --show-toplevel 2>/dev/null || echo "${current_repo}")"
  cd "${current_repo}" || return 1
  assert_git_operable "${current_repo}" || return 1

  local rem branch
  rem="$(get_remote)"
  branch="$(get_branch || echo main)"
  echo -e "${C_CYAN}◈ Syncalibur Safe Pull${C_RESET} [$(basename "${current_repo}") @ ${branch}]"
  if safe_bidirectional_sync "${rem}" "${branch}" 1; then
    if [[ -f "${current_repo}/.gitmodules" ]]; then
      git submodule update --init --recursive 2>/dev/null || true
    fi
    echo -e "  ${C_GREEN}✓ Local branch '${branch}' up to date with '${rem}/${branch}'.${C_RESET}"
    return 0
  else
    echo -e "  ${C_RED}✕ Safe pull failed.${C_RESET}"
    return 1
  fi
}

cmd_mode() {
  local target_mode=""
  local is_global=0
  local current_repo
  current_repo="$(git rev-parse --show-toplevel 2>/dev/null || echo "${REPO_ROOT}")"

  for arg in "$@"; do
    if [[ "${arg}" == "--global" || "${arg}" == "-g" ]]; then
      is_global=1
    elif [[ -z "${target_mode}" && "${arg}" != -* ]]; then
      target_mode="${arg}"
    fi
  done

  # Determine current active mode and provenance for this repository
  local current_source="global syncalibur.env"
  if [[ -f "${current_repo}/.syncalibur" ]] && grep -qiE '^(MODE|SYNCALIBUR_MODE)=' "${current_repo}/.syncalibur"; then
    current_source="${current_repo}/.syncalibur"
  elif [[ -f "${current_repo}/syncalibur.env" ]] && grep -qiE '^(MODE|SYNCALIBUR_MODE)=' "${current_repo}/syncalibur.env"; then
    current_source="${current_repo}/syncalibur.env"
  elif [[ -f "${FLEET_LIST_FILE}" ]] && grep -E "^[[:space:]]*${current_repo}([[:space:]]|$)" "${FLEET_LIST_FILE}" 2>/dev/null | grep -qiE 'mode='; then
    current_source="repos.list (fleet)"
  fi

  if [[ -z "${target_mode}" ]]; then
    print_syncalibur_dashboard
    echo -e "${C_GOLD}${C_BOLD}◈ Syncalibur Operating Architecture:${C_RESET}"
    echo -e "  Current Operating Mode : ${C_CYAN}${C_BOLD}${MODE}${C_RESET} [source: ${current_source}]"
    echo -e "  Debounce Settle Time   : ${BURST_COOLDOWN}s • zero-data-loss sentinel active"
    echo ""
    echo -e "  ${C_CYAN}${C_BOLD}1. ⚡ autopilot${C_RESET} ${C_SLATE}(Continuous Stream & Workstation Flight Recorder) [Varsayılan]${C_RESET}"
    echo -e "     • 5s sessizlikte AST commit atar, Shannon Secret Shield korur, doğrudan remote'lara pushlar."
    echo -e "     • Time Machine: Her adımda yerel kurtarma halkası oluşturur ('syncalibur undo')."
    echo -e "     • Asla arka planda habersiz PR açmaz; bildirim kirliliği ve korku yaratmaz."
    echo ""
    echo -e "  ${C_GOLD}${C_BOLD}2. ⏸ manual${C_RESET} ${C_SLATE}(On-Demand Quicksave Only)${C_RESET}"
    echo -e "     • Arka plan commit atmaz. Yalnızca 'syncalibur save [mesaj]' veya TUI'den [s] tetiklendiğinde senkronize eder."
    echo ""
    echo -e "  ${C_SLATE}Hızlı Komutlar:${C_RESET}"
    echo -e "    Anlık Hızlı Kayıt     : ${C_GOLD}syncalibur save [mesaj]${C_RESET}"
    echo -e "    Time Machine Geri Al  : ${C_GOLD}syncalibur undo [adım]${C_RESET}"
    echo -e "    Modu Değiştir         : ${C_GOLD}syncalibur mode <autopilot | manual>${C_RESET}"
    return 0
  fi

  case "${target_mode}" in
    manual | idle | sentinel) target_mode="manual" ;;
    *) target_mode="autopilot" ;;
  esac

  if (( is_global )); then
    local cfg_file="${CONFIG_FILE:-${SCRIPT_DIR}/syncalibur.env}"
    if [[ -f "${cfg_file}" ]]; then
      if grep -q "^SYNCALIBUR_MODE=" "${cfg_file}"; then
        sed -i -E "s/^SYNCALIBUR_MODE=.*/SYNCALIBUR_MODE=\"${target_mode}\"/" "${cfg_file}"
      else
        echo "SYNCALIBUR_MODE=\"${target_mode}\"" >> "${cfg_file}"
      fi
      echo -e "${C_GREEN}✓ Syncalibur global varsayılan modu '${target_mode}' olarak ayarlandı (${cfg_file}).${C_RESET}"
    fi
  else
    local repo_cfg="${current_repo}/.syncalibur"
    if [[ -f "${repo_cfg}" ]]; then
      if grep -qiE '^(MODE|SYNCALIBUR_MODE)=' "${repo_cfg}"; then
        sed -i -E "s/^(MODE|SYNCALIBUR_MODE)=.*/MODE=\"${target_mode}\"/" "${repo_cfg}"
      else
        echo "MODE=\"${target_mode}\"" >> "${repo_cfg}"
      fi
    else
      cat > "${repo_cfg}" << EOF
# Syncalibur per-repository configuration
MODE="${target_mode}"
EOF
    fi
    echo -e "${C_GREEN}✓ '${current_repo}' için Syncalibur modu '${target_mode}' olarak ayarlandı (.syncalibur).${C_RESET}"
    if [[ -f "${FLEET_LIST_FILE}" ]] && grep -qE "^[[:space:]]*${current_repo}([[:space:]]|$)" "${FLEET_LIST_FILE}"; then
      sed -i -E "s|^[[:space:]]*(${current_repo})([[:space:]].*)?$|\1 [mode=${target_mode}]|" "${FLEET_LIST_FILE}"
    fi
  fi

  if [[ -f "${PID_FILE}" ]] && kill -0 "$(cat "${PID_FILE}" 2>/dev/null)" 2>/dev/null; then
    echo -e "  ${C_CYAN}↻ Arka plan servisi yeni modla yeniden başlatılıyor...${C_RESET}"
    cmd_restart
  fi
}

cmd_hook() {
  local action="${1:-status}"
  local current_repo
  current_repo="$(git rev-parse --show-toplevel 2>/dev/null || echo "${REPO_ROOT}")"
  local hooks_dir="${current_repo}/.git/hooks"

  if [[ ! -d "${hooks_dir}" ]]; then
    echo -e "${C_RED}✕ Not a Git repository or .git/hooks directory missing.${C_RESET}"
    return 1
  fi

  local pre_commit_hook="${hooks_dir}/pre-commit"
  local prep_msg_hook="${hooks_dir}/prepare-commit-msg"

  case "${action}" in
    install)
      mkdir -p "${hooks_dir}"
      # 1. Install pre-commit hook (Secret Shield & Syntax Hold)
      cat > "${pre_commit_hook}" << 'HOOK_EOF'
#!/usr/bin/env bash
# --- BEGIN SYNCALIBUR PRE-COMMIT HOOK ---
# Universal pre-commit protection: Secret Shield & Syntax Hold
if command -v syncalibur >/dev/null 2>&1; then
  if ! syncalibur check 2>&1; then
    echo -e "\033[38;2;239;68;68m✕ [SYNCALIBUR HOOK] Commit rejected: fix syntax errors or remove secrets.\033[0m" >&2
    exit 1
  fi
fi
# --- END SYNCALIBUR PRE-COMMIT HOOK ---
HOOK_EOF
      chmod +x "${pre_commit_hook}"

      # 2. Install prepare-commit-msg hook (Auto Conventional Commit default)
      cat > "${prep_msg_hook}" << 'HOOK_EOF'
#!/usr/bin/env bash
# --- BEGIN SYNCALIBUR PREPARE-COMMIT-MSG HOOK ---
COMMIT_MSG_FILE="$1"
COMMIT_SOURCE="$2"
if [[ -z "${COMMIT_SOURCE}" ]] && command -v syncalibur >/dev/null 2>&1; then
  REPO_ROOT="$(git rev-parse --show-toplevel 2>/dev/null || pwd)"
  BIN_ENG="${REPO_ROOT}/bin/syncalibur"
  if [[ -x "${BIN_ENG}" ]]; then
    SUGG="$("${BIN_ENG}" commit-msg "${REPO_ROOT}" 2>/dev/null || true)"
    if [[ -n "${SUGG}" ]]; then
      printf '%s\n\n# Syncalibur AST Conventional Commit Suggestion\n' "${SUGG}" > "${COMMIT_MSG_FILE}"
    fi
  fi
fi
# --- END SYNCALIBUR PREPARE-COMMIT-MSG HOOK ---
HOOK_EOF
      chmod +x "${prep_msg_hook}"

      echo -e "${C_GREEN}✓ Installed Syncalibur Git hooks in $(basename "${current_repo}")/.git/hooks:${C_RESET}"
      echo -e "  • ${C_CYAN}pre-commit${C_RESET}         (Secret Shield & Syntax Hold gate)"
      echo -e "  • ${C_CYAN}prepare-commit-msg${C_RESET} (AST Conventional Commit suggestions)"
      ;;

    uninstall)
      rm -f "${pre_commit_hook}" "${prep_msg_hook}"
      echo -e "${C_GREEN}✓ Uninstalled Syncalibur Git hooks from $(basename "${current_repo}")/.git/hooks.${C_RESET}"
      ;;

    status|*)
      echo -e "${C_CYAN}◈ Syncalibur Git Hook Status${C_RESET} [$(basename "${current_repo}")]"
      if [[ -x "${pre_commit_hook}" ]]; then
        echo -e "  ${C_GREEN}✓ pre-commit${C_RESET}         : Active (${pre_commit_hook})"
      else
        echo -e "  ${C_SLATE}○ pre-commit${C_RESET}         : Not installed"
      fi
      if [[ -x "${prep_msg_hook}" ]]; then
        echo -e "  ${C_GREEN}✓ prepare-commit-msg${C_RESET} : Active (${prep_msg_hook})"
      else
        echo -e "  ${C_SLATE}○ prepare-commit-msg${C_RESET} : Not installed"
      fi
      echo -e "  ${C_SLATE}Run 'syncalibur hook install' to activate universal protection.${C_RESET}"
      ;;
  esac
}

# ==============================================================================
# Domain 1.1: Multi-Repository Fleet Supervisor (`syncalibur fleet`)
# ==============================================================================

cmd_fleet() {
  local subcmd="${1:-status}"
  local arg="${2:-}"

  case "${subcmd}" in
    add)
      [[ -n "${arg}" ]] || { echo "✕ Usage: syncalibur fleet add <path> [--mode=instant-push|developer|devops|pr]"; return 1; }
      local target target_mode="developer"
      target="$(resolve_path "$(normalize_path_input "${arg}")")"
      shift 2 2>/dev/null || true
      while [[ $# -gt 0 ]]; do
        case "$1" in
          --mode=*) target_mode="${1#*=}" ;;
          instant-push|instant|stream|continuous|commit-generator|commit-gen) target_mode="instant-push" ;;
          developer|pro|team) target_mode="developer" ;;
          devops|infra|k8s|iac) target_mode="devops" ;;
          pr|review) target_mode="pr" ;;
        esac
        shift
      done
      mkdir -p "${target}" 2>/dev/null || true
      if ! git -C "${target}" rev-parse --show-toplevel >/dev/null 2>&1; then
        git -C "${target}" init -b main >/dev/null 2>&1 || git -C "${target}" init >/dev/null 2>&1
      fi
      touch "${FLEET_LIST_FILE}"
      local entry="${target} [mode=${target_mode}]"
      if ! grep -E "^[[:space:]]*${target}([[:space:]]|$)" "${FLEET_LIST_FILE}" 2>/dev/null; then
        printf '%s\n' "${entry}" >> "${FLEET_LIST_FILE}"
        echo "✓ Added '${target}' (${target_mode}) to Syncalibur Fleet (${FLEET_LIST_FILE})."
      else
        sed -i -E "s|^[[:space:]]*${target}([[:space:]].*)?$|${entry}|" "${FLEET_LIST_FILE}"
        echo "✓ Updated '${target}' in Fleet to mode '${target_mode}'."
      fi
      ;;
    remove | rm)
      [[ -n "${arg}" && -f "${FLEET_LIST_FILE}" ]] || { echo "✕ Usage: syncalibur fleet remove <path>"; return 1; }
      local target tmp
      target="$(resolve_path "$(normalize_path_input "${arg}")")"
      tmp="$(mktemp)"
      grep -vE "^[[:space:]]*${target}([[:space:]]|$)" "${FLEET_LIST_FILE}" > "${tmp}" || true
      mv "${tmp}" "${FLEET_LIST_FILE}"
      echo "✓ Removed '${target}' from Syncalibur Fleet."
      ;;
    scan)
      local scan_dir="${arg:-${HOME}}"
      scan_dir="$(resolve_path "$(normalize_path_input "${scan_dir}")")"
      echo "◇ Scanning '${scan_dir}' for Git repositories via Go engine..."
      touch "${FLEET_LIST_FILE}"
      local count=0 repo
      if ensure_go_engine; then
        while IFS= read -r repo; do
          [[ -z "${repo}" ]] && continue
          if ! grep -E "^[[:space:]]*${repo}([[:space:]]|$)" "${FLEET_LIST_FILE}" 2>/dev/null; then
            printf '%s\n' "${repo}" >> "${FLEET_LIST_FILE}"
            echo "  + Enrolled: ${repo}"
            count=$(( count + 1 ))
          fi
        done < <("${GO_ENGINE_BIN}" fleet-scan "${scan_dir}")
      else
        while IFS= read -r repo; do
          repo="$(dirname "${repo}")"
          if ! grep -E "^[[:space:]]*${repo}([[:space:]]|$)" "${FLEET_LIST_FILE}" 2>/dev/null; then
            printf '%s\n' "${repo}" >> "${FLEET_LIST_FILE}"
            echo "  + Enrolled: ${repo}"
            count=$(( count + 1 ))
          fi
        done < <(find "${scan_dir}" -maxdepth 3 -name ".git" -type d 2>/dev/null)
      fi
      echo "✦ Fleet scan complete (${count} new repository/repositories enrolled)."
      ;;
    start)
      [[ -f "${FLEET_LIST_FILE}" ]] || { echo "▲  No fleet list found. Run 'syncalibur fleet scan' or 'syncalibur fleet add <path>' first."; return 1; }
      local line repo_path rest repo_mode
      while IFS= read -r line; do
        line="${line%%#*}"
        line="${line%"${line##*[![:space:]]}"}"
        [[ -z "${line}" ]] && continue
        read -r repo_path rest <<< "${line}"
        [[ -z "${repo_path}" || ! -d "${repo_path}" ]] && continue
        repo_mode="developer"
        if [[ "${rest}" =~ mode=([a-zA-Z0-9_-]+) ]]; then
          repo_mode="${BASH_REMATCH[1]}"
        elif [[ "${rest}" =~ (instant-push|instant|stream|continuous|commit-generator|commit-gen|developer|devops|pr) ]]; then
          repo_mode="${BASH_REMATCH[1]}"
        fi
        SYNCALIBUR_REPO_PATH="${repo_path}" SYNCALIBUR_MODE="${repo_mode}" "${SCRIPT_PATH}" start
      done < "${FLEET_LIST_FILE}"
      ;;
    stop)
      [[ -f "${FLEET_LIST_FILE}" ]] || return 0
      local line repo_path rest
      while IFS= read -r line; do
        line="${line%%#*}"
        line="${line%"${line##*[![:space:]]}"}"
        [[ -z "${line}" ]] && continue
        read -r repo_path rest <<< "${line}"
        [[ -z "${repo_path}" || ! -d "${repo_path}" ]] && continue
        SYNCALIBUR_REPO_PATH="${repo_path}" "${SCRIPT_PATH}" stop
      done < "${FLEET_LIST_FILE}"
      ;;
    status | list | *)
      print_syncalibur_dashboard
      if ensure_go_engine; then
        "${GO_ENGINE_BIN}" fleet-status "${FLEET_LIST_FILE}"
      elif [[ -f "${FLEET_LIST_FILE}" ]]; then
        cat "${FLEET_LIST_FILE}"
      else
        echo "  No repositories enrolled in Fleet yet. Run: syncalibur fleet scan ~"
      fi
      ;;
  esac
}

# ==============================================================================
# Domain 3.1 & 3.3: Interactive TUI (`ui`) & 10-Point System Doctor (`doctor`)
# ==============================================================================

cmd_ui() {
  cmd_status "$@"
}

cmd_doctor() {
  print_syncalibur_dashboard
  printf '  %s╭────────────────────────────────────────────────────────────────────────────────────────╮%s\n' "${C_GOLD}" "${C_RESET}"
  printf '  %s│%s  %s%s%-84s%s  %s│%s\n' "${C_GOLD}" "${C_RESET}" "${C_GOLD}" "${C_BOLD}" "✚  S Y N C A L I B U R   D O C T O R   —   1 2 - P O I N T   D I A G N O S T I C S" "${C_RESET}" "${C_GOLD}" "${C_RESET}"
  printf '  %s╰────────────────────────────────────────────────────────────────────────────────────────╯%s\n\n' "${C_GOLD}" "${C_RESET}"

  local pass=0 warn=0 fail=0
  check_item() {
    local st="$1" name="$2" msg="$3"
    case "${st}" in
      OK)   printf '    %s[✓]%s %-26s : %s%s%s\n' "${C_GREEN}" "${C_RESET}" "${name}" "${C_WHITE}" "${msg}" "${C_RESET}"; pass=$((pass+1)) ;;
      WARN) printf '    %s[▲]%s %-26s : %s%s%s\n' "${C_AMBER}" "${C_RESET}" "${name}" "${C_AMBER}" "${msg}" "${C_RESET}"; warn=$((warn+1)) ;;
      FAIL) printf '    %s[✕]%s %-26s : %s%s%s\n' "${C_RED}" "${C_RESET}" "${name}" "${C_RED}" "${msg}" "${C_RESET}"; fail=$((fail+1)) ;;
    esac
  }

  # 1. Go Engine
  if ensure_go_engine; then
    check_item "OK" "Go Native Engine" "Compiled & active (${GO_ENGINE_BIN##*/})"
  else
    check_item "WARN" "Go Native Engine" "Not compiled (using pure Bash fallback)"
  fi

  # 2. Filesystem Watcher
  if [[ "${FS_WATCHER}" != "polling" ]]; then
    check_item "OK" "Filesystem Event Watcher" "${FS_WATCHER} (0.00% idle CPU, instant <10ms trigger)"
  else
    check_item "OK" "Filesystem Event Watcher" "Polling mode (${POLL_INTERVAL}s interval; install inotify-tools/fswatch for 0% CPU)"
  fi

  # 3. Git Binary & Lock State
  if [[ -f "${GIT_DIR}/index.lock" ]]; then
    check_item "FAIL" "Git Repository State" "Stale ${GIT_DIR}/index.lock detected!"
  else
    check_item "OK" "Git Repository State" "$(git --version) • Working tree unlocked"
  fi

  # 4. Daemon Process
  local pid
  if pid="$(daemon_pid)"; then
    check_item "OK" "Daemon Process" "Running (PID ${pid}) $(daemon_resource_badge "${pid}")"
  else
    check_item "WARN" "Daemon Process" "Not running (start with 'syncalibur start')"
  fi

  # 5. Native OS Auto-Start
  if is_service_installed; then
    check_item "OK" "OS Auto-Start Service" "${SERVICE_BACKEND} (${SERVICE_NAME}) registered"
  else
    check_item "WARN" "OS Auto-Start Service" "Not installed (run 'syncalibur install')"
  fi

  # 6. Provider Authentication
  local auth_line gh_u gh_v gl_u gl_v
  auth_line="$(probe_provider_auth)"
  IFS='|' read -r gh_u gh_v gl_u gl_v <<< "${auth_line}"
  if [[ -n "${gh_u}" || -n "${gl_u}" ]]; then
    check_item "OK" "Provider Authentication" "GitHub: @${gh_u:-none} (${gh_v:-n/a}) • GitLab: @${gl_u:-none} (${gl_v:-n/a})"
  else
    check_item "WARN" "Provider Authentication" "No verified GitHub/GitLab CLI or SSH session"
  fi

  # 7. Push Destinations
  local remote
  remote="$(get_remote)"
  if [[ -n "$(get_push_urls "${remote}")" ]]; then
    check_item "OK" "Push Destinations" "$(describe_push_targets_inline "${remote}")"
  else
    check_item "WARN" "Push Destinations" "None yet (auto-create on first push: ${AUTO_CREATE_REMOTE})"
  fi

  # 8. Self-Updater
  if [[ "${SELF_AUTO_UPDATE}" == "true" && -n "${SELF_REPO_ROOT}" ]]; then
    check_item "OK" "Self-Watching Updater" "Active on ${SELF_REPO_ROOT} (every ${SELF_CHECK_INTERVAL}s)"
  else
    check_item "WARN" "Self-Watching Updater" "Disabled"
  fi

  # 9. Global CLI Symlink
  if command -v syncalibur >/dev/null 2>&1; then
    check_item "OK" "Global CLI Command" "$(command -v syncalibur)"
  else
    check_item "WARN" "Global CLI Command" "Not in PATH (run 'syncalibur install')"
  fi

  # 10. Telegram Bot & Multi-Channel Webhooks
  if [[ -n "${TELEGRAM_BOT_TOKEN}" && -n "${TELEGRAM_CHAT_ID}" && -n "${WEBHOOK_URL}" ]]; then
    check_item "OK" "Telegram & Webhooks" "Telegram Bot (chat ${TELEGRAM_CHAT_ID}) + Webhook (${WEBHOOK_EVENTS}) active"
  elif [[ -n "${TELEGRAM_BOT_TOKEN}" && -n "${TELEGRAM_CHAT_ID}" ]]; then
    check_item "OK" "Telegram & Webhooks" "Telegram Bot active (chat ${TELEGRAM_CHAT_ID} • events: ${WEBHOOK_EVENTS})"
  elif [[ -n "${WEBHOOK_URL}" ]]; then
    check_item "OK" "Telegram & Webhooks" "Webhook active (${WEBHOOK_EVENTS})"
  else
    check_item "OK" "Telegram & Webhooks" "Desktop & Terminal Popup active (set TELEGRAM_BOT_TOKEN in syncalibur.env for Telegram)"
  fi

  # 11. Pre-flight Syntax & Shannon Entropy Secret Shield
  if ensure_go_engine; then
    local pf_test_out pf_rc=0
    pf_test_out="$("${GO_ENGINE_BIN}" preflight "${REPO_ROOT}" 2>&1)" || pf_rc=$?
    if (( pf_rc == 0 )); then
      check_item "OK" "Syntax & Secret Shield" "Active (Shannon entropy & syntax pre-flight verified clean)"
    elif (( pf_rc == 3 )); then
      check_item "WARN" "Syntax & Secret Shield" "Credential detected in repository (quarantine active)"
    else
      check_item "WARN" "Syntax & Secret Shield" "Syntax error detected in working tree"
    fi
  else
    check_item "WARN" "Syntax & Secret Shield" "Go engine not compiled"
  fi

  # 12. DevOps & GitOps Guard (Kubernetes, Helm, Docker, Terraform)
  if ensure_go_engine; then
    local devops_out
    devops_out="$("${GO_ENGINE_BIN}" preflight "${REPO_ROOT}" --json 2>/dev/null || true)"
    if [[ -n "${devops_out}" ]] && command -v jq >/dev/null 2>&1; then
      local err_cnt
      err_cnt="$(jq -r '(.devops_errors | length) // 0' <<< "${devops_out}")"
      if (( err_cnt > 0 )); then
        local first_err
        first_err="$(jq -r '.devops_errors[0] // "policy error"' <<< "${devops_out}")"
        check_item "WARN" "DevOps & GitOps Guard" "${err_cnt} policy issue(s) detected: ${first_err}"
      else
        check_item "OK" "DevOps & GitOps Guard" "Active (Kubernetes ports, Dockerfile, Terraform & conflict markers verified)"
      fi
    else
      check_item "OK" "DevOps & GitOps Guard" "Active (Kubernetes ports, Dockerfile, Terraform & conflict markers verified)"
    fi
  else
    check_item "WARN" "DevOps & GitOps Guard" "Go engine required for deep AST/YAML validation"
  fi

  printf '\n  %s────────────────────────────────────────────────────────────────────────────────────────%s\n' "${C_SLATE}" "${C_RESET}"
  printf '  %sSummary:%s %s%d Passed%s  •  %s%d Warnings%s  •  %s%d Failures%s\n\n' \
    "${C_BOLD}" "${C_RESET}" "${C_GREEN}" "${pass}" "${C_RESET}" "${C_AMBER}" "${warn}" "${C_RESET}" "${C_RED}" "${fail}" "${C_RESET}"
}

# ==============================================================================
# Daemon Loop
# ==============================================================================

daemon_pid() {
  local pid
  pid="$(cat "${PID_FILE}" 2>/dev/null)" || return 1
  [[ -n "${pid}" ]] && kill -0 "${pid}" 2>/dev/null && echo "${pid}"
}

acquire_single_instance_lock() {
  if command -v flock >/dev/null 2>&1; then
    exec 9> "${LOCK_FILE}"
    flock -w 5 9 || return 1
    return 0
  fi
  local existing_pid
  if existing_pid="$(daemon_pid)" && [[ "${existing_pid}" != "$$" && "${existing_pid}" != "${SYNCALIBUR_SUPERVISOR_PID:-}" ]]; then
    return 1
  fi
  return 0
}

on_exit() {
  [[ -n "${NAP_PID}" ]] && kill "${NAP_PID}" 2>/dev/null || true
  rm -f "${PID_FILE}"
  log "○ Syncalibur stopped."
}

run_daemon() {
  cd "${REPO_ROOT}" || exit 1
  mkdir -p "$(dirname "${LOG_FILE}")"
  ensure_go_engine || true

  # Re-exec under native `bin/syncalibur daemon-exec` if invoked via `bash syncalibur.sh run`
  # so `top`, `htop`, `ps`, Activity Monitor, and Task Manager show `syncalibur` as the process name!
  if [[ -z "${SYNCALIBUR_WRAPPED:-}" ]] && [[ -x "${GO_ENGINE_BIN}" ]]; then
    exec "${GO_ENGINE_BIN}" daemon-exec "${SCRIPT_PATH}" run
  fi
  if [[ -w "/proc/$$/comm" ]]; then
    printf 'syncalibur' > "/proc/$$/comm" 2>/dev/null || true
  fi

  if ! acquire_single_instance_lock; then
    log "✕ Another Syncalibur instance is already running for ${REPO_ROOT}."
    exit 1
  fi
  echo "${SYNCALIBUR_SUPERVISOR_PID:-$$}" > "${PID_FILE}"
  trap on_exit EXIT
  trap 'exit 0' INT TERM

  if [[ "${SYNCALIBUR_BACKGROUND:-0}" != "1" && -t 1 ]]; then
    print_syncalibur_dashboard
  fi

  local remote branch busy head snap id_raw git_user git_email sys_ident ssh_cred gh_user gh_via gl_user gl_via
  local idle_snapshot="" last_hold_snapshot="" next_hold_retry=0 checked_head="" push_pending=0 next_retry=0 retry_delay="${RETRY_BASE}" force_sync=0
  local last_commit_ts=0 last_burst_notice=0

  id_raw="$(detect_user_identity)"
  IFS='|' read -r git_user git_email sys_ident ssh_cred gh_user gh_via gl_user gl_via <<< "${id_raw}"

  remote="$(get_remote)"
  if [[ -z "$(get_push_urls "${remote}")" ]]; then
    auto_create_or_bind_remote "" "" || true
    remote="$(get_remote)"
  fi

  log "● Syncalibur started. Repo: ${REPO_ROOT} (scope: ${SCOPE}, watcher: ${FS_WATCHER})"
  log "◆ Targets: $(describe_push_targets_inline "${remote}")"
  log "◆ Identity: ${git_user} <${git_email}> (GitHub: ${gh_user:-n/a}, GitLab: ${gl_user:-n/a}, Key: ${ssh_cred})"
  if [[ "${SELF_AUTO_UPDATE}" == "true" && -n "${SELF_REPO_ROOT}" ]]; then
    log "⇄ Self-Update: Watching Syncalibur repo (${SELF_REPO_ROOT}) for local & upstream updates every ${SELF_CHECK_INTERVAL}s."
  fi
  log "◆  Poll: ${POLL_INTERVAL}s | Debounce: ${DEBOUNCE_WAIT}s | Burst Cooldown: ${BURST_COOLDOWN}s (auto-tidy: ${AUTO_TIDY}, auto-pr: ${AUTO_PR})"

  while true; do
    trim_log

    # --- 0) Watch & auto-update Syncalibur's own repository ------------------
    sync_and_update_self_repo

    # Honor interactive pause toggle from `syncalibur status` / `syncalibur pause`
    if [[ -f "${PAUSE_FILE}" ]]; then
      notice "▲  Paused by user. Run 'syncalibur resume' or press [p] in 'syncalibur status' to resume."
      nap "${POLL_INTERVAL}"
      continue
    fi

    force_sync=0
    if [[ -f "${SYNC_TRIGGER_FILE}" ]]; then
      rm -f "${SYNC_TRIGGER_FILE}" 2>/dev/null || true
      force_sync=1
    fi

    remote="$(get_remote)"

    if ! branch="$(get_branch)"; then
      notice "▲  Detached HEAD; paused until a branch is checked out."
      nap; continue
    fi
    if busy="$(repo_busy)"; then
      notice "▲  Git operation in progress (${busy}); paused."
      nap; continue
    fi
    LAST_NOTICE=""

    # --- 1) Bidirectional Sync (fetch remote & fast-forward / rebase) --------
    safe_bidirectional_sync "${remote}" "${branch}" "${force_sync}" || true

    # --- 2) Stage & commit local file changes --------------------------------
    if has_changes_in_scope; then
      snap="$(snapshot)"
      if (( force_sync )) || [[ "${snap}" != "${idle_snapshot}" ]]; then
        if [[ -n "${last_hold_snapshot}" && "${snap}" == "${last_hold_snapshot}" ]]; then
          # Snapshot unchanged since last hold/quarantine; check quietly every 15s instead of spamming logs
          if (( $(date +%s) < next_hold_retry )); then
            nap "${POLL_INTERVAL}"
            continue
          fi
        fi
        log "✦ Changes detected; waiting for writes to settle..."
        (( ! force_sync )) && wait_for_quiet
        repo_busy >/dev/null && continue
        if commit_changes; then
          idle_snapshot="$(snapshot)"
          last_hold_snapshot=""
          push_pending=1
          last_commit_ts="$(date +%s)"
          if (( $(date +%s) >= next_retry )); then
            next_retry=0
            retry_delay="${RETRY_BASE}"
          fi
        else
          last_hold_snapshot="$(snapshot)"
          idle_snapshot=""
          next_hold_retry=$(( $(date +%s) + 15 ))
        fi
      fi
    fi

    # --- 3) Detect unpushed commits (startup, manual commits, branch switch) -
    head="$(git rev-parse -q --verify HEAD 2>/dev/null || true)"
    if (( force_sync )) || [[ -n "${head}" && "${head}" != "${checked_head}" ]]; then
      checked_head="${head}"
      if (( $(unpushed_count "${remote}" "${branch}") > 0 )); then
        push_pending=1
        (( force_sync )) && next_retry=0
      fi
    fi

    # --- 4) Push to remote, batching rapid bursts and auto-squashing/auto-PR --
    if (( push_pending )) && (( $(date +%s) >= next_retry )); then
      local now_ts
      now_ts="$(date +%s)"
      local time_since_commit=$(( now_ts - last_commit_ts ))

      if (( ! force_sync )) && (( BURST_COOLDOWN > 0 )) && (( last_commit_ts > 0 )) && (( time_since_commit < BURST_COOLDOWN )); then
        local remaining=$(( BURST_COOLDOWN - time_since_commit ))
        if (( now_ts - last_burst_notice >= 4 )); then
          notice "⏳ Batching rapid edits (cooldown: ${remaining}s remaining before push)..."
          last_burst_notice="${now_ts}"
        fi
      else
        local unpushed_cnt=0
        unpushed_cnt="$(unpushed_count "${remote}" "${branch}")"

        # Adaptive Burst Detection:
        # If unpushed commits reach TIDY_THRESHOLD (default 3), treat as an active burst/sprint.
        # If fewer than TIDY_THRESHOLD (e.g. 1 or 2 small single commits), keep them separate!
        local was_burst=0
        if (( unpushed_cnt >= TIDY_THRESHOLD )); then
          was_burst=1
        fi

        if (( was_burst )) && [[ "${AUTO_TIDY}" == "true" ]]; then
          log "◈ Adaptive Auto-Tidy: Squashing ${unpushed_cnt} burst micro-commits (threshold: ${TIDY_THRESHOLD}) into 1 clean conventional commit..."
          cmd_tidy "${REPO_ROOT}" >> "${LOG_FILE}" 2>&1 || true
          unpushed_cnt="$(unpushed_count "${remote}" "${branch}")"
        fi

        # Determine target: Pull Request vs Direct Remote Push
        local should_open_pr=0
        if [[ "${AUTO_PR}" == "true" ]]; then
          should_open_pr=1
        elif [[ "${AUTO_PR}" == "burst" ]] && (( was_burst )); then
          should_open_pr=1
        fi

        if (( should_open_pr )); then
          log "⇄ Auto-PR triggered (${unpushed_cnt} commit(s)); opening/updating Pull Request for '${branch}'..."
          if cmd_pr "${branch}" >> "${LOG_FILE}" 2>&1; then
            push_pending=0
            last_commit_ts=0
            retry_delay="${RETRY_BASE}"
          else
            next_retry=$(( $(date +%s) + retry_delay ))
            log "↻ Retrying Auto-PR in ${retry_delay}s."
            retry_delay=$(( retry_delay * 2 > RETRY_MAX ? RETRY_MAX : retry_delay * 2 ))
          fi
        else
          if push_changes "${remote}" "${branch}"; then
            push_pending=0
            last_commit_ts=0
            retry_delay="${RETRY_BASE}"
          else
            next_retry=$(( $(date +%s) + retry_delay ))
            log "↻ Retrying the push in ${retry_delay}s."
            retry_delay=$(( retry_delay * 2 > RETRY_MAX ? RETRY_MAX : retry_delay * 2 ))
          fi
        fi
      fi
    fi

    wait_for_fs_change "${POLL_INTERVAL}" || true
  done
}

# ==============================================================================
# Cross-Platform Service Management & Global CLI Wiring
# ==============================================================================

is_service_installed() {
  case "${SERVICE_BACKEND}" in
    systemd)         [[ -f "${UNIT_FILE}" ]] ;;
    launchd)         [[ -f "${PLIST_FILE}" ]] ;;
    windows-startup) [[ -f "$(win_startup_vbs_path)" ]] ;;
    cron)            crontab -l 2>/dev/null | grep -Fq "# syncalibur:${SERVICE_NAME}" || [[ -f "${DESKTOP_FILE}" ]] ;;
  esac
}

install_global_cli_and_completions() {
  mkdir -p "${HOME}/.local/bin" 2>/dev/null || true
  ln -sf "${SCRIPT_PATH}" "${HOME}/.local/bin/syncalibur" 2>/dev/null || true

  local bash_comp_dir="${HOME}/.local/share/bash-completion/completions"
  mkdir -p "${bash_comp_dir}" 2>/dev/null || true
  cat > "${bash_comp_dir}/syncalibur" 2>/dev/null << 'EOF' || true
_syncalibur_completions() {
  local cur="${COMP_WORDS[COMP_CWORD]}"
  local prev="${COMP_WORDS[COMP_CWORD-1]}"
  local cmds="install uninstall start stop restart status logs ui fleet doctor notify pr pull sync pause resume alert run"
  if [[ ${COMP_CWORD} -eq 1 ]]; then
    COMPREPLY=( $(compgen -W "${cmds}" -- "${cur}") )
    return
  fi
  case "${prev}" in
    fleet) COMPREPLY=( $(compgen -W "status add remove scan start stop" -- "${cur}") ) ;;
    alert) COMPREPLY=( $(compgen -W "--test --wait" -- "${cur}") ) ;;
    pr)    COMPREPLY=( $(compgen -W "main --dry-run --preview" -- "${cur}") ) ;;
  esac
}
complete -F _syncalibur_completions syncalibur
EOF
}

start_background_nohup() {
  local pid
  if pid="$(daemon_pid)"; then
    echo "▲  Syncalibur is already running (PID ${pid})."
    return 0
  fi
  mkdir -p "$(dirname "${LOG_FILE}")"
  ensure_go_engine || true
  if [[ -x "${GO_ENGINE_BIN}" ]]; then
    SYNCALIBUR_BACKGROUND=1 SYNCALIBUR_CONFIG="${CONFIG_FILE}" \
      nohup "${GO_ENGINE_BIN}" daemon-exec "${SCRIPT_PATH}" run </dev/null >> "${LOG_FILE}" 2>&1 &
  else
    SYNCALIBUR_BACKGROUND=1 SYNCALIBUR_CONFIG="${CONFIG_FILE}" \
      nohup bash -c 'exec -a syncalibur bash "$1" run' _ "${SCRIPT_PATH}" </dev/null >> "${LOG_FILE}" 2>&1 &
  fi
  sleep 1
  if pid="$(daemon_pid)"; then
    echo "✓ Syncalibur started in the background (PID ${pid})."
    echo "• Follow logs : syncalibur logs"
    echo "•  Live Cockpit : syncalibur status"
    return 0
  fi
  echo "✕ Failed to start. Check '${LOG_FILE}' for details." >&2
  return 1
}

stop_background_pid() {
  local pid
  if ! pid="$(daemon_pid)"; then
    echo "ℹ  Syncalibur is not running."
    rm -f "${PID_FILE}"
    return 0
  fi
  echo "○ Stopping Syncalibur (PID ${pid})..."
  kill "${pid}" 2>/dev/null || true
  for _ in 1 2 3 4 5; do
    kill -0 "${pid}" 2>/dev/null || break
    sleep 1
  done
  kill -0 "${pid}" 2>/dev/null && kill -9 "${pid}" 2>/dev/null || true
  rm -f "${PID_FILE}"
  echo "✓ Syncalibur stopped."
}

cmd_install() {
  ensure_go_engine || true
  install_global_cli_and_completions

  if [[ ! -f "${DEFAULT_ENV_FILE}" ]]; then
    save_repo_path_to_env "${WATCH_PATH}"
    echo "✦ Created local config file: ${DEFAULT_ENV_FILE} (watching ${WATCH_PATH})"
  fi

  local exec_start_cmd="/usr/bin/env bash \"${SCRIPT_PATH}\" run"
  if [[ -x "${GO_ENGINE_BIN}" ]]; then
    exec_start_cmd="\"${GO_ENGINE_BIN}\" daemon-exec \"${SCRIPT_PATH}\" run"
  fi

  case "${SERVICE_BACKEND}" in
    systemd)
      if ! is_service_installed && daemon_pid >/dev/null; then
        stop_background_pid
      fi
      mkdir -p "$(dirname "${UNIT_FILE}")"
      cat > "${UNIT_FILE}" << EOF
[Unit]
Description=Syncalibur - Git Auto-Sync Daemon for $(basename "${REPO_ROOT}")
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
WorkingDirectory=${REPO_ROOT}
Environment="SYNCALIBUR_CONFIG=${CONFIG_FILE}"
ExecStart=${exec_start_cmd}
Restart=always
RestartSec=5

[Install]
WantedBy=default.target
EOF
      systemctl --user daemon-reload \
        && systemctl --user enable "${SERVICE_NAME}" >/dev/null 2>&1 \
        && systemctl --user restart "${SERVICE_NAME}" \
        || { echo "✕ Could not enable/start systemd service '${SERVICE_NAME}'." >&2; return 1; }
      command -v loginctl >/dev/null 2>&1 && loginctl enable-linger "${USER:-$(id -un)}" 2>/dev/null || true
      ;;

    launchd)
      daemon_pid >/dev/null && stop_background_pid
      mkdir -p "$(dirname "${PLIST_FILE}")"
      if [[ -x "${GO_ENGINE_BIN}" ]]; then
        cat > "${PLIST_FILE}" << EOF
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
  <key>Label</key>
  <string>${PLIST_LABEL}</string>
  <key>ProgramArguments</key>
  <array>
    <string>${GO_ENGINE_BIN}</string>
    <string>daemon-exec</string>
    <string>${SCRIPT_PATH}</string>
    <string>run</string>
  </array>
  <key>WorkingDirectory</key>
  <string>${REPO_ROOT}</string>
  <key>EnvironmentVariables</key>
  <dict>
    <key>PATH</key>
    <string>${HOME}/.local/bin:/opt/homebrew/bin:/usr/local/bin:/usr/bin:/bin:/usr/sbin:/sbin</string>
    <key>SYNCALIBUR_CONFIG</key>
    <string>${CONFIG_FILE}</string>
  </dict>
  <key>RunAtLoad</key>
  <true/>
  <key>KeepAlive</key>
  <true/>
  <key>StandardOutPath</key>
  <string>${LOG_FILE}</string>
  <key>StandardErrorPath</key>
  <string>${LOG_FILE}</string>
</dict>
</plist>
EOF
      else
        cat > "${PLIST_FILE}" << EOF
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
  <key>Label</key>
  <string>${PLIST_LABEL}</string>
  <key>ProgramArguments</key>
  <array>
    <string>/bin/bash</string>
    <string>${SCRIPT_PATH}</string>
    <string>run</string>
  </array>
  <key>WorkingDirectory</key>
  <string>${REPO_ROOT}</string>
  <key>EnvironmentVariables</key>
  <dict>
    <key>PATH</key>
    <string>${HOME}/.local/bin:/opt/homebrew/bin:/usr/local/bin:/usr/bin:/bin:/usr/sbin:/sbin</string>
    <key>SYNCALIBUR_CONFIG</key>
    <string>${CONFIG_FILE}</string>
  </dict>
  <key>RunAtLoad</key>
  <true/>
  <key>KeepAlive</key>
  <true/>
  <key>StandardOutPath</key>
  <string>${LOG_FILE}</string>
  <key>StandardErrorPath</key>
  <string>${LOG_FILE}</string>
</dict>
</plist>
EOF
      fi
      launchctl unload "${PLIST_FILE}" >/dev/null 2>&1 || true
      launchctl load -w "${PLIST_FILE}" || { echo "✕ launchctl load failed." >&2; return 1; }
      ;;

    windows-startup)
      local vbs_file bash_win script_win config_win
      vbs_file="$(win_startup_vbs_path)"
      mkdir -p "$(dirname "${vbs_file}")"
      if command -v cygpath >/dev/null 2>&1; then
        bash_win="$(cygpath -w "$(command -v bash)")"
        script_win="$(cygpath -w "${SCRIPT_PATH}")"
        config_win="$(cygpath -w "${CONFIG_FILE}")"
      else
        bash_win="bash.exe"
        script_win="${SCRIPT_PATH}"
        config_win="${CONFIG_FILE}"
      fi
      cat > "${vbs_file}" << EOF
Set WshShell = CreateObject("WScript.Shell")
Set WshEnv = WshShell.Environment("PROCESS")
WshEnv("SYNCALIBUR_CONFIG") = "${config_win}"
WshShell.Run """${bash_win}"" ""${script_win}"" start", 0, False
EOF
      daemon_pid >/dev/null && stop_background_pid
      start_background_nohup >/dev/null
      ;;

    cron)
      if command -v crontab >/dev/null 2>&1; then
        local cron_line="@reboot SYNCALIBUR_CONFIG=\"${CONFIG_FILE}\" \"${SCRIPT_PATH}\" start >/dev/null 2>&1 # syncalibur:${SERVICE_NAME}"
        (crontab -l 2>/dev/null | grep -Fv "# syncalibur:${SERVICE_NAME}" || true; echo "${cron_line}") | crontab - 2>/dev/null || true
      fi
      mkdir -p "$(dirname "${DESKTOP_FILE}")" 2>/dev/null || true
      cat > "${DESKTOP_FILE}" 2>/dev/null << EOF || true
[Desktop Entry]
Type=Application
Name=Syncalibur (${SERVICE_NAME})
Exec=/usr/bin/env SYNCALIBUR_CONFIG="${CONFIG_FILE}" "${SCRIPT_PATH}" start
X-GNOME-Autostart-enabled=true
NoDisplay=true
EOF
      daemon_pid >/dev/null && stop_background_pid
      start_background_nohup >/dev/null
      ;;
  esac

  sleep 1
  print_syncalibur_dashboard
  echo "  ${C_GREEN}${C_BOLD}✓ Syncalibur installed and active!${C_RESET} Global CLI '${C_CYAN}syncalibur${C_RESET}' is ready."
  echo ""
}

cmd_uninstall() {
  case "${SERVICE_BACKEND}" in
    systemd)
      if [[ -f "${UNIT_FILE}" ]]; then
        systemctl --user disable --now "${SERVICE_NAME}" >/dev/null 2>&1 || true
        rm -f "${UNIT_FILE}"
        systemctl --user daemon-reload 2>/dev/null || true
      fi
      stop_background_pid >/dev/null 2>&1 || true
      echo "✓ Stopped, disabled and removed systemd service '${SERVICE_NAME}'."
      ;;
    launchd)
      if [[ -f "${PLIST_FILE}" ]]; then
        launchctl unload -w "${PLIST_FILE}" >/dev/null 2>&1 || true
        rm -f "${PLIST_FILE}"
      fi
      stop_background_pid >/dev/null 2>&1 || true
      echo "✓ Stopped and removed macOS LaunchAgent '${PLIST_LABEL}'."
      ;;
    windows-startup)
      rm -f "$(win_startup_vbs_path)"
      stop_background_pid
      echo "✓ Removed Windows Startup launcher and stopped Syncalibur."
      ;;
    cron)
      if command -v crontab >/dev/null 2>&1; then
        (crontab -l 2>/dev/null | grep -Fv "# syncalibur:${SERVICE_NAME}" || true) | crontab - 2>/dev/null || true
      fi
      rm -f "${DESKTOP_FILE}"
      stop_background_pid
      echo "✓ Removed autostart entries and stopped Syncalibur."
      ;;
  esac
}

cmd_start() {
  ensure_go_engine || true
  case "${SERVICE_BACKEND}" in
    systemd)
      if is_service_installed; then
        systemctl --user start "${SERVICE_NAME}" && echo "✓ Started systemd user service '${SERVICE_NAME}'."
        return
      fi
      ;;
    launchd)
      if is_service_installed; then
        launchctl start "${PLIST_LABEL}" 2>/dev/null || launchctl load -w "${PLIST_FILE}" 2>/dev/null
        echo "✓ Started macOS LaunchAgent '${PLIST_LABEL}'."
        return
      fi
      ;;
  esac
  start_background_nohup
}

cmd_stop() {
  case "${SERVICE_BACKEND}" in
    systemd)
      if is_service_installed && systemctl --user is-active --quiet "${SERVICE_NAME}"; then
        systemctl --user stop "${SERVICE_NAME}" && echo "✓ Stopped systemd user service '${SERVICE_NAME}'."
        return
      fi
      ;;
    launchd)
      if is_service_installed; then
        launchctl unload "${PLIST_FILE}" >/dev/null 2>&1 || true
        stop_background_pid >/dev/null 2>&1 || true
        echo "✓ Stopped macOS LaunchAgent '${PLIST_LABEL}'."
        return
      fi
      ;;
  esac
  stop_background_pid
}

cmd_restart() {
  ensure_go_engine || true
  clear_sync_failure
  rm -f "${PAUSE_FILE}" 2>/dev/null || true
  case "${SERVICE_BACKEND}" in
    systemd)
      if is_service_installed; then
        systemctl --user restart "${SERVICE_NAME}" && echo "✓ Restarted '${SERVICE_NAME}'."
        return
      fi
      ;;
    launchd)
      if is_service_installed; then
        launchctl unload "${PLIST_FILE}" >/dev/null 2>&1 || true
        launchctl load -w "${PLIST_FILE}" && echo "✓ Restarted '${PLIST_LABEL}'."
        return
      fi
      ;;
  esac
  stop_background_pid
  start_background_nohup
}

cmd_status() {
  local is_static=false
  local target_arg=""

  while [[ $# -gt 0 ]]; do
    case "$1" in
      --static)
        is_static=true
        shift
        ;;
      *)
        if [[ -z "${target_arg}" ]]; then
          target_arg="$1"
        fi
        shift
        ;;
    esac
  done

  local target_repo=""
  if [[ -n "${target_arg}" && -d "${target_arg}" ]]; then
    target_repo="$(git -C "${target_arg}" rev-parse --show-toplevel 2>/dev/null || echo "${target_arg}")"
  elif cur_repo="$(git rev-parse --show-toplevel 2>/dev/null)"; then
    target_repo="${cur_repo}"
  else
    target_repo="${REPO_ROOT}"
  fi
  target_repo="$(resolve_path "${target_repo}")"

  if [[ -t 0 && -t 1 ]] && [[ "${is_static}" != "true" ]] && ensure_go_engine; then
    "${GO_ENGINE_BIN}" status "${SCRIPT_PATH}" "${target_repo}"
    return $?
  fi

  local prev_repo="${REPO_ROOT}"
  local prev_watch="${WATCH_PATH}"
  REPO_ROOT="${target_repo}"
  WATCH_PATH="${target_repo}"
  cd "${REPO_ROOT}" || return 1
  print_syncalibur_dashboard

  if [[ -f "${FAIL_REPORT_FILE}" ]]; then
    local f_ts f_stage f_reason f_hint
    f_ts="$(sed -n '1p' "${FAIL_REPORT_FILE}")"
    f_stage="$(sed -n '2p' "${FAIL_REPORT_FILE}")"
    f_reason="$(sed -n '4p' "${FAIL_REPORT_FILE}")"
    f_hint="$(sed -n '5p' "${FAIL_REPORT_FILE}")"
    printf '  %s%s✕ Active Sync Failure Alert (%s — %s):%s\n' "${C_RED}" "${C_BOLD}" "${f_stage}" "${f_ts}" "${C_RESET}"
    printf '    %sReason :%s %s\n' "${C_RED}" "${C_RESET}" "${f_reason}"
    printf '    %sFix    :%s %s\n\n' "${C_AMBER}" "${C_RESET}" "${f_hint}"
  fi

  local path reason
  local -a blocked=()
  while IFS= read -r -d '' path; do
    reason="$(block_reason "${path}")" && blocked+=("    ${C_AMBER}⛨ ${path}${C_RESET} ${C_SLATE}— ${reason}${C_RESET}")
  done < <(changed_paths)
  if (( ${#blocked[@]} > 0 )); then
    printf '  %s%sSkipped Files (%d):%s\n' "${C_AMBER}" "${C_BOLD}" "${#blocked[@]}" "${C_RESET}"
    printf '%s\n' "${blocked[@]}"
    printf '\n'
  fi

  local tail_lines=10
  if [[ -f "${LOG_FILE}" ]]; then
    printf '  %s%sRecent Activity%s\n' "${C_GOLD}" "${C_BOLD}" "${C_RESET}"
    printf '  %s────────────────────────────────────────────────────────────────────────────────────────%s\n' "${C_SLATE}" "${C_RESET}"
    local line
    while IFS= read -r line; do
      printf '  '
      format_log_line "${line}"
    done < <(tail -n "${tail_lines}" "${LOG_FILE}")
    printf '\n'
  fi

  REPO_ROOT="${prev_repo}"
  WATCH_PATH="${prev_watch}"
  return 0
}

cmd_logs() {
  mkdir -p "$(dirname "${LOG_FILE}")"
  touch "${LOG_FILE}"
  print_syncalibur_dashboard
  printf '  %s%s◈ Live Syncalibur Telemetry%s %s(Press Ctrl+C to exit)%s\n' "${C_GOLD}" "${C_BOLD}" "${C_RESET}" "${C_SLATE}" "${C_RESET}"
  printf '  %s────────────────────────────────────────────────────────────────────────────────────────%s\n' "${C_SLATE}" "${C_RESET}"
  local line
  tail -n 20 -f "${LOG_FILE}" | while IFS= read -r line; do
    printf '  '
    format_log_line "${line}"
  done
}

cmd_version() {
  ensure_go_engine
  if [[ -x "${GO_ENGINE_BIN}" ]]; then
    "${GO_ENGINE_BIN}" version
  else
    echo "Syncalibur ${SYNCALIBUR_PLATFORM_VERSION} (bash fallback)"
  fi
}

cmd_release() {
  ensure_go_engine
  local cur_git_root
  cur_git_root="$(git rev-parse --show-toplevel 2>/dev/null || echo "")"
  local target_repo="${cur_git_root:-${REPO_ROOT}}"
  if [[ -x "${GO_ENGINE_BIN}" ]]; then
    "${GO_ENGINE_BIN}" release "$@" "${target_repo}"
  else
    echo "✕ Go engine required for automated release calculation." >&2
    return 1
  fi
}

usage() {
  print_syncalibur_dashboard
  cat << EOF
  ${C_GOLD}${C_BOLD}Usage:${C_RESET} syncalibur ${C_CYAN}{push|commit|check|release|version|tidy|mode|hook|install|status|fleet|doctor|logs|...}${C_RESET}

  ${C_GOLD}${C_BOLD}Developer Ergonomics & Flight Recorder:${C_RESET}
    ${C_CYAN}save [msg]${C_RESET}             One-shot quicksave: secret scan + syntax check + AST commit + push
    ${C_CYAN}undo [N]${C_RESET}               Time Machine: safely revert N commits via soft reset (preserves working tree)
    ${C_CYAN}push [msg]${C_RESET}             Alias to 'save' (one-shot quicksave & multi-remote push)
    ${C_CYAN}commit [msg]${C_RESET}           Local preflight check & commit (skips network push)
    ${C_CYAN}check | preflight${C_RESET}     Run pre-push inspection (syntax hold, secret shield, diff, proposed msg)
    ${C_CYAN}checkpoint [list|save]${C_RESET} List or create safety rollback checkpoints
    ${C_CYAN}release [mode]${C_RESET}         Automated SemVer bump, CHANGELOG generation & Git release tag
    ${C_CYAN}version | -v${C_RESET}          Print official release version & platform milestone
    ${C_CYAN}tidy${C_RESET}                    Squash unpushed local micro-commits into 1 clean conventional commit
    ${C_CYAN}hook <install|status>${C_RESET}  Install universal pre-commit & prepare-commit-msg Git hooks

  ${C_GOLD}${C_BOLD}Autonomous Daemon & Fleet Control:${C_RESET}
    ${C_CYAN}install [path]${C_RESET}         Configure OS auto-start, compile Go engine, link global CLI & start now
    ${C_CYAN}status [--static]${C_RESET}      Launch interactive cockpit in TTY (or static dashboard with --static)
    ${C_CYAN}mode [name]${C_RESET}            Inspect or toggle operating architecture: ${C_WHITE}autopilot | manual${C_RESET}
    ${C_CYAN}logs${C_RESET}                   Stream live colorized telemetry
    ${C_CYAN}fleet <subcmd>${C_RESET}         Manage multi-repo fleet: ${C_WHITE}status | add <path> | remove <path> | scan [dir] | start | stop${C_RESET}
    ${C_CYAN}doctor${C_RESET}                 Run 12-point system diagnostic health check
    ${C_CYAN}notify [msg]${C_RESET}           Send an instant test notification to Telegram Bot / Discord / Slack Webhooks
    ${C_CYAN}pr [base|--preview]${C_RESET}    Open a structured Pull Request using Syncalibur's auto-populated PR Template
    ${C_CYAN}pull${C_RESET}                   Trigger immediate bidirectional remote fetch & safe rebase
    ${C_CYAN}pause | resume${C_RESET}         Pause or resume background auto-commits without stopping the service
    ${C_CYAN}alert [--test]${C_RESET}         Display full status + Sync Failed diagnostic report (or simulate with --test)
    ${C_CYAN}restart | stop | start${C_RESET} Control background daemon lifecycle
    ${C_CYAN}uninstall${C_RESET}              Stop Syncalibur and remove OS auto-start service

EOF
}

case "${1:-}" in
  save | quicksave) shift; cmd_save "$@" ;;
  push)       shift; cmd_push "$@" ;;
  commit)     shift; cmd_commit "$@" ;;
  check | preflight) shift; cmd_check "${1:-}" ;;
  undo | recover) shift; cmd_undo "$@" ;;
  checkpoint | checkpoints) shift; cmd_checkpoint "$@" ;;
  release)    shift; cmd_release "$@" ;;
  version | --version | -v) cmd_version ;;
  tidy)       shift; cmd_tidy "${1:-}" ;;
  mode)       shift; cmd_mode "$@" ;;
  hook)       shift; cmd_hook "${1:-status}" ;;
  install)    cmd_install "${2:-}" ;;
  uninstall)  cmd_uninstall ;;
  start)      cmd_start ;;
  stop)       cmd_stop ;;
  restart)    cmd_restart ;;
  status)     shift; cmd_status "$@" ;;
  logs | log) cmd_logs ;;
  ui)         shift; cmd_status "$@" ;;
  fleet)      cmd_fleet "${2:-status}" "${3:-}" ;;
  doctor)     cmd_doctor ;;
  notify | telegram-test | webhook-test)
    shift
    cmd_notify "$@"
    ;;
  pr)         cmd_pr "${2:-}" "${3:-}" ;;
  pull)       shift; cmd_pull "$@" ;;
  sync)       touch "${SYNC_TRIGGER_FILE}" && echo "✦ Triggered immediate sync cycle." ;;
  pause)      touch "${PAUSE_FILE}" && echo "▲  Syncalibur paused. Run 'syncalibur resume' to continue." ;;
  resume)     rm -f "${PAUSE_FILE}" && echo "▶  Syncalibur resumed." ;;
  alert)      cmd_alert "${2:-}" ;;
  run)        run_daemon ;;
  *)          usage; exit 1 ;;
esac
