#!/bin/bash

# Copyright 2022 the Kubeapps contributors.
# SPDX-License-Identifier: Apache-2.0

CONTROL_PLANE_CONTAINER=${CONTROL_PLANE_CONTAINER:-"kubeapps-ci-control-plane"}
REGISTRY_NS=ci
DOCKER_REGISTRY_HOST=local-docker-registry
DOCKER_REGISTRY_PORT=5600
DOCKER_REGISTRY_URL="https://$DOCKER_REGISTRY_HOST:$DOCKER_REGISTRY_PORT"
DOCKER_USERNAME=testuser
DOCKER_PASSWORD=testpassword

installLocalRegistry() {
    if [ -z "$DOCKER_REGISTRY_VERSION" ]; then
      echo "No Docker registry version supplied"
      exit 1
    fi
    if [ -z "$1" ]; then
      echo "No project path supplied"
      exit 1
    fi
    local PROJECT_PATH=$1

    echo "Installing local Docker registry v${DOCKER_REGISTRY_VERSION} in control plane '${CONTROL_PLANE_CONTAINER}'"

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
    envsubst < "${PROJECT_PATH}/integration/registry/local-registry.yaml" | kubectl apply -f -

    # Wait for deployment to be ready
    kubectl rollout status -w deployment/private-registry -n $REGISTRY_NS

    # Add registry to node hosts.
    # Haven't found a way to use the cluster DNS from the node, 
    # following https://github.com/kubernetes/kubernetes/issues/8735#issuecomment-148800699
    REGISTRY_IP=$(kubectl -n $REGISTRY_NS get service/docker-registry -o jsonpath='{.spec.clusterIP}')
    docker exec --user root $CONTROL_PLANE_CONTAINER sh -c "echo '$REGISTRY_IP  $DOCKER_REGISTRY_HOST' >> /etc/hosts"

    echo "Installing Ingress for Docker registry with access through host ${DOCKER_REGISTRY_HOST}"
    kubectl apply -f - -o yaml << EOF
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  annotations:
    nginx.ingress.kubernetes.io/proxy-body-size: "0"
    nginx.ingress.kubernetes.io/proxy-read-timeout: "600"
    nginx.ingress.kubernetes.io/proxy-send-timeout: "600"
    nginx.ingress.kubernetes.io/backend-protocol: "HTTPS"
  name: docker-registry
  namespace: ${REGISTRY_NS}
spec:
  ingressClassName: nginx
  tls:
  - hosts:
    - ${DOCKER_REGISTRY_HOST}
    secretName: registry-tls
  rules:
  - host: ${DOCKER_REGISTRY_HOST}
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: docker-registry
            port:
              number: 5600
EOF
  sleep 10
}

pushContainerToLocalRegistry() {
    # Access through Ingress TLS
    DOCKER_REGISTRY="$DOCKER_REGISTRY_HOST:443"

    echo "127.0.0.1  $DOCKER_REGISTRY_HOST" | sudo tee -a /etc/hosts

    docker pull nginx
    docker tag nginx $DOCKER_REGISTRY/nginx

    docker login $DOCKER_REGISTRY -u=testuser -p=testpassword
    docker push $DOCKER_REGISTRY/nginx
    docker logout $DOCKER_REGISTRY
}

# Scans for opened port during max. 10 seconds
waitForPort() {
  HOST_NAME=$1
  PORT=$2

  # shellcheck disable=SC2016
  timeout 10 sh -c 'until nc -z $0 $1; do sleep 1; done' "$HOST_NAME" "$PORT"
}

uninstallLocalRegistry() {
  if [ -z "$DOCKER_REGISTRY_VERSION" ]; then
    echo "No Docker registry version supplied"
    exit 1
  fi
  if [ -z "$1" ]; then
    echo "No project path supplied"
    exit 1
  fi
  local PROJECT_PATH=$1
  
  envsubst < "${PROJECT_PATH}/integration/registry/local-registry.yaml" | kubectl delete -f -
  kubectl -n ${REGISTRY_NS} delete ingress docker-registry
  kubectl -n ${REGISTRY_NS} delete secret registry-tls
}

case $1 in

  install)
    installLocalRegistry $2
    ;;

  pushNginx)
    pushContainerToLocalRegistry
    ;;

  uninstall)
    uninstallLocalRegistry $2
    ;;

esac
