# Copyright 2021-2022 the Kubeapps contributors.
# SPDX-License-Identifier: Apache-2.0

# Deploy a dev environment of Kubeapps using OIDC for authentication with a
# local dex as the provider.
#
# Targets in this helper assume that kubectl is configured with a cluster
# that has been setup with OIDC support (see ./cluster-kind.mk)

# Have a look at /docs/reference/developer/pinniped-proxy.md for instructions on how to run this makefile

PINNIPED_VERSION ?= v0.18.0

deploy-dex-for-pinniped: devel/dex.crt-for-pinniped devel/dex.key-for-pinniped
	kubectl --kubeconfig=${CLUSTER_CONFIG_FOR_PINNIPED} create namespace dex
	kubectl --kubeconfig=${CLUSTER_CONFIG_FOR_PINNIPED} -n dex create secret tls dex-web-server-tls \
		--key ./devel/dex.key \
		--cert ./devel/dex.crt
	helm --kubeconfig=${CLUSTER_CONFIG_FOR_PINNIPED} repo add dex https://charts.dexidp.io
	helm --kubeconfig=${CLUSTER_CONFIG_FOR_PINNIPED} install dex dex/dex --version 0.5.0 --namespace dex --values ./site/content/docs/latest/reference/manifests/kubeapps-local-dev-dex-values.yaml

deploy-openldap-for-pinniped:
	kubectl --kubeconfig=${CLUSTER_CONFIG_FOR_PINNIPED} create namespace ldap
	helm --kubeconfig=${CLUSTER_CONFIG_FOR_PINNIPED} repo add stable https://charts.helm.sh/stable
	helm --kubeconfig=${CLUSTER_CONFIG_FOR_PINNIPED} install ldap stable/openldap --namespace ldap \
		--values ./site/content/docs/latest/reference/manifests/kubeapps-local-dev-openldap-values.yaml

deploy-dependencies-for-pinniped: deploy-dex-for-pinniped deploy-openldap-for-pinniped devel/localhost-cert.pem deploy-pinniped
	kubectl --kubeconfig=${CLUSTER_CONFIG_FOR_PINNIPED} create namespace kubeapps
	kubectl --kubeconfig=${CLUSTER_CONFIG_FOR_PINNIPED} -n kubeapps create secret tls localhost-tls \
		--key ./devel/localhost-key.pem \
		--cert ./devel/localhost-cert.pem
	kubectl --kubeconfig=${CLUSTER_CONFIG_FOR_PINNIPED} -n kubeapps create secret generic postgresql-db \
		--from-literal=postgres-postgres-password=dev-only-fake-password \
		--from-literal=postgres-password=dev-only-fake-password

deploy-pinniped:
	kubectl --kubeconfig=${CLUSTER_CONFIG_FOR_PINNIPED} apply -f https://get.pinniped.dev/${PINNIPED_VERSION}/install-pinniped-concierge-crds.yaml
	kubectl --kubeconfig=${CLUSTER_CONFIG_FOR_PINNIPED} apply -f https://get.pinniped.dev/${PINNIPED_VERSION}/install-pinniped-concierge-resources.yaml

deploy-pinniped-additional:
	kubectl --kubeconfig=${ADDITIONAL_CLUSTER_CONFIG_FOR_PINNIPED} apply -f https://get.pinniped.dev/${PINNIPED_VERSION}/install-pinniped-concierge-crds.yaml
	kubectl --kubeconfig=${ADDITIONAL_CLUSTER_CONFIG_FOR_PINNIPED} apply -f https://get.pinniped.dev/${PINNIPED_VERSION}/install-pinniped-concierge-resources.yaml

add-pinniped-jwt-authenticator:
	kubectl --kubeconfig=${CLUSTER_CONFIG_FOR_PINNIPED} apply -f ./site/content/docs/latest/reference/manifests/kubeapps-pinniped-jwt-authenticator.yaml

add-pinniped-jwt-authenticator-additional:
	kubectl --kubeconfig=${ADDITIONAL_CLUSTER_CONFIG_FOR_PINNIPED} apply -f ./site/content/docs/latest/reference/manifests/kubeapps-pinniped-jwt-authenticator.yaml

delete-pinniped-jwt-authenticator:
	kubectl --kubeconfig=${CLUSTER_CONFIG_FOR_PINNIPED} delete -f ./site/content/docs/latest/reference/manifests/kubeapps-pinniped-jwt-authenticator.yaml
	kubectl --kubeconfig=${ADDITIONAL_CLUSTER_CONFIG_FOR_PINNIPED} delete -f ./site/content/docs/latest/reference/manifests/kubeapps-pinniped-jwt-authenticator.yaml

delete-pinniped:
	kubectl --kubeconfig=${CLUSTER_CONFIG_FOR_PINNIPED} delete -f https://get.pinniped.dev/${PINNIPED_VERSION}/install-pinniped-concierge-resources.yaml
	kubectl --kubeconfig=${ADDITIONAL_CLUSTER_CONFIG_FOR_PINNIPED} delete -f https://get.pinniped.dev/${PINNIPED_VERSION}/install-pinniped-concierge-crds.yaml

deploy-dev-kubeapps-for-pinniped:
	helm --kubeconfig=${CLUSTER_CONFIG_FOR_PINNIPED} upgrade --install kubeapps ./chart/kubeapps --namespace kubeapps --create-namespace \
		--values ./site/content/docs/latest/reference/manifests/kubeapps-local-dev-values.yaml \
		--set pinnipedProxy.enabled=true \
		--values ./site/content/docs/latest/reference/manifests/kubeapps-local-dev-auth-proxy-values.yaml \
		--values ./site/content/docs/latest/reference/manifests/kubeapps-local-dev-additional-kind-cluster-for-pinniped.yaml

deploy-dev-for-pinniped: deploy-dependencies-for-pinniped deploy-dev-kubeapps-for-pinniped
	@echo "\nYou can now simply open your browser at https://localhost/ to access Kubeapps!"
	@echo "When logging in, you will be redirected to dex (with a self-signed cert) and can login with email as either of"
	@echo "  kubeapps-operator@example.com:password"
	@echo "  kubeapps-user@example.com:password"
	@echo "or with LDAP as either of"
	@echo "  kubeapps-operator-ldap@example.org:password"
	@echo "  kubeapps-user-ldap@example.org:password"
	@echo "to authenticate with the corresponding permissions."

.PHONY: deploy-dex deploy-dependencies deploy-dev deploy-openldap reset-dev
