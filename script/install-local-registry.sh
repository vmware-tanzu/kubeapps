#!/bin/bash

createLocalRegistry() {
    KIND_CLUSTER=$1
    REGISTRY_NS=$2
    PROJECT_PATH=$3

    CONTROL_PLANE="$KIND_CLUSTER-control-plane"
    kubectl create namespace $REGISTRY_NS

    # Add our CA to the node
    docker cp "$(mkcert -CAROOT)"/rootCA.pem "$CONTROL_PLANE:/usr/share/ca-certificates/local-rootCA.pem"
    docker exec --user root $CONTROL_PLANE sh -c "echo 'local-rootCA.pem' >> /etc/ca-certificates.conf"
    docker exec --user root $CONTROL_PLANE update-ca-certificates -f

    # Restart containerd to make the CA addition effective
    docker exec --user root $CONTROL_PLANE systemctl restart containerd

    # Generate new certificate for registry
    mkcert -key-file $PROJECT_PATH/devel/localhost-key.pem -cert-file $PROJECT_PATH/devel/docker-registry-cert.pem docker-registry "docker-registry.$REGISTRY_NS.svc.cluster.local"
    kubectl -n $REGISTRY_NS delete secret registry-tls --ignore-not-found=true
    kubectl -n $REGISTRY_NS create secret tls registry-tls --key $PROJECT_PATH/devel/localhost-key.pem --cert $PROJECT_PATH/devel/docker-registry-cert.pem

    # Prepare registry image
    docker pull registry:2
    kind load docker-image --name $KIND_CLUSTER registry:2

    # Create registry resources
    kubectl apply -f ${PROJECT_PATH}/integration/registry/local-registry.yaml

    # Wait for deployment to be ready
    kubectl rollout status -w deployment/private-registry -n $REGISTRY_NS

    # Add registry to node hosts
    REGISTRY_IP=$(kubectl get service/docker-registry -n $REGISTRY_NS -o jsonpath='{.spec.clusterIP}')
    docker exec --user root $CONTROL_PLANE sh -c "echo '$REGISTRY_IP  docker-registry' >> /etc/hosts"
}

pushContainerToLocalRegistry() {
    DOCKER_REGISTRY=$1
    REGISTRY_NS=$2

    echo "127.0.0.1  docker-registry" | sudo tee -a /etc/hosts

    docker pull nginx
    docker tag nginx $DOCKER_REGISTRY/nginx

    /bin/sh -c "kubectl -n ${REGISTRY_NS} port-forward service/docker-registry 5600:5600 &"
    sleep 2

    docker login $DOCKER_REGISTRY -u=testuser -p=testpassword
    docker push $DOCKER_REGISTRY/nginx
    docker logout $DOCKER_REGISTRY

    # End port forward
    pkill -f "kubectl -n ${REGISTRY_NS} port-forward service/docker-registry 5600:5600"
}

pushChartToChartMuseum() {
    CHART_MUSEUM_NS=$1
    CM_USER=$2
    CM_PASSWORD=$3
    CHART_NAME=$4
    CHART_VERSION=$5

    local POD_NAME=$(kubectl get pods --namespace ${CHART_MUSEUM_NS} -l "app=chartmuseum" -l "release=chartmuseum" -o jsonpath="{.items[0].metadata.name}")
    /bin/sh -c "kubectl port-forward $POD_NAME 8080:8080 --namespace ${CHART_MUSEUM_NS} &"
    sleep 2

    CHART_EXISTS=$(curl -u "${CM_USER}:${CM_PASSWORD}" -X GET http://localhost:8080/api/charts/${CHART_NAME}/${CHART_VERSION} | jq -r 'any([ .error] ; . > 0)')
    if [ "$CHART_EXISTS" == "true" ]; then
        # Charts exists and must be deleted first
        curl -u "${CM_USER}:${CM_PASSWORD}" -X DELETE http://localhost:8080/api/charts/${CHART_NAME}/${CHART_VERSION}
    fi

    # Upload chart to Chart Museum
    FILE_NAME="${CHART_NAME}-${CHART_VERSION}.tgz"
    curl -u "${CM_USER}:${CM_PASSWORD}" --data-binary "@${FILE_NAME}" http://localhost:8080/api/charts  

    # End port forward
    pkill -f "kubectl port-forward $POD_NAME 8080:8080 --namespace ${CHART_MUSEUM_NS}"
}

REGISTRY_NS=ci
KIND_CLUSTER=kubeapps
DOCKER_REGISTRY=docker-registry:5600
PROJECT_PATH="/Users/rcastelblanq/work/kubeapps/repo"

#createLocalRegistry $KIND_CLUSTER $REGISTRY_NS "/Users/rcastelblanq/work/kubeapps/repo"
#pushContainerToLocalRegistry $DOCKER_REGISTRY $REGISTRY_NS

helm package $PROJECT_PATH/integration/charts/simplechart
pushChartToChartMuseum chart-museum admin password simplechart "0.1.0"
