#!/bin/bash

# Copyright 2021-2022 the Kubeapps contributors.
# SPDX-License-Identifier: Apache-2.0

# this is used to build an image that can be used to stand-up a pod that serves static test-data in 
# local kind cluster. Used by the integration tests. This script needs to be run once before the running
# the test(s) 
set -o errexit
set -o nounset
set -o pipefail

OCI_REGISTRY_REMOTE_PORT=5000
OCI_REGISTRY_LOCAL_PORT=5000
#OCI_REGISTRY_REMOTE_PORT=443
#OCI_REGISTRY_LOCAL_PORT=52443

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
  kubectl create secret tls registry-tls --key ./cert/server-key.pem --cert ./cert/ssl-bundle.pem
  kubectl apply -f registry-app.yaml
  kubectl expose deployment registry-app --port=$OCI_REGISTRY_REMOTE_PORT --target-port=$OCI_REGISTRY_REMOTE_PORT --name=registry-app-svc --context kind-kubeapps
  local max=20
  local n=0
  while [[ $(kubectl get pods -l app=registry-app -o 'jsonpath={..status.conditions[?(@.type=="Ready")].status}') != "True" ]]; do
    n=$((n+1))
    if [[ $n -ge $max ]]; then
      echo "registry-app pod did not reach Ready state within expected time limit..."
      exit 1
    fi
    echo "Waiting 1s for registry-app to reach Ready state [$n/$max]..."
    sleep 1
  done

  # ref https://stackoverflow.com/questions/67415637/kubectl-port-forward-reliably-in-a-shell-script
  kubectl -n default port-forward svc/registry-app-svc $OCI_REGISTRY_LOCAL_PORT:$OCI_REGISTRY_REMOTE_PORT --context kind-kubeapps &
  
  pid=$!

  # kill the port-forward regardless of how this script exits
  trap '{
    echo Killing process kubectl port-forward to [$OCI_REGISTRY_LOCAL_PORT:$OCI_REGISTRY_REMOTE_PORT] with id [$pid]...
    kill $pid
  }' EXIT  

  # wait for $OCI_REGISTRY_LOCAL_PORT to become available
  while ! nc -vz localhost $OCI_REGISTRY_LOCAL_PORT > /dev/null 2>&1 ; do
    echo Sleeping 1s until local port [$OCI_REGISTRY_LOCAL_PORT] becomes available...
    sleep 1
  done
  
  # This should show that the port is open, like this:
  # PORT      STATE SERVICE
  # 52443/tcp open  unknown
  nmap -sT -p $OCI_REGISTRY_LOCAL_PORT localhost
  
  max=5  
  n=0
  until [ $n -ge $max ]
  do
   helm registry login -u foo localhost:$OCI_REGISTRY_LOCAL_PORT -p bar && break
   n=$((n+1)) 
   echo "Retrying helm login in 5s [$n/$max]..."
   sleep 5
  done
  if [[ $n -ge $max ]]; then
    echo "Failed to login to helm registry [localhost:$OCI_REGISTRY_LOCAL_PORT] with [$max] attempts. Exiting script..."
    exit 1
  fi

  # these .tgz files were pulled from https://stefanprodan.github.io/podinfo/ 
  echo about to run command helm push charts/podinfo-6.0.0.tgz oci://localhost:$OCI_REGISTRY_LOCAL_PORT/helm-charts...
  helm push charts/podinfo-6.0.0.tgz oci://localhost:$OCI_REGISTRY_LOCAL_PORT/helm-charts 
  # currently fails due to server error 
  # time="2022-07-17T04:42:10.0289284Z" level=warning msg="error authorizing context: basic 
  # authentication challenge for realm "Registry Realm": invalid authorization credential" 
  # go.version=go1.16.15 http.request.host="localhost:52443" 
  # http.request.id=52e52475-08f0-4bb2-bc8a-9e8d09b93291 
  # http.request.method=GET http.request.remoteaddr="127.0.0.1:36168" 
  # http.request.uri="/v2/" http.request.useragent="Helm/3.9.1" 
  #
  # I think this is https://github.com/helm/helm/issues/6324
  # How can I push chart to an insecure OCI registry with helm v3 
  #
  echo command completed 
  helm show all oci://localhost:$OCI_REGISTRY_LOCAL_PORT/helm-charts/podinfo | head -9
  helm registry logout localhost:$OCI_REGISTRY_LOCAL_PORT
}

function undeploy {
  set +e
  pid="$(ps -ef | grep port-forward | grep $OCI_REGISTRY_LOCAL_PORT | awk '{print $2}')"
  if [[ "$pid" != "" ]]; then
    echo "Killing process 'kubectl port-forward to [$OCI_REGISTRY_LOCAL_PORT]' with id [$pid]..."
    kill $pid
  fi

  kubectl delete svc/fluxv2plugin-testdata-svc
  kubectl delete svc/fluxv2plugin-testdata-ssl-svc
  kubectl delete svc/registry-app-svc
  kubectl delete configmap registry-configmap  
  kubectl delete deployment/registry-app --context kind-kubeapps 
  kubectl delete deployment fluxv2plugin-testdata-app --context kind-kubeapps 
  kubectl delete secret registry-tls --ignore-not-found=true
  set -e
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
