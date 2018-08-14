#!/usr/bin/env bash

# Copyright (c) 2018 Bitnami
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

set -e

ROOT_DIR=`cd "$( dirname "${BASH_SOURCE[0]}" )/.." >/dev/null && pwd`

source $ROOT_DIR/script/libtest.sh

# Add admin permissions to default user in kube-system namespace
kubectl get clusterrolebinding kube-dns-admin >& /dev/null || \
    kubectl create clusterrolebinding kube-dns-admin --serviceaccount=kube-system:default --clusterrole=cluster-admin 

# Wait for Tiller
k8s_wait_for_pod_ready kube-system app=helm,name=tiller
wait_for_tiller

# Install Kubeapps
# TODO: Use just built images
helm dep up $ROOT_DIR/chart/kubeapps/
helm install --name kubeapps-ci --namespace kubeapps $ROOT_DIR/chart/kubeapps

# Wait for Kubeapps Pods
k8s_wait_for_pod_ready kubeapps app=kubeapps-ci
k8s_wait_for_pod_ready kubeapps app=kubeapps-ci-apprepository-controller
k8s_wait_for_pod_ready kubeapps app=kubeapps-ci-chartsvc
k8s_wait_for_pod_ready kubeapps app=kubeapps-ci-tiller-proxy
k8s_wait_for_pod_ready kubeapps app=mongodb
k8s_wait_for_pod_completed kubeapps apprepositories.kubeapps.com/repo-name=stable

# Run helm tests
helm test --cleanup kubeapps-ci
