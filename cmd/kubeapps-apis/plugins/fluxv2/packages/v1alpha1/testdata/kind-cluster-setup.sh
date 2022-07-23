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
OCI_REGISTRY_USER=foo
OCI_REGISTRY_PWD=bar

function pushChartToLocalRegistryUsingHelmCLI() {
  max=5  
  n=0
  until [ $n -ge $max ]
  do
   helm registry login -u $OCI_REGISTRY_USER localhost:$OCI_REGISTRY_LOCAL_PORT -p $OCI_REGISTRY_PWD && break
   n=$((n+1)) 
   echo "Retrying helm login in 5s [$n/$max]..."
   sleep 5
  done
  if [[ $n -ge $max ]]; then
    echo "Failed to login to helm registry [localhost:$OCI_REGISTRY_LOCAL_PORT] after [$max] attempts. Exiting..."
    exit 1
  fi

  # these .tgz files were pulled from https://stefanprodan.github.io/podinfo/ 
  CMD="helm push charts/podinfo-6.0.0.tgz oci://localhost:$OCI_REGISTRY_LOCAL_PORT/helm-charts"
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
      echo kubectl port-forward to [$OCI_REGISTRY_LOCAL_PORT] failed to respond within expected time limit...
      exit 1
    fi
    echo Sleeping 1s until kubectl port-forward process starts responding on port [$OCI_REGISTRY_LOCAL_PORT] [$n/$max]...
    sleep 1
  done
}

# the goal is to create an OCI registry whose contents I completely control and will modify 
# by running integration tests. Therefore 'pushChartToMyGitHubRegistry'
# ref https://docs.github.com/en/packages/working-with-a-github-packages-registry/working-with-the-container-registry
# TODO gfichtenholtz By default the new OCI registry is private. Attempts to pull a chart down
# from this repo from the code fails with:
# I0720 03:07:23.083292       1 oci_repo.go:269] helmGetter.Get(ghcr.io/gfichtenholt/helm-charts/podinfo:6.1.6) returned error: failed to authorize: failed to fetch anonymous token: unexpected status: 401 Unauthorized 
# Even though the Helm getter is programmatically configured with GITHUB_USER/GITHUB_TOKEN
# In order to make progress I made my OCI registry public per
# https://docs.github.com/en/packages/learn-github-packages/configuring-a-packages-access-control-and-visibility
function pushChartToMyGitHubRegistry() {
  if [ $# -lt 1 ]
  then
    echo "Usage : $0 chart.tgz"
    exit
  fi

  max=5  
  n=0
  until [ $n -ge $max ]
  do
   helm registry login ghcr.io -u $GITHUB_USER -p $GITHUB_TOKEN && break
   n=$((n+1)) 
   echo "Retrying helm login in 5s [$n/$max]..."
   sleep 5
  done
  if [[ $n -ge $max ]]; then
    echo "Failed to login to helm registry [ghcr.io] after [$max] attempts. Exiting..."
    exit 1
  fi

  trap '{
    helm registry logout ghcr.io
  }' EXIT  

  CMD="helm push $1 oci://ghcr.io/gfichtenholt/helm-charts"
  echo Starting command: $CMD...
  $CMD
  echo Command completed

  # sanity checks
  # TODO (gfichtenholt) display all versions. 'show all' only shows the latest 
  helm show all oci://ghcr.io/gfichtenholt/helm-charts/podinfo | head -9

  echo 
  echo You can see all packages at [https://github.com/gfichtenholt?tab=packages]
  echo You can see and delete any specific package version at 
  echo  [https://github.com/users/gfichtenholt/packages/container/helm-charts%2Fpodinfo/versions]
  echo
}

function deleteChartVersionFromMyGitHubRegistry() {
  if [ $# -lt 1 ]
  then
    echo "Usage : $0 version"
    exit
  fi

  # ref https://docs.github.com/en/rest/packages#get-all-package-versions-for-a-package-owned-by-the-authenticated-user
  PACKAGE_VERSION_ID=$(gh api   -H "Accept: application/vnd.github+json" /users/gfichtenholt/packages/container/helm-charts%2Fpodinfo/versions | jq -rc '.[] | select(.metadata.container.tags[] | contains("6.1.6")) | .id')
  if [[ "$PACKAGE_VERSION_ID" != "" ]]; then
     # ref https://docs.github.com/en/rest/packages#delete-a-package-version-for-the-authenticated-user
     # ref https://github.com/cli/cli/issues/3937
     echo -n | gh api   --method DELETE   -H "Accept: application/vnd.github+json"   /user/packages/container/helm-charts%2Fpodinfo/versions/$PACKAGE_VERSION_ID --input -
     # one can verify that the version has been deleted on web portal
     # https://github.com/users/gfichtenholt/packages/container/helm-charts%2Fpodinfo/versions
  fi

  # TODO (gfichtenholt) sanity check 
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
  local max=20
  local n=0
  while [[ $(kubectl get pods -l app=registry-app -o 'jsonpath={..status.conditions[?(@.type=="Ready")].status}') != "True" ]]; do
    n=$((n+1))
    if [[ $n -ge $max ]]; then
      echo "registry-app pod did not reach Ready state within expected time limit..."
      exit 1
    fi
    echo "Waiting 1s for registry-app pod to reach Ready state [$n/$max]..."
    sleep 1
  done

  # can't quite get local registry to work yet. Want to move on. I will come back to 
  # this as time allows
  # portForwardToLocalRegistry
  # pushChartToLocalRegistryUsingHelmCLI

  VERSION=6.1.5 

  # these .tgz files were pulled from https://stefanprodan.github.io/podinfo/ + .tgz
  pushChartToMyGitHubRegistry "charts/podinfo-$VERSION.tgz"

  # TODO (gfichtenholt) automate this so it does not require interaction
  # ref https://docs.github.com/en/rest/packages#about-the-github-packages-api
  open https://github.com/users/gfichtenholt/packages/container/helm-charts%2Fpodinfo/versions
  echo 
  echo Make sure that: 
  echo  - package visibility for [podinfo] is [public] 
  echo    at https://github.com/users/gfichtenholt/packages/container/helm-charts%2Fpodinfo/settings 
  echo - the only package version is [$VERSION] 
  echo    at https://github.com/users/gfichtenholt/packages/container/helm-charts%2Fpodinfo/versions
  echo before running the integration tests
  echo 
  read -p "Press any key to resume ..."

  # TODO (gfichtenholt) sanity check
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


if [ $# -lt 1 ]
then
  echo "Usage : $0 deploy|undeploy|redeploy|shell|logs|pushChartToMyGithub|deleteChartVersionFromMyGitHub"
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
*) echo "Invalid command: $1"
   ;;
esac
