# Migration to v1.0.0-alpha.5

The release includes several breaking changes that should be handled carefully if you are updating Kubeapps from a version prior to v1.0.0-alpha.5. As a summary this release include the following breaking changes:

 - The recommended way of installing Kubeapps is through its helm chart.
 - The `kubeapps` cli is now deprecated. **It won't be included in future releases**.
 - Kubeapps no longer includes Tiller nor Kubeless by default.
 - The component `Helm CRD` has been substitued with a secure proxy to the Tiller server.

These are the steps you need to follow to upgrade Kubeapps to this version.

## Install Tiller

Please follow the steps in [this guide](./securing-kubeapps.md) to install Tiller securely. Don't install the Kubeapps chart yet since it will fail because it will find resources that already exist. Once the new Tiller instance is ready you can migrate the existing releases using the utility included in `kubeapps` 1.0.0-alpha.5:

```
$ kubeapps migrate-configmaps-to-secrets
2018/08/03 16:00:35 Migrated foo.v1 as a secret
2018/08/03 16:00:35 Done. ConfigMaps are left in the namespace kubeapps to debug possible errors. Please delete them manually
```

If you list the releases you should be able to see all of them:

```
$ helm list --tls --tls-ca-cert cert/kubeapps-ca.pem --tls-cert cert/kubeapps.crt --tls-key cert/kubeapps.key
NAME               	REVISION	UPDATED                 	STATUS	 CHART          	NAMESPACE
foo                	1       	Fri Aug  3 15:42:42 2018	DEPLOYED aerospike-0.1.7	default
```

## Delete the previous Kubeapps installation

Now that we have backed up the releases we should existing Kubeapps resources. To do so execute:

```
kubeapps down
kubectl delete crd helmreleases.helm.bitnami.com sealedsecrets.bitnami.com
kubectl delete -f https://github.com/bitnami-labs/sealed-secrets/releases/download/v0.7.0/controller.yaml
kubectl get helmreleases -o=name --all-namespaces | xargs kubectl patch $1 --type merge -p '{ "metadata": { "finalizers": [] } }'
```

### Delete Kubeless

If you want to delete Kubeless (if you are not using it) you can delete it executing the following command:

```
kubectl delete -f https://github.com/kubeless/kubeless/releases/download/v0.6.0/kubeless-v0.6.0-alpha.7.yaml
```

## Install the Kubeapps chart

Now you can install the new version of Kubeapps using the Helm chart included in this repository (this command assumes that you have followed the guide [here](./securing-kubeapps.md):

```
helm install \
  --tls --tls-ca-cert kubeapps-ca.pem --tls-cert kubeapps.crt --tls-key kubeapps.key \
  --set tillerProxy.tls.ca="$(cat kubeapps-ca.pem)" \
  --set tillerProxy.tls.key="$(cat kubeapps.key)" \
  --set tillerProxy.tls.cert="$(cat kubeapps.crt)" \
  --namespace kubeapps \
  ./chart/kubeapps
```

When the chart is finally ready you can access the application showed after installing the chart.

If you still want to access Kubeless functions using Kubeapps you need to manually write the URL including the namespace of them. For example `http://localhost:8080/functions/ns/default`.
