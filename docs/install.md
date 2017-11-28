# Installation

Kubeapps assumes a working Kubernetes (v1.7+) with RBAC enabled and [Kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/) installed and configured to talk to your Kubernetes cluster.

Kubeapps has been tested with both `minikube` and Google Kubernetes Engine (GKE). 

> On GKE, you must either be an "Owner" or have the "Container Engine Admin" role in order to install Kubeapps.

## Install pre-built binary

* Download a binary version of the latest Kubeapps Installer for your platform from the [release page](https://github.com/kubeapps/kubeapps/releases). Currently, the Kubeapps Installer is distributed in binary form for two platforms: Linux amd64 and OS X amd64.
* Make the binary executable.

For example, to install 0.0.2 release on Linux, use this command:

```
sudo curl -L https://github.com/kubeapps/installer/releases/download/v0.0.2/kubeapps-linux-amd64 -o /usr/local/bin/kubeapps && sudo chmod +x /usr/local/bin/kubeapps
```

## Build binary from source

The Kubeapps Installer is a CLI tool written in Go that will deploy the Kubeapps components into your cluster.
You can build the latest Kubeapps Installer from source by following the steps below: 

* Visit [the Go website](https://golang.org), download the most recent [binary distribution of Go](https://golang.org/dl/) and install it following the [official instructions](https://golang.org/doc/install). 

  The remainder of this section assumes that Go is installed in `/usr/local/go`. Update the paths in subsequent commands if you used a different location.

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

