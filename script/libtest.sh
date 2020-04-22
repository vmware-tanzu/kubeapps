#!/usr/bin/env bash
#
# Test functions Library

# Copyright (c) 2018-2020 Bitnami
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


# Load Generic Libraries
# shellcheck disable=SC1090
. "${ROOT_DIR}/script/liblog.sh"
# shellcheck disable=SC1090
. "${ROOT_DIR}/script/libutil.sh"

export TEST_MAX_RETRIES=600
export TEXT_TIME_STEP=5

########################
# Wait for a deployment to be rolled out
# Arguments:
#   $1 - Namespace
#   $2 - Deployment name
# Returns:
#   Integer - ExitCode
#########################
k8s_wait_for_deployment() {
    namespace=${1:?namespace is missing}
    deployment=${2:?deployment name is missing}
    local -i exit_code=0
    
    debug "Waiting for deployment ${deployment} to be successfully rolled out..."
    # Avoid to exit the function if the rollout fails
    silence kubectl rollout status --namespace "$namespace" deployment "$deployment" || exit_code=$?
    debug "Rollout exit code: '${exit_code}'"
    return $exit_code
}

########################
# Checks if a service has N endpoints 
# Arguments:
#   $1 - Namespace
#   $2 - Service name
#   $3 - Number of endpoints
# Returns:
#   Integer - ExitCode
#########################
k8s_svc_endpoints() {
    namespace=${1:?namespace is missing}
    svc=${2:?service is missing}
    number_of_endpoints=${3:?number of endpoints is missing}
    local -i exit_code=0

    silence kubectl get ep "$svc" -n "$namespace" -o jsonpath="{.subsets[0].addresses[$((number_of_endpoints - 1))]}" || exit_code=$?
    [[ "$exit_code" -eq 0 ]] && info "Endpoint ready!"
    
    return $exit_code
}

########################
# Wait for a service to have N endpoints 
# Arguments:
#   $1 - Namespace
#   $2 - Service name
#   $3 - Number of endpoints
# Returns:
#   Integer - ExitCode
#########################
k8s_wait_for_endpoints() {
    namespace=${1:?namespace is missing}
    svc=${2:?service is missing}
    number_of_endpoints=${3:?number of endpoints is missing}
    local -i exit_code=0

    debug "Waiting for the endpoints of ${svc} to be at least ${number_of_endpoints}..."
    retry_while "k8s_svc_endpoints $namespace $svc $number_of_endpoints" "$TEST_MAX_RETRIES" "$TEXT_TIME_STEP" || exit_code=$?
    
    return $exit_code
}

########################
# Checks if a deployment uses an image that matches a given pattern
# Arguments:
#   $1 - Namespace
#   $2 - Deployment
#   $3 - Expected pattern
#   $4 - (optional) jsonpath to image name
# Returns:
#   Integer - ExitCode
#########################
k8s_ensure_image() {
    namespace=${1:?namespace is missing}
    deployment=${2:?deployment name is missing}
    expectedPattern=${3:?expected pattern is missing}
    jsonpath=${4:-'{.spec.template.spec.containers[0].image}'}
    local -i exit_code=0
    
    debug "Checking that $deployment uses an image matching $expectedPattern..."
    kubectl get deployment "$deployment" -n "$namespace" -o jsonpath="$jsonpath" | grep "$expectedPattern" || exit_code=$?

    return $exit_code
}

########################
# Checks if a job is completed
# Arguments:
#   $1 - Namespace
#   $2 - Labels selector
# Returns:
#   Integer - ExitCode
#########################
k8s_job_completed() {
    namespace=${1:?namespace is missing}
    label_selector=${2:?labels selector is missing}
    local -i exit_code=0

    kubectl get jobs -l "$label_selector" -n "$namespace" -o jsonpath='{.items[*].status.conditions[?(@.type=="Complete")].status}' | grep "True" || exit_code=$?
    [[ "$exit_code" -eq 0 ]] && debug "Job completed!"
    
    return $exit_code
}

########################
# Wait for a job to be completed
# Arguments:
#   $1 - Namespace
#   $2 - Labels selector
# Returns:
#   Integer - ExitCode
#########################
k8s_wait_for_job_completed() {
    namespace=${1:?namespace is missing}
    label_selector=${2:?labels selector is missing}
    local -i exit_code=0

    debug "Wait for job completion..."
    retry_while "k8s_job_completed $namespace $label_selector" "$TEST_MAX_RETRIES" "$TEXT_TIME_STEP" || exit_code=$?
    
    return $exit_code
}
