# This file provides targets which create a local k8s cluster setup
# with OIDC integration for development and testing.
KUBE ?= ${HOME}/.kube
CLUSTER_NAME ?= kubeapps

CLUSTER_CONFIG = ${KUBE}/kind-config-${CLUSTER_NAME}

devel/local-dev-apiserver-config.json:
	cat docs/user/manifests/kubeapps-local-dev-apiserver-config.json | \
	jq ".nodes[0].extraMounts[0].hostPath = \"${PWD}/script/test-certs/ca.cert.pem\"" > \
	$@

${CLUSTER_CONFIG}: devel/local-dev-apiserver-config.json
	kind create cluster \
		--name ${CLUSTER_NAME} \
		--config=./devel/local-dev-apiserver-config.json \
		--retain

cluster-kind: ${CLUSTER_CONFIG}

delete-cluster-kind:
	kind delete cluster --name ${CLUSTER_NAME}

.PHONY: cluster-kind cluster-kind-delete
