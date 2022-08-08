# Copyright 2021-2022 the Kubeapps contributors.
# SPDX-License-Identifier: Apache-2.0

# this Dockerfile is for building a docker image for a pod that can be deployed into a k8s cluster
# that serves some of the content of test-data (index.yaml, charts .tgz)
FROM nginx:1.23.1
COPY ./nginx.conf /etc/nginx/nginx.conf

# only has a single user: foo, password: bar
COPY ./md5.htpasswd /etc/apache2/.htpasswd

COPY ./charts/podinfo-index.yaml /usr/share/nginx/html/podinfo/index.yaml
COPY ./charts/podinfo-6.0.0.tgz /usr/share/nginx/html/podinfo/
COPY ./charts/podinfo-5.2.1.tgz /usr/share/nginx/html/podinfo/

COPY ./charts/podinfo-basic-auth-index.yaml /usr/share/nginx/html/podinfo-basic-auth/index.yaml
COPY ./charts/podinfo-6.0.0.tgz /usr/share/nginx/html/podinfo-basic-auth/
COPY ./charts/podinfo-5.2.1.tgz /usr/share/nginx/html/podinfo-basic-auth/

COPY ./cert/ssl-bundle.pem /etc/ssl/certs/
COPY ./cert/server-key.pem /etc/ssl/certs/
COPY ./charts/podinfo-tls-index.yaml /usr/share/nginx/html/podinfo-tls/index.yaml
COPY ./charts/podinfo-6.0.0.tgz /usr/share/nginx/html/podinfo-tls/
COPY ./charts/podinfo-5.2.1.tgz /usr/share/nginx/html/podinfo-tls/

RUN mkdir /usr/share/nginx/html/bitnami/ ; curl "https://charts.bitnami.com/bitnami/index.yaml" -o /usr/share/nginx/html/bitnami/index.yaml
