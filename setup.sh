#!/usr/bin/env bash
set -euo pipefail

readonly SCRIPT_DIRECTORY="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "${SCRIPT_DIRECTORY}"

readonly TEMPLATE_MODULE_PATH="github.com/eminbekov/fiber-v3-template"
readonly REQUIRED_GO_VERSION="1.26"
readonly ENV_EXAMPLE_FILE=".env.example"
readonly ENV_FILE=".env"

readonly MODULE_KEYS=(
  "nats"
  "grpc"
  "websocket"
  "admin"
  "web"
  "i18n"
  "storage"
  "cron"
  "console"
  "monitoring"
  "swagger"
  "views"
)

readonly MARKER_FILES=(
  "cmd/server/main.go"
  "internal/router/router.go"
  "internal/config/config.go"
  ".env.example"
  "Makefile"
  "deploy/docker/Dockerfile"
  "deploy/docker/docker-compose.yml"
  "deploy/docker/docker-compose.dev.yml"
)

declare -A MODULE_DESCRIPTIONS=(
  ["nats"]="NATS messaging and consumers"
  ["grpc"]="gRPC API server and protobuf generated code"
  ["websocket"]="WebSocket realtime channel support"
  ["admin"]="Admin HTML auth and dashboard pages"
  ["web"]="Public HTML landing page"
  ["i18n"]="Language detection and locale translations"
  ["storage"]="File upload/download and signed URLs"
  ["cron"]="Dedicated cron binary and in-app scheduler wiring"
  ["console"]="Console CLI admin commands (create-admin, assign-role, cache-clear, export-users)"
  ["monitoring"]="Prometheus/Loki/Tempo/Grafana compose services"
  ["swagger"]="Generated OpenAPI docs and Swagger route"
)

declare -A MODULE_PATHS=(
  ["nats"]="internal/nats"
  ["grpc"]="internal/grpc gen proto"
  ["websocket"]="internal/websocket"
  ["admin"]="internal/handler/admin views/admin views/layouts/auth.html views/layouts/base.html views/partials/sidebar.html"
  ["web"]="internal/handler/web views/public views/layouts/public.html views/partials/public_header.html views/partials/public_footer.html"
  ["i18n"]="internal/i18n internal/middleware/language.go"
  ["storage"]="internal/storage uploads internal/middleware/signed_url.go"
  ["cron"]="internal/cron cmd/cron"
  ["console"]="internal/console cmd/console"
  ["monitoring"]="monitoring"
  ["swagger"]="docs"
)

declare -A KEEP_MODULES=()

readonly COLOR_RED='\033[0;31m'
readonly COLOR_GREEN='\033[0;32m'
readonly COLOR_YELLOW='\033[1;33m'
readonly COLOR_CYAN='\033[0;36m'
readonly COLOR_BOLD='\033[1m'
readonly COLOR_RESET='\033[0m'

log_info() {
  printf "${COLOR_GREEN}[INFO]${COLOR_RESET} %s\n" "$1"
}

log_warning() {
  printf "${COLOR_YELLOW}[WARN]${COLOR_RESET} %s\n" "$1"
}

log_error() {
  printf "${COLOR_RED}[ERROR]${COLOR_RESET} %s\n" "$1" >&2
}

print_section() {
  printf "\n${COLOR_BOLD}${COLOR_CYAN}=== %s ===${COLOR_RESET}\n\n" "$1"
}

require_command() {
  local command_name="$1"
  if ! command -v "${command_name}" >/dev/null 2>&1; then
    log_error "Required command is missing: ${command_name}"
    exit 1
  fi
}

strip_to_major_minor() {
  local version_string="$1"
  printf '%s' "${version_string}" | awk -F. '{print $1 "." $2}'
}

ensure_go_version() {
  local go_version_output
  local current_go_version
  local current_major_minor
  local required_major_minor

  go_version_output="$(go version)"
  current_go_version="$(printf '%s\n' "${go_version_output}" | awk '{print $3}' | sed 's/^go//')"
  current_major_minor="$(strip_to_major_minor "${current_go_version}")"
  required_major_minor="$(strip_to_major_minor "${REQUIRED_GO_VERSION}.0")"

  if [[ "${current_major_minor}" < "${required_major_minor}" ]]; then
    log_error "Go ${REQUIRED_GO_VERSION}+ is required, found ${current_go_version}"
    exit 1
  fi

  log_info "Go version: ${current_go_version}"
}

replace_in_file() {
  local file_path="$1"
  local search_value="$2"
  local replace_value="$3"
  local temporary_file

  temporary_file="$(mktemp)"
  sed "s|${search_value}|${replace_value}|g" "${file_path}" > "${temporary_file}"
  mv "${temporary_file}" "${file_path}"
}

replace_module_path() {
  local new_module_path="$1"

  print_section "Module Path"
  log_info "Replacing module path: ${TEMPLATE_MODULE_PATH} -> ${new_module_path}"

  while IFS= read -r go_file_path; do
    replace_in_file "${go_file_path}" "${TEMPLATE_MODULE_PATH}" "${new_module_path}"
  done < <(find . -type f -name "*.go" -not -path "./.git/*" -print)

  replace_in_file "go.mod" "${TEMPLATE_MODULE_PATH}" "${new_module_path}"
  replace_in_file "deploy/docker/docker-compose.yml" "${TEMPLATE_MODULE_PATH}" "${new_module_path}"
}

remove_marker_block() {
  local file_path="$1"
  local module_key="$2"
  local temporary_file

  temporary_file="$(mktemp)"
  awk "
    \$0 ~ /\\[module:${module_key}:start\\]/ { skipping=1; next }
    \$0 ~ /\\[module:${module_key}:end\\]/   { skipping=0; next }
    !skipping { print \$0 }
  " "${file_path}" > "${temporary_file}"
  mv "${temporary_file}" "${file_path}"
}

remove_marker_comments_only() {
  local file_path="$1"
  local module_key="$2"
  local temporary_file

  temporary_file="$(mktemp)"
  awk "
    \$0 ~ /\\[module:${module_key}:start\\]/ { next }
    \$0 ~ /\\[module:${module_key}:end\\]/   { next }
    { print \$0 }
  " "${file_path}" > "${temporary_file}"
  mv "${temporary_file}" "${file_path}"
}

strip_markers_for_module() {
  local module_key="$1"
  local file_path

  for file_path in "${MARKER_FILES[@]}"; do
    if [[ -f "${file_path}" ]]; then
      if [[ "${KEEP_MODULES[${module_key}]:-0}" == "1" ]]; then
        remove_marker_comments_only "${file_path}" "${module_key}"
      else
        remove_marker_block "${file_path}" "${module_key}"
      fi
    fi
  done
}

delete_paths_for_module() {
  local module_key="$1"
  local module_paths_string
  local path_value

  module_paths_string="${MODULE_PATHS[${module_key}]:-}"
  if [[ -z "${module_paths_string}" ]]; then
    return
  fi

  for path_value in ${module_paths_string}; do
    if [[ -e "${path_value}" ]]; then
      rm -rf "${path_value}"
      log_info "Deleted ${path_value}"
    fi
  done
}

ask_yes_no_default_yes() {
  local prompt_message="$1"
  local answer_value

  read -r -p "${prompt_message} [Y/n]: " answer_value
  answer_value="${answer_value:-y}"
  [[ "${answer_value,,}" == "y" || "${answer_value,,}" == "yes" ]]
}

select_modules() {
  local module_key
  local keep_response

  print_section "Optional Modules"
  log_info "Press Enter to keep a module (default). Type 'n' to remove."

  for module_key in "${MODULE_KEYS[@]}"; do
    if [[ "${module_key}" == "views" ]]; then
      continue
    fi
    read -r -p "$(printf "  %-11s %s [Y/n]: " "${module_key}" "${MODULE_DESCRIPTIONS[${module_key}]}")" keep_response
    keep_response="${keep_response:-y}"
    if [[ "${keep_response,,}" == "y" || "${keep_response,,}" == "yes" ]]; then
      KEEP_MODULES["${module_key}"]="1"
    else
      KEEP_MODULES["${module_key}"]="0"
    fi
  done
}

apply_module_selection() {
  local module_key

  print_section "Applying Module Selection"

  for module_key in "${MODULE_KEYS[@]}"; do
    if [[ "${module_key}" == "views" ]]; then
      continue
    fi
    if [[ "${KEEP_MODULES[${module_key}]:-0}" == "0" ]]; then
      log_info "Removing module: ${module_key}"
      delete_paths_for_module "${module_key}"
    else
      log_info "Keeping module: ${module_key}"
    fi
    strip_markers_for_module "${module_key}"
  done

  if [[ "${KEEP_MODULES[admin]:-0}" == "0" && "${KEEP_MODULES[web]:-0}" == "0" ]]; then
    KEEP_MODULES["views"]="0"
    strip_markers_for_module "views"
    if [[ -d "views" ]]; then
      rm -rf "views"
      log_info "Removed views directory because both admin and web modules are disabled"
    fi
  else
    KEEP_MODULES["views"]="1"
    strip_markers_for_module "views"
  fi
}

build_env_file() {
  local line_value
  local variable_name
  local default_value
  local user_value

  print_section "Environment Setup"
  : > "${ENV_FILE}"

  while IFS= read -r line_value || [[ -n "${line_value}" ]]; do
    if [[ -z "${line_value}" || "${line_value}" =~ ^[[:space:]]*# ]]; then
      printf "%s\n" "${line_value}" >> "${ENV_FILE}"
      continue
    fi

    variable_name="${line_value%%=*}"
    default_value="${line_value#*=}"

    read -r -p "$(printf "%-30s [%s]: " "${variable_name}" "${default_value}")" user_value
    user_value="${user_value:-${default_value}}"
    printf "%s=%s\n" "${variable_name}" "${user_value}" >> "${ENV_FILE}"
  done < "${ENV_EXAMPLE_FILE}"

  log_info "Created ${ENV_FILE}"
}

maybe_remove_template_files() {
  if ask_yes_no_default_yes "Remove template-only docs (GO_FIBER_PROJECT_GUIDE.md, TASKS.md)?"; then
    rm -f GO_FIBER_PROJECT_GUIDE.md TASKS.md
    log_info "Removed template-only docs"
  fi
}

maybe_reinitialize_git() {
  local reinitialize_answer
  read -r -p "Reinitialize git history for a fresh project? [y/N]: " reinitialize_answer
  reinitialize_answer="${reinitialize_answer:-n}"

  if [[ "${reinitialize_answer,,}" == "y" || "${reinitialize_answer,,}" == "yes" ]]; then
    rm -rf .git
    git init
    git add -A
    git commit -m "feat: initialize project from fiber-v3-template"
    log_info "Initialized new git history"
  fi
}

maybe_remove_setup_script() {
  if ask_yes_no_default_yes "Remove setup.sh after installation?"; then
    log_info "Removing setup.sh"
    rm -f "./setup.sh"
  fi
}

run_finalize_commands() {
  print_section "Finalize"
  go mod tidy
  gofmt -s -w .
  log_info "Ran go mod tidy and gofmt"
}

print_next_steps() {
  print_section "Done"
  printf "Project setup is complete.\n\n"
  printf "Next commands:\n"
  printf "  make docker-dev   # start Postgres + Redis"
  if [[ "${KEEP_MODULES[nats]:-0}" == "1" ]]; then
    printf " + NATS"
  fi
  printf "\n"
  printf "  make run          # run server locally\n"
  printf "  make up           # full docker stack\n"
  printf "  make help         # list make targets\n"
}

main() {
  local new_module_path

  print_section "Prerequisites"
  require_command "go"
  require_command "git"
  ensure_go_version

  if ! command -v docker >/dev/null 2>&1; then
    log_warning "docker is not installed. Docker workflows will be unavailable."
  fi
  if ! command -v make >/dev/null 2>&1; then
    log_warning "make is not installed. Makefile targets will be unavailable."
  fi

  read -r -p "Go module path (example: github.com/yourname/myproject): " new_module_path
  if [[ -z "${new_module_path}" ]]; then
    log_error "Module path cannot be empty"
    exit 1
  fi

  replace_module_path "${new_module_path}"
  select_modules
  apply_module_selection
  build_env_file
  run_finalize_commands
  maybe_remove_template_files
  maybe_reinitialize_git
  print_next_steps
  maybe_remove_setup_script
}

main "$@"
