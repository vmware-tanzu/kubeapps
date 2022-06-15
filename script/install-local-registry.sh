#!/bin/bash

CONTROL_PLANE_CONTAINER="kubeapps-ci-control-plane"
REGISTRY_NS=ci
DOCKER_REGISTRY_HOST=local-docker-registry
DOCKER_REGISTRY_PORT=5600
DOCKER_USERNAME=testuser
DOCKER_PASSWORD=testpassword

installLocalRegistry() {
    PROJECT_PATH=$1

    kubectl create namespace $REGISTRY_NS

    # Add our CA to the node
    docker cp "$(mkcert -CAROOT)"/rootCA.pem "$CONTROL_PLANE_CONTAINER:/usr/share/ca-certificates/local-rootCA.pem"
    docker exec --user root $CONTROL_PLANE_CONTAINER sh -c "echo 'local-rootCA.pem' >> /etc/ca-certificates.conf"
    docker exec --user root $CONTROL_PLANE_CONTAINER update-ca-certificates -f

    # Restart containerd to make the CA addition effective
    docker exec --user root $CONTROL_PLANE_CONTAINER systemctl restart containerd

    # Generate new certificate for registry
    mkcert -key-file $PROJECT_PATH/devel/localhost-key.pem -cert-file $PROJECT_PATH/devel/docker-registry-cert.pem $DOCKER_REGISTRY_HOST "docker-registry.$REGISTRY_NS.svc.cluster.local"
    kubectl -n $REGISTRY_NS delete secret registry-tls --ignore-not-found=true
    kubectl -n $REGISTRY_NS create secret tls registry-tls --key $PROJECT_PATH/devel/localhost-key.pem --cert $PROJECT_PATH/devel/docker-registry-cert.pem

    # Create registry resources
    kubectl apply -f ${PROJECT_PATH}/integration/registry/local-registry.yaml

    # Wait for deployment to be ready
    kubectl rollout status -w deployment/private-registry -n $REGISTRY_NS

    # Add registry to node hosts.
    # Haven't found a way to use the cluster DNS from the node, 
    # following https://github.com/kubernetes/kubernetes/issues/8735#issuecomment-148800699
    REGISTRY_IP=$(kubectl -n $REGISTRY_NS get service/docker-registry -o jsonpath='{.spec.clusterIP}')
    docker exec --user root $CONTROL_PLANE_CONTAINER sh -c "echo '$REGISTRY_IP  $DOCKER_REGISTRY_HOST' >> /etc/hosts"
}

pushContainerToLocalRegistry() {
    DOCKER_REGISTRY="$DOCKER_REGISTRY_HOST:$DOCKER_REGISTRY_PORT"

    echo "127.0.0.1  $DOCKER_REGISTRY_HOST" | sudo tee -a /etc/hosts

    docker pull nginx
    docker tag nginx $DOCKER_REGISTRY/nginx

    /bin/sh -c "kubectl -n ${REGISTRY_NS} port-forward service/docker-registry 5600:5600 &"
    sleep 4

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
