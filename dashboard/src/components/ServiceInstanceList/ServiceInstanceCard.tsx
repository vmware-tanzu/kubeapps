import * as React from "react";

import { PreformattedCard } from "../Card";
import "../ChartList/ChartListItem.css";
import "./ServiceInstanceCard.css";

export interface IServiceInstanceCardProps {
  name: string;
  namespace: string;
  servicePlanName: string;
  statusReason: string | undefined;
  link?: string;
  icon?: string;
}

function generalizeStatus(status: string) {
  if (status.match(/Provisioned|Success/)) {
    return "Provisioned";
  }
  if (status.match(/Failed|Error/)) {
    return "Failed";
  }
  return status;
}

const ServiceInstanceCard: React.SFC<IServiceInstanceCardProps> = props => {
  const { name, namespace, link, icon, servicePlanName, statusReason } = props;
  return (
    <PreformattedCard
      title={name}
      link={link}
      icon={icon}
      info={servicePlanName}
      tag1Content={namespace}
      tag2Class={statusReason && generalizeStatus(statusReason)}
      tag2Content={statusReason && generalizeStatus(statusReason)}
    />
  );
};

export default ServiceInstanceCard;
