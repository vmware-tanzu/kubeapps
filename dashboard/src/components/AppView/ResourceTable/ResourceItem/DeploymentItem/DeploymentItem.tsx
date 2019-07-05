import * as React from "react";

import { IDeploymentStatus, IResource } from "shared/types";

interface IDeploymentItemRow {
  resource: IResource;
}

export const DeploymentColumns: React.SFC = () => {
  return (
    <React.Fragment>
      <th className="col-6">NAME</th>
      <th className="col-2">DESIRED</th>
      <th className="col-2">UP-TO-DATE</th>
      <th className="col-2">AVAILABLE</th>
    </React.Fragment>
  );
};

const DeploymentItemRow: React.SFC<IDeploymentItemRow> = props => {
  const status = props.resource.status as IDeploymentStatus;
  return (
    <React.Fragment>
      <td className="col-6">{props.resource.metadata.name}</td>
      <td className="col-2">{status.replicas || 0}</td>
      <td className="col-2">{status.updatedReplicas || 0}</td>
      <td className="col-2">{status.availableReplicas || 0}</td>
    </React.Fragment>
  );
};

export default DeploymentItemRow;
