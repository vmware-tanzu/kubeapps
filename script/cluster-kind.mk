# This file provides targets which create a local k8s cluster setup
# with OIDC integration for development and testing.
KUBE ?= ${HOME}/.kube
CLUSTER_NAME ?= kubeapps
ADDITIONAL_CLUSTER_NAME ?= kubeapps-additional

CLUSTER_CONFIG = ${KUBE}/kind-config-${CLUSTER_NAME}
ADDITIONAL_CLUSTER_CONFIG = ${KUBE}/kind-config-${ADDITIONAL_CLUSTER_NAME}

${CLUSTER_CONFIG}:
	kind create cluster \
		--kubeconfig ${CLUSTER_CONFIG} \
		--name ${CLUSTER_NAME} \
		--config=./docs/user/manifests/kubeapps-local-dev-apiserver-config.json \
		--retain
	kubectl apply --kubeconfig=${CLUSTER_CONFIG} -f ./docs/user/manifests/kubeapps-local-dev-users-rbac.yaml

cluster-kind: ${CLUSTER_CONFIG}

# dex will be running on the same node as the API server in the dev environment, so we can
# reuse the key and cert from the apiserver, which already includes v3 extensions
# for the correct alternative name (using the IP address).
devel/dex.crt:
	kubectl -n kube-system cp kube-apiserver-kubeapps-control-plane:etc/kubernetes/pki/apiserver.crt ./devel/dex.crt

devel/dex.key:
	kubectl -n kube-system cp kube-apiserver-kubeapps-control-plane:etc/kubernetes/pki/apiserver.key ./devel/dex.key

${ADDITIONAL_CLUSTER_CONFIG}: devel/dex.crt
	kind create cluster \
		--kubeconfig ${ADDITIONAL_CLUSTER_CONFIG} \
		--name ${ADDITIONAL_CLUSTER_NAME} \
		--config=./docs/user/manifests/kubeapps-local-dev-additional-apiserver-config.json \
		--retain
	kubectl apply --kubeconfig=$ADDITIONAL_CLUSTER_CONFIG -f ./docs/user/manifests/kubeapps-local-dev-users-rbac.yaml

additional-cluster-kind: ${ADDITIONAL_CLUSTER_CONFIG}

delete-cluster-kind:
	kind delete cluster --name ${CLUSTER_NAME} || true
	kind delete cluster --name ${ADDITIONAL_CLUSTER_NAME} || true
	rm ${CLUSTER_CONFIG}
	rm ${ADDITIONAL_CLUSTER_CONFIG} || true
	rm devel/dex.* || true

.PHONY: additional-cluster-kind cluster-kind cluster-kind-delete
