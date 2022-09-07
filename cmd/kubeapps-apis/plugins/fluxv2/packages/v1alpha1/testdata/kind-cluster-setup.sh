#!/bin/bash

# Copyright 2021-2022 the Kubeapps contributors.
# SPDX-License-Identifier: Apache-2.0

# This script is used to:
# - build an image that can be used to stand-up a pod that serves static test-data in 
# local kind cluster. 
# - create and seed an OCI registry on ghcr.io
# - create and seed an OCI registry on demo.goharbor.io
# - create and seed an OCI registry on gcr.io
# These are usedby the integration tests. 
# This script needs to be run once before the running the test(s).
set -o errexit
set -o nounset
set -o pipefail

# see
# https://stackoverflow.com/questions/5947742/how-to-change-the-output-color-of-echo-in-linux
RED='\033[0;31m'
GREEN='\033[0;32m'
L_GREEN='\033[1;32m'
CYAN='\033[0;36m'
BLUE='\033[0;34m'
L_BLUE='\033[1;34m'
L_GRAY='\033[0;37m'
L_YELLOW='\033[0;33m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

#!/bin/bash 
# Absolute path to this script, e.g. /home/user/bin/foo.sh
SCRIPT=$(readlink -f "$0")
# Absolute path this script is in, thus /home/user/bin
SCRIPTPATH=$(dirname "$SCRIPT")

# An error exit function
function error_exit()
{
  echo "******************************************************************" 1>&2
	echo -e ${RED}"Error:${NC} $@. Exiting!" 1>&2
  echo "******************************************************************" 1>&2
	exit 1
}

OCI_REGISTRY_REMOTE_PORT=5000
OCI_REGISTRY_LOCAL_PORT=5000
LOCAL_OCI_REGISTRY_USER=foo
LOCAL_OCI_REGISTRY_PWD=bar

# this is the only package version used to seed podinfo OCI repository
OCI_PODINFO_CHART_VERSION=6.1.5

function pushChartToLocalRegistryUsingHelmCLI() {
  max=5  
  n=0
  until [ $n -ge $max ]
  do
   helm registry login -u $LOCAL_OCI_REGISTRY_USER localhost:$OCI_REGISTRY_LOCAL_PORT -p $LOCAL_OCI_REGISTRY_PWD && break
   n=$((n+1)) 
   echo "Retrying helm login in 5s [$n/$max]..."
   sleep 5
  done
  if [[ $n -ge $max ]]; then
    error_exit "Failed to login to helm registry [localhost:$OCI_REGISTRY_LOCAL_PORT] after [$max] attempts. Exiting..."
  fi

  # these .tgz files were pulled from https://stefanprodan.github.io/podinfo/ 
  CMD="helm push charts/podinfo-$OCI_PODINFO_CHART_VERSION.tgz oci://localhost:$OCI_REGISTRY_LOCAL_PORT/helm-charts"
  echo Starting command: $CMD...
  $CMD
  echo Command completed

  # currently fails on the client with 
  # Error: unexpected status: 400 Bad Request
  # due to server-side error:
  # time="2022-07-17T04:42:10.0289284Z" level=warning msg="error authorizing context: basic 
  # authentication challenge for realm "Registry Realm": invalid authorization credential" 
  # go.version=go1.16.15 http.request.host="localhost:52443" 
  # http.request.id=52e52475-08f0-4bb2-bc8a-9e8d09b93291 
  # http.request.method=GET http.request.remoteaddr="127.0.0.1:36168" 
  # http.request.uri="/v2/" http.request.useragent="Helm/3.9.1" 
  #
  # I think this might be due to https://github.com/helm/helm/issues/6324
  # How can I push chart to an insecure OCI registry with helm v3 
  #
  # Methinks the problem is the registry host uses a self-signed,
  # a.k.a. custom Certificate Authority and is therefore "insecure"
  # Any request to it needs to be encrypted with the custom CA, 
  # but Helm CLI 'push' command appears not to have any option to provide 
  # a CA for the registry host
  # 
  # TODO (gfichtenholt): 
  # How is it that I am able to do 'helm login' successfully without 
  # specifying custom CA, but 'docker login' only works if I point to it?

  helm show all oci://localhost:$OCI_REGISTRY_LOCAL_PORT/helm-charts/podinfo | head -9
  helm registry logout localhost:$OCI_REGISTRY_LOCAL_PORT
}

function portForwardToLocalRegistry() {
  # ref https://stackoverflow.com/questions/67415637/kubectl-port-forward-reliably-in-a-shell-script
  kubectl -n default port-forward svc/registry-svc $OCI_REGISTRY_LOCAL_PORT:$OCI_REGISTRY_REMOTE_PORT --context kind-kubeapps &
  
  pid=$!

  # kill the port-forward regardless of how this script exits
  trap '{
    echo Killing process kubectl port-forward to [$OCI_REGISTRY_LOCAL_PORT:$OCI_REGISTRY_REMOTE_PORT] with id [$pid]...
    kill $pid
  }' EXIT  

  # wait for kubectl port-forward to start responding on $OCI_REGISTRY_LOCAL_PORT
  local max=20
  local n=0
  while ! nc -vz localhost $OCI_REGISTRY_LOCAL_PORT > /dev/null 2>&1 ; do
    n=$((n+1))
    if [[ $n -ge $max ]]; then
      error_exit kubectl port-forward to [$OCI_REGISTRY_LOCAL_PORT] failed to respond within expected time limit...
    fi
    echo Sleeping 1s until kubectl port-forward process starts responding on port [$OCI_REGISTRY_LOCAL_PORT] [$n/$max]...
    sleep 1
  done
}

function deploy {
  TAG=0.0.11
  docker build -t kubeapps/fluxv2plugin-testdata:$TAG .
  # "kubeapps" is the name of the kind cluster
  kind load docker-image kubeapps/fluxv2plugin-testdata:$TAG --name kubeapps
  kubectl create deployment --image=kubeapps/fluxv2plugin-testdata:$TAG fluxv2plugin-testdata-app --context kind-kubeapps
  kubectl set env deployment/fluxv2plugin-testdata-app DOMAIN=cluster --context kind-kubeapps
  kubectl expose deployment fluxv2plugin-testdata-app --port=80 --target-port=80 --name=fluxv2plugin-testdata-svc --context kind-kubeapps
  kubectl expose deployment fluxv2plugin-testdata-app --port=443 --target-port=443 --name=fluxv2plugin-testdata-ssl-svc --context kind-kubeapps
  # set up a local OCI registry
  # ref https://helm.sh/docs/topics/registries/
  kubectl create secret tls registry-tls --key ./cert/server-key.pem --cert ./cert/ssl-bundle.pem
  kubectl apply -f registry-app.yaml
  local max=25
  local n=0
  while [[ $(kubectl get pods -l app=registry-app -o 'jsonpath={..status.conditions[?(@.type=="Ready")].status}') != "True" ]]; do
    n=$((n+1))
    if [[ $n -ge $max ]]; then
      error_exit "registry-app pod did not reach Ready state within expected time limit..."
    fi
    echo "Waiting 1s for registry-app pod to reach Ready state [$n/$max]..."
    sleep 1
  done

  # can't quite get local registry to work yet. See comment in 
  # TestKindClusterAvailablePackageEndpointsForOCI.
  # Want to move on. I will come back to this as time allows

  # portForwardToLocalRegistry
  # pushChartToLocalRegistryUsingHelmCLI

  pushChartToMyGitHubRegistry "$OCI_PODINFO_CHART_VERSION"

  # sanity checks
  myGithubRegistrySanityCheck
  
  setupGithubStefanProdanClone
  setupHarborStefanProdanClone
  setupGcrStefanProdanClone
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
  kubectl delete -f registry-app.yaml
  kubectl delete deployment fluxv2plugin-testdata-app --context kind-kubeapps 
  kubectl delete secret registry-tls --ignore-not-found=true
  
  deleteChartFromMyGithubRegistry
  set -e
}

function redeploy {
   undeploy
   deploy
}

function shell {
  #kubectl exec --stdin --tty fluxv2plugin-testdata-app-74766cf559-695qg -- /bin/bash --context kind-kubeapps 
  RANDOM=$$
  kubectl run -i --rm --tty centos-$RANDOM --image=centos --restart=Never -- /bin/bash
}

function logs {
  kubectl logs pod/$(kubectl get pod -n default | grep fluxv2plugin | head -n 1 | awk '{print $1}') -n default --context kind-kubeapps 
}

. ./ghcr-util.sh
. ./harbor-util.sh
. ./gcloud-util.sh

if [ $# -lt 1 ]
then
  echo "Usage : $0 deploy|undeploy|redeploy|shell|logs|pushChartToMyGithub|deleteChartVersionFromMyGitHub|setupGithubStefanProdanClone|setupHarborStefanProdanClone|setupGcrStefanProdanClone"
  exit
fi

case "$1" in
deploy) deploy
    ;;
undeploy) undeploy
    ;;
redeploy) redeploy
    ;;
shell) shell
    ;;
logs) logs
    ;;
# this is for integration tests TestKindClusterAddTagsToOciRepository and 
# TestKindClusterAutoUpdateInstalledPackageFromOciRepo
pushChartToMyGithub) pushChartToMyGitHubRegistry $2
    ;;
# this is for integration tests TestKindClusterAddTagsToOciRepository and 
# TestKindClusterAutoUpdateInstalledPackageFromOciRepo
deleteChartVersionFromMyGitHub) deleteChartVersionFromMyGitHubRegistry $2
    ;;
setupGithubStefanProdanClone) setupGithubStefanProdanClone
    ;;
setupHarborStefanProdanClone) setupHarborStefanProdanClone ${2: }
    ;;
setupHarborRobotAccount) setupHarborRobotAccount
    ;;
setupGcrStefanProdanClone) setupGcrStefanProdanClone
    ;;
*) error_exit "Invalid command: $1"
   ;;
esac
