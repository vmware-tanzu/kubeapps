This folder contains dummy self-signed certificates for CI tests.

To generate them execute:

```
openssl genrsa -out ./ca.key.pem 4096
cat <<EOF >> tls_config
[ req ]
distinguished_name="req_distinguished_name"
prompt="no"

[ req_distinguished_name ]
C="ES"
ST="Andalucia"
L="Sevilla"
O="Kubeapps"
CN="localhost"
EOF
openssl req -key ca.key.pem -new -x509 -days 7300 -sha256 -out ca.cert.pem -config tls_config

## server key
openssl genrsa -out ./tiller.key.pem 4096
## client key
openssl genrsa -out ./helm.key.pem 4096

openssl req -days 7300 -key tiller.key.pem -new -sha256 -out tiller.csr.pem -config tls_config
openssl req -days 7300 -key helm.key.pem -new -sha256 -out helm.csr.pem -config tls_config

openssl x509 -days 7300 -req -CA ca.cert.pem -CAkey ca.key.pem -CAcreateserial -in tiller.csr.pem -out tiller.cert.pem
openssl x509 -days 7300 -req -CA ca.cert.pem -CAkey ca.key.pem -CAcreateserial -in helm.csr.pem -out helm.cert.pem

# clean up unnecessary files
rm ca.key.pem ca.srl helm.csr.pem tiller.csr.pem tls_config
```
