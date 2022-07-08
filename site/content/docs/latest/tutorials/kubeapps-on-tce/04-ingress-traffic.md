# Routing traffic to Kubeapps

## Using port-forward

Only for development or demo purposes!
## Using Contour

## Using Nginx

```bash
kubectl -n kubeapps create secret tls localhost-tls \
      --key $PROJECT_PATH/devel/localhost-key.pem \
      --cert $PROJECT_PATH/devel/localhost-cert.pem
    kubectl -n kubeapps create secret generic postgresql-db \
      --from-literal=postgres-postgres-password=dev-only-fake-password \
      --from-literal=postgres-password=dev-only-fake-password
```

- If using OauthProxy (usually with OIDC), add patch to increase proxy buffers:
https://github.com/vmware-tanzu/kubeapps/pull/1944/files

- Otherwise: https://raw.githubusercontent.com/kubernetes/ingress-nginx/main/deploy/static/provider/kind/deploy.yaml

## Using other ingresses

Comment the possibility
