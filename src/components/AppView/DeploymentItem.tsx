import * as React from "react";

import { IDeploymentStatus, IResource } from "../../shared/types";

interface IDeploymentItemProps {
  deployment: IResource;
}

class DeploymentItem extends React.Component<IDeploymentItemProps> {
  public render() {
    const { deployment } = this.props;
    const status: IDeploymentStatus = deployment.status;
    return (
      <tr>
        <td>{deployment.metadata.name}</td>
        <td>{status.replicas}</td>
        <td>{status.updatedReplicas}</td>
        <td>{status.availableReplicas || 0}</td>
      </tr>
    );
  }
}

export default DeploymentItem;
