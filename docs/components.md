# Components

## Helm / Tiller

Kubeapps installs Tiller ([Helm](https://helm.sh/) in-cluster component) which will allow you to easily installed Kubernetes ready applications packaged as charts.

## Kubeless

[Kubeless](http://kubeless.io/) is a Kubernetes-native serverless framework. It runs on top of your Kubernetes cluster and allows you to deploy small unit of code without having to build container images. With Kubeless you can build advanced applications that tie together services using functions.

## Kubeapps Dashboard

Kubeapps comes with an in-cluster web dashboard that allows you to easily deploy applications and functions packaged as charts and to create, edit and run Kubeless functions.
Check the [main documentation for the Kubeapps Dashboard](dashboard.md).

## Sealed Secrets

[Sealed Secrets](https://github.com/bitnami/sealed-secrets) are a "one-way" encrypted Secret that can be created by anyone, but can only be decrypted by the controller running in the target cluster. The Sealed Secret is safe to share publicly, upload to git repositories, post to twitter, etc. Once the SealedSecret is safely uploaded to the target Kubernetes cluster, the sealed secrets controller will decrypt it and recover the original Secret.