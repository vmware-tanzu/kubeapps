# The Kubeapps Components

### Kubeapps dashboard

The dashboard is the main UI component of the Kubeapps project. Written in Javascript, the dashboard uses the React Javascript library for the frontend.

Please refer to the [Kubeapps Dashboard Developer Guide](dashboard.md) for the developer setup.

### chartsvc

The `chartsvc` component is a micro-service that creates a API endpoint for accessing the metadata for charts in Helm chart repositories that's populated in a MongoDB server.

<!-- TODO: Point to developer guide in Monocular repo  -->

### chart-repo

The `chart-repo` component is tool that scans a Helm chart repository and populates chart metadata in a MongoDB server. This metadata is then served by the `chartsvc` component.

<!-- TODO: Point to developer guide in Monocular repo  -->

### tiller-proxy

The `tiller-proxy` component is a service used both as a client for Tiller but also to provide a way to authorize users to deploy, upgrade and delete charts in different namespaces.

Please refer to the [Kubeapps tiller-proxy Developer Guide](tiller-proxy.md) for the developer setup.
