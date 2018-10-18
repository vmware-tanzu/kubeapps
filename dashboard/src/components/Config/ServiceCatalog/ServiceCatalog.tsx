import * as React from "react";

import LoadingWrapper from "../../../components/LoadingWrapper";
import { IServiceCatalogState } from "../../../reducers/catalog";
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
        <LoadingWrapper loaded={!brokers.isFetching}>
          {!isInstalled ? (
            <ServiceCatalogNotInstalledAlert />
          ) : (
            <div>
              <ServiceBrokerList errors={errors} brokers={brokers} sync={sync} />
            </div>
          )}
        </LoadingWrapper>
      </div>
    );
  }
}

export default ServiceCatalog;
