# This Makefile aims to help local development with an OpenShift/Minishift cluster. It is not for production
# use, but does document requirements for OpenShift. The targets assume that you have:
# 1) helm installed 
# 2) minishift installed and a cluster started (which automatically updates your KUBECONFIG).
# 3) the `oc` cli setup (ie. you've run `eval $(minishift oc-env)`) and are authed as `system:admin`
KUBEAPPS_NAMESPACE ?= kubeapps

devel/openshift-kubeapps-project-created: 
	oc new-project ${KUBEAPPS_NAMESPACE}
	touch $@

chart/kubeapps/charts/mongodb-${MONGODB_CHART_VERSION}.tgz:
	helm dep build ./chart/kubeapps

devel/openshift-kubeapps-installed:
	@oc project ${KUBEAPPS_NAMESPACE}
	helm install ./chart/kubeapps -n ${KUBEAPPS_NAMESPACE} \
		--values ./docs/user/manifests/kubeapps-local-dev-values.yaml

openshift-kubeapps: devel/openshift-kubeapps-installed

openshift-kubeapps-reset:
	oc delete project ${KUBEAPPS_NAMESPACE} || true
	oc delete customresourcedefinition apprepositories.kubeapps.com || true
	rm devel/openshift-* || true

.PHONY: openshift-kubeapps openshift-kubeapps-reset
