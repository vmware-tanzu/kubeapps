# This file provides targets which create a local k8s cluster setup
# with OIDC integration for development and testing.
KUBE ?= ${HOME}/.kube
CLUSTER_NAME ?= kubeapps

CLUSTER_CONFIG = ${KUBE}/kind-config-${CLUSTER_NAME}

${CLUSTER_CONFIG}:
	kind create cluster \
		--kubeconfig ${CLUSTER_CONFIG} \
		--name ${CLUSTER_NAME} \
		--config=./docs/user/manifests/kubeapps-local-dev-apiserver-config.json \
		--retain

cluster-kind: ${CLUSTER_CONFIG}

delete-cluster-kind:
	kind delete cluster --name ${CLUSTER_NAME} || true
	rm ${CLUSTER_CONFIG}

.PHONY: cluster-kind cluster-kind-delete
