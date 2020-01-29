#!/usr/bin/env bash
#
# Logging Library

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

# Color palette
declare -r RESET='\033[0m'
declare -r GREEN='\033[38;5;2m'
declare -r RED='\033[38;5;1m'
declare -r YELLOW='\033[38;5;3m'

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