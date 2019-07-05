import * as React from "react";

import LoadingWrapper from "../../../components/LoadingWrapper";
import PageHeader from "../../../components/PageHeader";
import { IServiceCatalogState } from "../../../reducers/catalog";
import { IServiceBroker } from "../../../shared/ServiceCatalog";
import { IRBACRole } from "../../../shared/types";
import { CardGrid } from "../../Card";
import {
  ErrorSelector,
  ServiceBrokersNotFoundAlert,
  ServiceCatalogNotInstalledAlert,
} from "../../ErrorAlert";
import ServiceBrokerItem from "./ServiceBrokerItem";

interface IServiceBrokerListProps {
  errors: {
    fetch?: Error;
    update?: Error;
  };
  getBrokers: () => Promise<any>;
  brokers: IServiceCatalogState["brokers"];
  sync: (broker: IServiceBroker) => Promise<any>;
  checkCatalogInstalled: () => Promise<any>;
  isInstalled: boolean;
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

class ServiceBrokerList extends React.Component<IServiceBrokerListProps> {
  public componentDidMount() {
    this.props.checkCatalogInstalled();
    this.props.getBrokers();
  }

  public render() {
    const { brokers, errors, sync, isInstalled } = this.props;
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
          {!isInstalled ? <ServiceCatalogNotInstalledAlert /> : <main>{body}</main>}
        </LoadingWrapper>
      </section>
    );
  }
}

export default ServiceBrokerList;
