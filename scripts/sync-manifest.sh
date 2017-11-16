#!/bin/bash -xe

MANIFEST_REPO=${1:-../manifest}

KUBECFG_JPATH=$MANIFEST_REPO/lib:$MANIFEST_REPO/vendor/kubecfg/lib:$MANIFEST_REPO/vendor/ksonnet-lib
export KUBECFG_JPATH

kubecfg show $MANIFEST_REPO/kubeapps.jsonnet > static/kubeapps-objs.yaml

cat >> static/kubeapps-objs.yaml <<EOF
---
apiVersion: v1
data:
  mongodb-password: MjNneWZ3ZWZoZzkyOA==
  mongodb-root-password: MjNneWZ3ZWZoZzkyOA==
kind: Secret
metadata:
  annotations: {}
  name: mongodb
  namespace: kubeapps
type: Opaque
EOF
