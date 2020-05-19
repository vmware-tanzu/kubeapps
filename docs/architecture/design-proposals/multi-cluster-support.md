# Multi-cluster support for Kubeapps

## Objective

Enable Kubeapps users to be able to **install apps into other Kubernetes clusters** in addition to the cluster on which Kubeapps is installed.

This design aims to utilize the fact that to install apps on additional clusters, all that is required are user credentials with the required RBAC and network access to the Kubernetes API of the additional cluster. With Helm 3, we no longer need to run any infrastructure (ie. tiller or kubeapps services) on the additional cluster(s) to be able to deploy applications, as we are just sending YAML to the api server using the users credentials.

## User Stories

* As an operator of multiple Kubernetes clusters, I want to install, configure and maintain Kubeapps **on one cluster** so that users can deploy applications (initially Helm charts) **to the configured multiple clusters** to which they have access.
* As a user of Kubeapps, I want to login to my teams' Kubeapps instance and install ApplicationX into my namespace on the teams sandbox cluster to test before installing on the staging cluster.

## Explicit Constraints

* At least initially, we **will not support K8s service account tokens as a user authentication option** for users in multiple clusters. Kubeapps initially supported only service tokens for user authentication until support for OIDC/SSO with an appropriately configured Kubernetes API server was later added. At least initially, the multi-cluster support will **only be available when using SSO such that a single user can be authenticated with multiple clusters using the same credential** (eg. an OIDC id_token). We may later decide to support service account tokens for users, but it is not recommended for use (even with a single cluster).
* **We will not run parts of Kubeapps infrastructure in each additional cluster**: Ideally we will be able to achieve our objective without the requirement that human operators deploy and maintain extra Kubeapps services in additional clusters, which would move away significantly from `helm install`ing Kubeapps. Kubeapps allows easy configuration of Kubernetes Apps, resulting in YAML which can be applied to the K8s API server with the users credentials.
* **Network access from the Kubeapps cluster to the additional clusters’ API server**. For Kubeapps to support an additional cluster, the cluster operator would need to ensure that the additional cluster's API server is reachable from the Kubeapps’ pods on the initial cluster. This is normally the case for hosted clusters which have public endpoints requiring authorization or private endpoints within a common private network or even on multiple private networks which can be bridged.

## Design overview

The overview displayed below includes an indication of the multi-cluster support for private app repositories, which is discussed further in the design doc below.

![Kubeapps Multi-cluster Overview](img/Kubeapps-Multi-cluster-simple.png)

## Details and discussion

More details, design considerations and discussion is in the separate [design doc](https://docs.google.com/document/d/1-6cKxOsW6K5u3lK7Om2zQeVYVPxzHT6dVwej5wy3_9A/edit).