import * as React from "react";

import { IClusterServiceClass } from "../../shared/ClusterServiceClass";
import { IServiceInstance } from "../../shared/ServiceInstance";
import { Card, CardContainer } from "../Card";

export interface InstanceCardListProps {
  classes: IClusterServiceClass[];
  instances: IServiceInstance[];
}

export const InstanceCardList = (props: InstanceCardListProps) => {
  const { instances, classes } = props;
  return (
    <div className="InstanceCardList">
      <section>
        <CardContainer>
          {instances.length > 0 &&
            instances.map(instance => {
              const conditions = [...instance.status.conditions];
              const status = conditions.shift(); // first in list is most recent
              const message = status ? status.message : "";
              const svcClass = classes.find(
                potential => potential.metadata.name === instance.spec.clusterServiceClassRef.name,
              );
              const broker = svcClass && svcClass.spec.clusterServiceBrokerName;
              const icon =
                svcClass &&
                svcClass.spec.externalMetadata &&
                svcClass.spec.externalMetadata.imageUrl;

              const card = (
                <Card
                  key={instance.metadata.uid}
                  header={
                    <span>
                      {instance.metadata.namespace}/{instance.metadata.name}
                    </span>
                  }
                  icon={icon}
                  body={message}
                  buttonText="Details"
                  linkTo={`/services/brokers/${broker}/instances/${instance.metadata.namespace}/${
                    instance.metadata.name
                  }/`}
                  notes={<span>{instance.spec.clusterServicePlanExternalName}</span>}
                />
              );
              return card;
            })}
        </CardContainer>
      </section>
    </div>
  );
};
