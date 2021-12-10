# this Dockerfile is for building a docker image for a pod that can be deployed into a k8s cluster
# that serves some of the content of test-data (index.yaml, charts .tgz)
FROM nginx:1.21.3
COPY ./nginx.conf /etc/nginx/nginx.conf

# only has a single user: foo, password: bar
COPY ./.htpasswd /etc/apache2/.htpasswd

COPY ./podinfo-index.yaml /usr/share/nginx/html/podinfo/index.yaml
COPY ./podinfo-6.0.0.tgz /usr/share/nginx/html/podinfo/
COPY ./podinfo-5.2.1.tgz /usr/share/nginx/html/podinfo/

COPY ./podinfo-basic-auth-index.yaml /usr/share/nginx/html/podinfo-basic-auth/index.yaml
COPY ./podinfo-6.0.0.tgz /usr/share/nginx/html/podinfo-basic-auth/
COPY ./podinfo-5.2.1.tgz /usr/share/nginx/html/podinfo-basic-auth/
