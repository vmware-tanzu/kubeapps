# Deploy a dev environment of Kubeapps using OIDC for authentication with a
# local dex as the provider.
#
# Targets in this helper assume that kubectl is configured with a cluster
# that has been setup with OIDC support (see ./cluster-kind.mk)

deploy-dex: devel/dex.crt devel/dex.key
	kubectl --kubeconfig=${CLUSTER_CONFIG} create namespace dex
	kubectl --kubeconfig=${CLUSTER_CONFIG} -n dex create secret tls dex-web-server-tls \
		--key ./devel/dex.key \
		--cert ./devel/dex.crt
	helm --kubeconfig=${CLUSTER_CONFIG} repo add stable https://charts.helm.sh/stable
	helm --kubeconfig=${CLUSTER_CONFIG} install dex stable/dex --namespace dex  --values ./docs/user/manifests/kubeapps-local-dev-dex-values.yaml

deploy-openldap:
	kubectl --kubeconfig=${CLUSTER_CONFIG} create namespace ldap
	helm --kubeconfig=${CLUSTER_CONFIG} repo add stable https://charts.helm.sh/stable
	helm --kubeconfig=${CLUSTER_CONFIG} install ldap stable/openldap --namespace ldap \
		--values ./docs/user/manifests/kubeapps-local-dev-openldap-values.yaml

# Get mkcert from https://github.com/FiloSottile/mkcert/releases
devel/localhost-cert.pem:
	mkcert -key-file ./devel/localhost-key.pem -cert-file ./devel/localhost-cert.pem localhost 172.18.0.2

deploy-dependencies: deploy-dex deploy-openldap devel/localhost-cert.pem
	kubectl --kubeconfig=${CLUSTER_CONFIG} create namespace kubeapps
	kubectl --kubeconfig=${CLUSTER_CONFIG} -n kubeapps create secret tls localhost-tls \
		--key ./devel/localhost-key.pem \
		--cert ./devel/localhost-cert.pem
	kubectl --kubeconfig=${CLUSTER_CONFIG} -n kubeapps create secret generic postgresql-db \
		--from-literal=postgresql-postgres-password=dev-only-fake-password \
		--from-literal=postgresql-password=dev-only-fake-password

deploy-dev-kubeapps:
	helm --kubeconfig=${CLUSTER_CONFIG} upgrade --install kubeapps ./chart/kubeapps --namespace kubeapps --create-namespace \
		--values ./docs/user/manifests/kubeapps-local-dev-values.yaml \
		--values ./docs/user/manifests/kubeapps-local-dev-auth-proxy-values.yaml \
		--values ./docs/user/manifests/kubeapps-local-dev-additional-kind-cluster.yaml

deploy-dev: deploy-dependencies deploy-dev-kubeapps
	@echo "\nYou can now simply open your browser at https://localhost/ to access Kubeapps!"
	@echo "When logging in, you will be redirected to dex (with a self-signed cert) and can login with email as either of"
	@echo "  kubeapps-operator@example.com:password"
	@echo "  kubeapps-user@example.com:password"
	@echo "or with LDAP as either of"
	@echo "  kubeapps-operator-ldap@example.org:password"
	@echo "  kubeapps-user-ldap@example.org:password"
	@echo "to authenticate with the corresponding permissions."

reset-dev-kubeapps:
	kubectl delete namespace --wait kubeapps

# The kapp-controller support for the new Package and PackageRepository CRDs is currently
# only available in an alpha release.
deploy-kapp-controller:
	kubectl --kubeconfig=${CLUSTER_CONFIG} apply -f https://raw.githubusercontent.com/vmware-tanzu/carvel-kapp-controller/develop/alpha-releases/v0.20.0-rc.1.yml

# Add the flux controllers used for testing the kubeapps-apis integration.
deploy-flux-controllers:
	kubectl --kubeconfig=${CLUSTER_CONFIG} apply -f https://github.com/fluxcd/flux2/releases/download/v0.16.0/install.yaml

reset-dev:
	helm --kubeconfig=${CLUSTER_CONFIG} -n kubeapps delete kubeapps  || true
	helm --kubeconfig=${CLUSTER_CONFIG} -n dex delete dex  || true
	helm --kubeconfig=${CLUSTER_CONFIG} -n ldap delete ldap || true
	kubectl delete namespace --wait dex ldap kubeapps || true

.PHONY: deploy-dex deploy-dependencies deploy-dev deploy-openldap reset-dev
