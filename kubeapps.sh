#!/bin/sh
#
# mock implementation

GCTAG=kubeapps

mydir=${0%/*}
KUBECFG_JPATH=$mydir/lib:$mydir/vendor/kubecfg/lib:$mydir/vendor/ksonnet-lib
export KUBECFG_JPATH

subcommand="${1?Missing subcommand}"; shift

mkpw() {
    python -c 'import random, string; print "".join(random.choice(string.lowercase) for i in xrange(20))'
}

set -x -e

case "$subcommand" in
    up)
        if ! kubectl get -n kubeapps secret/mongodb >/dev/null 2>&1; then
            set +x
            echo "Creating mongodb password.."
            kubectl create namespace kubeapps || :
            kubectl -n kubeapps create secret generic mongodb \
                    --from-literal=mongodb-password=$(mkpw) \
                    --from-literal=mongodb-root-password=$(mkpw)
            set -x
        fi
        kubecfg update -v --gc-tag=$GCTAG kubeapps.jsonnet "$@"
        # TODO(gus): Actually implement `kubecfg update --wait`
        kubectl rollout status -n kubeapps deployment/kubeapps-hub-ui
        kubectl rollout status -n kubeapps deployment/kubeapps-hub-api
        kubectl rollout status -n kube-system deployment/nginx-ingress-controller
        ;;
    down)
        # This assumes kubeapps.jsonnet is in sync with what's
        # currently running.
        # FIXME(gus): add support for deletion using the garbage
        # collection mechanism.
        kubectl --ignore-not-found -n kubeapps delete secret mongodb
        kubecfg delete -v kubeapps.jsonnet "$@"
        ;;
    show)
        exec kubecfg show kubeapps.jsonnet "$@"
        ;;
    dashboard)
        echo "NB until front page is updated: kubeless-ui is at /kubeless" >&2
        exec minikube service -n kube-system nginx-ingress "$@"
        ;;
    *)
        echo "Unknown subcommand: $subcommand" >&2
        exit 1
esac
