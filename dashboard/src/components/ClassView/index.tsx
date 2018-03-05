import * as React from "react";
import { RouterAction } from "react-router-redux";

import { IClusterServiceClass } from "../../shared/ClusterServiceClass";
import { IServicePlan } from "../../shared/ServiceCatalog";
import ProvisionButton from "../ProvisionButton";

interface IClassViewProps {
  classes: IClusterServiceClass[];
  classname: string;
  getCatalog: () => Promise<any>;
  plans: IServicePlan[];
  provision: (
    instanceName: string,
    namespace: string,
    className: string,
    planName: string,
    parameters: {},
  ) => Promise<any>;
  push: (location: string) => RouterAction;
  svcClass: IClusterServiceClass | undefined;
}

export class ClassView extends React.Component<IClassViewProps> {
  public componentDidMount() {
    this.props.getCatalog();
  }

  public render() {
    const { classes, classname, plans, provision, push, svcClass } = this.props;
    const classPlans = svcClass
      ? plans.filter(plan => plan.spec.clusterServiceClassRef.name === svcClass.metadata.name)
      : [];

    return (
      <div className="class-view">
        <h1>Plans: {classname}</h1>
        <p>Service Plans available for provisioning under {classname}</p>
        <table className="striped">
          <thead>
            <tr>
              <th>Name</th>
              <th>Specs</th>
              <th />
            </tr>
          </thead>
          <tbody>
            {svcClass &&
              classPlans.map(plan => {
                const serviceClass = classes.find(
                  potential => potential.metadata.name === plan.spec.clusterServiceClassRef.name,
                );
                const { spec } = plan;
                const { externalMetadata } = spec;
                const name = externalMetadata ? externalMetadata.displayName : spec.externalName;
                const description =
                  externalMetadata && externalMetadata.bullets
                    ? externalMetadata.bullets
                    : [spec.description];
                const bullets = (
                  <div>
                    <ul className="margin-reset">
                      {description.map(bullet => <li key={bullet}>{bullet}</li>)}
                    </ul>
                  </div>
                );

                return (
                  <tr key={plan.spec.externalID}>
                    <td>
                      <b>{name}</b>
                    </td>
                    <td>{bullets}</td>
                    <td className="text-r">
                      <ProvisionButton
                        selectedClass={serviceClass}
                        selectedPlan={plan}
                        plans={plans}
                        classes={classes}
                        provision={provision}
                        push={push}
                      />
                    </td>
                  </tr>
                );
              })}
          </tbody>
        </table>
      </div>
    );
  }
}
