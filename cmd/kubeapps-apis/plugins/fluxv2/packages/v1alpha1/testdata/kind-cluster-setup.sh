#!/bin/bash

# Copyright 2021-2022 the Kubeapps contributors.
# SPDX-License-Identifier: Apache-2.0

# This script is used to:
# - build an image that can be used to stand-up a pod that serves static test-data in 
# local kind cluster. 
# - create and seed an OCI registry on ghcr.io
# These are usedby the integration tests. 
# This script needs to be run once before the running the test(s).
# This script requires GitHub CLI (gh) to be installed locally. On MacOS you can
# install it via 'brew install gh'. gh releases page: https://github.com/cli/cli/releases 
set -o errexit
set -o nounset
set -o pipefail

OCI_REGISTRY_REMOTE_PORT=5000
OCI_REGISTRY_LOCAL_PORT=5000
LOCAL_OCI_REGISTRY_USER=foo
LOCAL_OCI_REGISTRY_PWD=bar
GITHUB_OCI_REGISTRY_URL=oci://ghcr.io/gfichtenholt/helm-charts
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
    echo "Failed to login to helm registry [localhost:$OCI_REGISTRY_LOCAL_PORT] after [$max] attempts. Exiting..."
    exit 1
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
#
# Note the difference between flux/helm and GitHub terminology:
#   1) flux and helm uses the following terms:
#      Given a URL like "oci://ghcr.io/stefanprodan/charts" and 
#      a repository like "podinfo" and 
#      an image/version like "podinfo:6.1.5", the terms are
#      - oci://              - URL scheme, indicating this is an OCI registry, compared with HTTP registry
#      - ghcr.io             - registry host
#      - stefanprodan/charts - registry path
#      - podinfo             - repository
#      - podinfo:6.1.5       - application, a.k.a. package and version, a.k.a. tag
#      "oci://ghcr.io/stefanprodan/charts" is the registry URL
#      A given registry may have multiple repositories 
#      A given repository may have multiple packages
#      A given package may have multiple versions
#   2) GitHub Container Registry WebPortal and API do not use the term "OCI registry" and "OCI repository":  
#      - oci://ghcr.io         - host or "base", always the same
#        - stefanprodan        - owner
#        - charts/podinfo      - package, whose type is "container". 
#                                In some docs, also referred to by IMAGE_NAME
#        - 6.1.5               - package version a.k.a. tag
#      A given owner may have mutiple packages, e.g. "nginx/nginx", "charts/podinfo", etc
#      A given package may have multiple versions
#      There is no concept of a single repository containing multiple packages
function pushChartToMyGitHubRegistry() {
  if [ $# -lt 1 ]
  then
    echo "Usage : $0 version"
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

  # these .tgz files were originally sourced from https://stefanprodan.github.io/podinfo/podinfo-{version}.tgz
  CMD="helm push charts/podinfo-$1.tgz $GITHUB_OCI_REGISTRY_URL"
  echo Starting command: $CMD...
  $CMD
  echo Command completed

  # sanity checks
 
  # Display all chart versions
 
  # TODO (gfichtenholt) currently prints:
  # github.com
  # ✓ Logged in to github.com as gfichtenholt (GITHUB_TOKEN)
  # ✓ Git operations for github.com configured to use https protocol.
  # ✓ Token: *******************
  # find out if/when I need to login and logout via
  # gh auth login --hostname github.com --web --scopes read:packages
  # gh auth logout --hostname github.com
  gh auth status

  # ref https://docs.github.com/en/rest/packages#get-all-package-versions-for-a-package-owned-by-the-authenticated-user
  ALL_VERSIONS=$(gh api \
  -H "Accept: application/vnd.github+json" \
  /user/packages/container/helm-charts%2Fpodinfo/versions | jq -rc '.[].metadata.container.tags[]')
  echo
  echo Remote Repository aka Package [$GITHUB_OCI_REGISTRY_URL/podinfo] / All Versions 
  echo ================================================================================
  echo "$ALL_VERSIONS"
  echo ================================================================================
  # You can also see all package versions on GitHub web portal at 
  # [https://github.com/users/gfichtenholt/packages/container/helm-charts%2Fpodinfo/versions]
  # change dots to dashes for grep to work on whole words
  ALL_VERSIONS_DASHES=${ALL_VERSIONS//./-}
  VERSION_DASHES=${1//./-}
  EXPECTED_VERSION=$(echo $ALL_VERSIONS_DASHES | grep -w $VERSION_DASHES)
  if [[ "$EXPECTED_VERSION" == "" ]]; then
    echo Expected version [$1] missing from the remote [$GITHUB_OCI_REGISTRY_URL/podinfo]. Exiting...
    exit 1
  fi
}

function deleteChartVersionFromMyGitHubRegistry() {
  if [ $# -lt 1 ]
  then
    echo "Usage : $0 version"
    exit
  fi

  while true; do
    # ref https://docs.github.com/en/rest/packages#get-all-package-versions-for-a-package-owned-by-the-authenticated-user
    PACKAGE_VERSION_ID=$(gh api -H "Accept: application/vnd.github+json" /users/gfichtenholt/packages/container/helm-charts%2Fpodinfo/versions | jq --arg arg1 $1 -rc '.[] | select(.metadata.container.tags[] | contains($arg1)) | .id')
    if [[ "$PACKAGE_VERSION_ID" != "" ]]; then
      echo Deleting package version [$1] from remote [$GITHUB_OCI_REGISTRY_URL/podinfo]...
      # ref https://docs.github.com/en/rest/packages#delete-a-package-version-for-the-authenticated-user
      # ref https://github.com/cli/cli/issues/3937
      echo -n | gh api   --method DELETE   -H "Accept: application/vnd.github+json"   /user/packages/container/helm-charts%2Fpodinfo/versions/$PACKAGE_VERSION_ID --input -
      # one can verify that the version has been deleted on web portal
      # https://github.com/users/gfichtenholt/packages/container/helm-charts%2Fpodinfo/versions
    else 
      break
    fi
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

  pushChartToMyGitHubRegistry "$OCI_PODINFO_CHART_VERSION"

  # sanity checks
  
  # 1. check the version we pushed exists on the remote and is the only version avaialble
  ALL_VERSIONS=$(gh api \
  -H "Accept: application/vnd.github+json" \
  /user/packages/container/helm-charts%2Fpodinfo/versions | jq -rc '.[].metadata.container.tags[]')
  # GH web portal: You can see all package versions at 
  # [https://github.com/users/gfichtenholt/packages/container/helm-charts%2Fpodinfo/versions]
  NUM_VERSIONS=$(echo $ALL_VERSIONS | wc -l)
  if [ $NUM_VERSIONS != 1 ] 
  then
    echo "Expected exactly [1] version on the remote [$GITHUB_OCI_REGISTRY_URL/podinfo], got [$NUM_VERSIONS]. Exiting..."
    exit 1
  fi

  # 2. By default the new OCI registry is private. 
  # In order for the intergration test to work the OCI registry needs 'public' visibility
  # https://docs.github.com/en/packages/learn-github-packages/configuring-a-packages-access-control-and-visibility
  # API ref https://docs.github.com/en/rest/packages#get-a-package-for-the-authenticated-user
  # GitHub web portal: https://github.com/users/gfichtenholt/packages/container/helm-charts%2Fpodinfo/settings 
  while true; do
    VISIBILITY=$(gh api   -H "Accept: application/vnd.github+json"   /user/packages/container/helm-charts%2Fpodinfo | jq -rc '.visibility')
    if [[ "$VISIBILITY" != "public" ]]; then
      # TODO (gfichtenholt) can't seem to find docs for an API to change the package visibility on 
      # https://docs.github.com/en/rest/packages, so for now just ask to do this in web portal 
      # ref https://github.com/cli/cli/discussions/6003
      echo "Please change package [helm-charts/podinfo] visibility from [$VISIBILITY] to [public] on [https://github.com/users/gfichtenholt/packages/container/helm-charts%2Fpodinfo/settings]..." 
      open https://github.com/users/gfichtenholt/packages/container/helm-charts%2Fpodinfo/settings
      read -p "Press any key to continue..."
    else 
      break
    fi
  done
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
  while true; do 
    # GitHub API ref https://docs.github.com/en/rest/packages#list-packages-for-the-authenticated-users-namespace
    # GitHub web portal: https://github.com/gfichtenholt?tab=packages&ecosystem=container
    ALL_PACKAGES=$(gh api -H "Accept: application/vnd.github+json" /user/packages?package_type=container | jq '.[].name')
    echo Remote Repository [$GITHUB_OCI_REGISTRY_URL] / All Packages 
    echo ================================================================================
    echo "$ALL_PACKAGES"
    echo ================================================================================
    PODINFO_EXISTS=$(echo $ALL_PACKAGES | grep -sw 'helm-charts/podinfo')
    if [[ "$PODINFO_EXISTS" != "" ]]; then
      echo Deleting package [podinfo] from [$GITHUB_OCI_REGISTRY_URL]...
      # GitHub API ref https://docs.github.com/en/rest/packages#delete-a-package-for-the-authenticated-user
      # GitHub web portal: https://github.com/users/gfichtenholt/packages/container/helm-charts%2Fpodinfo/settings 
      echo -n | gh api --method DELETE -H "Accept: application/vnd.github+json" /user/packages/container/helm-charts%2Fpodinfo --input -
    else 
      break  
    fi
  done
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
