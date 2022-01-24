#!/usr/bin/env bash

# Copyright 2018-2022 the Kubeapps contributors.
# SPDX-License-Identifier: Apache-2.0

# Util functions Library

########################
# Retries a command a given number of times
# Arguments:
#   $1 - cmd (as a string)
#   $2 - max retries. Default: 12
#   $3 - sleep between retries (in seconds). Default: 5
# Returns:
#   Boolean
#########################
retry_while() {
    local -r cmd="${1:?cmd is missing}"
    local -r retries="${2:-12}"
    local -r sleep_time="${3:-5}"
    local return_value=1

    read -r -a command <<<"$cmd"
    for ((i = 1; i <= retries; i += 1)); do
        "${command[@]}" && return_value=0 && break
        sleep "$sleep_time"
    done
    return $return_value
}

#########################
# Redirects output to /dev/null unless debug mode is enabled
# Globals:
#   DEBUG_MODE
# Arguments:
#   $@ - Command to run
# Returns:
#   None
#########################
silence() {
    if ${DEBUG_MODE:-false}; then
        "$@"
    else
        "$@" >/dev/null 2>&1
    fi
}
