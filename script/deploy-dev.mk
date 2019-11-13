# Deploy a dev environment of Kubeapps using OIDC for authentication with a
# local dex as the provider.
#
# Targets in this helper assume that kubectl is configured with a cluster
# that has been setup with OIDC support (see ./cluster-kind.mk)

deploy-helm:
	kubectl apply -f ./docs/user/manifests/kubeapps-local-dev-tiller-rbac.yaml
	helm init --service-account tiller --wait

deploy-dex: deploy-helm
	kubectl create namespace dex
	kubectl -n dex create secret tls dex-web-server-tls \
		--key ./script/test-certs/dex.key.pem \
		--cert ./script/test-certs/dex.cert.pem
	helm install stable/dex --namespace dex --name dex --version 2.4.0 \
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

deploy-dev: deploy-dex update-apiserver-etc-hosts
	helm install ./chart/kubeapps --namespace kubeapps --name kubeapps \
		--values ./docs/user/manifests/kubeapps-local-dev-values.yaml \
		--values ./docs/user/manifests/kubeapps-local-dev-auth-proxy-values.yaml
	kubectl apply -f ./docs/user/manifests/kubeapps-local-dev-users-rbac.yaml
	@echo "\nEnsure you have the entry '127.0.0.1 dex.dex' in your /etc/hosts, then run\n"
	@echo "kubectl -n dex port-forward svc/dex 32000\n"
	@echo "and in another terminal using the same cluster,\n"
	@echo "kubectl -n kubeapps port-forward svc/kubeapps 3000:80\n"
	@echo "You can then open http://localhost:3000 and login as either of"
	@echo "  kubeapps-operator@example.com:password"
	@echo "  kubeapps-user@example.com:password"
	@echo "to authenticate with the corresponding permissions."

reset-dev:
	helm delete --purge kubeapps || true
	helm delete --purge dex || true
	kubectl delete namespace dex kubeapps || true
	helm reset || true
	kubectl delete -f ./docs/user/manifests/kubeapps-local-dev-tiller-rbac.yaml || true
	kubectl delete -f ./docs/user/manifests/kubeapps-local-dev-users-rbac.yaml

.PHONY: deploy-dex deploy-dev reset-dev update-apiserver-etc-hosts
