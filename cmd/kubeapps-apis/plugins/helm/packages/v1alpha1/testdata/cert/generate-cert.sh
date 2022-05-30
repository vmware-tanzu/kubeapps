#!/bin/bash

# Copyright 2021-2022 the Kubeapps contributors.
# SPDX-License-Identifier: Apache-2.0

set -o errexit
set -o nounset
set -o pipefail

# inspired by https://github.com/fluxcd/source-controller/tree/main/controllers/testdata/certs
# cfssl tool is kind of like openssl, but a little friendlier for simple stuff
# you can install it with brew on MacOS

# this will generate root cert authority cert (ca.pem) and private key (ca-key.pem)
cfssl gencert -initca ca-csr.json | cfssljson -bare ca -
# this will generate server cert (server.pem) and private key (server-key.pem)
cfssl gencert -ca=ca.pem -ca-key=ca-key.pem -config=ca-config.json -profile=web-servers server-csr.json | cfssljson -bare server
# this will generate the bundle that should be used for nginx ssl server config
cat server.pem ca.pem > ssl-bundle.pem
