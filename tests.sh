#!/bin/sh

url=${1?Missing URL arg}

# This is a (super basic!) smoke test.

set -e -x

curl -fv $url/
# FIXME: We don't wait for the API to rollout as it times out
# curl -fv $url/api/healthz
curl -fv $url/kubeless
