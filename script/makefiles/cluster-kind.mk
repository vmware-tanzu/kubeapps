# Copyright 2021-2022 the Kubeapps contributors.
# SPDX-License-Identifier: Apache-2.0

# This file provides targets which create a local k8s cluster setup
# with OIDC integration for development and testing.
KUBE ?= ${HOME}/.kube
CLUSTER_NAME ?= kubeapps
ADDITIONAL_CLUSTER_NAME ?= kubeapps-additional
IMAGE ?= kindest/node:v1.24.0@sha256:0866296e693efe1fed79d5e6c7af8df71fc73ae45e3679af05342239cdc5bc8e

CLUSTER_CONFIG = ${KUBE}/kind-config-${CLUSTER_NAME}
ADDITIONAL_CLUSTER_CONFIG = ${KUBE}/kind-config-${ADDITIONAL_CLUSTER_NAME}

# The --wait 10s in the create cluster calls is not sufficient for the control-plane node to be ready,
# but is sufficient for the pod to be created so that we can copy the certs below.
${CLUSTER_CONFIG}:
	kind create cluster \
		--image ${IMAGE} \
		--kubeconfig ${CLUSTER_CONFIG} \
		--name ${CLUSTER_NAME} \
		--config=./site/content/docs/latest/reference/manifests/kubeapps-local-dev-apiserver-config.yaml \
		--retain \
		--wait 10s
	kubectl apply --kubeconfig=${CLUSTER_CONFIG} -f ./site/content/docs/latest/reference/manifests/kubeapps-local-dev-users-rbac.yaml
	kubectl apply --kubeconfig=${CLUSTER_CONFIG} -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/main/deploy/static/provider/kind/deploy.yaml
	# TODO: need to add wait for condition=exists or similar - https://github.com/kubernetes/kubernetes/issues/83242
	sleep 5
	kubectl wait --kubeconfig=${CLUSTER_CONFIG} --namespace ingress-nginx \
		--for=condition=ready pod \
		--selector=app.kubernetes.io/component=controller \
		--timeout=120s

cluster-kind: ${CLUSTER_CONFIG}

# dex will be running on the same node as the API server in the dev environment, so we can
# reuse the key and cert from the apiserver, which already includes v3 extensions
# for the correct alternative name (using the IP address).
devel/dex.crt:
	docker cp ${CLUSTER_NAME}-control-plane:/etc/kubernetes/pki/apiserver.crt ./devel/dex.crt

devel/dex.key:
	docker cp ${CLUSTER_NAME}-control-plane:/etc/kubernetes/pki/apiserver.key ./devel/dex.key

${ADDITIONAL_CLUSTER_CONFIG}: devel/dex.crt
	kind create cluster \
		--kubeconfig ${ADDITIONAL_CLUSTER_CONFIG} \
		--name ${ADDITIONAL_CLUSTER_NAME} \
		--config=./site/content/docs/latest/reference/manifests/kubeapps-local-dev-additional-apiserver-config.yaml \
		--retain \
		--wait 10s
	kubectl apply --kubeconfig=${ADDITIONAL_CLUSTER_CONFIG} -f ./site/content/docs/latest/reference/manifests/kubeapps-local-dev-users-rbac.yaml
	kubectl apply --kubeconfig=${ADDITIONAL_CLUSTER_CONFIG} -f ./site/content/docs/latest/reference/manifests/kubeapps-local-dev-namespace-discovery-rbac.yaml

additional-cluster-kind: ${ADDITIONAL_CLUSTER_CONFIG}

multi-cluster-kind: cluster-kind additional-cluster-kind

delete-cluster-kind:
	kind delete cluster --name ${CLUSTER_NAME} || true
	kind delete cluster --name ${ADDITIONAL_CLUSTER_NAME} || true
	rm ${CLUSTER_CONFIG}
	rm ${ADDITIONAL_CLUSTER_CONFIG} || true
	rm devel/dex.* || true

.PHONY: additional-cluster-kind cluster-kind cluster-kind-delete multi-cluster-kind pause
