import * as React from "react";
import { Link } from "react-router-dom";

import { IClusterServiceClass } from "../../shared/ClusterServiceClass";
import { IServiceBinding } from "../../shared/ServiceBinding";
import { IServiceBroker, IServicePlan } from "../../shared/ServiceCatalog";
import { IServiceInstance } from "../../shared/ServiceInstance";
import SyncButton from "../SyncButton";

export interface IBrokerViewProps {
  bindings: IServiceBinding[];
  broker: IServiceBroker | undefined;
  classes: IClusterServiceClass[];
  getCatalog: () => Promise<any>;
  instances: IServiceInstance[];
  plans: IServicePlan[];
  sync: (broker: IServiceBroker) => Promise<any>;
  deprovision: (instance: IServiceInstance) => Promise<any>;
}

export class BrokerView extends React.PureComponent<IBrokerViewProps> {
  public async componentDidMount() {
    this.props.getCatalog();
  }

  public render() {
    const { broker, instances } = this.props;

    return (
      <div className="BrokerView container">
        {broker && (
          <div>
            <h1>{broker.metadata.name}</h1>
            <div>Catalog last updated at {broker.status.lastCatalogRetrievalTime}</div>
            <div
              style={{
                display: "flex",
                flexDirection: "row",
              }}
            >
              <Link to={window.location.pathname + "/classes"}>
                <button className="button button-primary">Provision New Service</button>
              </Link>
              <SyncButton sync={this.props.sync} broker={broker} />
            </div>
            <h3>Service Instances</h3>
            <p>Most recent statuses for your brokers:</p>
            <table>
              <thead>
                <tr>
                  <th>Instance</th>
                  <th>Status</th>
                  <th>Message</th>
                  <th />
                </tr>
              </thead>
              <tbody>
                {instances.map(instance => {
                  const conditions = [...instance.status.conditions];
                  const status = conditions.shift(); // first in list is most recent
                  const reason = status ? status.reason : "";
                  const message = status ? status.message : "";
                  const { name, namespace } = instance.metadata;

                  return (
                    <tr key={instance.metadata.uid}>
                      <td key={instance.metadata.name}>
                        <strong>
                          {instance.metadata.namespace}/{instance.metadata.name}
                        </strong>
                      </td>
                      <td key={reason}>
                        <code>{reason}</code>
                      </td>
                      <td key={message}>
                        <code>{message}</code>
                      </td>
                      <td>
                        <div className="button-list" style={{ display: "flex" }}>
                          <Link to={location.pathname + `/instances/${namespace}/${name}`}>
                            <button className="button button-primary button-small">View</button>
                          </Link>
                        </div>
                      </td>
                    </tr>
                  );
                })}
              </tbody>
            </table>
          </div>
        )}
      </div>
    );
  }
}
