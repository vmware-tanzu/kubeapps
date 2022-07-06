#!/bin/bash

# Copyright 2021-2022 the Kubeapps contributors.
# SPDX-License-Identifier: Apache-2.0

# this is used to build an image that can be used to stand-up a pod that serves static test-data in 
# local kind cluster. Used by the integration tests. This script needs to be run once before the running
# the test(s) 
set -o errexit
set -o nounset
set -o pipefail

function deploy {
  TAG=0.0.11
  docker build -t kubeapps/fluxv2plugin-testdata:$TAG .
  # "kubeapps" is the name of the kind cluster
  kind load docker-image kubeapps/fluxv2plugin-testdata:$TAG --name kubeapps
  kubectl create deployment --image=kubeapps/fluxv2plugin-testdata:$TAG fluxv2plugin-testdata-app --context kind-kubeapps
  kubectl set env deployment/fluxv2plugin-testdata-app DOMAIN=cluster --context kind-kubeapps
  kubectl expose deployment fluxv2plugin-testdata-app --port=80 --target-port=80 --name=fluxv2plugin-testdata-svc --context kind-kubeapps
  kubectl expose deployment fluxv2plugin-testdata-app --port=443 --target-port=443 --name=fluxv2plugin-testdata-ssl-svc --context kind-kubeapps
  # set up OCI registry
  # ref https://helm.sh/docs/topics/registries/
  # only has a single user: foo, password: bar
  kubectl create configmap registry-configmap --from-file ./bcrypt.htpasswd
  kubectl apply -f registry-app.yaml
  kubectl expose deployment registry-app --port=5000 --target-port=5000 --name=registry-app-svc --context kind-kubeapps
  while [[ $(kubectl get pods -l app=registry-app -o 'jsonpath={..status.conditions[?(@.type=="Ready")].status}') != "True" ]]; do
    echo "Waiting for registry-app to reach Ready state..."
    sleep 1
  done
  # TODO gfichtenholt this needs to be done asynchronously
  kubectl -n default port-forward svc/registry-app-svc 5000:5000 --context kind-kubeapps 

  local max=5  
  local n=0
  until [ "$n" -ge "$max" ]
  do
   helm registry login -u foo localhost:5000 -p bar && break
   n=$((n+1)) 
   echo "Retrying helm login in 5s [$n/5]..."
   sleep 5
  done
  if [[ $n -ge $max ]]; then
    echo "Failed to login with [5] attempts.Exiting script..."
    exit 1
  fi

  # these .tgz files were pulled from https://stefanprodan.github.io/podinfo/ 
  helm push charts/podinfo-6.0.0.tgz oci://localhost:5000/helm-charts 
  helm show all oci://localhost:5000/helm-charts/podinfo | head -9
  helm registry logout localhost:5000
}

function undeploy {
  kubectl delete svc/fluxv2plugin-testdata-svc
  kubectl delete svc/fluxv2plugin-testdata-ssl-svc
  kubectl delete svc/registry-app-svc
  kubectl delete deployment/registry-app --context kind-kubeapps 
  kubectl delete deployment fluxv2plugin-testdata-app --context kind-kubeapps 
}

function redeploy {
   undeploy
   deploy
}


function portforward {
  kubectl -n default port-forward svc/fluxv2plugin-testdata-svc 8081:80 --context kind-kubeapps 
  #kubectl -n default port-forward svc/fluxv2plugin-testdata-ssl-svc 8081:443 --context kind-kubeapps 
}

function shell {
  #kubectl exec --stdin --tty fluxv2plugin-testdata-app-74766cf559-695qg -- /bin/bash --context kind-kubeapps 
  RANDOM=$$
  kubectl run -i --rm --tty centos-$RANDOM --image=centos --restart=Never -- /bin/bash
}

function logs {
  kubectl logs pod/$(kubectl get pod -n default | grep fluxv2plugin | head -n 1 | awk '{print $1}') -n default --context kind-kubeapps 
}


if [ $# -lt 1 ]
then
  echo "Usage : $0 deploy|undeploy|redeploy|portforward|shell|logs"
  exit
fi

case "$1" in
deploy) deploy
    ;;
undeploy) undeploy
    ;;
redeploy) redeploy
    ;;
portforward) portforward
    ;;
shell) shell
    ;;
logs) logs
    ;;
*) echo "Invalid command: $1"
   ;;
esac
