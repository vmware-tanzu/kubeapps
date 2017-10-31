# Kubeapps Installer

## Build

```
$ go build -o kubeapps main.go
```

## Usage

**Deploy with kubeapps.jsonnet**

```
$ d=$PWD/kubeapps-manifest
$ git clone --recurse-submodules https://github.com/kubeapps/manifest $d
$ export KUBECFG_JPATH=$d/lib:$d/vendor/kubecfg/lib:$d/vendor/ksonnet-lib
$ kubeapps up --file $d/kubeapps.jsonnet
$ kubeapps down --file $d/kubeapps.jsonnet
```

**Deploy with yaml manifests**

Put yaml manifests in a particular folder (default to `~/.kubeapps`)
```
# Bring up cluster
$ kubeapps up
# Or providing yaml manifests folder
$ kubeapps up --path /path/to/manifests

# Bring down cluster
$ kubeapps down

# Update the newer manifests (after modification)
$ kubeapps up
```
