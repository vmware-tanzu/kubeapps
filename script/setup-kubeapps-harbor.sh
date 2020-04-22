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

info "Updating chart repositories..."
silence helm repo update
echo

# Install Harbor
info "Installing Harbor in namespace 'harbor'..."
harbor_values="$(cat << 'EOF'
service:
  tls:
    enabled: false
EOF
)"
info "Using the values.yaml below:"
printf '#####\n\n%s\n\n#####\n' "$harbor_values"
silence kubectl create ns harbor
silence helm install harbor \
  --namespace harbor \
  -f <(echo "$harbor_values") \
  bitnami/harbor
# Wait for Harbor components
info "Waiting for Harbor components to be ready..."
harbor_deployments=(
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
for dep in "${harbor_deployments[@]}"; do
  k8s_wait_for_deployment harbor "$dep"
  info "Deployment ${dep} ready"
done
echo

# Install Kubeapps
info "Installing Kubeapps in namespace 'kubeapps'..."
kubeapps_values="$(cat << 'EOF'
useHelm3: true
apprepository:
  initialRepos:
    - name: bitnami
      url: https://charts.bitnami.com/bitnami
    - name: harbor-library
      url: http://harbor.harbor.svc.cluster.local/chartrepo/library
EOF
)"
info "Using the values.yaml below:"
printf '#####\n\n%s\n\n#####\n' "$kubeapps_values"
silence kubectl create ns kubeapps
silence helm install kubeapps \
  --namespace kubeapps \
  -f <(echo "$kubeapps_values") \
  bitnami/kubeapps
# Wait for Kubeapps components
info "Waiting for Kubeapps components to be ready..."
kubeapps_deployments=(
  "kubeapps"
  "kubeapps-internal-apprepository-controller"
  "kubeapps-internal-assetsvc"
  "kubeapps-internal-dashboard"
)
for dep in "${kubeapps_deployments[@]}"; do
  k8s_wait_for_deployment kubeapps "$dep"
  info "Deployment ${dep} ready"
done
echo

# Create serviceAccount
info "Creating 'example' serviceAccount and adding RBAC permissions for 'default' namespace..."
silence kubectl create serviceaccount example --namespace default
silence kubectl apply -f https://raw.githubusercontent.com/kubeapps/kubeapps/master/docs/user/manifests/kubeapps-applications-read.yaml
silence kubectl create -n default rolebinding example-view --clusterrole=kubeapps-applications-read --serviceaccount default:example
silence kubectl create -n default rolebinding example-edit --clusterrole=edit --serviceaccount default:example
silence kubectl create -n kubeapps rolebinding example-kubeapps-repositories-read --role=kubeapps-repositories-read --serviceaccount default:example
silence kubectl create -n kubeapps rolebinding example-kubeapps-repositories-write --role=kubeapps-repositories-write --serviceaccount default:example
echo

info "Use this command for port-forwading to Harbor:"
info "kubectl port-forward --namespace harbor svc/harbor 8888:80 >/dev/null 2>&1 &"
info "Harbor URL: http://127.0.0.1:8888/"
info "Harbor credentials"
info "  - username: admin"
info "  - Password: $(kubectl get secret harbor-core-envvars --namespace harbor -o jsonpath="{.data.HARBOR_ADMIN_PASSWORD}" | base64 --decode)"
echo
info "Use this command for port forwading to Kubeapps Dashboard:"
info "kubectl port-forward --namespace kubeapps svc/kubeapps 8080:80 >/dev/null 2>&1 &"
info "Kubeapps URL: http://127.0.0.1:8080"
info "Kubeppas API Token:"
kubectl get -n default secret "$(kubectl get serviceaccount example --namespace default -o jsonpath='{.secrets[].name}')" -o go-template='{{.data.token | base64decode}}' && echo
