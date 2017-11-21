#!/bin/bash -xe
# Copyright (c) 2017 Bitnami
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

MANIFEST_REPO=${1:-../manifest}

KUBECFG_JPATH=$MANIFEST_REPO/lib:$MANIFEST_REPO/vendor/kubecfg/lib:$MANIFEST_REPO/vendor/ksonnet-lib
export KUBECFG_JPATH

kubecfg show $MANIFEST_REPO/kubeapps.jsonnet > static/kubeapps-objs.yaml

mkpasswd() {
    echo $(LC_ALL=C tr -dc 'A-Za-z0-9!"#$%&'\''()*+,-./:;<=>?@[\]^_`{|}~' </dev/urandom | head -c 10 | base64)
}

MONGODB_PW=$(mkpasswd)

cat >> static/kubeapps-objs.yaml <<EOF
---
apiVersion: v1
data:
  mongodb-password: $MONGODB_PW
  mongodb-root-password: $MONGODB_PW
kind: Secret
metadata:
  annotations: {}
  name: mongodb
  namespace: kubeapps
type: Opaque
EOF