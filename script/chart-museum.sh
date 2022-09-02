#!/bin/bash

# Copyright 2022 the Kubeapps contributors.
# SPDX-License-Identifier: Apache-2.0

CHARTMUSEUM_PORT=${CHARTMUSEUM_PORT:-8090}
CHARTMUSEUM_USER=${CHARTMUSEUM_USER:-"admin"}
CHARTMUSEUM_PWD=${CHARTMUSEUM_PWD:-"password"}
CHARTMUSEUM_NS=${CHARTMUSEUM_NS:-"chart-museum"}
CHARTMUSEUM_VERSION=${CHARTMUSEUM_VERSION:-"3.9.0"}

# Pull a Bitnami chart to a local TGZ file
# Arguments:
#   $1: Chart name
#   $2: Chart version
pullBitnamiChart() {
  if [ -z "$1" ]; then
    echo "No Bitnami chart name supplied"
    exit 1
  fi
  if [ -z "$2" ]; then
    echo "No Bitnami chart version supplied"
    exit 1
  fi

  local CHART_NAME=$1
  local CHART_VERSION=$2

  echo "Pulling Bitnami chart ${CHART_NAME} v${CHART_VERSION} to local"

  CHART_FILE="${CHART_NAME}-${CHART_VERSION}.tgz"
  CHART_URL="https://charts.bitnami.com/bitnami/${CHART_FILE}"
  echo ">> Adding ${CHART_NAME}-${CHART_VERSION} to ChartMuseum from URL $CHART_URL"
  curl -LO "${CHART_URL}"
}

# Push a local chart to ChartsMuseum
# Arguments:
#   $1: Chart name
#   $2: Chart version
#   $3: Chart file
pushChartToChartMuseum() {
  if [ -z "$1" ]; then
    echo "No chart name supplied"
    exit 1
  fi
  if [ -z "$2" ]; then
    echo "No chart version supplied"
    exit 1
  fi
  if [ -z "$3" ]; then
    echo "No chart file supplied"
    exit 1
  fi

  local CHART_NAME=$1
  local CHART_VERSION=$2
  local CHART_FILE=$3

  echo "Pushing chart ${CHART_NAME} v${CHART_VERSION} to chart museum"

  local CHARTMUSEUM_POD_NAME=$(kubectl get pods --namespace ${CHARTMUSEUM_NS} -l "app.kubernetes.io/name=chartmuseum" -o jsonpath="{.items[0].metadata.name}")
  /bin/sh -c "kubectl port-forward $CHARTMUSEUM_POD_NAME ${CHARTMUSEUM_PORT}:8080 --namespace ${CHARTMUSEUM_NS} &"
  sleep 2

  CHART_EXISTS=$(curl -u "${CHARTMUSEUM_USER}:${CHARTMUSEUM_PWD}" -X GET http://localhost:${CHARTMUSEUM_PORT}/api/charts/${CHART_NAME}/${CHART_VERSION} | jq -r 'any([ .error] ; . > 0)')
  if [ "$CHART_EXISTS" == "true" ]; then
    echo ">> CHART EXISTS: deleting"
    curl -u "${CHARTMUSEUM_USER}:${CHARTMUSEUM_PWD}" -X DELETE http://localhost:${CHARTMUSEUM_PORT}/api/charts/${CHART_NAME}/${CHART_VERSION}
  fi
  
  echo ">> Uploading chart from file ${CHART_FILE}"
  curl -u "${CHARTMUSEUM_USER}:${CHARTMUSEUM_PWD}" --data-binary "@${CHART_FILE}" http://localhost:${CHARTMUSEUM_PORT}/api/charts  
  
  # End port forward
  pkill -f "kubectl port-forward $CHARTMUSEUM_POD_NAME ${CHARTMUSEUM_PORT}:8080 --namespace ${CHARTMUSEUM_NS}"

  rm ${CHART_FILE}
}

# Install ChartsMuseum
installChartMuseum() {
  echo "Installing ChartMuseum ${CHARTMUSEUM_VERSION}..."
  helm install chartmuseum --namespace ${CHARTMUSEUM_NS} --create-namespace "https://github.com/chartmuseum/charts/releases/download/chartmuseum-${CHARTMUSEUM_VERSION}/chartmuseum-${CHARTMUSEUM_VERSION}.tgz" \
    --set env.open.DISABLE_API=false \
    --set persistence.enabled=true \
    --set env.secret.BASIC_AUTH_USER=$CHARTMUSEUM_USER \
    --set env.secret.BASIC_AUTH_PASS=$CHARTMUSEUM_PWD
  info "Waiting for ChartMuseum to be ready..."
  kubectl rollout status -w deployment/chartmuseum --namespace=${CHARTMUSEUM_NS}

  echo "Chart museum v${CHARTMUSEUM_VERSION} installed in namespace ${CHARTMUSEUM_NS}"
  echo "Credentials: ${CHARTMUSEUM_USER} / ${CHARTMUSEUM_PWD}"
  echo "Cluster internal URL: "
  echo "    http://chartmuseum.${CHARTMUSEUM_NS}.svc.cluster.local:8080/"
}

# Uninstall ChartsMuseum
uninstallChartMuseum() {
  echo "Uninstalling ChartMuseum..."
  helm uninstall chartmuseum --namespace ${CHARTMUSEUM_NS}
}

case $1 in

  install)
    installChartMuseum
    ;;

  uninstall)
    uninstallChartMuseum
    ;;

  pullBitnamiChart)
    pullBitnamiChart $2 $3
    ;;

  pushChart)
    pushChartToChartMuseum $2 $3 $4
    ;;

  *)
    ;;
esac
