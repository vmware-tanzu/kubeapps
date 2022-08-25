#!/bin/bash

# Copyright 2022 the Kubeapps contributors.
# SPDX-License-Identifier: Apache-2.0

# This script requires GitHub CLI (gh) to be installed locally. On MacOS you can
# install it via 'brew install gh'. gh releases page: https://github.com/cli/cli/releases 

# the goal is to create an OCI registry whose contents I completely control and will modify 
# by running integration tests. Therefore 'pushChartToMyGitHubRegistry'
# ref https://docs.github.com/en/packages/working-with-a-github-packages-registry/working-with-the-container-registry
GITHUB_OCI_REGISTRY_URL=oci://ghcr.io/gfichtenholt/helm-charts

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
    error_exit "Failed to login to helm registry [ghcr.io] after [$max] attempts. Exiting..."
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
  echo -e Remote Repository aka Package [${L_YELLOW}$GITHUB_OCI_REGISTRY_URL/podinfo${NC}] / All Versions 
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

function stefanProdanCloneRegistrySanityCheck() {
  echo
  echo Running sanity checks...
  echo
  gh auth status
  ALL_VERSIONS=$(gh api \
    -H "Accept: application/vnd.github+json" \
    /user/packages/container/stefanprodan-podinfo-clone%2Fpodinfo/versions | jq -rc '.[].metadata.container.tags[]')
  echo
  echo Remote Repository aka Package [$DEST_URL/podinfo] / All Versions 
  echo ================================================================================
  echo "$ALL_VERSIONS"
  echo ================================================================================  
  # GH web portal: You can see all package versions at 
  # [https://github.com/users/gfichtenholt/packages/container/stefanprodan-podinfo-clone%2Fpodinfo/versions]
  NUM_VERSIONS=$(echo "$ALL_VERSIONS" | wc -l)
  if [ $NUM_VERSIONS != 9 ] 
  then
    echo "Expected exactly [9] versions on the remote [$DEST_URL/podinfo], got [$NUM_VERSIONS]. Exiting..."
    exit 1
  fi
  while true; do
    VISIBILITY=$(gh api   -H "Accept: application/vnd.github+json"   /user/packages/container/stefanprodan-podinfo-clone%2Fpodinfo | jq -rc '.visibility')
    if [[ "$VISIBILITY" != "public" ]]; then
      # TODO (gfichtenholt) can't seem to find docs for an API to change the package visibility on 
      # https://docs.github.com/en/rest/packages, so for now just ask to do this in web portal 
      # ref https://github.com/cli/cli/discussions/6003
      echo -e "Please change package [${L_YELLOW}stefanprodan-podinfo-clone/podinfo${NC}] visibility from [$VISIBILITY] to [public] on [https://github.com/users/gfichtenholt/packages/container/stefanprodan-podinfo-clone%2Fpodinfo/settings]..." 
      open https://github.com/users/gfichtenholt/packages/container/stefanprodan-podinfo-clone%2Fpodinfo/settings
      read -p "Press any key to continue..."
    else 
      break
    fi
  done
}

function myGithubRegistrySanityCheck() {
  # 1. check the version we pushed exists on the remote and is the only version avaialble
  # ref https://docs.github.com/en/rest/packages#get-all-package-versions-for-a-package-owned-by-the-authenticated-user
  ALL_VERSIONS=$(gh api \
  -H "Accept: application/vnd.github+json" \
  /user/packages/container/helm-charts%2Fpodinfo/versions | jq -rc '.[].metadata.container.tags[]')
  # GH web portal: You can see all package versions at 
  # [https://github.com/users/gfichtenholt/packages/container/helm-charts%2Fpodinfo/versions]
  NUM_VERSIONS=$(echo "$ALL_VERSIONS" | wc -l)
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
      echo -e "Please change package [${L_YELLOW}helm-charts/podinfo${NC}] visibility from [$VISIBILITY] to [public] on [https://github.com/users/gfichtenholt/packages/container/helm-charts%2Fpodinfo/settings]..." 
      open https://github.com/users/gfichtenholt/packages/container/helm-charts%2Fpodinfo/settings
      read -p "Press any key to continue..."
    else 
      break
    fi
  done
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
      echo -n | gh api --method DELETE   -H "Accept: application/vnd.github+json"   /user/packages/container/helm-charts%2Fpodinfo/versions/$PACKAGE_VERSION_ID --input -
      # one can verify that the version has been deleted on web portal
      # https://github.com/users/gfichtenholt/packages/container/helm-charts%2Fpodinfo/versions
    else 
      break
    fi
  done
}

function deleteChartFromMyGithubRegistry() {
  while true; do 
    # GitHub API ref https://docs.github.com/en/rest/packages#list-packages-for-the-authenticated-users-namespace
    # GitHub web portal: https://github.com/gfichtenholt?tab=packages&ecosystem=container
    ALL_PACKAGES=$(gh api -H "Accept: application/vnd.github+json" /user/packages?package_type=container | jq '.[].name')
    echo -e Remote Repository [${L_YELLOW}$GITHUB_OCI_REGISTRY_URL${NC}] / All Packages 
    echo ================================================================================
    echo "$ALL_PACKAGES"
    echo ================================================================================
    PODINFO_EXISTS=$(echo $ALL_PACKAGES | grep -sw 'helm-charts/podinfo')
    if [[ "$PODINFO_EXISTS" != "" ]]; then
      echo -e Deleting package [${L_YELLOW}podinfo${NC}] from [${L_YELLOW}$GITHUB_OCI_REGISTRY_URL${NC}]...
      # GitHub API ref https://docs.github.com/en/rest/packages#delete-a-package-for-the-authenticated-user
      # GitHub web portal: https://github.com/users/gfichtenholt/packages/container/helm-charts%2Fpodinfo/settings 
      echo -n | gh api --method DELETE -H "Accept: application/vnd.github+json" /user/packages/container/helm-charts%2Fpodinfo --input -
    else 
      break  
    fi
  done
}

# $GITHUB_USER/$GITHUB_TOKEN is a secret for authentication with GitHub (ghcr.io)
#	I used my GitHub handle [gfichtenholt@vmware.com] and 
# personal access token [ghp_...] can be seen on https://github.com/settings/tokens
# and has scopes:
# "repo, workflow, write:packages, delete:packages, read:org, admin:repo_hook, delete_repo"
# Current token expires Nov 
# current token expires Mon, Nov 21 2022
function setupGithubStefanProdanClone {
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
    error_exit "Failed to login to helm registry [ghcr.io] after [$max] attempts. Exiting..."
  fi

  trap '{
    helm registry logout ghcr.io
  }' EXIT  

  pushd $SCRIPTPATH/charts
  trap '{
    popd
  }' EXIT  

  # this creates a clone of what was out on "oci://ghcr.io/stefanprodan/charts" as of Jul 28 2022
  # to oci://ghcr.io/gfichtenholt/stefanprodan-podinfo-clone
  ALL_VERSIONS=("6.1.0" "6.1.1" "6.1.2" "6.1.3" "6.1.4" "6.1.5" "6.1.6" "6.1.7" "6.1.8")
  DEST_URL=oci://ghcr.io/gfichtenholt/stefanprodan-podinfo-clone
  for v in ${ALL_VERSIONS[@]}; do
    helm push podinfo-$v.tgz $DEST_URL
  done
  
  stefanProdanCloneRegistrySanityCheck
}
