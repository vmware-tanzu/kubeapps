#!/bin/bash

# Copyright 2022 the Kubeapps contributors.
# SPDX-License-Identifier: Apache-2.0

# This script requires gcloud CLI (gcloud) to be installed locally.

function deleteArtifactRegistry()
{
  # sanity check
  if [[ "$#" -lt 1 ]]; then
    error_exit "Usage: deleteArtifactRegistry name"
  fi

  local REGISTRY_NAME=$1
  local REGISTRY_FULL_PATH=projects/vmware-kubeapps-ci/locations/$FLUX_TEST_GCP_LOCATION/repositories/$REGISTRY_NAME

  echo
  echo -e Checking if artifact registry [${L_YELLOW}$REGISTRY_NAME${NC}] exists...
  RESP=$(gcloud artifacts repositories list --location=$FLUX_TEST_GCP_LOCATION --format "json" | jq --arg REGISTRY_FULL_PATH "$REGISTRY_FULL_PATH" '.[] | select(.name==$REGISTRY_FULL_PATH)')
  if [[ "$RESP" != "" ]] ; then
    echo -e Deleting repository [${L_YELLOW}$REGISTRY_NAME${NC}]...
    gcloud artifacts repositories delete $REGISTRY_NAME --location=$FLUX_TEST_GCP_LOCATION -q 
  fi
}

function createArtifactRegistry()
{
  # sanity check
  if [[ "$#" -lt 1 ]]; then
    error_exit "Usage: createArtifactRegistry name"
  fi

  local REGISTRY_NAME=$1
  local REGISTRY_FULL_PATH=projects/vmware-kubeapps-ci/locations/$FLUX_TEST_GCP_LOCATION/repositories/$REGISTRY_NAME

  echo
  echo -e Checking if artifact registry [${L_YELLOW}$REGISTRY_NAME${NC}] exists...
  RESP=$(gcloud artifacts repositories list --location=$FLUX_TEST_GCP_LOCATION --format "json" | jq --arg REGISTRY_FULL_PATH "$REGISTRY_FULL_PATH" '.[] | select(.name==$REGISTRY_FULL_PATH)')
  if [[ "$RESP" != "" ]] ; then
       echo -e "Artifact registry [${L_YELLOW}${$REGISTRY_NAME}${NC}] already exists in harbor..."
  else 
    gcloud artifacts repositories create $REGISTRY_NAME --repository-format=docker --location=$FLUX_TEST_GCP_LOCATION --description="Helm repository for kubeapps flux plugin integration testing"
  fi

  # configure Docker with a credential helper to authenticate with Artifact Registry
  gcloud auth configure-docker $FLUX_TEST_GCP_REGISTRY_DOMAIN
}
