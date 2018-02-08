import * as React from "react";
import { Link } from "react-router-dom";

import { ServiceBrokerList } from "../../components/ServiceBrokerList";
import { IServiceCatalogState } from "../../reducers/catalog";

export interface IServiceCatalogDispatch {
  checkCatalogInstalled: () => Promise<boolean>;
  getCatalog: () => Promise<any>;
}

export class ServiceCatalogView extends React.Component<
  IServiceCatalogDispatch & IServiceCatalogState
> {
  public async componentDidMount() {
    this.props.checkCatalogInstalled();
    this.props.getCatalog();
  }

  public render() {
    const { brokers, isInstalled } = this.props;

    return (
      <div className="service-list-container">
        <h1>Service Catalog</h1>
        {!isInstalled ? (
          <div>
            <p>Service Catalog not installed.</p>
            <div className="padding-normal">
              <Link className="button button-primary" to={`/charts`}>
                Install Catalog
              </Link>
            </div>
          </div>
        ) : (
          <div>
            <ServiceBrokerList brokers={brokers} />
          </div>
        )}
      </div>
    );
  }
}
