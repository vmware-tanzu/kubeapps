import * as React from "react";

import { Link } from "react-router-dom";
import { IClusterServiceClass } from "../../shared/ClusterServiceClass";
import { IServiceBroker, IServicePlan } from "../../shared/ServiceCatalog";
import { IServiceInstance } from "../../shared/ServiceInstance";
import { Card, CardContainer } from "../Card";

export interface InstanceListViewProps {
  brokers: IServiceBroker[];
  classes: IClusterServiceClass[];
  getCatalog: () => Promise<any>;
  instances: IServiceInstance[];
  plans: IServicePlan[];
}

export class InstanceListView extends React.PureComponent<InstanceListViewProps> {
  public async componentDidMount() {
    this.props.getCatalog();
  }

  public render() {
    const { brokers, instances, classes } = this.props;

    return (
      <div className="InstanceList">
        {brokers && (
          <div>
            <div className="row">
              <div className="col-8">
                <h3>Service Instances</h3>
                <p>Service instances from your brokers:</p>
              </div>
              <div className="col-2 margin-normal">
                <Link to={`/services/classes`}>
                  <button className="button button-primary">Provision New Service</button>
                </Link>
              </div>
            </div>
            <table>
              <tbody>
                <CardContainer>
                  {instances.length > 0 &&
                    instances.map(instance => {
                      const conditions = [...instance.status.conditions];
                      const status = conditions.shift(); // first in list is most recent
                      const message = status ? status.message : "";
                      const svcClass = classes.find(
                        potential =>
                          potential.metadata.name === instance.spec.clusterServiceClassRef.name,
                      );
                      const broker = svcClass && svcClass.spec.clusterServiceBrokerName;
                      const icon =
                        svcClass &&
                        svcClass.spec.externalMetadata &&
                        svcClass.spec.externalMetadata.imageUrl;

                      const card = (
                        <Card
                          key={instance.metadata.uid}
                          header={
                            <span>
                              {instance.metadata.namespace}/{instance.metadata.name}
                            </span>
                          }
                          icon={icon}
                          body={message}
                          buttonText="Details"
                          linkTo={`/services/brokers/${broker}/instances/${
                            instance.metadata.namespace
                          }/${instance.metadata.name}/`}
                          notes={<span>{instance.spec.clusterServicePlanExternalName}</span>}
                        />
                      );
                      return card;
                    })}
                </CardContainer>
              </tbody>
            </table>
          </div>
        )}
      </div>
    );
  }
}
