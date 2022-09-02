# Copyright 2021-2022 the Kubeapps contributors.
# SPDX-License-Identifier: Apache-2.0

# This file provides targets which create a local k8s cluster setup
# with OIDC integration for development and testing.

# Have a look at /docs/reference/developer/pinniped-proxy.md for instructions on how to run this makefile

KUBE ?= ${HOME}/.kube
CLUSTER_NAME_FOR_PINNIPED ?= kubeapps-for-pinniped
ADDITIONAL_CLUSTER_NAME_FOR_PINNIPED ?= kubeapps-for-pinniped-additional

CLUSTER_CONFIG_FOR_PINNIPED = ${KUBE}/kind-config-${CLUSTER_NAME_FOR_PINNIPED}
ADDITIONAL_CLUSTER_CONFIG_FOR_PINNIPED = ${KUBE}/kind-config-${ADDITIONAL_CLUSTER_NAME_FOR_PINNIPED}

${CLUSTER_CONFIG_FOR_PINNIPED}-for-pinniped:
	kind create cluster \
		--kubeconfig ${CLUSTER_CONFIG_FOR_PINNIPED} \
		--name ${CLUSTER_NAME_FOR_PINNIPED} \
		--config=./site/content/docs/latest/reference/manifests/kubeapps-local-dev-apiserver-no-oidc-config.yaml \
		--retain \
		--wait 10s
	kubectl apply --kubeconfig=${CLUSTER_CONFIG_FOR_PINNIPED} -f ./site/content/docs/latest/reference/manifests/kubeapps-local-dev-users-rbac.yaml
	kubectl apply --kubeconfig=${CLUSTER_CONFIG_FOR_PINNIPED} -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/main/deploy/static/provider/kind/deploy.yaml
	# TODO: need to add wait for condition=exists or similar - https://github.com/kubernetes/kubernetes/issues/83242
	sleep 5
	kubectl wait --kubeconfig=${CLUSTER_CONFIG_FOR_PINNIPED} --namespace ingress-nginx \
		--for=condition=ready pod \
		--selector=app.kubernetes.io/component=controller \
		--timeout=120s

cluster-kind-for-pinniped: ${CLUSTER_CONFIG_FOR_PINNIPED}-for-pinniped

devel/dex.crt-for-pinniped:
	docker cp kubeapps-for-pinniped-control-plane:/etc/kubernetes/pki/apiserver.crt ./devel/dex.crt

devel/dex.key-for-pinniped:
	docker cp kubeapps-for-pinniped-control-plane:/etc/kubernetes/pki/apiserver.key ./devel/dex.key

${ADDITIONAL_CLUSTER_CONFIG_FOR_PINNIPED}-for-pinniped: devel/dex.crt-for-pinniped
	kind create cluster \
		--kubeconfig ${ADDITIONAL_CLUSTER_CONFIG_FOR_PINNIPED} \
		--name ${ADDITIONAL_CLUSTER_NAME_FOR_PINNIPED} \
		--config=./site/content/docs/latest/reference/manifests/kubeapps-local-dev-additional-apiserver-config-for-pinniped.yaml \
		--retain \
		--wait 10s
	kubectl apply --kubeconfig=${ADDITIONAL_CLUSTER_CONFIG_FOR_PINNIPED} -f ./site/content/docs/latest/reference/manifests/kubeapps-local-dev-users-rbac.yaml
	kubectl apply --kubeconfig=${ADDITIONAL_CLUSTER_CONFIG_FOR_PINNIPED} -f ./site/content/docs/latest/reference/manifests/kubeapps-local-dev-namespace-discovery-rbac.yaml

additional-cluster-kind-for-pinniped: ${ADDITIONAL_CLUSTER_CONFIG_FOR_PINNIPED}-for-pinniped

multi-cluster-kind-for-pinniped: cluster-kind-for-pinniped additional-cluster-kind-for-pinniped

delete-cluster-kind-for-pinniped:
	kind delete cluster --name ${CLUSTER_NAME_FOR_PINNIPED} || true
	kind delete cluster --name ${ADDITIONAL_CLUSTER_NAME_FOR_PINNIPED} || true
	rm ${CLUSTER_CONFIG_FOR_PINNIPED}
	rm ${ADDITIONAL_CLUSTER_CONFIG_FOR_PINNIPED} || true
	rm devel/dex.* || true

.PHONY: additional-cluster-kind-for-pinniped cluster-kind-for-pinniped cluster-kind-delete-for-pinniped multi-cluster-kind-for-pinniped pause
