#!/bin/sh

url=${1?Missing URL arg}

# This is a (super basic!) smoke test.

set -e -x

curl -fv $url/
curl -fv $url/api/healthz
curl -fv $url/kubeless
