import * as React from "react";

import Check from "../../icons/Check";
import Compass from "../../icons/Compass";
import { IDeploymentStatus, IResource } from "../../shared/types";
import "./AppStatus.css";

interface IAppStatusProps {
  deployments: Map<string, IResource>;
}

class AppStatus extends React.Component<IAppStatusProps> {
  public render() {
    return this.isReady() ? this.renderSuccessStatus() : this.renderPendingStatus();
  }

  private renderSuccessStatus() {
    return (
      <span className="AppStatus AppStatus--success">
        <Check className="icon padding-t-tiny" /> Deployed
      </span>
    );
  }

  private renderPendingStatus() {
    return (
      <span className="AppStatus AppStatus--pending">
        <Compass className="icon padding-t-tiny" /> Deploying
      </span>
    );
  }

  private isReady() {
    const { deployments } = this.props;
    if (Object.keys(deployments).length > 0) {
      return Object.keys(deployments).every(k => {
        const dStatus: IDeploymentStatus = deployments[k].status;
        return dStatus.availableReplicas === dStatus.replicas;
      });
    } else {
      // if there are no deployments, then the app is considered "ready"
      // TODO: this currently does not distinguish between deployments not
      // loaded yet and no deployments
      return true;
    }
  }
}

export default AppStatus;
