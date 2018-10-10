# Migration to v1.0.0-alpha.5

The release includes several breaking changes that should be handled carefully if you are updating Kubeapps from a version prior to v1.0.0-alpha.5. As a summary this release includes the following breaking changes:

- The recommended way of installing Kubeapps is through its Helm chart.
- The `kubeapps` CLI is now deprecated. **It won't be included in future releases**.
- Kubeapps no longer installs Tiller, Kubeless and SealedSecrets by default.
- The [experimental Helm CRD controller](https://github.com/bitnami-labs/helm-crd) has been replaced with a secure REST proxy to the Tiller server. More info about this proxy [here](../../cmd/tiller-proxy/README.md).

These are the steps you need to follow to upgrade Kubeapps to this version.

## Install Tiller

Please follow the steps in [this guide](./securing-kubeapps.md) to install Tiller securely. Don't install the Kubeapps chart yet since it will fail because it will find resources that already exist. Once the new Tiller instance is ready you can migrate the existing releases using the utility command included in `kubeapps` 1.0.0-alpha.5:

```
$ kubeapps migrate-configmaps-to-secrets --target-tiller-namespace kube-system
2018/08/06 12:24:23 Migrated foo.v1 as a secret
2018/08/06 12:24:23 Done. ConfigMaps are left in the namespace kubeapps to debug possible errors. Please delete them manually
```

**NOTE**: The tool asumes that you have deployed Helm storing releases as secrets. If that is not the case you can still migrate the releases executing:

```
kubectl get configmaps -n kubeapps -o yaml -l OWNER=TILLER | sed 's/namespace: kubeapps/namespace: kube-system/g'  | kubectl create -f -
```

If you list the releases you should be able to see all of them:

```
$ helm ls --tls --tls-ca-cert ca.cert.pem --tls-cert helm.cert.pem --tls-key helm.key.pem
NAME	REVISION	UPDATED                 	STATUS  	CHART          	NAMESPACE
foo 	1       	Mon Aug  6 12:10:07 2018	DEPLOYED	aerospike-0.1.7	default
```

**NOTE**: You can skip the TLS flags if you have not installed Helm with a TLS certificate.

## Delete the previous Kubeapps installation

Now that we have backed up the releases we should delete existing Kubeapps resources. To do so execute:

```
kubeapps down
kubectl delete crd helmreleases.helm.bitnami.com sealedsecrets.bitnami.com
kubectl delete -f https://github.com/bitnami-labs/sealed-secrets/releases/download/v0.7.0/controller.yaml
kubectl get helmreleases -o=name --all-namespaces | xargs kubectl patch $1 --type merge -p '{ "metadata": { "finalizers": [] } }'
```

Wait until everything in the namespace of Kubeapps has been deleted:

```
$ kubectl get all --namespace kubeapps
No resources found.
```

### Delete Kubeless

If you want to delete Kubeless (if you are not using it) you can delete it executing the following command:

```
kubectl delete -f https://github.com/kubeless/kubeless/releases/download/v0.6.0/kubeless-v0.6.0.yaml
```

## Install the Kubeapps chart

Now you can install the new version of Kubeapps using the Helm chart included in this repository:

```
helm repo add bitnami https://charts.bitnami.com/bitnami
helm install \
  --tls --tls-ca-cert ca.cert.pem --tls-cert helm.cert.pem --tls-key helm.key.pem \
  --set tillerProxy.tls.ca="$(cat ca.cert.pem)" \
  --set tillerProxy.tls.key="$(cat helm.key.pem)" \
  --set tillerProxy.tls.cert="$(cat helm.cert.pem)" \
  --namespace kubeapps \
  bitnami/kubeapps
```

**NOTE**: You can skip the TLS flags if you have not installed Helm with a TLS certificate.

When the chart is finally ready you can access the application and you will see your previous applications.
