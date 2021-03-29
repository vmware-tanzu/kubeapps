# Deploy a dev environment of Kubeapps using OIDC for authentication with a
# local dex as the provider.
#
# Targets in this helper assume that kubectl is configured with a cluster
# that has been setup with OIDC support (see ./cluster-kind.mk)

# Have a look at /docs/developer/pinniped-proxy.md for instructions on how to run this makefile

deploy-dex-for-pinniped: devel/dex.crt-for-pinniped devel/dex.key-for-pinniped
	kubectl --kubeconfig=${CLUSTER_CONFIG_FOR_PINNIPED} create namespace dex
	kubectl --kubeconfig=${CLUSTER_CONFIG_FOR_PINNIPED} -n dex create secret tls dex-web-server-tls \
		--key ./devel/dex.key \
		--cert ./devel/dex.crt
	helm --kubeconfig=${CLUSTER_CONFIG_FOR_PINNIPED} repo add stable https://charts.helm.sh/stable
	helm --kubeconfig=${CLUSTER_CONFIG_FOR_PINNIPED} install dex stable/dex --namespace dex  --values ./docs/user/manifests/kubeapps-local-dev-dex-values.yaml

deploy-openldap-for-pinniped:
	kubectl --kubeconfig=${CLUSTER_CONFIG_FOR_PINNIPED} create namespace ldap
	helm --kubeconfig=${CLUSTER_CONFIG_FOR_PINNIPED} repo add stable https://charts.helm.sh/stable
	helm --kubeconfig=${CLUSTER_CONFIG_FOR_PINNIPED} install ldap stable/openldap --namespace ldap \
		--values ./docs/user/manifests/kubeapps-local-dev-openldap-values.yaml

deploy-dependencies-for-pinniped: deploy-dex-for-pinniped deploy-openldap-for-pinniped devel/localhost-cert.pem deploy-pinniped
	kubectl --kubeconfig=${CLUSTER_CONFIG_FOR_PINNIPED} create namespace kubeapps

deploy-pinniped:
	kubectl --kubeconfig=${CLUSTER_CONFIG_FOR_PINNIPED} apply -f https://github.com/vmware-tanzu/pinniped/releases/download/v0.6.0/install-pinniped-concierge.yaml

deploy-pinniped-additional:
	kubectl --kubeconfig=${ADDITIONAL_CLUSTER_CONFIG_FOR_PINNIPED} apply -f https://github.com/vmware-tanzu/pinniped/releases/download/v0.6.0/install-pinniped-concierge.yaml

add-pinniped-jwt-authenticator:
	kubectl --kubeconfig=${CLUSTER_CONFIG_FOR_PINNIPED} apply -f ./docs/user/manifests/kubeapps-pinniped-jwt-authenticator.yaml
	kubectl --kubeconfig=${ADDITIONAL_CLUSTER_CONFIG_FOR_PINNIPED} apply -f ./docs/user/manifests/kubeapps-pinniped-jwt-authenticator.yaml

delete-pinniped-jwt-authenticator:
	kubectl --kubeconfig=${CLUSTER_CONFIG_FOR_PINNIPED} delete -f ./docs/user/manifests/kubeapps-pinniped-jwt-authenticator.yaml
	kubectl --kubeconfig=${ADDITIONAL_CLUSTER_CONFIG_FOR_PINNIPED} delete -f ./docs/user/manifests/kubeapps-pinniped-jwt-authenticator.yaml

delete-pinniped:
	kubectl --kubeconfig=${CLUSTER_CONFIG_FOR_PINNIPED} delete -f https://github.com/vmware-tanzu/pinniped/releases/download/v0.6.0/install-pinniped-concierge.yaml
	kubectl --kubeconfig=${ADDITIONAL_CLUSTER_CONFIG_FOR_PINNIPED} delete -f https://github.com/vmware-tanzu/pinniped/releases/download/v0.6.0/install-pinniped-concierge.yaml

deploy-dev-kubeapps-for-pinniped:
	kubectl --kubeconfig=${CLUSTER_CONFIG_FOR_PINNIPED} -n kubeapps delete secret localhost-tls  || true
	helm --kubeconfig=${CLUSTER_CONFIG_FOR_PINNIPED} install kubeapps ./chart/kubeapps --namespace kubeapps --create-namespace \
		--values ./docs/user/manifests/kubeapps-local-dev-values-for-pinniped.yaml \
		--values ./docs/user/manifests/kubeapps-local-dev-auth-proxy-values.yaml \
		--values ./docs/user/manifests/kubeapps-local-dev-additional-kind-cluster-for-pinniped.yaml
	kubectl --kubeconfig=${CLUSTER_CONFIG_FOR_PINNIPED} -n kubeapps delete secret localhost-tls 
	kubectl --kubeconfig=${CLUSTER_CONFIG_FOR_PINNIPED} -n kubeapps create secret tls localhost-tls \
		--key ./devel/localhost-key.pem \
		--cert ./devel/localhost-cert.pem

deploy-dev-for-pinniped: deploy-dependencies-for-pinniped deploy-dev-kubeapps-for-pinniped deploy-pinniped-additional
	@echo "\nYou can now simply open your browser at https://localhost/ to access Kubeapps!"
	@echo "When logging in, you will be redirected to dex (with a self-signed cert) and can login with email as either of"
	@echo "  kubeapps-operator@example.com:password"
	@echo "  kubeapps-user@example.com:password"
	@echo "or with LDAP as either of"
	@echo "  kubeapps-operator-ldap@example.org:password"
	@echo "  kubeapps-user-ldap@example.org:password"
	@echo "to authenticate with the corresponding permissions."

.PHONY: deploy-dex deploy-dependencies deploy-dev deploy-openldap reset-dev
