#!/bin/bash

# Copyright 2022 the Kubeapps contributors.
# SPDX-License-Identifier: Apache-2.0

# This script requires gcloud CLI (gcloud) to be installed locally.

FLUX_TEST_GCP_LOCATION=us-west1
FLUX_TEST_GCP_REGISTRY_DOMAIN=us-west1-docker.pkg.dev

function deleteGcpArtifactRepository()
{
  # sanity check
  if [[ "$#" -lt 1 ]]; then
    error_exit "Usage: deleteGcpArtifactRepository name"
  fi

  local REGISTRY_NAME=$1
  local REGISTRY_FULL_PATH=projects/vmware-kubeapps-ci/locations/$FLUX_TEST_GCP_LOCATION/repositories/$REGISTRY_NAME

  echo
  echo -e Checking if artifact repository [${L_YELLOW}$REGISTRY_NAME${NC}] exists...
  RESP=$(gcloud artifacts repositories list --location=$FLUX_TEST_GCP_LOCATION --format "json" | jq --arg REGISTRY_FULL_PATH "$REGISTRY_FULL_PATH" '.[] | select(.name==$REGISTRY_FULL_PATH)')
  if [[ "$RESP" != "" ]] ; then
    echo -e Deleting repository [${L_YELLOW}$REGISTRY_NAME${NC}]...
    gcloud artifacts repositories delete $REGISTRY_NAME --location=$FLUX_TEST_GCP_LOCATION -q 
  fi
}

function createGcpArtifactRepository()
{
  # sanity check
  if [[ "$#" -lt 1 ]]; then
    error_exit "Usage: createGcpArtifactRepository name"
  fi

  local REGISTRY_NAME=$1
  local REGISTRY_FULL_PATH=projects/vmware-kubeapps-ci/locations/$FLUX_TEST_GCP_LOCATION/repositories/$REGISTRY_NAME

  echo
  echo -e Checking if artifact repository [${L_YELLOW}$REGISTRY_NAME${NC}] exists...
  RESP=$(gcloud artifacts repositories list --location=$FLUX_TEST_GCP_LOCATION --format "json" | jq --arg REGISTRY_FULL_PATH "$REGISTRY_FULL_PATH" '.[] | select(.name==$REGISTRY_FULL_PATH)')
  if [[ "$RESP" != "" ]] ; then
       echo -e "Artifact repository [${L_YELLOW}${$REGISTRY_NAME}${NC}] already exists in harbor..."
  else 
    gcloud artifacts repositories create $REGISTRY_NAME --repository-format=docker --location=$FLUX_TEST_GCP_LOCATION --description="Helm repository for kubeapps flux plugin integration testing"
  fi

  # configure Docker with a credential helper to authenticate with Artifact Registry
  gcloud auth configure-docker $FLUX_TEST_GCP_REGISTRY_DOMAIN
}

function setupGcrStefanProdanClone {
  # this creates a clone of what was out on "oci://ghcr.io/stefanprodan/charts" as of Jul 28 2022
  # to oci://demo.goharbor.io/stefanprodan-podinfo-clone
  
  # TODO commands below require
  #  $ gcloud auth login
  # to have run first and 
  #  $ gcloud auth revoke 
  # when done. In other words, my personal GCP login. Need to switch to use a service account 
  
  local REPOSITORY_NAME=stefanprodan-podinfo-clone
  deleteGcpArtifactRepository $REPOSITORY_NAME
  createGcpArtifactRepository $REPOSITORY_NAME

  gcloud auth print-access-token | helm registry login -u oauth2accesstoken \
    --password-stdin $FLUX_TEST_GCP_REGISTRY_DOMAIN
  
  trap '{
    helm registry logout $FLUX_TEST_GCP_REGISTRY_DOMAIN
  }' EXIT  

  pushd $SCRIPTPATH/charts
  trap '{
    popd
  }' EXIT  

  ALL_VERSIONS=("6.1.0" "6.1.1" "6.1.2" "6.1.3" "6.1.4" "6.1.5" "6.1.6" "6.1.7" "6.1.8")
  DEST_URL=oci://$FLUX_TEST_GCP_REGISTRY_DOMAIN/vmware-kubeapps-ci/$REPOSITORY_NAME
  for v in ${ALL_VERSIONS[@]}; do
    helm push podinfo-$v.tgz $DEST_URL
    # this should result in something like
    # Pushed: us-west1-docker.pkg.dev/vmware-kubeapps-ci/stefanprodan-podinfo-clone/podinfo:6.1.0
  done
  
  echo
  echo Running sanity checks...
  echo TODO 
  echo
}
