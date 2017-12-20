# AppRepository Controller

The controller for managing app (Helm) repository syncs for
[Kubeapps](https://kubeapps.com).

An AppRepository resource looks like this:

```
apiVersion: v1
items:
apiVersion: kubeapps.com/v1alpha1
kind: AppRepository
metadata:
  name: bitnami
spec:
  url: https://charts.bitnami.com/incubator
  type: helm
```

This controller will monitor resources of the above type and create [Kubernetes
CronJobs](https://kubernetes.io/docs/concepts/workloads/controllers/cron-jobs/)
to schedule the repository to be synced to the database. This is a component of
Kubeapps and is intended to be used with it.

Based off the [Kubernetes Sample
Controller](https://github.com/kubernetes/sample-controller).
