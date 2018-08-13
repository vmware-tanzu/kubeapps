import * as React from "react";

import { AlertTriangle } from "react-feather";
import Check from "../../icons/Check";
import Compass from "../../icons/Compass";
import { IDeploymentStatus, IResource } from "../../shared/types";
import "./DeploymentStatus.css";

interface IDeploymentStatusProps {
  deployments: IResource[];
  deleted?: boolean;
}

class DeploymentStatus extends React.Component<IDeploymentStatusProps> {
  public render() {
    if (this.props.deleted) {
      return this.renderDeletedStatus();
    }
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

  private renderDeletedStatus() {
    return (
      <span className="DeploymentStatus DeploymentStatus--deleted">
        <AlertTriangle className="icon" style={{ bottom: "-0.425em" }} /> Deleted
      </span>
    );
  }
}

export default DeploymentStatus;
