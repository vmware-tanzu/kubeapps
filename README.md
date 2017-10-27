# Setup

TODO: vendor some of this, or otherwise reduce the external
dependencies.

```
# Install kubecfg somewhere in $PATH
# NB: v0.5.0, not master
wget -O $HOME/bin/kubecfg https://github.com/ksonnet/kubecfg/releases/download/v0.5.0/kubecfg-linux-amd64
chmod +x $HOME/bin/kubecfg

# Install kubecfg jsonnet lib somewhere
kubecfglib_dir=$PWD
wget https://raw.githubusercontent.com/ksonnet/kubecfg/master/lib/kubecfg.libsonnet

# Install ksonnet-lib somewhere
ksonnetlib_dir=$PWD/ksonnet-lib
git clone github.com/ksonnet/ksonnet-lib.git $ksonnetlib_dir

# Doesn't have to be here, but kubeapps' `kubeless.jsonnet` lazily
# assumes this structure.  Complain if you want something else.
cd $GOPATH/src
git clone github.com/kubeless/kubeless.git github.com/kubeless/kubeless
git clone github.com/kubeapps/manifest.git github.com/kubeapps/manifest

kubeapps=$GOPATH/src/github.com/kubeapps/manifest
export KUBECFG_JPATH=$kubeapps/lib:$ksonnetlib_dir:$kubecfglib_dir

# Make sure your ~/.kube/config points to a working cluster
# If required: minikube start
kubectl cluster-info
```

# Usage

```
# Bring up cluster
./kubeapps.sh up

# Bring down cluster
./kubeapps.sh down

# Update to newer manifests (after modification)
./kubeapps.sh up

# "I just want to see the YAML"
kubecfg show kubeapps.jsonnet
```

# Notes

Installs into `kubeapps`, `kube-system`, and `kubeless` namespaces.

There's a whole mash of different jsonnet styles in this repo,
optimised for least change from existing files rather than any attempt
at long-term maintenance/cleanliness.  Enjoy :/
