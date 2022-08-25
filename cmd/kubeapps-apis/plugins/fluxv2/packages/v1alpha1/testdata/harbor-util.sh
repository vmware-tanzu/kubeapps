#!/bin/bash

# Copyright 2022 the Kubeapps contributors.
# SPDX-License-Identifier: Apache-2.0

FLUX_TEST_HARBOR_HOST=demo.goharbor.io
FLUX_TEST_HARBOR_URL=https://${FLUX_TEST_HARBOR_HOST}
# admin/Harbor12345 is a well known default login for harbor registries
FLUX_TEST_HARBOR_ADMIN_USER=admin
FLUX_TEST_HARBOR_ADMIN_PWD=Harbor12345

function createHarborProject()
{
  # sanity check
  if [[ "$#" -lt 1 ]]; then
    error_exit "Usage: createHarborProject name"
  fi

  local PROJECT_NAME=$1
  local status_code=$(curl -L --write-out %{http_code} \
                      --silent --output /dev/null \
                      --head ${FLUX_TEST_HARBOR_URL}/api/v2.0/projects?project_name=${PROJECT_NAME} \
                      -u $FLUX_TEST_HARBOR_ADMIN_USER:$FLUX_TEST_HARBOR_ADMIN_PWD)
  if [[ "$status_code" -eq 200 ]] ; then
    echo -e "Project [${L_YELLOW}${PROJECT_NAME}${NC}] already exists in harbor..."
  elif [[ "$status_code" -eq 404 ]] ; then
    echo -e "Creating public project [${L_YELLOW}$PROJECT_NAME${NC}] in harbor..."
    local payload=$(sed "s/\$NAME/${PROJECT_NAME}/g" $SCRIPTPATH/harbor-create-project.json)
    local status_code=$(curl -L --write-out %{http_code} --silent \
                        --output /dev/null \
                        -X POST ${FLUX_TEST_HARBOR_URL}/api/v2.0/projects \
                        -H 'Content-Type: application/json' \
                        --data "${payload}" \
                        -u $FLUX_TEST_HARBOR_ADMIN_USER:$FLUX_TEST_HARBOR_ADMIN_PWD)
    if [[ "$status_code" -eq 201 ]] ; then
      echo -e "Project [${L_YELLOW}${PROJECT_NAME}${NC}] successfully created..."
      #
      # todo set up some project quotas (tag retention policies)
    else
      error_exit "Unexpected HTTP status creating project [$PROJECT_NAME]: [$status_code]"
    fi
  else
    error_exit "Unexpected HTTP status checking if project [$PROJECT_NAME] exists: [$status_code]"
  fi
}

function deleteHarborProject()
{
  # sanity check
  if [[ "$#" -lt 1 ]]; then
    error_exit "Usage: deleteHarborProject name"
  fi
  local PROJECT_NAME=$1
  echo
  echo -e Checking if harbor project [${L_YELLOW}$PROJECT_NAME${NC}] exists...
  local status_code=$(curl -L --write-out %{http_code} \
                      --silent --output /dev/null \
                      --show-error \
                      --head ${FLUX_TEST_HARBOR_URL}/api/v2.0/projects?project_name=${PROJECT_NAME} \
                      -u $FLUX_TEST_HARBOR_ADMIN_USER:$FLUX_TEST_HARBOR_ADMIN_PWD)
  if [[ "$status_code" -eq 200 ]] ; then
    echo -e "Project [${L_YELLOW}$PROJECT_NAME${NC}] exists in harbor. This script will now delete it..."
    CMD="curl -L --silent --show-error \
           ${FLUX_TEST_HARBOR_URL}/api/v2.0/projects/$PROJECT_NAME/repositories \
           -u $FLUX_TEST_HARBOR_ADMIN_USER:$FLUX_TEST_HARBOR_ADMIN_PWD"
    RESP=$($CMD)
    RESP=$(echo "$RESP" | jq .[].name | tr -d '"')
    if [[ ! -z "$RESP" ]] ; then
      IFS='/' read -ra SEGMENTS <<< "$RESP"
      for n in "${SEGMENTS[1]}"
      do
        echo -e Deleting repository [${L_YELLOW}$n${NC}]...
        status_code=$(curl -L --write-out %{http_code} --silent \
              --show-error -X DELETE --output /dev/null \
              ${FLUX_TEST_HARBOR_URL}/api/v2.0/projects/$PROJECT_NAME/repositories/$n \
              -u $FLUX_TEST_HARBOR_ADMIN_USER:$FLUX_TEST_HARBOR_ADMIN_PWD)
        if [[ "$status_code" -eq 200 ]] ; then
            echo -e Repository [${L_YELLOW}$n${NC}] deleted
        else
            error_exit "Failed to delete repository [$n] due to HTTP status: [$status_code]"
        fi
      done
    fi
    status_code=$(curl -L --write-out %{http_code} --silent \
          --show-error -X DELETE \
          --output /dev/null \
           ${FLUX_TEST_HARBOR_URL}/api/v2.0/projects/${PROJECT_NAME} \
           -u $FLUX_TEST_HARBOR_ADMIN_USER:$FLUX_TEST_HARBOR_ADMIN_PWD)
    if [[ "$status_code" -eq 200 ]] ; then
        echo -e Project [${L_YELLOW}${PROJECT_NAME}${NC}] deleted
    else
        error_exit "Failed to delete project [$PROJECT_NAME] due to HTTP status: [$status_code]"
    fi
  elif [[ "$status_code" -ne 404 ]] ; then
    error_exit "Unexpected HTTP status checking if project [$PROJECT_NAME] exists: [$status_code]"
  fi
}

function setupHarborStefanProdanClone {
  # this creates a clone of what was out on "oci://ghcr.io/stefanprodan/charts" as of Jul 28 2022
  # to oci://demo.goharbor.io/stefanprodan-podinfo-clone
  local PROJECT_NAME=stefanprodan-podinfo-clone
  deleteHarborProject $PROJECT_NAME
  createHarborProject $PROJECT_NAME
  
  helm registry login $FLUX_TEST_HARBOR_HOST -u $FLUX_TEST_HARBOR_ADMIN_USER -p $FLUX_TEST_HARBOR_ADMIN_PWD
  trap '{
    helm registry logout $FLUX_TEST_HARBOR_HOST 
  }' EXIT  

  pushd $SCRIPTPATH/charts
  trap '{
    popd
  }' EXIT  

  SRC_URL_PREFIX=https://stefanprodan.github.io/podinfo
  ALL_VERSIONS=("6.1.0" "6.1.1" "6.1.2" "6.1.3" "6.1.4" "6.1.5" "6.1.6" "6.1.7" "6.1.8")
  DEST_URL=oci://demo.goharbor.io/$PROJECT_NAME
  for v in ${ALL_VERSIONS[@]}; do
    helm push podinfo-$v.tgz $DEST_URL
  done
  
  echo
  echo Running sanity checks...
  echo TODO 
  echo
}
