# Migration to v1.0.0

This release includes several breaking changes that can make upgrading fairly
difficult. Helm 2.10 and above is also now required to install Kubeapps.

If you have difficulty upgrading to the v1.0.0 release, we recommend backing up
any AppRepository objects (custom repositories) you may have added and perform a
clean install of Kubeapps.

To backup a custom repository, run the following command for each repository:

```
kubectl get apprepository -o yaml <repo name> > <repo name>.yaml
```

**Note**: you do not need to backup the `stable`, `incubator`, `bitnami` or
`svc-cat` repositories, as these will be recreated when reinstalling Kubeapps.

After backing up your custom repositories, run the following command to remove
and reinstall Kubeapps:

```
helm delete --purge kubeapps
helm install bitnami/kubeapps --version 1.0.0
```

To recover your custom repository backups, run the following command for each
repository:

```
kubectl apply -f <repo name>.yaml
```
