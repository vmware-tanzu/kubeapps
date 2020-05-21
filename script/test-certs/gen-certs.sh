#!/bin/bash

cat << EOF > tls_config
[req]
req_extensions = v3_req
distinguished_name = req_distinguished_name

[req_distinguished_name]

[ v3_req ]
basicConstraints = CA:FALSE
keyUsage = nonRepudiation, digitalSignature, keyEncipherment
subjectAltName = @alt_names

[alt_names]
DNS.1 = localhost
EOF

openssl genrsa -out ./ca.key.pem 4096
openssl req -key ca.key.pem -new -x509 -days 7300 -sha256 -out ca.cert.pem -subj "/CN=kubeapps-ca"

## tiller server key
openssl genrsa -out ./tiller.key.pem 4096
## helm client key
openssl genrsa -out ./helm.key.pem 4096

openssl req -key tiller.key.pem -new -sha256 -out tiller.csr.pem -config tls_config -subj "/CN=kubeapps-ca"
openssl req -key helm.key.pem   -new -sha256 -out helm.csr.pem   -config tls_config -subj "/CN=kubeapps-ca"

openssl x509 -days 7300 -req -CA ca.cert.pem -CAkey ca.key.pem -CAcreateserial -in tiller.csr.pem -out tiller.cert.pem
openssl x509 -days 7300 -req -CA ca.cert.pem -CAkey ca.key.pem -CAcreateserial -in helm.csr.pem -out helm.cert.pem

# clean up unnecessary files
rm ca.key.pem ca.cert.srl helm.csr.pem tiller.csr.pem tls_config
