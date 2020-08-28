import * as React from "react";

import { Link } from "react-router-dom";
import * as url from "shared/url";
import LoadingWrapper from "../../../components/LoadingWrapper";
import PageHeader from "../../../components/PageHeader";
import { IServiceCatalogState } from "../../../reducers/catalog";
import { IServiceBroker } from "../../../shared/ServiceCatalog";
import { IRBACRole } from "../../../shared/types";
import { CardGrid } from "../../Card";
import {
  ErrorSelector,
  MessageAlert,
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
  cluster: string;
  kubeappsCluster: string;
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
    const { brokers, errors, sync } = this.props;
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
        <LoadingWrapper loaded={!brokers.isFetching}>{this.renderBody(body)}</LoadingWrapper>
      </section>
    );
  }

  private renderBody(body: React.ReactFragment) {
    const { cluster, isInstalled, kubeappsCluster } = this.props;
    if (cluster !== kubeappsCluster) {
      if (kubeappsCluster) {
        return (
          <MessageAlert header="Service brokers can be created on the cluster on which Kubeapps is installed only">
            <div>
              <p className="margin-v-normal">
                Kubeapps' Service Broker support enables the addition of{" "}
                <Link to={url.app.config.brokers(kubeappsCluster)}>
                  service brokers on the cluster on which Kubeapps is installed only
                </Link>
                .
              </p>
            </div>
          </MessageAlert>
        );
      } else {
        return (
          <MessageAlert header="Service brokers are not supported on this installation">
            <div>
              <p className="margin-v-normal">
                Kubeapps' Service Broker support enables the addition of service brokers on the
                cluster on which Kubeapps is installed only. This installation of Kubeapps is
                configured without access to that cluster.
              </p>
            </div>
          </MessageAlert>
        );
      }
    }

    if (!isInstalled) {
      return <ServiceCatalogNotInstalledAlert />;
    }

    return <main>{body}</main>;
  }
}

export default ServiceBrokerList;
