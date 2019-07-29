#!/bin/bash

set -e -x

updateVersions() {
    local targetRepo=${1:?}
    local targetTag=${2:?}
    local targetChartPath="${targetRepo}/${CHART_REPO_PATH}"
    local chartYaml="${targetChartPath}/Chart.yaml"
    if [ ! -f "${chartYaml}" ]; then
        echo "Wrong repo path. You should provide the root of the repository" > /dev/stderr
        return 1
    fi
    # DANGER: This replaces any tag marked as latest in the values.yaml
    local tagWithoutV=$(echo $targetTag | tr -d v)
    sed -i.bk 's/tag: latest/tag: '"${tagWithoutV}"'/g' "${targetChartPath}/values.yaml"
    # Use bitnami images
    #sed -i.bk 's/repository: kubeapps\/\(.*\)/repository: bitnami\/kubeapps-\1/g' "${targetChartPath}/values.yaml"
    rm "${targetChartPath}/values.yaml.bk"
}

docker tag kubeapps/apprepository-controller:${VERSION} ${REGISTRY}/${PROJECT_NAME}/apprepository-controller:latest
docker push ${REGISTRY}/${PROJECT_NAME}/apprepository-controller:latest
docker tag kubeapps/apprepository-controller:${VERSION} ${REGISTRY}/${PROJECT_NAME}/apprepository-controller:${VERSION}
docker push ${REGISTRY}/${PROJECT_NAME}/apprepository-controller:${VERSION}

docker tag kubeapps/dashboard:${VERSION} ${REGISTRY}/${PROJECT_NAME}/dashboard:latest
docker push ${REGISTRY}/${PROJECT_NAME}/dashboard:latest
docker tag kubeapps/dashboard:${VERSION} ${REGISTRY}/${PROJECT_NAME}/dashboard:${VERSION}
docker push ${REGISTRY}/${PROJECT_NAME}/dashboard:${VERSION}

export GATE_POD=$(kubectl get pods -n ${NAMESPACE} -l "component=gate" -o jsonpath="{.items[0].metadata.name}")
kubectl port-forward -n ${NAMESPACE} ${GATE_POD} 8084:8084 >> /dev/null &
kubectl port-forward -n cicd svc/cicd-chartmuseum 9090:8080 >> /dev/null &

sleep 5

updateVersions ./chart/kubeapps ${VERSION}
helm repo add chartmuseum http://localhost:9090

helm package -u --app-version ${VERSION} --version ${VERSION} ./chart/kubeapps

#Push Chart
helm push "kubeapps-${VERSION}.tgz" chartmuseum

#Trigger Deployment
curl -X POST http://localhost:8084/webhooks/webhook/deploy-infinity-service -H "Content-Type: application/json" -d '{ "serviceName": "'"${REPO}"'", "parameters": { "version": "'"${VERSION}"'" } }'
