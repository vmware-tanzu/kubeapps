// Copyright 2020-2023 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import AlertGroup from "components/AlertGroup";
import { findOwnedKind, getIcon } from "shared/Operators";
import { IClusterServiceVersion, IResource } from "shared/types";
import { app } from "shared/url";
import { getPluginIcon } from "shared/utils";
import InfoCard from "../InfoCard/InfoCard";
interface ICustomResourceListItemProps {
  cluster: string;
  resource: IResource;
  csv: IClusterServiceVersion;
}

function CustomResourceListItem(props: ICustomResourceListItemProps) {
  const { cluster, resource, csv } = props;
  const crd = findOwnedKind(csv, resource.kind);
  if (!crd) {
    // Unexpected error, CRD should be present if resource is defined
    return (
      <AlertGroup status="danger">
        {`Unable to retrieve the CustomResourceDefinition for ${resource.kind} in ${csv.metadata.name}`}
      </AlertGroup>
    );
  }
  const icon = getIcon(csv);
  return (
    <InfoCard
      key={resource.metadata.name + "_" + resource.metadata.namespace}
      link={app.operatorInstances.view(
        cluster,
        resource.metadata.namespace,
        csv.metadata.name,
        crd.name,
        resource.metadata.name,
      )}
      title={resource.metadata.name}
      icon={icon}
      description={crd.description}
      info={
        <>
          <div>App: {resource.kind}</div>
          <div>Operator: {csv.spec.version || "-"}</div>
          <div>Namespace: {resource.metadata.namespace || "-"}</div>
        </>
      }
      bgIcon={getPluginIcon("operator")}
    />
  );
}

export default CustomResourceListItem;
