import * as React from "react";

import { IResource } from "../../shared/types";

interface IDeploymentItemProps {
  deployment: IResource;
}

class DeploymentItem extends React.Component<IDeploymentItemProps> {
  public render() {
    const { deployment } = this.props;
    return (
      <tr>
        <td>{deployment.metadata.name}</td>
        <td>{deployment.status.replicas}</td>
        <td>{deployment.status.updatedReplicas}</td>
        <td>{deployment.status.availableReplicas || 0}</td>
      </tr>
    );
  }
}

export default DeploymentItem;
