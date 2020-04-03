# Deploy a dev environment of Kubeapps using OIDC for authentication with a
# local dex as the provider.
#
# Targets in this helper assume that kubectl is configured with a cluster
# that has been setup with OIDC support (see ./cluster-kind.mk)

deploy-dex:
	kubectl create namespace dex
	kubectl -n dex create secret tls dex-web-server-tls \
		--key ./script/test-certs/dex.key.pem \
		--cert ./script/test-certs/dex.cert.pem
	helm install dex stable/dex --namespace dex --version 2.4.0 \
		--values ./docs/user/manifests/kubeapps-local-dev-dex-values.yaml

# The api server does not have service dns entries (in kind or vanilla k8s), so
# dex is normally required to be available on an external host. Short-circuit
# that requirement by ensuring dex.dex resolves to the internal IP address on
# the apiserver so that it can initialise the oidc plugin.
update-apiserver-etc-hosts:
	while ! kubectl -n kube-system get po kube-apiserver-kubeapps-control-plane; do \
		echo "Waiting for api server" && sleep 1; \
	done
	kubectl -n kube-system exec kube-apiserver-kubeapps-control-plane -- \
		sh -c "echo '$(shell kubectl -n dex get svc -o=jsonpath='{.items[0].spec.clusterIP}') dex.dex' >> /etc/hosts"

deploy-openldap:
	kubectl create namespace ldap
	helm install ldap stable/openldap --namespace ldap \
		--values ./docs/user/manifests/kubeapps-local-dev-openldap-values.yaml

deploy-dev: deploy-dex deploy-openldap update-apiserver-etc-hosts
	kubectl create namespace kubeapps
	helm install kubeapps ./chart/kubeapps --namespace kubeapps \
		--values ./docs/user/manifests/kubeapps-local-dev-values.yaml \
		--values ./docs/user/manifests/kubeapps-local-dev-auth-proxy-values.yaml \
		--set useHelm3=true \
		--set postgresql.enabled=true \
		--set postgresql.volumePermissions.enabled=false \
		--set postgresql.volumePermissions.image.pullPolicy=IfNotPresent \
		--set mongodb.enabled=false
	kubectl apply -f ./docs/user/manifests/kubeapps-local-dev-users-rbac.yaml
	@echo "\nEnsure you have the entry '127.0.0.1 dex.dex' in your /etc/hosts, then run\n"
	@echo "kubectl -n dex port-forward svc/dex 32000\n"
	@echo "and in another terminal using the same cluster,\n"
	@echo "kubectl -n kubeapps port-forward svc/kubeapps 3000:80\n"
	@echo "You can then open http://localhost:3000 and login with email as either of"
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
