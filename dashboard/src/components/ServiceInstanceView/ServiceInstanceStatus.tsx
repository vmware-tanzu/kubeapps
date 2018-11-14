import * as React from "react";

import { AlertTriangle } from "react-feather";
import { IServiceInstance } from "shared/ServiceInstance";
import Check from "../../icons/Check";
import Compass from "../../icons/Compass";
import "../DeploymentStatus/DeploymentStatus.css";

interface IServiceInstanceStatusProps {
  instance: IServiceInstance;
}

class ServiceInstanceStatus extends React.Component<IServiceInstanceStatusProps> {
  public render() {
    const { instance } = this.props;
    const status = instance.status.conditions[0];
    if (status) {
      if (status.reason.match(/Provisioning/)) {
        return this.renderProvisioningStatus();
      }
      if (status.reason.match(/Provisioned|Success/)) {
        return this.renderProvisionedStatus();
      }
      if (status.reason.match(/Failed|Error/)) {
        return this.renderFailedStatus();
      }
      if (status.reason.match(/Deprovisioning/)) {
        return this.renderDeprovisioningStatus();
      }
    }
    return this.renderUnknownStatus();
  }

  private renderProvisionedStatus() {
    return (
      <span className="DeploymentStatus DeploymentStatus--success">
        <Check className="icon padding-t-tiny" /> Provisioned
      </span>
    );
  }

  private renderProvisioningStatus() {
    return (
      <span className="DeploymentStatus DeploymentStatus--pending">
        <Compass className="icon padding-t-tiny" /> Provisioning
      </span>
    );
  }

  private renderDeprovisioningStatus() {
    return (
      <span className="DeploymentStatus DeploymentStatus--pending">
        <Compass className="icon padding-t-tiny" /> Deprovisioning
      </span>
    );
  }

  private renderFailedStatus() {
    return (
      <span className="DeploymentStatus">
        {/* TODO(prydonius): move style to CSS once we switch all icons to feather icons */}
        <AlertTriangle className="icon" style={{ bottom: "-0.3em" }} /> Failed
      </span>
    );
  }

  private renderUnknownStatus() {
    return (
      <span className="DeploymentStatus">
        {/* TODO(prydonius): move style to CSS once we switch all icons to feather icons */}
        <AlertTriangle className="icon" style={{ bottom: "-0.3em" }} /> Unknown
      </span>
    );
  }
}

export default ServiceInstanceStatus;
