#!/bin/bash

CM_PORT=${CM_PORT:-8090}
CM_USER=${CM_USER:-"admin"}
CM_PWD=${CM_PWD:-"password"}
CHART_MUSEUM_NS=${CHART_MUSEUM_NS:-"chart-museum"}
CM_VERSION=${CM_VERSION:-"3.9.0"}

# Pull a Bitnami chart to a local TGZ file
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

  local CM_POD_NAME=$(kubectl get pods --namespace ${CHART_MUSEUM_NS} -l "app.kubernetes.io/name=chartmuseum" -o jsonpath="{.items[0].metadata.name}")
  /bin/sh -c "kubectl port-forward $CM_POD_NAME ${CM_PORT}:8080 --namespace ${CHART_MUSEUM_NS} &"
  sleep 2

  CHART_EXISTS=$(curl -u "${CM_USER}:${CM_PWD}" -X GET http://localhost:${CM_PORT}/api/charts/${CHART_NAME}/${CHART_VERSION} | jq -r 'any([ .error] ; . > 0)')
  if [ "$CHART_EXISTS" == "true" ]; then
    echo ">> CHART EXISTS: deleting"
    curl -u "${CM_USER}:${CM_PWD}" -X DELETE http://localhost:${CM_PORT}/api/charts/${CHART_NAME}/${CHART_VERSION}
  fi
  
  echo ">> Uploading chart from file ${CHART_FILE}"
  curl -u "${CM_USER}:${CM_PWD}" --data-binary "@${CHART_FILE}" http://localhost:${CM_PORT}/api/charts  
  
  # End port forward
  pkill -f "kubectl port-forward $CM_POD_NAME ${CM_PORT}:8080 --namespace ${CHART_MUSEUM_NS}"

  rm ${CHART_FILE}
}

# Install ChartsMuseum
installChartMuseum() {
  echo "Installing ChartMuseum ${CM_VERSION}..."
  helm install chartmuseum --namespace ${CHART_MUSEUM_NS} --create-namespace "https://github.com/chartmuseum/charts/releases/download/chartmuseum-${CM_VERSION}/chartmuseum-${CM_VERSION}.tgz" \
    --set env.open.DISABLE_API=false \
    --set persistence.enabled=true \
    --set env.secret.BASIC_AUTH_USER=$CM_USER \
    --set env.secret.BASIC_AUTH_PASS=$CM_PWD
  kubectl rollout status -w deployment/chartmuseum --namespace=${CHART_MUSEUM_NS}

  echo "Chart museum v${CM_VERSION} installed in namespace ${CHART_MUSEUM_NS}"
  echo "Credentials: ${CM_USER} / ${CM_PWD}"
  echo "Cluster internal URL: "
  echo "    http://chartmuseum.${CHART_MUSEUM_NS}.svc.cluster.local:8080/"
}

# Uninstall ChartsMuseum
uninstallChartMuseum() {
  echo "Uninstalling ChartMuseum..."
  helm uninstall chartmuseum --namespace ${CHART_MUSEUM_NS}
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
