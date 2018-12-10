import * as React from "react";

import { AlertTriangle } from "react-feather";
import isSomeResourceLoading from "../../components/AppView/helpers";
import Check from "../../icons/Check";
import Compass from "../../icons/Compass";
import { hapi } from "../../shared/hapi/release";
import { IDeploymentStatus, IKubeItem, IResource } from "../../shared/types";
import "./DeploymentStatus.css";

interface IDeploymentStatusProps {
  deployments: Array<IKubeItem<IResource>>;
  info?: hapi.release.IInfo;
}

class DeploymentStatus extends React.Component<IDeploymentStatusProps> {
  public render() {
    if (isSomeResourceLoading(this.props.deployments)) {
      return <span className="DeploymentStatus">Loading...</span>;
    }
    if (this.props.info && this.props.info.deleted) {
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
      // Check if all the deployments has the same number of
      // desired and available replicas.
      return deployments.every(d => {
        if (d.item) {
          const status: IDeploymentStatus = d.item.status;
          return status.availableReplicas === status.replicas;
        }
        return false;
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
