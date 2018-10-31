import * as React from "react";

import { CardFooter } from "../../../components/Card";
import InfoCard from "../../../components/InfoCard";
import azureIcon from "../../../icons/azure.png";
import gcpIcon from "../../../icons/gcp.png";
import placeholder from "../../../placeholder.png";
import { IServiceBroker } from "../../../shared/ServiceCatalog";
import SyncButton from "./SyncButton";

interface IServiceBrokerItemProps {
  broker: IServiceBroker;
  sync: (broker: IServiceBroker) => Promise<any>;
}

function getIcon(brokerURL: string) {
  const parser = document.createElement("a");
  parser.href = brokerURL;
  if (parser.hostname.match(/azure/)) {
    return azureIcon;
  } else if (parser.hostname.match(/servicebroker\.googleapis/)) {
    return gcpIcon;
  }
  return placeholder;
}

const ServiceBrokerItem: React.SFC<IServiceBrokerItemProps> = props => {
  const { broker, sync } = props;
  return (
    <InfoCard
      title={broker.metadata.name}
      info={`Last updated ${broker.status.lastCatalogRetrievalTime}`}
      icon={getIcon(broker.spec.url)}
    >
      <CardFooter className="text-c">
        <SyncButton sync={sync} broker={broker} />
      </CardFooter>
    </InfoCard>
  );
};

export default ServiceBrokerItem;
