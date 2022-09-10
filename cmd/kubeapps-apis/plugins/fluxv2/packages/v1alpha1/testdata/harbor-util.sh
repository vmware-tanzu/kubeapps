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
  if [[ "$#" -lt 5 ]]; then
    error_exit "Usage: createHarborProject host user password project_name is_public"
  fi

  local HOST=$1
  local USER=$2
  local PWD=$3
  local PROJECT_NAME=$4
  local PUBLIC=$5
  local URL=https://${HOST}
  local status_code=$(curl -L --write-out %{http_code} \
                      --silent --output /dev/null \
                      --head ${URL}/api/v2.0/projects?project_name=${PROJECT_NAME} \
                      -u $USER:$PWD)
  if [[ "$status_code" -eq 200 ]] ; then
    echo -e "Project [${L_YELLOW}${PROJECT_NAME}${NC}] already exists on [${L_YELLOW}${HOST}${NC}] ..."
  elif [[ "$status_code" -eq 404 ]] ; then
    if [[ $PUBLIC == "true" ]]; then 
      echo -e "Creating public project [${L_YELLOW}$PROJECT_NAME${NC}] on [${L_YELLOW}${HOST}${NC}] ..."
    elif [[ $PUBLIC == "false" ]]; then 
      echo -e "Creating private project [${L_YELLOW}$PROJECT_NAME${NC}] on [${L_YELLOW}${HOST}${NC}] ..."
    else 
      error_exit "Unsupported value for public: $PUBLIC"
    fi
    local payload=$(sed "s/\$NAME/${PROJECT_NAME}/g" $SCRIPTPATH/harbor-create-project.json)
    payload=$(echo $payload | sed "s/\$PUBLIC/${PUBLIC}/g")
    local status_code=$(curl -L --write-out %{http_code} --silent \
                        --output /dev/null \
                        -X POST ${URL}/api/v2.0/projects \
                        -H 'Content-Type: application/json' \
                        --data "${payload}" \
                        -u $USER:$PWD)
    if [[ "$status_code" -eq 201 ]] ; then
      echo -e "Project [${L_YELLOW}${PROJECT_NAME}${NC}] successfully created..."
      #
      # todo set up some project quotas (tag retention policies)
    else
      error_exit "Unexpected HTTP status creating project [$PROJECT_NAME]: [$status_code]"
    fi
  else
    error_exit "Unexpected HTTP status checking whether project [$PROJECT_NAME] exists: [$status_code]"
  fi
}

function deleteHarborProjectRepositories()
{
  if [[ "$#" -lt 4 ]]; then
    error_exit "Usage: deleteHarborProjectRepositories host user password project_name"
  fi
  local HOST=$1
  local USER=$2
  local PWD=$3
  local PROJECT_NAME=$4
  local URL=https://${HOST}

  RESP=$(curl -L --silent --show-error \
        ${URL}/api/v2.0/projects/$PROJECT_NAME/repositories \
        -u $USER:$PWD)
  RESP=$(echo "$RESP" | jq -r .[].name | tr -d '"')
  IFS=$'\n' RESP=($RESP)
  if [[ ! -z "$RESP" ]] ; then
    for (( i=0; i<${#RESP[@]}; i++ ))
    do
      IFS='/' read -ra SEGMENTS <<< "${RESP[$i]}"
      local n=
      for (( j=1; j<${#SEGMENTS[@]}; j++ ))
      do
        if [[ $j > 1 ]] ; then 
          n=$n/
        fi
        n=$n${SEGMENTS[j]}
      done
      n=$(echo $n | sed 's|/|%252F|g')
      echo -e Deleting repository [${L_YELLOW}$n${NC}] on [${L_YELLOW}$HOST${NC}]...
      status_code=$(curl -L --write-out %{http_code} --silent \
                  --show-error -X DELETE --output /dev/null \
                  ${URL}/api/v2.0/projects/$PROJECT_NAME/repositories/$n \
                  -u $USER:$PWD)
      if [[ "$status_code" -eq 200 ]] ; then
          echo -e Repository [${L_YELLOW}$n${NC}] deleted
      else
          error_exit "Failed to delete repository [$n] due to HTTP status: [$status_code]"
      fi
    done
  fi
}

function deleteHarborProject()
{
  # sanity check
  if [[ "$#" -lt 4 ]]; then
    error_exit "Usage: deleteHarborProject host user password project_name"
  fi
  local HOST=$1
  local USER=$2
  local PWD=$3
  local PROJECT_NAME=$4
  local URL=https://${HOST}
  echo
  echo -e Checking whether project [${L_YELLOW}$PROJECT_NAME${NC}] exists on [${L_YELLOW}$HOST${NC}] ...
  local status_code=$(curl -L --write-out %{http_code} \
                      --silent --output /dev/null \
                      --show-error \
                      --head ${URL}/api/v2.0/projects?project_name=${PROJECT_NAME} \
                      -u $USER:$PWD)
  if [[ "$status_code" -eq 200 ]] ; then
    echo -e "Project [${L_YELLOW}$PROJECT_NAME${NC}] exists on [${L_YELLOW}$HOST${NC}]. This script will now delete it..."
    deleteHarborProjectRepositories $*
    status_code=$(curl -L --write-out %{http_code} --silent \
          --show-error -X DELETE \
          --output /dev/null \
           ${URL}/api/v2.0/projects/${PROJECT_NAME} \
           -u $USER:$PWD)
    if [[ "$status_code" -eq 200 ]] ; then
        echo -e Project [${L_YELLOW}${PROJECT_NAME}${NC}] deleted
    else
        error_exit "Failed to delete project [$PROJECT_NAME] due to HTTP status: [$status_code]"
    fi
  elif [[ "$status_code" -ne 404 ]] ; then
    error_exit "Unexpected HTTP status checking whether project [$PROJECT_NAME] exists: [$status_code]"
  fi
}

function pushChartsToHarborProject() 
{
  if [[ "$#" -lt 4 ]]; then
    error_exit "Usage: pushChartsToHarbor host user password project_name"
  fi
  local HOST=$1
  local USER=$2
  local PWD=$3
  local PROJECT_NAME=$4
  local URL=https://${HOST}

  helm registry login $HOST -u $USER -p $PWD
  trap '{
    helm registry logout $HOST 
  }' EXIT  

  pushd $SCRIPTPATH/charts
  trap '{
    popd
  }' EXIT  

  ALL_VERSIONS=("6.1.0" "6.1.1" "6.1.2" "6.1.3" "6.1.4" "6.1.5" "6.1.6" "6.1.7" "6.1.8")
  DEST_URL=oci://$HOST/$PROJECT_NAME
  for v in ${ALL_VERSIONS[@]}; do
    helm push podinfo-$v.tgz $DEST_URL
  done
}

# shortcut to only look at the project existence and if so assume all is well
function quickCheckProjectExists()
{
  if [[ "$#" -lt 5 ]]; then
    error_exit "Usage: quickCheckProjectExist host user password project_name result_var"
  fi
  local HOST=$1
  local USER=$2
  local PWD=$3
  local PROJECT_NAME=$4
  local  __resultvar=$5
  local URL=https://${HOST}
  echo
  echo -e Checking whether harbor project [${L_YELLOW}$PROJECT_NAME${NC}] exists on [${L_YELLOW}$HOST${NC}]...
  local status_code=$(curl -L --write-out %{http_code} \
                      --silent --output /dev/null \
                      --show-error \
                      --head ${URL}/api/v2.0/projects?project_name=${PROJECT_NAME} \
                      -u $USER:$PWD)
  if [[ "$status_code" -eq 200 ]] ; then
    echo -e "Project [${L_YELLOW}$PROJECT_NAME${NC}] exists on [${L_YELLOW}$HOST${NC}]"
    # here we assume that since project exists, it contains all the charts
    eval $__resultvar="true"
  else 
    eval $__resultvar="false"
  fi
}

# special case: in VMware corp harbor a project is created via filing a request
function setupVMwareHarborStefanProdanClone {
  # sanity check
  if [[ "$#" -lt 1 ]]; then
    error_exit "Usage: setupVMwareHarborStefanProdanClone project_name [--quick]"
  fi

  local PROJECT_NAME=$1
  if [ -z ${HARBOR_VMWARE_CORP_HOST+x} ]; then
    error_exit "Environment variable [HARBOR_VMWARE_CORP_HOST] must be set"
  fi
  if [ -z ${HARBOR_VMWARE_CORP_ROBOT_USER+x} ]; then
    error_exit "Environment variable [HARBOR_VMWARE_CORP_ROBOT_USER] must be set"
  fi
  if [ -z ${HARBOR_VMWARE_CORP_ROBOT_SECRET+x} ]; then
    error_exit "Environment variable [HARBOR_VMWARE_CORP_ROBOT_SECRET] must be set"
  fi
  local HOST=$HARBOR_VMWARE_CORP_HOST
  local USER=$HARBOR_VMWARE_CORP_ROBOT_USER
  local PWD=$HARBOR_VMWARE_CORP_ROBOT_SECRET
  local URL=https://${HOST}

  # unfortunately cannot re-use setupHarborStefanProdanCloneInProject because the 
  # remote is configured so that createProject fails with a 403
  if [ "$#" -gt 1 ]; then
    if [ "$2" == "--quick" ]; then
      # shortcut to only look at the project existence and if so assume all is well
      quickCheckProjectExists $HOST $USER $PWD $PROJECT_NAME EXISTS
      if [[ "$EXISTS" == "true" ]]; then
        return
      fi
    fi
  fi

  deleteHarborProjectRepositories $HOST $USER $PWD $PROJECT_NAME
  pushChartsToHarborProject $HOST $USER $PWD $PROJECT_NAME
}

function setupHarborStefanProdanCloneInProject {
  # sanity check
  if [[ "$#" -lt 5 ]]; then
    echo "args=$*"
    error_exit "Usage: setupHarborStefanProdanCloneInProject host user password project_name is_public [--quick]"
  fi

  local HOST=$1
  local USER=$2
  local PWD=$3
  local PROJECT_NAME=$4
  local PUBLIC=$5
  local URL=https://${HOST}

  if [ "$#" -gt 5 ]; then
    if [ "$6" == "--quick" ]; then
      quickCheckProjectExists $HOST $USER $PWD $PROJECT_NAME EXISTS
      if [[ "$EXISTS" == "true" ]]; then
        return
      fi
    fi
  fi

  deleteHarborProject $HOST $USER $PWD $PROJECT_NAME
  createHarborProject $HOST $USER $PWD $PROJECT_NAME $PUBLIC
  pushChartsToHarborProject $HOST $USER $PWD $PROJECT_NAME
  
  echo
  echo Running sanity checks...
  echo TODO 
  echo
}

function setupHarborStefanProdanClone {
  # this creates a clone of what was out on "oci://ghcr.io/stefanprodan/charts" as of Jul 28 2022
  # to oci://demo.goharbor.io/stefanprodan-podinfo-clone
  setupHarborStefanProdanCloneInProject \
     $FLUX_TEST_HARBOR_HOST \
     $FLUX_TEST_HARBOR_ADMIN_USER \
     $FLUX_TEST_HARBOR_ADMIN_PWD \
    stefanprodan-podinfo-clone true $*
  
  setupHarborStefanProdanCloneInProject \
     $FLUX_TEST_HARBOR_HOST \
     $FLUX_TEST_HARBOR_ADMIN_USER \
     $FLUX_TEST_HARBOR_ADMIN_PWD \
     stefanprodan-podinfo-clone-private false $*

  setupVMwareHarborStefanProdanClone kubeapps_flux_integration $*
}

function deleteHarborRobotAccount()
{
  # sanity check
  if [[ "$#" -lt 1 ]]; then
    error_exit "Usage: deleteHarborRobotAccount name"
  fi
  local ACCOUNT_NAME=$1
  echo
  echo -e Checking whether harbor robot account [${L_YELLOW}$ACCOUNT_NAME${NC}] exists...
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
  if [[ "$#" -lt 3 ]]; then
    error_exit "Usage: createHarborRobotAccount name project_name_1 project_name_2"
  fi
  local ACCOUNT_NAME=$1
  local PROJECT_NAME_1=$2
  local PROJECT_NAME_2=$3

  echo -e "Creating robot account [${L_YELLOW}$ACCOUNT_NAME${NC}] in harbor..."
  local payload=$(sed "s/\$NAME/${ACCOUNT_NAME}/g" $SCRIPTPATH/harbor-create-robot-account.json)
  payload=$(echo $payload | sed "s/\$PROJECT_NAME_1/${PROJECT_NAME_1}/g")
  payload=$(echo $payload | sed "s/\$PROJECT_NAME_2/${PROJECT_NAME_2}/g")
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
  local PROJECT_NAME_1=stefanprodan-podinfo-clone
  local PROJECT_NAME_2=stefanprodan-podinfo-clone-private

  deleteHarborRobotAccount $ACCOUNT_NAME
  createHarborRobotAccount $ACCOUNT_NAME $PROJECT_NAME_1 $PROJECT_NAME_2
}
