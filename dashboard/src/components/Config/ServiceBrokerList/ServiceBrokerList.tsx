import * as React from "react";

import LoadingWrapper from "../../../components/LoadingWrapper";
import PageHeader from "../../../components/PageHeader";
import { IServiceCatalogState } from "../../../reducers/catalog";
import { IServiceBroker } from "../../../shared/ServiceCatalog";
import { IRBACRole } from "../../../shared/types";
import { CardGrid } from "../../Card";
import { ErrorSelector, ServiceBrokersNotFoundAlert } from "../../ErrorAlert";
import ServiceBrokerItem from "./ServiceBrokerItem";

interface IServiceBrokerListProps {
  errors: {
    fetch?: Error;
    update?: Error;
  };
  brokers: IServiceCatalogState["brokers"];
  sync: (broker: IServiceBroker) => Promise<any>;
}

export const RequiredRBACRoles: { [s: string]: IRBACRole[] } = {
  resync: [
    {
      apiGroup: "servicecatalog.k8s.io",
      clusterWide: true,
      resource: "clusterservicebrokers",
      verbs: ["patch"],
    },
  ],
  view: [
    {
      apiGroup: "servicecatalog.k8s.io",
      clusterWide: true,
      resource: "clusterservicebrokers",
      verbs: ["list"],
    },
  ],
};

const ServiceBrokerList: React.SFC<IServiceBrokerListProps> = props => {
  const { brokers, errors, sync } = props;
  let body = <span />;
  if (errors.fetch) {
    body = (
      <ErrorSelector
        error={errors.fetch}
        resource="Service Brokers"
        action="view"
        defaultRequiredRBACRoles={RequiredRBACRoles}
      />
    );
  } else {
    if (brokers.list.length > 0) {
      if (errors.update) {
        body = (
          <ErrorSelector
            error={errors.update}
            resource="Service Brokers"
            action="resync"
            defaultRequiredRBACRoles={RequiredRBACRoles}
          />
        );
      } else {
        body = (
          <CardGrid className="BrokerList">
            {brokers.list.map(broker => (
              <ServiceBrokerItem key={broker.metadata.uid} broker={broker} sync={sync} />
            ))}
          </CardGrid>
        );
      }
    } else {
      body = <ServiceBrokersNotFoundAlert />;
    }
  }
  return (
    <section className="AppList">
      <PageHeader>
        <h1>Service Brokers</h1>
      </PageHeader>
      <LoadingWrapper loaded={!brokers.isFetching}>
        <main>{body}</main>
      </LoadingWrapper>
    </section>
  );
};

export default ServiceBrokerList;
