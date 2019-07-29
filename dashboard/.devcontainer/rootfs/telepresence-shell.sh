#!/bin/bash

telepresence \
  --namespace "${KUBEAPPS_NAMESPACE:-kubeapps}" \
  --method inject-tcp \
  --swap-deployment "${KUBEAPPS_DASHBOARD_DEPLOYMENT:-kubeapps-internal-dashboard}" \
  --expose 3000:8080 --run-shell
