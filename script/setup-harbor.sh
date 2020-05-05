#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

# Constants
ROOT_DIR="$(cd "$( dirname "${BASH_SOURCE[0]}" )/.." >/dev/null && pwd)"
RESET='\033[0m'
GREEN='\033[38;5;2m'
RED='\033[38;5;1m'
YELLOW='\033[38;5;3m'

# Load Libraries
# shellcheck disable=SC1090
. "${ROOT_DIR}/script/libtest.sh"
# shellcheck disable=SC1090
. "${ROOT_DIR}/script/liblog.sh"

# Axiliar functions
print_menu() {
    local script
    script=$(basename "${BASH_SOURCE[0]}")
    log "${RED}NAME${RESET}"
    log "    $(basename -s .sh "${BASH_SOURCE[0]}")"
    log ""
    log "${RED}SINOPSIS${RESET}"
    log "    $script [${YELLOW}-dh${RESET}] [${YELLOW}-n ${GREEN}\"namespace\"${RESET}]"
    log ""
    log "${RED}DESCRIPTION${RESET}"
    log "    Script to setup Harbor on your K8s cluster."
    log ""
    log "    The options are as follow:"
    log ""
    log "      ${YELLOW}-n, --namespace ${GREEN}[namespace]${RESET}           Namespace to use for Harbor."
    log "      ${YELLOW}-h, --help${RESET}                            Print this help"
    log "      ${YELLOW}-u, --dry-run${RESET}                         Enable \"dry run\" mode"
    log ""
    log "${RED}EXAMPLES${RESET}"
    log "      $script --help"
    log "      $script --namespace \"harbor\""
    log ""
}

namespace="harbor"
help_menu=0
dry_run=0
while [[ "$#" -gt 0 ]]; do
    case "$1" in
        -h|--help)
            help_menu=1
            ;;
        -u|--dry-run)
            dry_run=1
            ;;
        -n|--namespace)
            shift; namespace="${1:?missing namespace}"
            ;;
        *)
            error "Invalid command line flag $1" >&2
            exit 1
            ;;
    esac
    shift
done

if [[ "$help_menu" -eq 1 ]]; then
    print_menu
    exit 0
fi

# Harbor values
values="$(cat << EOF
service:
  tls:
    enabled: false
EOF
)"

if [[ "$dry_run" -eq 1 ]]; then
    info "DRY RUN mode enabled!"
    info "Namespace: $namespace"
    info "Generated values.yaml:"
    printf '#####\n\n%s\n\n#####\n' "$values"
    exit 0
fi

# Install Harbor
info "Using the values.yaml below:"
printf '#####\n\n%s\n\n#####\n' "$values"
info "Installing Harbor in namespace '$namespace'..."
silence kubectl create ns "$namespace"
silence helm install harbor \
    --namespace "$namespace" \
    -f <(echo "$values") \
    bitnami/harbor
# Wait for Harbor components
info "Waiting for Harbor components to be ready..."
deployments=(
    "harbor-chartmuseum"
    "harbor-clair"
    "harbor-core"
    "harbor-jobservice"
    "harbor-nginx"
    "harbor-notary-server"
    "harbor-notary-signer"
    "harbor-portal"
    "harbor-registry"
)

for dep in "${deployments[@]}"; do
    k8s_wait_for_deployment "$namespace" "$dep"
    info "Deployment ${dep} ready!"
done
echo
    
info "Use this command for port-forwading to Harbor:"
info "kubectl port-forward --namespace $namespace svc/harbor 8888:80 >/dev/null 2>&1 &"
info "Harbor URL: http://127.0.0.1:8888/"
info "Harbor credentials"
info "  - username: admin"
info "  - Password: $(kubectl get secret harbor-core-envvars --namespace "$namespace" -o jsonpath="{.data.HARBOR_ADMIN_PASSWORD}" | base64 --decode)"
echo
