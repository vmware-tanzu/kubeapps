#!/bin/bash

# Copyright 2018-2021 VMware. All Rights Reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -o errexit
set -o nounset
set -o pipefail

CLUSTER=${1:?}
ZONE=${2:?}
BRANCH=${3:?}
ADMIN=${4:?}

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
        gcloud container clusters delete "${CLUSTER}" --zone "${ZONE}"
    fi
fi

echo "Creating cluster ${CLUSTER} in ${ZONE} (v$BRANCH)"
gcloud container clusters create --cluster-version="${BRANCH}" --zone "${ZONE}" "${CLUSTER}" --num-nodes 2 --machine-type=n1-standard-2 --preemptible --labels=team=kubeapps
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
kubectl create clusterrolebinding kubeapps-cluster-admin --clusterrole=cluster-admin --user=$ADMIN
