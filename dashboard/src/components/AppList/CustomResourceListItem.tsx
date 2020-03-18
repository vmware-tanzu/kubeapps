import * as React from "react";

import placeholder from "../../placeholder.png";
import { IClusterServiceVersion, IResource } from "../../shared/types";
import InfoCard from "../InfoCard";

interface ICustomResourceListItemProps {
  resource: IResource;
  csv?: IClusterServiceVersion;
}

class CustomResourceListItem extends React.Component<ICustomResourceListItemProps> {
  public render() {
    const { resource, csv } = this.props;
    const icon = csv?.spec.icon
      ? `data:${csv.spec.icon[0].mediatype};base64,${csv.spec.icon[0].base64data}`
      : placeholder;
    return (
      <InfoCard
        key={resource.metadata.name}
        link={`/operators-instances/ns/${resource.metadata.namespace}/${resource.metadata.name}`}
        title={resource.metadata.name}
        icon={icon}
        info={`${resource.kind} v${csv?.spec.version || "-"}`}
        tag1Content={resource.metadata.namespace}
        tag2Content="operator"
      />
    );
  }
}

export default CustomResourceListItem;
