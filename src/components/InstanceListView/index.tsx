import * as React from "react";

import { Link } from "react-router-dom";
import { IClusterServiceClass } from "../../shared/ClusterServiceClass";
import { IServiceBroker, IServicePlan } from "../../shared/ServiceCatalog";
import { IServiceInstance } from "../../shared/ServiceInstance";
import { InstanceCardList } from "./InstanceCardList";

export interface InstanceListViewProps {
  brokers: IServiceBroker[];
  classes: IClusterServiceClass[];
  getCatalog: () => Promise<any>;
  checkCatalogInstalled: () => Promise<any>;
  instances: IServiceInstance[];
  plans: IServicePlan[];
  isInstalled: boolean;
}

export class InstanceListView extends React.PureComponent<InstanceListViewProps> {
  public async componentDidMount() {
    this.props.checkCatalogInstalled();
    this.props.getCatalog();
  }

  public render() {
    const { isInstalled, brokers, instances, classes } = this.props;

    return (
      <div className="InstanceList">
        <h3>Service Instances</h3>

        {isInstalled ? (
          <div>
            {brokers.length > 0 ? (
              <div>
                <div className="row">
                  <div className="col-8">
                    <p>Service instances from your brokers:</p>
                  </div>
                  <div className="col-4 text-r">
                    <Link to={`/services/classes`}>
                      <button className="button button-accent">Provision New Service</button>
                    </Link>
                  </div>
                </div>
                {instances.length > 0 ? (
                  <InstanceCardList instances={instances} classes={classes} />
                ) : (
                  <div>No service instances are found.</div>
                )}
              </div>
            ) : (
              <div>
                No service brokers are installed in your cluster. Please ask an administrator to
                install it.
              </div>
            )}
          </div>
        ) : (
          <div>
            <div>No Service Catalog is installed.</div>
            <Link className="button button-primary" to={`/charts/svc-cat/catalog`}>
              Install Catalog
            </Link>
          </div>
        )}
      </div>
    );
  }
}
