import * as React from "react";
import ResourceRef from "shared/ResourceRef";

export interface IOtherResourceProps {
  resource: ResourceRef;
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
  return (
    <tr key={`otherResources/${resource.kind}/${resource.name}`} className="flex">
      <td className="col-6">{resource.name}</td>
      <td className="col-6">{resource.kind}</td>
    </tr>
  );
};

export default OtherResourceItem;
