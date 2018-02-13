import * as React from "react";
import { Link } from "react-router-dom";

import { IClusterServiceClass } from "../../shared/ClusterServiceClass";
import { IServiceBinding, ServiceBinding } from "../../shared/ServiceBinding";
import { IServiceBroker, IServicePlan } from "../../shared/ServiceCatalog";
import { IServiceInstance } from "../../shared/ServiceInstance";
import { Card, CardContainer } from "../Card";
import DeprovisionButton from "../DeprovisionButton";
import SyncButton from "../SyncButton";
import { AddBindingButton } from "./AddBindingButton";
import { RemoveBindingButton } from "./RemoveBindingButton";

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
    const { bindings, broker, instances, deprovision } = this.props;

    return (
      <div className="broker">
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
                          <DeprovisionButton deprovision={deprovision} instance={instance} />
                          <AddBindingButton
                            bindingName={instance.metadata.name + "-binding"}
                            instanceRefName={instance.metadata.name}
                            namespace={instance.metadata.namespace}
                            addBinding={this.addbinding}
                          />
                        </div>
                      </td>
                    </tr>
                  );
                })}
              </tbody>
            </table>

            <h3>Bindings</h3>
            <CardContainer>
              {bindings.length > 0 &&
                bindings.map(binding => {
                  const {
                    instanceRef,
                    secretName,
                    secretDatabase,
                    secretHost,
                    secretPassword,
                    secretPort,
                    secretUsername,
                  } = binding.spec;
                  const statuses: Array<[string, string | undefined]> = [
                    ["Instance", instanceRef.name],
                    ["Secret", secretName],
                    ["Database", secretDatabase],
                    ["Host", secretHost],
                    ["Password", secretPassword],
                    ["Port", secretPort],
                    ["Username", secretUsername],
                  ];
                  const condition = [...binding.status.conditions].shift();
                  const currentStatus = condition ? (
                    <div className="condition">
                      <div>
                        <strong>{condition.type}</strong>: <code>{condition.status}</code>
                      </div>
                      <code>{condition.message}</code>
                    </div>
                  ) : (
                    undefined
                  );

                  const body = (
                    <div style={{ display: "flex", flexWrap: "wrap", flexDirection: "column" }}>
                      {currentStatus}
                      {statuses.map(statusPair => {
                        const [key, value] = statusPair;
                        return (
                          <div key={key} style={{ display: "flex" }}>
                            <strong key={key} style={{ flex: "0 0 5em" }}>
                              {key}:
                            </strong>
                            <code
                              key={value || "null"}
                              style={{ flex: "1 1", wordBreak: "break-all" }}
                            >
                              {value}
                            </code>
                          </div>
                        );
                      })}
                    </div>
                  );
                  const card = (
                    <Card
                      key={binding.metadata.name}
                      header={binding.metadata.name}
                      body={body}
                      button={
                        <RemoveBindingButton
                          binding={binding}
                          onRemoveComplete={this.props.getCatalog}
                        />
                      }
                    />
                  );
                  return card;
                })}
            </CardContainer>
          </div>
        )}
      </div>
    );
  }

  private addbinding = async (bindingName: string, instanceName: string, namespace: string) => {
    const binding = await ServiceBinding.create(bindingName, instanceName, namespace);
    await this.props.getCatalog();
    return binding;
  };
}
