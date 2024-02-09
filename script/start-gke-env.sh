#!/bin/bash

# Copyright 2018-2024 the Kubeapps contributors.
# SPDX-License-Identifier: Apache-2.0

set -o errexit
set -o nounset
set -o pipefail

CLUSTER=${1:?}
ZONE=${2:?}
GKE_VERSION=${3:?}
ADMIN=${4:?}
GKE_RELEASE_CHANNEL=${5:?}
ALLOW_GKE_VERSION_FALLBACK=${6:?}
DEBUG_MODE=${DEBUG_MODE:-"false"}

[[ "${DEBUG_MODE}" == "true" ]] && set -x

if ! gcloud container clusters list; then
    echo "Unable to access gcloud project"
    exit 1
fi

# Check if the cluster is already running
if [[ $(gcloud container clusters list --filter="name:${CLUSTER}") ]]; then
    if gcloud container clusters list --filter="name:${CLUSTER}" | grep "STOPPING"; then
        cnt=300
        while gcloud container clusters list | grep "${CLUSTER}"; do
            ((cnt = cnt - 1)) || (echo "Waited 5m but cluster is still being deleted" && exit 1)
            sleep 1
        done
    else
        echo "GKE cluster already exits. Deleting it"
        gcloud container clusters delete "${CLUSTER}" --zone "${ZONE}" --quiet
    fi
fi

# Check if the version is allowed in the current channel
gcloud container get-server-config --location "${ZONE}" >list.yaml
for i in {0..2}; do
    channel=$(yq .channels[$i].channel <list.yaml)
    if [[ "${channel,,}" == "${GKE_RELEASE_CHANNEL,,}" ]]; then
        defaultVersion=$(yq .channels[$i].defaultVersion <list.yaml)
        valid_versions=$(yq -o=j -I=0 .channels[$i].validVersions <list.yaml)
        if [[ $valid_versions =~ $GKE_VERSION ]]; then
            echo "[INFO] The version '${GKE_VERSION}' is one of the currently allowed versions in GKE channel '${channel}'"
            echo "[INFO] The default version for GKE channel '${channel}' is: '${defaultVersion}'"
            break
        else
            echo "[WARNING] The version '${GKE_VERSION}' is not allowed in GKE channel '${channel}'"
            echo "[INFO] The allowed versions for GKE channel '${channel}' (default: ${defaultVersion}) are: ${valid_versions}"
            if ! ${ALLOW_GKE_VERSION_FALLBACK}; then
                echo "[ERROR] Please, change the 'GKE_VERSION' variable in the script to one of the allowed versions, or set the 'ALLOW_GKE_VERSION_FALLBACK' variable to 'true'"
                exit 1
            else
                echo "[WARNING] Falling back to default version '${defaultVersion}, to change this behavior, set the 'ALLOW_GKE_VERSION_FALLBACK' variable to 'false'"
                GKE_VERSION=$defaultVersion
            fi
        fi
    fi
done

echo "Creating cluster ${CLUSTER} in ${ZONE} (v$GKE_VERSION)"

# TODO(bjesus) Removing the use of --labels=team=kubeapps due to a bug in gcloud cli: https://issuetracker.google.com/issues/255708239
#gcloud container clusters create --cluster-version="${GKE_VERSION}" --zone "${ZONE}" "${CLUSTER}" --num-nodes 2 --machine-type=n1-standard-2 --preemptible --labels=team=kubeapps --quiet
gcloud container clusters create --cluster-version="${GKE_VERSION}" --zone "${ZONE}" "${CLUSTER}" --num-nodes 2 --machine-type=n1-standard-2 --preemptible --quiet
echo "Waiting for the cluster to respond..."
cnt=20
until kubectl get pods >/dev/null 2>&1; do
    ((cnt = cnt - 1))
    if [[ "$cnt" -eq 0 ]]; then
        echo "Tried 20 times but the cluster is not reachable"
        exit 1
    fi
    sleep 1
done

# Set the current user as admin
kubectl create clusterrolebinding kubeapps-cluster-admin --clusterrole=cluster-admin --user="$ADMIN"
