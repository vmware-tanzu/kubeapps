#!/bin/bash

# Copyright 2020 Bitnami.
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

# From https://github.com/kubernetes/sample-controller#when-using-go-111-modules
# To regenerate the api code:
#
# 1. update the .gitmodules in the root directory with the correct branch of code-generator
# and then run: git submodule update --remote
# 2. Run the this script from the apprepository-controller directory: ./hack/update-codegen.sh
# 3. Move the newly generated files over the old ones:
#    mv github.com/kubeapps/kubeapps/cmd/apprepository-controller/pkg/apis/apprepository/v1alpha1/zz_generated.deepcopy.go ./pkg/apis/apprepository/v1alpha1/zz_generated.deepcopy.go
#    rm -rf pkg/client && mv github.com/kubeapps/kubeapps/cmd/apprepository-controller/pkg/client ./pkg
#
# from slack: 
# - what are the situations when one needs to run update-codegen.sh manually after modifying 
# types.go in apprepository-controller?
#  Michael Nelson: This is following the example from the Kubernetes repository for a sample
# Kubernetes controller written in Go. From memory, whenever we update the client-go library 
# (to a new K8s version) on which the sample depends (client-go provides "go clients for talking 
# to a kubernetes cluster"), we've had to use the update-codegen.sh to get the versioned client 
# sets for talking to the cluster.
#
set -o errexit
set -o nounset
set -o pipefail

SCRIPT_ROOT=$(dirname ${BASH_SOURCE})/..
CODEGEN_PKG=${CODEGEN_PKG:-$(cd ${SCRIPT_ROOT}; ls -d -1 ./vendor/k8s.io/code-generator 2>/dev/null || echo ../../../k8s.io/code-generator)}

bash "${CODEGEN_PKG}"/generate-groups.sh "deepcopy,client,informer,lister" \
  github.com/kubeapps/kubeapps/cmd/apprepository-controller/pkg/client github.com/kubeapps/kubeapps/cmd/apprepository-controller/pkg/apis \
  apprepository:v1alpha1 \
  --output-base "${SCRIPT_ROOT}" \
  --go-header-file ${SCRIPT_ROOT}/hack/custom-boilerplate.go.txt
