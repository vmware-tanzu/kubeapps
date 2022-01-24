// Copyright 2020-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import Tabs from "components/Tabs";
import ResourceTable from "components/AppView/ResourceTable";
import { ResourceRef } from "gen/kubeappsapis/core/packages/v1alpha1/packages";

interface IAppViewResourceRefs {
  deployments: ResourceRef[];
  statefulsets: ResourceRef[];
  daemonsets: ResourceRef[];
  services: ResourceRef[];
  secrets: ResourceRef[];
  otherResources: ResourceRef[];
}

export default function ResourceTabs({
  deployments,
  statefulsets,
  daemonsets,
  secrets,
  services,
  otherResources,
}: IAppViewResourceRefs) {
  const columns = [];
  const data = [];
  if (deployments.length) {
    columns.push("Deployments");
    data.push(<ResourceTable resourceRefs={deployments} key="deployments" id="deployments" />);
  }
  if (statefulsets.length) {
    columns.push("StatefulSets");
    data.push(<ResourceTable resourceRefs={statefulsets} key="statefulsets" id="statefulsets" />);
  }
  if (daemonsets.length) {
    columns.push("DaemonSets");
    data.push(<ResourceTable resourceRefs={daemonsets} key="daemonsets" id="daemonsets" />);
  }
  if (secrets.length) {
    columns.push("Secrets");
    data.push(<ResourceTable resourceRefs={secrets} key="secrets" id="secrets" />);
  }
  if (services.length) {
    columns.push("Services");
    data.push(<ResourceTable resourceRefs={services} key="services" id="services" />);
  }
  if (otherResources.length) {
    columns.push("Other Resources");
    data.push(
      <ResourceTable resourceRefs={otherResources} key="otherResources" id="otherResources" />,
    );
  }
  return (
    <section aria-labelledby="resources-table">
      <h3 className="section-title" id="resources-table">
        Application Resources
      </h3>
      <Tabs id="resource-table-tabs" columns={columns} data={data} />
    </section>
  );
}
