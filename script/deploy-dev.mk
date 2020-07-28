# Deploy a dev environment of Kubeapps using OIDC for authentication with a
# local dex as the provider.
#
# Targets in this helper assume that kubectl is configured with a cluster
# that has been setup with OIDC support (see ./cluster-kind.mk)

deploy-dex: devel/dex.crt devel/dex.key
	kubectl create namespace dex
	kubectl -n dex create secret tls dex-web-server-tls \
		--key ./devel/dex.key \
		--cert ./devel/dex.crt
	helm install dex stable/dex --namespace dex --version 2.4.0 \
		--values ./docs/user/manifests/kubeapps-local-dev-dex-values.yaml

deploy-openldap:
	kubectl create namespace ldap
	helm install ldap stable/openldap --namespace ldap \
		--values ./docs/user/manifests/kubeapps-local-dev-openldap-values.yaml

deploy-dev: deploy-dex deploy-openldap
	kubectl create namespace kubeapps
	helm install kubeapps ./chart/kubeapps --namespace kubeapps \
		--values ./docs/user/manifests/kubeapps-local-dev-values.yaml \
		--values ./docs/user/manifests/kubeapps-local-dev-auth-proxy-values.yaml \
		--values ./docs/user/manifests/kubeapps-local-dev-additional-kind-cluster.yaml \
		--set useHelm3=true
	@echo "\nYou can now simply open your browser at http://172.18.0.2:30000 to access Kubeapps!"
	@echo "When logging in, you will be redirected to dex (with a self-signed cert) and can login with email as either of"
	@echo "  kubeapps-operator@example.com:password"
	@echo "  kubeapps-user@example.com:password"
	@echo "or with LDAP as either of"
	@echo "  kubeapps-operator-ldap@example.org:password"
	@echo "  kubeapps-user-ldap@example.org:password"
	@echo "to authenticate with the corresponding permissions."

reset-dev:
	helm -n kubeapps delete kubeapps || true
	helm -n dex delete dex || true
	helm -n ldap delete ldap || true
	# In case helm installations fail, still delete non-namespaced resources.
	kubectl delete clusterrole dex kubeapps:controller:apprepository-reader-kubeapps || true
	kubectl delete clusterrolebinding dex kubeapps:controller:apprepository-reader-kubeapps || true
	kubectl delete namespace --wait dex ldap kubeapps || true
	kubectl delete --wait -f ./docs/user/manifests/kubeapps-local-dev-users-rbac.yaml || true

.PHONY: deploy-dex deploy-dev deploy-openldap reset-dev update-apiserver-etc-hosts
