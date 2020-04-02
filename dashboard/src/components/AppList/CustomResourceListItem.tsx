import * as React from "react";

import operatorIcon from "../../icons/operator-framework.svg";
import placeholder from "../../placeholder.png";
import { IClusterServiceVersion, IResource } from "../../shared/types";
import UnexpectedErrorPage from "../ErrorAlert/UnexpectedErrorAlert";
import InfoCard from "../InfoCard";

interface ICustomResourceListItemProps {
  resource: IResource;
  csv: IClusterServiceVersion;
}

class CustomResourceListItem extends React.Component<ICustomResourceListItemProps> {
  public render() {
    const { resource, csv } = this.props;
    const icon = csv.spec.icon
      ? `data:${csv.spec.icon[0].mediatype};base64,${csv.spec.icon[0].base64data}`
      : placeholder;
    const crd = csv.spec.customresourcedefinitions.owned.find(c => c.kind === resource.kind);
    if (!crd) {
      // Unexpected error, CRD should be present if resource is defined
      return (
        <UnexpectedErrorPage
          text={`Unable to retrieve the CustomResourceDefinition for ${resource.kind} in ${csv.metadata.name}`}
        />
      );
    }
    return (
      <InfoCard
        key={resource.metadata.name}
        link={`/ns/${resource.metadata.namespace}/operators-instances/${csv.metadata.name}/${crd.name}/${resource.metadata.name}`}
        title={resource.metadata.name}
        icon={icon}
        info={`${resource.kind} v${csv.spec.version || "-"}`}
        tag1Content={resource.metadata.namespace}
        tag2Content={csv.metadata.name.split(".")[0]}
        subIcon={operatorIcon}
      />
    );
  }
}

export default CustomResourceListItem;
