import { lowerFirst } from "lodash";
import * as React from "react";

import { IClusterServiceClass } from "shared/ClusterServiceClass";
import { IServiceInstance } from "shared/ServiceInstance";
import { CardGrid } from "../Card";
import "../Catalog/CatalogItem.css";
import ServiceInstanceCard from "./ServiceInstanceCard";
import "./ServiceInstanceCard.css";

export interface IServiceInstanceCardListProps {
  classes: IClusterServiceClass[];
  instances: IServiceInstance[];
}

const ServiceInstanceCardList: React.SFC<IServiceInstanceCardListProps> = props => {
  const { instances, classes } = props;
  return (
    <div className="ServiceInstanceCardList">
      <section>
        <CardGrid>
          {instances.length > 0 &&
            instances.map(instance => {
              const conditions = [...instance.status.conditions];
              const status = conditions.shift(); // first in list is most recent
              const svcClass = classes.find(
                potential =>
                  !!instance.spec.clusterServiceClassRef &&
                  potential.metadata.name === instance.spec.clusterServiceClassRef.name,
              );
              const broker = svcClass && svcClass.spec.clusterServiceBrokerName;
              const icon =
                svcClass &&
                svcClass.spec.externalMetadata &&
                svcClass.spec.externalMetadata.imageUrl;
              const link =
                broker &&
                `/services/brokers/${broker}/instances/ns/${instance.metadata.namespace}/${
                  instance.metadata.name
                }`;

              return (
                <ServiceInstanceCard
                  key={instance.metadata.uid}
                  name={instance.metadata.name}
                  namespace={instance.metadata.namespace}
                  link={link}
                  icon={icon}
                  serviceClassName={(svcClass && svcClass.spec.externalName) || "-"}
                  servicePlanName={instance.spec.clusterServicePlanExternalName}
                  statusReason={status && lowerFirst(status.reason)}
                />
              );
            })}
        </CardGrid>
      </section>
    </div>
  );
};

export default ServiceInstanceCardList;
