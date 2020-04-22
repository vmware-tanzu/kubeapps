#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

# Constants
ROOT_DIR="$(cd "$( dirname "${BASH_SOURCE[0]}" )/.." >/dev/null && pwd)"

# Load Libraries
# shellcheck disable=SC1090
. "${ROOT_DIR}/script/libtest.sh"
# shellcheck disable=SC1090
. "${ROOT_DIR}/script/liblog.sh"


# Uninstall Harbor
info "Uninstalling Harbor in namespace 'harbor'..."
silence helm uninstall harbor -n harbor
silence kubectl delete pvc -n harbor $(kubectl get pvc -n harbor -o jsonpath='{.items[*].metadata.name}')
info "Deleting 'harbor' namespace..."
silence kubectl delete ns harbor

# Uninstall Kubeapps
info "Uninstalling Kubeapps in namespace 'kubeapps'..."
silence helm uninstall kubeapps -n kubeapps
silence kubectl delete rolebinding example-kubeapps-repositories-read -n kubeapps
silence kubectl delete rolebinding example-kubeapps-repositories-write -n kubeapps
info "Deleting 'kubeapps' namespace..."
silence delete ns kubeapps 

# Delete serviceAccount
info "Deleting 'example' serviceAccount and related RBAC objects..."
silence kubectl delete serviceaccount example --namespace default
silence kubectl delete clusterrole kubeapps-applications-read
silence kubectl delete rolebinding example-view
silence kubectl delete rolebinding example-edit
