# Components

## Helm / Tiller

Tiller is [Helm](https://helm.sh/)'s in-cluster component that enables easy installation of Kubernetes-ready applications packaged as charts.

## Kubeless

[Kubeless](http://kubeless.io/) is a Kubernetes-native serverless framework. It runs on top of a Kubernetes cluster and allows you to deploy small units of code without having to build container images. With Kubeless, you can build advanced applications that tie together services using functions.

## Kubeapps Dashboard

Kubeapps comes with an in-cluster web dashboard that allows you easy deployment of applications and functions packaged as charts and to create, edit and run Kubeless functions. [Learn more about the Kubeapps Dashboard](dashboard.md).

## Sealed Secrets

[Sealed Secrets](https://github.com/bitnami/sealed-secrets) are "one-way" encrypted Secrets that can be created by anyone, but can only be decrypted by the controller running in the target cluster. A Sealed Secret is safe to share publicly, upload to Git repositories, post to Twitter, etc. Once the SealedSecret is safely uploaded to the target Kubernetes cluster, the Sealed Secrets controller will decrypt it and recover the original Secret.

For more information, check the [latest version of the Kubernetes manifest](https://github.com/kubeapps/kubeapps/blob/master/static/kubeapps-objs.yaml) that Kubeapps will deploy for you.