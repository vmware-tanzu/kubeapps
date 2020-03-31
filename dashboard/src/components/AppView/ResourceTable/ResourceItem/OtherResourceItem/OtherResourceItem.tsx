import * as React from "react";
import ResourceRef from "../../../../../shared/ResourceRef";
import { IResource } from "../../../../../shared/types";

export interface IOtherResourceProps {
  resource: IResource | ResourceRef;
}

export const OtherResourceColumns: React.SFC = () => {
  return (
    <React.Fragment>
      <th className="col-6">NAME</th>
      <th className="col-6">KIND</th>
    </React.Fragment>
  );
};

const OtherResourceItem: React.SFC<IOtherResourceProps> = props => {
  const { resource } = props;
  let name;
  const fullResource = resource as IResource;
  if (fullResource.metadata && fullResource.metadata.name) {
    name = fullResource.metadata.name;
  } else {
    const resourceRef = resource as ResourceRef;
    name = resourceRef.name;
  }
  return (
    <>
      <td className="col-6">{name}</td>
      <td className="col-6">{resource.kind}</td>
    </>
  );
};

export default OtherResourceItem;
