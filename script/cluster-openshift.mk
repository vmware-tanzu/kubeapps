# This Makefile aims to help local development with an OpenShift/Minishift cluster. It is not for production
# use, but does document requirements for OpenShift. The targets assume that you have:
# 1) helm installed 
# 2) minishift installed and a cluster started (which automatically updates your KUBECONFIG).
# 3) the `oc` cli setup (ie. you've run `eval $(minishift oc-env)`) and are authed as `system:admin`
TILLER_NAMESPACE ?= tiller
KUBEAPPS_NAMESPACE ?= kubeapps

MONGODB_CHART_VERSION = $(strip $(shell cat chart/kubeapps/requirements.lock | grep version | cut --delimiter=":" -f2))

devel/openshift-tiller-project-created:
	oc new-project ${TILLER_NAMESPACE}
	touch $@

# TODO: The following three targets create enable tiller to create CRDs
# cluster-wide as well as app repositories in the TILLER_NAMESPACE. They should
# not be specific to OpenShift, but the default for any development
# environment, though the importance of this will vanish with Helm3.
devel/openshift-tiller-with-crd-rbac.yaml: devel/openshift-tiller-project-created
	@oc process -f ./docs/user/manifests/openshift-tiller-with-crd-rbac.yaml \
		-p TILLER_NAMESPACE="${TILLER_NAMESPACE}" \
		-o yaml \
	> $@

devel/openshift-tiller-with-apprepository-rbac.yaml: devel/openshift-tiller-with-crd-rbac.yaml
	@oc process -f ./docs/user/manifests/openshift-tiller-with-apprepository-rbac.yaml \
		-p TILLER_NAMESPACE="${TILLER_NAMESPACE}" \
		-p KUBEAPPS_NAMESPACE="${KUBEAPPS_NAMESPACE}" \
		-o yaml \
	> $@

openshift-install-tiller: devel/openshift-tiller-with-crd-rbac.yaml devel/openshift-tiller-with-apprepository-rbac.yaml devel/openshift-kubeapps-project-created
	kubectl --namespace ${TILLER_NAMESPACE} apply -f devel/openshift-tiller-with-crd-rbac.yaml --wait=true
	kubectl --namespace ${KUBEAPPS_NAMESPACE} apply -f devel/openshift-tiller-with-apprepository-rbac.yaml
	helm init --tiller-namespace ${TILLER_NAMESPACE} --service-account tiller --wait

devel/openshift-kubeapps-project-created: devel/openshift-tiller-project-created
	oc new-project ${KUBEAPPS_NAMESPACE}
	oc policy add-role-to-user edit "system:serviceaccount:${TILLER_NAMESPACE}:tiller"
	touch $@

chart/kubeapps/charts/mongodb-${MONGODB_CHART_VERSION}.tgz:
	helm dep build ./chart/kubeapps

devel/openshift-kubeapps-installed: openshift-install-tiller chart/kubeapps/charts/mongodb-${MONGODB_CHART_VERSION}.tgz
	@oc project ${KUBEAPPS_NAMESPACE}
	helm --tiller-namespace=${TILLER_NAMESPACE} install ./chart/kubeapps -n ${KUBEAPPS_NAMESPACE} \
		--set tillerProxy.host=tiller-deploy.${TILLER_NAMESPACE}:44134 \
		--values ./docs/user/manifests/kubeapps-local-dev-values.yaml

# Due to openshift having multiple secrets for the service account, the code is slightly different from
# that at https://github.com/kubeapps/kubeapps/blob/master/docs/user/getting-started.md#on-linuxmacos
# TODO: update this target to use a kubeapps user, rather than tiller service account.
openshift-tiller-token:
	@kubectl get secret -n "${TILLER_NAMESPACE}" \
		$(shell kubectl get serviceaccount -n "${TILLER_NAMESPACE}" tiller -o jsonpath='{range .secrets[*]}{.name}{"\n"}{end}' | grep tiller-token) \
		-o go-template='{{.data.token | base64decode}}' && echo

openshift-kubeapps: devel/openshift-kubeapps-installed

openshift-kubeapps-reset:
	oc delete project ${KUBEAPPS_NAMESPACE} || true
	oc delete project ${TILLER_NAMESPACE} || true
	oc delete -f devel/openshift-tiller-with-crd-rbac.yaml || true
	oc delete -f devel/openshift-tiller-with-apprepository-rbac.yaml || true
	oc delete customresourcedefinition apprepositories.kubeapps.com || true
	rm devel/openshift-* || true

.PHONY: openshift-install-tiller openshift-kubeapps openshift-kubeapps-reset
