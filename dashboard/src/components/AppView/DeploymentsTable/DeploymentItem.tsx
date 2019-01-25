import * as React from "react";
import { AlertTriangle } from "react-feather";

import LoadingWrapper, { LoaderType } from "../../../components/LoadingWrapper";
import { IDeploymentStatus, IKubeItem, IResource } from "../../../shared/types";

interface IDeploymentItemProps {
  name: string;
  deployment?: IKubeItem<IResource>;
  getDeployment: () => void;
}

class DeploymentItem extends React.Component<IDeploymentItemProps> {
  public componentDidMount() {
    this.props.getDeployment();
  }

  public render() {
    const { name, deployment } = this.props;
    return (
      <tr>
        <td>{name}</td>
        {this.renderDeploymentInfo(deployment)}
      </tr>
    );
  }

  private renderDeploymentInfo(deployment?: IKubeItem<IResource>) {
    if (deployment === undefined || deployment.isFetching) {
      return (
        <td colSpan={3}>
          <LoadingWrapper type={LoaderType.Placeholder} />
        </td>
      );
    }
    if (deployment.error) {
      return (
        <td colSpan={3}>
          <span className="flex">
            <AlertTriangle />
            <span className="flex margin-l-normal">Error: {deployment.error.message}</span>
          </span>
        </td>
      );
    }
    if (deployment.item) {
      const status: IDeploymentStatus = deployment.item.status;
      return (
        <React.Fragment>
          <td>{status.replicas}</td>
          <td>{status.updatedReplicas}</td>
          <td>{status.availableReplicas || 0}</td>
        </React.Fragment>
      );
    }
    return null;
  }
}

export default DeploymentItem;
