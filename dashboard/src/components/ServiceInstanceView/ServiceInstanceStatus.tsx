import * as React from "react";

import { AlertTriangle } from "react-feather";
import { IServiceInstance } from "shared/ServiceInstance";
import Check from "../../icons/Check";
import Compass from "../../icons/Compass";
import "../DeploymentStatus/DeploymentStatus.css";

interface IServiceInstanceStatusProps {
  instance: IServiceInstance;
}

const ServiceInstanceStatus: React.SFC<IServiceInstanceStatusProps> = props => {
  const { instance } = props;
  const status = instance.status.conditions[0];
  if (status) {
    if (status.reason.match(/Provisioning/)) {
      return renderProvisioningStatus();
    }
    if (status.reason.match(/Provisioned|Success/)) {
      return renderProvisionedStatus();
    }
    if (status.reason.match(/Failed|Error/)) {
      return renderFailedStatus();
    }
    if (status.reason.match(/Deprovisioning/)) {
      return renderDeprovisioningStatus();
    }
  }
  return renderUnknownStatus();
};

const renderProvisionedStatus = () => {
  return (
    <span className="DeploymentStatus DeploymentStatus--success">
      <Check className="icon padding-t-tiny" /> Provisioned
    </span>
  );
};

const renderProvisioningStatus = () => {
  return (
    <span className="DeploymentStatus DeploymentStatus--pending">
      <Compass className="icon padding-t-tiny" /> Provisioning
    </span>
  );
};

const renderDeprovisioningStatus = () => {
  return (
    <span className="DeploymentStatus DeploymentStatus--pending">
      <Compass className="icon padding-t-tiny" /> Deprovisioning
    </span>
  );
};

const renderFailedStatus = () => {
  return (
    <span className="DeploymentStatus">
      {/* TODO(prydonius): move style to CSS once we switch all icons to feather icons */}
      <AlertTriangle className="icon" style={{ bottom: "-0.3em" }} /> Failed
    </span>
  );
};

const renderUnknownStatus = () => {
  return (
    <span className="DeploymentStatus">
      {/* TODO(prydonius): move style to CSS once we switch all icons to feather icons */}
      <AlertTriangle className="icon" style={{ bottom: "-0.3em" }} /> Unknown
    </span>
  );
};

export default ServiceInstanceStatus;
