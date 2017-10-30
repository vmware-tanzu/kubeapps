# Kubeapps Installer

## Build

```
$ go build -o kubeapps main.go
```

## Usage

Put manifests into `~/.kubeapps` folder.

```
# Bring up cluster
$ kubeapps up

# Bring down cluster
$ kubeapps down

# Update the newer manifests (after modification)
$ kubeapps up
```
