import * as React from "react";
import { IClusterServiceClass } from "../../shared/ClusterServiceClass";
import { IServiceBroker, IServicePlan } from "../../shared/ServiceCatalog";

interface IServiceListProps {
  brokers: IServiceBroker[];
  classes: IClusterServiceClass[];
  plans: IServicePlan[];
  getCatalog: () => Promise<any>;
}

export class ServiceList extends React.Component<IServiceListProps> {
  public async componentDidMount() {
    this.props.getCatalog();
  }

  public render() {
    const { brokers, classes, plans } = this.props;
    return (
      <div className="service-list-container">
        <h2>Brokers</h2>
        <dl>
          {brokers.length > 0 &&
            brokers.map(broker => {
              return [
                <dt key={broker.metadata.name}>{broker.metadata.name}</dt>,
                <dd key={broker.spec.url}>{broker.spec.url}</dd>,
              ];
            })}
        </dl>
        <h2>Classes</h2>
        <dl>
          {classes.length > 0 &&
            classes.map(serviceClass => {
              return [
                <dt key={serviceClass.spec.externalName}>{serviceClass.spec.externalName}</dt>,
                <dd key={serviceClass.spec.description}>{serviceClass.spec.description}</dd>,
              ];
            })}
        </dl>
        <h2>Plans</h2>
        <dl>
          {plans.length > 0 &&
            plans.map(plan => {
              return [
                <dt key={plan.spec.externalName}>{plan.spec.externalName}</dt>,
                <dd key={plan.spec.description}>{plan.spec.description}</dd>,
              ];
            })}
        </dl>
      </div>
    );
  }
}
