import * as React from "react";

import Check from "../../icons/Check";
import Compass from "../../icons/Compass";
import { IDeploymentStatus, IResource } from "../../shared/types";
import "./DeploymentStatus.css";

interface IDeploymentStatusProps {
  deployments: IResource[];
}

class DeploymentStatus extends React.Component<IDeploymentStatusProps> {
  public render() {
    return this.isReady() ? this.renderSuccessStatus() : this.renderPendingStatus();
  }

  private renderSuccessStatus() {
    return (
      <span className="DeploymentStatus DeploymentStatus--success">
        <Check className="icon padding-t-tiny" /> Deployed
      </span>
    );
  }

  private renderPendingStatus() {
    return (
      <span className="DeploymentStatus DeploymentStatus--pending">
        <Compass className="icon padding-t-tiny" /> Deploying
      </span>
    );
  }

  private isReady() {
    const { deployments } = this.props;
    if (deployments.length > 0) {
      return deployments.every(d => {
        const status: IDeploymentStatus = d.status;
        return status.availableReplicas === status.replicas;
      });
    } else {
      // if there are no deployments, then the app is considered "ready"
      // TODO: this currently does not distinguish between deployments not
      // loaded yet and no deployments
      return true;
    }
  }
}

export default DeploymentStatus;
