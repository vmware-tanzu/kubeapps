import * as React from "react";

import { IServiceBroker } from "../../../shared/ServiceCatalog";
import { Card, CardContainer } from "../../Card";
import SyncButton from "../../SyncButton";

interface IServiceBrokerListProps {
  brokers: IServiceBroker[];
  sync: (broker: IServiceBroker) => Promise<any>;
}

export const ServiceBrokerList = (props: IServiceBrokerListProps) => {
  const { brokers, sync } = props;
  return (
    <div className="service-broker-list">
      <h3>Brokers</h3>
      {brokers.length > 0 ? (
        <CardContainer>
          {brokers.map(broker => {
            const card = (
              <div>
                <Card
                  key={broker.metadata.uid}
                  header={broker.metadata.name}
                  body={broker.spec.url}
                  notes={`Last updated ${broker.status.lastCatalogRetrievalTime}`}
                  button={<SyncButton sync={sync} broker={broker} />}
                />
              </div>
            );
            return card;
          })}
        </CardContainer>
      ) : (
        <div>
          No service brokers are installed in your cluster. Please ask an administrator to install
          it.
        </div>
      )}
    </div>
  );
};
