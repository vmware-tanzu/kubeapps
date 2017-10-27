# Setup

TODO: vendor some of this, or otherwise reduce the external
dependencies.

```
# NB: need submodules
d=kubeapps-manifest
git clone --recurse-submodules https://github.com/kubeapps/manifest $d

# Install kubecfg somewhere in $PATH
# NB: v0.5.0, not master
wget -O $HOME/bin/kubecfg https://github.com/ksonnet/kubecfg/releases/download/v0.5.0/kubecfg-linux-amd64
chmod +x $HOME/bin/kubecfg

export KUBECFG_JPATH=$d/lib:$d/vendor/kubecfg/lib:$d/ksonnet-lib

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
