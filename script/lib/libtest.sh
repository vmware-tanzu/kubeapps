#!/usr/bin/env bash

# Copyright 2018-2022 the Kubeapps contributors.
# SPDX-License-Identifier: Apache-2.0

# Test functions Library

# Load Generic Libraries
# shellcheck disable=SC1090
. "${ROOT_DIR}/script/lib/liblog.sh"
# shellcheck disable=SC1090
. "${ROOT_DIR}/script/lib/libutil.sh"

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
    local retries=5
    local retriesWait=10

    info "Checking rollout status in deployment ${deployment} in ns ${namespace}"
    until [[ $retries == 0 ]]; do
        silence kubectl rollout status --namespace "${namespace}" deployment "${deployment}" -w --timeout=60s || exit_code=$?
        if [[ $exit_code -eq 0 ]]; then
            break
        fi
        info "Attempt failed, retrying after ${retriesWait}... (remaining attempts: ${retries})"
        sleep $retriesWait
        retries=$((retries - 1))
    done
    if [ $retries == 0 ]; then
        info "Error while rolling out deployment ${deployment} in ns ${namespace}"
        exit 1
    fi
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
