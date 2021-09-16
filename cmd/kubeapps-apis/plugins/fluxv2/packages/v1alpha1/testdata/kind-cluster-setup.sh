#!/bin/bash
# this is used to build an image that can be used to stand-up a pod that serves static test-data in 
# local kind cluster. Used by the integration tests. This script needs to be run once before the running
# the test(s) 
set -o errexit
set -o nounset
set -o pipefail

TAG=0.0.6

function deploy {
  docker build -t kubeapps/fluxv2plugin-testdata:$TAG .
  kind load docker-image kubeapps/fluxv2plugin-testdata:$TAG --name kubeapps
  kubectl create deployment --image=kubeapps/fluxv2plugin-testdata:$TAG fluxv2plugin-testdata-app --context kind-kubeapps
  kubectl set env deployment/fluxv2plugin-testdata-app DOMAIN=cluster --context kind-kubeapps
  kubectl expose deployment fluxv2plugin-testdata-app --port=80 --target-port=80 --name=fluxv2plugin-testdata-svc --context kind-kubeapps
}

function undeploy {
   kubectl delete svc/fluxv2plugin-testdata-svc
   kubectl delete deployment fluxv2plugin-testdata-app --context kind-kubeapps 
}

function portforward {
  kubectl -n default port-forward svc/fluxv2plugin-testdata-svc 8081:80 --context kind-kubeapps 
}

function shell {
  kubectl exec --stdin --tty fluxv2plugin-testdata-app-74766cf559-695qg -- /bin/bash --context kind-kubeapps 
}

if [ $# -lt 1 ]
then
  echo "Usage : $0 deploy|undeploy|portforward|shell"
  exit
fi

case "$1" in
deploy) deploy
    ;;
undeploy) undeploy
    ;;
portforward) portforward
    ;;
shell) shell
    ;;
*) echo "Invalid option"
   ;;
esac


