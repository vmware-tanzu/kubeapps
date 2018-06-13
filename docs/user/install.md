# Installation

Kubeapps assumes a working Kubernetes cluster (v1.8+) and [`kubectl`](https://kubernetes.io/docs/tasks/tools/install-kubectl/) installed and configured to talk to your Kubernetes cluster. Kubeapps binaries are available for Linux, OS X and Windows, and Kubeapps has been tested with Azure Kubernetes Service (AKS), Google Kubernetes Engine (GKE), `minikube` and Docker for Desktop Kubernetes. Kubeapps works on RBAC-enabled clusters and this configuration is encouraged for a more secure install.

> On GKE, you must either be an "Owner" or have the "Container Engine Admin" role in order to install Kubeapps.

## Install pre-built binary

* Download a binary version of the latest Kubeapps Installer for your platform from the [release page](https://github.com/kubeapps/kubeapps/releases). Currently, the Kubeapps Installer is distributed in binary form for Linux, OS X and Windows (64-bit).
* Make the binary executable.

For example, to install the latest binary release on Linux or OS X, use this command:

```
curl -s https://api.github.com/repos/kubeapps/kubeapps/releases/latest | grep -i $(uname -s) | grep browser_download_url | cut -d '"' -f 4 | wget -i -
sudo mv kubeapps-$(uname -s| tr '[:upper:]' '[:lower:]')-amd64 /usr/local/bin/kubeapps
sudo chmod +x /usr/local/bin/kubeapps
```

## Build binary from source

The Kubeapps Installer is a CLI tool written in Go that will deploy the Kubeapps components into your cluster.
You can build the latest Kubeapps Installer from source by following the steps below:

* Visit [the Go website](https://golang.org), download the most recent [binary distribution of Go](https://golang.org/dl/) and install it following the [official instructions](https://golang.org/doc/install).

  > The remainder of this section assumes that Go is installed in `/usr/local/go`. Update the paths in subsequent commands if you used a different location.

* Set the Go environment variables:

  ```
  export GOROOT=/usr/local/go
  export GOPATH=/usr/local/go
  export PATH=$GOPATH/bin:$PATH
  ```

* Create a working directory for the project:

  ```
  working_dir=$GOPATH/src/github.com/kubeapps/
  mkdir -p $working_dir
  ```

* Clone the Kubeapps source repository:

  ```
  cd $working_dir
  git clone https://github.com/kubeapps/kubeapps
  ```

* Build the Kubeapps binary and move it to a location in your path:

  ```
  cd kubeapps
  make binary
  cp kubeapps /usr/local
  ```

# Next Steps

Confirm that [`kubectl`](https://kubernetes.io/docs/tasks/tools/install-kubectl/) is installed and verify the Kubernetes version:

```
kubectl version
Client Version: version.Info{Major:"1", Minor:"7", GitVersion:"v1.7.0", GitCommit:"d3ada0119e776222f11ec7945e6d860061339aad", GitTreeState:"clean", BuildDate:"2017-06-29T23:15:59Z", GoVersion:"go1.8.3", Compiler:"gc", Platform:"darwin/amd64"}
Server Version: version.Info{Major:"1", Minor:"8", GitVersion:"v1.8.0", GitCommit:"0b9efaeb34a2fc51ff8e4d34ad9bc6375459c4a4", GitTreeState:"dirty", BuildDate:"2017-10-17T15:09:55Z", GoVersion:"go1.8.3", Compiler:"gc", Platform:"linux/amd64"}
```

Use the Kubeapps Installer to deploy Kubeapps and launch a browser with the Kubeapps dashboard.

```
kubeapps up
kubeapps dashboard
```

To remove Kubeapps, use this command:

```
kubeapps down
```

# Exposing Externally

The main Kubernetes Service used to access Kubeapps is the `kubeapps` Service in the `kubeapps` namespace. To expose the dashboard for external access, you should setup an Ingress resource to point to the `kubeapps` Service and use an [Ingress Controller](https://kubernetes.io/docs/concepts/services-networking/ingress/#ingress-controllers) to expose it.

An alternative way is to edit the `kubeapps` Service and change it's type from `ClusterIP` to `LoadBalancer`, if your cloud platform supports provisioning LoadBalancer IPs or Hostnames for Services. This solution is not recommended as updates to Kubeapps will likely revert the Service back to a `ClusterIP` type.

# Useful Resources

* [Walkthrough for new users](getting-started.md)
* [Kubeapps Dashboard documentation](dashboard.md)
* [Kubeapps components](components.md)
