# Kubeapps Installer

## Build

```
$ go build -o kubeapps main.go
```

## Usage

Put manifests in a particular folder (default to `~/.kubeapps`)
```
# Bring up cluster
$ kubeapps up
# Or providing manifests folder
$ kubeapps up --path /path/to/manifests

# Bring down cluster
$ kubeapps down

# Update the newer manifests (after modification)
$ kubeapps up
```
