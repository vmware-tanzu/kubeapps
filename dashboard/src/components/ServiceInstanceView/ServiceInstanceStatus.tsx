import * as React from "react";

import { AlertTriangle } from "react-feather";
import { IServiceInstance } from "shared/ServiceInstance";
import "../../components/ApplicationStatus/ApplicationStatus.css";
import Check from "../../icons/Check";
import Compass from "../../icons/Compass";

interface IServiceInstanceStatusProps {
  instance: IServiceInstance;
}

const checkIcon = <Check className="icon padding-t-tiny" />;
const compassIcon = <Compass className="icon padding-t-tiny" />;
// TODO(prydonius): move style to CSS once we switch all icons to feather icons
const alertIcon = <AlertTriangle className="icon" style={{ bottom: "-0.3em" }} />;

const ServiceInstanceStatus: React.SFC<IServiceInstanceStatusProps> = props => {
  const { instance } = props;
  const status = instance.status.conditions[0];
  if (status) {
    if (status.reason.match(/Provisioning/)) {
      return renderStatus("Provisioning", "pending", compassIcon);
    }
    if (status.reason.match(/Provisioned|Success/)) {
      return renderStatus("Provisioned", "success", checkIcon);
    }
    if (status.reason.match(/Failed|Error/)) {
      return renderStatus("Failed", null, alertIcon);
    }
    if (status.reason.match(/Deprovisioning/)) {
      return renderStatus("Deprovisioning", "pending", compassIcon);
    }
  }
  return renderStatus("Unknown", null, alertIcon);
};

const renderStatus = (label: string, modifierClass: string | null, icon: JSX.Element) => (
  <span
    className={`ApplicationStatus${modifierClass ? ` ApplicationStatus--${modifierClass}` : ""}`}
  >
    {icon} {label}
  </span>
);

export default ServiceInstanceStatus;
