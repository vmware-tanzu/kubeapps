# The Kubeapps Components

### Kubeapps installer

The Kubeapps installer is a command-line tool for installing, upgrading and uninstalling the Kubeapps in-cluster components. The tool is written using the Go programming language and the Kubernetes manifests are written in the Jsonnet data templating language.

Please refer to the [Kubeapps Installer Developer Guide](kubeapps.md) for the developer setup.

### Kubeapps dashboard

The dashboard is the main UI component of the Kubeapps project. Written in Javascript, the dashboard uses the React Javascript library for the frontend.

Please refer to the [Kubeapps Dashboard Developer Guide](dashboard.md) for the developer setup.

### chart-svc

The `chartsvc` component is a micro-service that creates a API endpoint for accessing the metadata for charts in Helm chart repositories that's populated in a MongoDB server.

Please refer to the [Kubeapps ChartSVC Developer Guide](chartsvc.md) for the developer setup.

### chart-repo

The `chart-repo` component is tool that scans a Helm chart repository and populates chart metadata in a MongoDB server. This metadata is then served by the `chartsvc` component.

Please refer to the [Kubeapps chart-repo Developer Guide](chart-repo.md) for the developer setup.
