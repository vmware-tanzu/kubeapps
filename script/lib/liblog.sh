#!/usr/bin/env bash

# Copyright 2018-2022 the Kubeapps contributors.
# SPDX-License-Identifier: Apache-2.0

# Logging Library

# Color palette
RESET='\033[0m'
GREEN='\033[38;5;2m'
RED='\033[38;5;1m'
YELLOW='\033[38;5;3m'
MAGENTA='\033[38;5;5m'

# Functions

########################
# Log message to stderr
# Arguments:
#   $1 - Message to log
#########################
log() {
  printf "%b\n" "${*}" >&2
}

########################
# Log info message
# Arguments:
#   $1 - Message to log
#########################
info() {
  log "${GREEN}INFO ${RESET} ==> ${*}"
}

########################
# Log warning message
# Arguments:
#   $1 - Message to log
#########################
warn() {
  log "${YELLOW}WARN ${RESET} ==> ${*}"
}

########################
# Log error message
# Arguments:
#   $1 - Message to log
#########################
error() {
  log "${RED}ERROR ${RESET} ==> ${*}"
}

########################
# Log a 'debug' message
# Globals:
#   DEBUG_MODE
# Arguments:
#   None
# Returns:
#   None
#########################
debug() {
  local -r bool="${DEBUG_MODE:-false}"
  # comparison is performed without regard to the case of alphabetic characters
  shopt -s nocasematch
  if [[ "$bool" = 1 || "$bool" =~ ^(yes|true)$ ]]; then
    log "${MAGENTA}DEBUG${RESET} ==> ${*}"
  fi
}
