import * as React from "react";

import { IServiceCatalogState } from "reducers/catalog";
import { IServiceBroker } from "../../../shared/ServiceCatalog";
import { ServiceCatalogNotInstalledAlert } from "../../ErrorAlert";
import ServiceBrokerList from "../ServiceBrokerList";

export interface IServiceCatalogProps {
  errors: {
    fetch?: Error;
    update?: Error;
  };
  brokers: IServiceCatalogState["brokers"];
  checkCatalogInstalled: () => Promise<any>;
  isInstalled: boolean;
  getBrokers: () => Promise<any>;
  sync: (broker: IServiceBroker) => Promise<any>;
}

class ServiceCatalog extends React.Component<IServiceCatalogProps> {
  public async componentDidMount() {
    this.props.checkCatalogInstalled();
    this.props.getBrokers();
  }

  public render() {
    const { brokers, errors, isInstalled, sync } = this.props;

    return (
      <div className="service-list-container">
        {!isInstalled ? (
          <ServiceCatalogNotInstalledAlert />
        ) : (
          <div>
            {/* TODO: Check if isFetching */}
            <ServiceBrokerList errors={errors} brokers={brokers.list} sync={sync} />
          </div>
        )}
      </div>
    );
  }
}

export default ServiceCatalog;
