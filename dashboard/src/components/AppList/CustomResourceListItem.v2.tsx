import * as React from "react";

import { app } from "shared/url";
import operatorIcon from "../../icons/operator-framework.svg";
import { findOwnedKind, getIcon } from "../../shared/Operators";
import { IClusterServiceVersion, IResource } from "../../shared/types";
import InfoCard from "../InfoCard/InfoCard.v2";
import Alert from "../js/Alert";

interface ICustomResourceListItemProps {
  resource: IResource;
  csv: IClusterServiceVersion;
}

function CustomResourceListItem(props: ICustomResourceListItemProps) {
  const { resource, csv } = props;
  const crd = findOwnedKind(csv, resource.kind);
  if (!crd) {
    // Unexpected error, CRD should be present if resource is defined
    return (
      <Alert theme="danger">
        {`Unable to retrieve the CustomResourceDefinition for ${resource.kind} in ${csv.metadata.name}`}
      </Alert>
    );
  }
  const icon = getIcon(csv);
  return (
    <InfoCard
      key={resource.metadata.name}
      link={app.operatorInstances.view(
        resource.metadata.namespace,
        csv.metadata.name,
        crd.name,
        resource.metadata.name,
      )}
      title={resource.metadata.name}
      icon={icon}
      description={crd.description}
      info={`${resource.kind} v${csv.spec.version || "-"}`}
      tag1Content={csv.metadata.name.split(".")[0]}
      subIcon={operatorIcon}
    />
  );
}

export default CustomResourceListItem;
