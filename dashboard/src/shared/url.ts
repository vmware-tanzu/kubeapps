// Copyright 2018-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import {
  AvailablePackageReference,
  InstalledPackageReference,
} from "gen/kubeappsapis/core/packages/v1alpha1/packages";

export const app = {
  apps: {
    new: (
      cluster: string,
      namespace: string,
      availablePackageReference: AvailablePackageReference,
      version: string,
    ) => {
      const pkgPluginName = availablePackageReference?.plugin?.name;
      const pkgPluginVersion = availablePackageReference?.plugin?.version;
      const pkgId = availablePackageReference?.identifier || "";
      // Some plugins may not be cluster-aware nor support multi-cluster, so
      // if the returned available package ref doesn't set cluster, use the current
      // one.
      const pkgCluster = availablePackageReference?.context?.cluster || cluster;
      const pkgNamespace = availablePackageReference?.context?.namespace;
      return `/c/${cluster}/ns/${namespace}/apps/new/${pkgPluginName}/${pkgPluginVersion}/${pkgCluster}/${pkgNamespace}/${encodeURIComponent(
        pkgId,
      )}/versions/${version}`;
    },
    list: (cluster?: string, namespace?: string) => `/c/${cluster}/ns/${namespace}/apps`,
    get: (installedPackageReference: InstalledPackageReference) => {
      const pkgCluster = installedPackageReference?.context?.cluster;
      const pkgNamespace = installedPackageReference?.context?.namespace;
      const pkgPluginName = installedPackageReference?.plugin?.name;
      const pkgPluginVersion = installedPackageReference?.plugin?.version;
      const pkgId = installedPackageReference?.identifier || "";
      return `${app.apps.list(
        pkgCluster,
        pkgNamespace,
      )}/${pkgPluginName}/${pkgPluginVersion}/${pkgId}`;
    },
    upgrade: (ref: InstalledPackageReference) => `${app.apps.get(ref)}/upgrade`,
    upgradeTo: (ref: InstalledPackageReference, version?: string) =>
      `${app.apps.get(ref)}/upgrade/${version}`,
  },
  catalog: (cluster: string, namespace: string) => `/c/${cluster}/ns/${namespace}/catalog`,
  packages: {
    get: (
      cluster: string,
      namespace: string,
      availablePackageReference: AvailablePackageReference,
    ) => {
      const pkgPluginName = availablePackageReference?.plugin?.name;
      const pkgPluginVersion = availablePackageReference?.plugin?.version;
      const pkgId = availablePackageReference?.identifier || "";
      // Some plugins may not be cluster-aware nor support multi-cluster, so
      // if the returned available package ref doesn't set cluster, use the current
      // one.
      const pkgCluster = availablePackageReference?.context?.cluster || cluster;
      const pkgNamespace = availablePackageReference?.context?.namespace;
      return `/c/${cluster}/ns/${namespace}/packages/${pkgPluginName}/${pkgPluginVersion}/${pkgCluster}/${pkgNamespace}/${encodeURIComponent(
        pkgId,
      )}`;
    },
  },
  operators: {
    view: (cluster: string, namespace: string, name: string) =>
      `/c/${cluster}/ns/${namespace}/operators/${name}`,
    list: (cluster: string, namespace: string) => `/c/${cluster}/ns/${namespace}/operators`,
    new: (cluster: string, namespace: string, name: string) =>
      `/c/${cluster}/ns/${namespace}/operators/new/${name}`,
  },
  operatorInstances: {
    view: (
      cluster: string,
      namespace: string,
      csvName: string,
      crdName: string,
      resourceName: string,
    ) => `/c/${cluster}/ns/${namespace}/operators-instances/${csvName}/${crdName}/${resourceName}`,
    update: (
      cluster: string,
      namespace: string,
      csvName: string,
      crdName: string,
      instanceName: string,
    ) =>
      `/c/${cluster}/ns/${namespace}/operators-instances/${csvName}/${crdName}/${instanceName}/update`,
    new: (cluster: string, namespace: string, csvName: string, crdName: string) =>
      `/c/${cluster}/ns/${namespace}/operators-instances/new/${csvName}/${crdName}`,
  },
  config: {
    pkgrepositories: (cluster: string, namespace: string) =>
      `/c/${cluster}/ns/${namespace}/config/repos`,
    operators: (cluster: string, namespace: string) => `/c/${cluster}/ns/${namespace}/operators`,
  },
};

function withNS(namespace: string) {
  return namespace ? `namespaces/${namespace}/` : "";
}

export const api = {
  // URLs which are accessing the k8s API server directly are grouped together
  // so we can clearly differentiate and possibly begin to remove.
  // Note that this list is not yet exhaustive (search for APIBase to find other call-sites which
  // access the k8s api server directly).
  k8s: {
    base: (cluster: string) => `api/clusters/${cluster}`,
    apis: (cluster: string) => `${api.k8s.base(cluster)}/apis`,
    v1: (cluster: string) => `${api.k8s.base(cluster)}/api/v1`,
    groupVersion: (cluster: string, groupVersion: string) =>
      `${api.k8s.base(cluster)}/apis/${groupVersion}`,
    operators: {
      operators: (cluster: string, namespace: string) =>
        `${api.k8s.base(cluster)}/apis/packages.operators.coreos.com/v1/${withNS(
          namespace,
        )}packagemanifests`,
      operator: (cluster: string, namespace: string, name: string) =>
        `${api.k8s.base(
          cluster,
        )}/apis/packages.operators.coreos.com/v1/namespaces/${namespace}/packagemanifests/${name}`,
      clusterServiceVersions: (cluster: string, namespace: string) =>
        `${api.k8s.base(cluster)}/apis/operators.coreos.com/v1alpha1/${withNS(
          namespace,
        )}clusterserviceversions`,
      clusterServiceVersion: (cluster: string, namespace: string, name: string) =>
        `${api.k8s.base(
          cluster,
        )}/apis/operators.coreos.com/v1alpha1/namespaces/${namespace}/clusterserviceversions/${name}`,
      resources: (cluster: string, namespace: string, apiVersion: string, resource: string) =>
        `${api.k8s.base(cluster)}/apis/${apiVersion}/${withNS(namespace)}${resource}`,
      resource: (
        cluster: string,
        namespace: string,
        apiVersion: string,
        resource: string,
        name: string,
      ) =>
        `${api.k8s.base(cluster)}/apis/${apiVersion}/namespaces/${namespace}/${resource}/${name}`,
      operatorGroups: (cluster: string, namespace: string) =>
        `${api.k8s.base(
          cluster,
        )}/apis/operators.coreos.com/v1/namespaces/${namespace}/operatorgroups`,
      subscriptions: (cluster: string, namespace: string) =>
        `${api.k8s.base(
          cluster,
        )}/apis/operators.coreos.com/v1alpha1/namespaces/${namespace}/subscriptions`,
      subscription: (cluster: string, namespace: string, name: string) =>
        `${api.k8s.base(
          cluster,
        )}/apis/operators.coreos.com/v1alpha1/namespaces/${namespace}/subscriptions/${name}`,
    },
  },

  operators: {
    operatorIcon: (_cluster: string, namespace: string, name: string) =>
      `apis/operators/namespaces/${namespace}/operator/${name}/logo`,
  },

  kubeappsapis: "apis",
  config: "config.json",
  custom_locale: "custom_locale.json",
};
