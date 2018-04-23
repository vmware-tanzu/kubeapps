import * as React from "react";
import { Link } from "react-router-dom";

import { IServiceBroker } from "../../../shared/ServiceCatalog";
import ServiceBrokerList from "../ServiceBrokerList";

export interface IServiceCatalogProps {
  brokers: IServiceBroker[];
  checkCatalogInstalled: () => Promise<any>;
  isInstalled: boolean;
  getBrokers: () => Promise<any>;
  sync: (broker: IServiceBroker) => Promise<any>;
}

export class ServiceCatalogView extends React.Component<IServiceCatalogProps> {
  public async componentDidMount() {
    this.props.checkCatalogInstalled();
    this.props.getBrokers();
  }

  public render() {
    const { brokers, isInstalled, sync } = this.props;

    return (
      <div className="service-list-container">
        {!isInstalled ? (
          <div>
            <p>Service Catalog not installed.</p>
            <div className="padding-normal">
              <Link className="button button-primary" to={`/charts/svc-cat/catalog`}>
                Install Catalog
              </Link>
            </div>
          </div>
        ) : (
          <div>
            <ServiceBrokerList brokers={brokers} sync={sync} />
          </div>
        )}
      </div>
    );
  }
}
