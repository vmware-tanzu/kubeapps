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

  if [ "$#" -gt 0 ]; then
    if [ "$1" == "--quick" ]; then
      # short to only look at the project existence and if so assume all is well
      echo
      echo -e Checking if harbor project [${L_YELLOW}$PROJECT_NAME${NC}] exists...
      local status_code=$(curl -L --write-out %{http_code} \
                          --silent --output /dev/null \
                          --show-error \
                          --head ${FLUX_TEST_HARBOR_URL}/api/v2.0/projects?project_name=${PROJECT_NAME} \
                          -u $FLUX_TEST_HARBOR_ADMIN_USER:$FLUX_TEST_HARBOR_ADMIN_PWD)
      if [[ "$status_code" -eq 200 ]] ; then
        echo -e "Project [${L_YELLOW}$PROJECT_NAME${NC}] exists in harbor."
        # here we assume that since project exists, it contains all the charts
        exit 0
      fi
    fi
  fi

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

function deleteHarborRobotAccount()
{
  # sanity check
  if [[ "$#" -lt 1 ]]; then
    error_exit "Usage: deleteHarborRobotAccount name"
  fi
  local ACCOUNT_NAME=$1
  echo
  echo -e Checking if harbor robot account [${L_YELLOW}$ACCOUNT_NAME${NC}] exists...
  local CMD="curl -L --silent --show-error \
          ${FLUX_TEST_HARBOR_URL}/api/v2.0/robots \
          -u $FLUX_TEST_HARBOR_ADMIN_USER:$FLUX_TEST_HARBOR_ADMIN_PWD"
  local RESP=$($CMD)
  local ID=$(echo "$RESP" | jq --arg NAME "robot\$$ACCOUNT_NAME" '.[] | select(.name == $NAME) | .id')
  if [[ "$ID" != "" ]] ; then
    echo -e "Deleting robot account [${L_YELLOW}$ACCOUNT_NAME${NC}] in harbor..."
    status_code=$(curl -L --write-out %{http_code} --silent \
          --show-error -X DELETE --output /dev/null \
          ${FLUX_TEST_HARBOR_URL}/api/v2.0/robots/$ID \
          -u $FLUX_TEST_HARBOR_ADMIN_USER:$FLUX_TEST_HARBOR_ADMIN_PWD)
    if [[ "$status_code" -eq 200 ]] ; then
        echo -e Robot account [${L_YELLOW}$ACCOUNT_NAME${NC}] deleted
    else
        error_exit "Failed to delete robot account [$ACCOUNT_NAME] due to HTTP status: [$status_code]"
    fi
  fi
}

function createHarborRobotAccount()
{
  # sanity check
  if [[ "$#" -lt 2 ]]; then
    error_exit "Usage: createHarborRobotAccount name project_name"
  fi
  local ACCOUNT_NAME=$1
  local PROJECT_NAME=$2

  echo -e "Creating robot account [${L_YELLOW}$ACCOUNT_NAME${NC}] in harbor..."
  local payload=$(sed "s/\$NAME/${ACCOUNT_NAME}/g" $SCRIPTPATH/harbor-create-robot-account.json)
  payload=$(echo $payload | sed "s/\$PROJECT_NAME/${PROJECT_NAME}/g")
  local RESP=$(curl -L --silent --show-error \
                -X POST \
                -H 'Content-Type: application/json' \
                --data "${payload}" \
                 ${FLUX_TEST_HARBOR_URL}/api/v2.0/robots \
                -u $FLUX_TEST_HARBOR_ADMIN_USER:$FLUX_TEST_HARBOR_ADMIN_PWD)
  local RESP2=$(echo "$RESP" | jq -r '. | {name,secret} | join(" ")')
  if [[ "$RESP2" == robot* ]] ; then
    echo -e "Robot account successfully created: [$RESP2]"
  else
    error_exit "Unexpected HTTP response creating robot account [$ACCOUNT_NAME]: $RESP"
  fi
}

function setupHarborRobotAccount()
{
  local ACCOUNT_NAME=kubeapps-flux-plugin
  local PROJECT_NAME=stefanprodan-podinfo-clone

  deleteHarborRobotAccount $ACCOUNT_NAME
  createHarborRobotAccount $ACCOUNT_NAME $PROJECT_NAME
}


