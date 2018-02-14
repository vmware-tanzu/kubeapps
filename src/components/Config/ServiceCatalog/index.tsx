import * as React from "react";
import { Link } from "react-router-dom";

import { IServiceCatalogState } from "../../../reducers/catalog";
import { ServiceBrokerList } from "../ServiceBrokerList";

export interface IServiceCatalogDispatch {
  checkCatalogInstalled: () => Promise<boolean>;
  getCatalog: () => Promise<any>;
  sync: () => Promise<any>;
}

export class ServiceCatalogView extends React.Component<
  IServiceCatalogDispatch & IServiceCatalogState
> {
  public async componentDidMount() {
    this.props.checkCatalogInstalled();
    this.props.getCatalog();
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
