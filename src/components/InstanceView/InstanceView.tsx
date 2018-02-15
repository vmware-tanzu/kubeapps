import * as React from "react";

import { IClusterServiceClass } from "../../shared/ClusterServiceClass";
import { IServiceBinding, ServiceBinding } from "../../shared/ServiceBinding";
import { IServicePlan } from "../../shared/ServiceCatalog";
import { IServiceInstance } from "../../shared/ServiceInstance";
import { BindingList } from "../BindingList/BindingList";
import { Card, CardContainer } from "../Card";
import DeprovisionButton from "../DeprovisionButton";
import { AddBindingButton } from "./AddBindingButton";

interface IInstanceViewProps {
  instance: IServiceInstance | undefined;
  bindings: IServiceBinding[];
  svcClass: IClusterServiceClass | undefined;
  svcPlan: IServicePlan | undefined;
  getCatalog: () => Promise<any>;
  deprovision: (instance: IServiceInstance) => Promise<any>;
}

export class InstanceView extends React.Component<IInstanceViewProps> {
  public async componentDidMount() {
    await this.props.getCatalog();
  }

  public render() {
    const { instance, bindings, svcClass, svcPlan, getCatalog, deprovision } = this.props;

    let body = <span />;
    if (instance) {
      const conditions = [...instance.status.conditions];
      body = (
        <div>
          <div>
            <h3>Status</h3>
            <table>
              <thead>
                <tr>
                  <th>Type</th>
                  <th>Status</th>
                  <th>Last Transition Time</th>
                  <th>Reason</th>
                  <th>Message</th>
                </tr>
              </thead>
              <tbody>
                {conditions.length > 0 ? (
                  conditions.map(condition => {
                    return (
                      <tr key={condition.lastTransitionTime}>
                        <td>{condition.type}</td>
                        <td>{condition.status}</td>
                        <td>{condition.lastTransitionTime}</td>
                        <td>
                          <code>{condition.reason}</code>
                        </td>
                        <td>{condition.message}</td>
                      </tr>
                    );
                  })
                ) : (
                  <tr>
                    <td colSpan={5}>
                      <p>No statuses</p>
                    </td>
                  </tr>
                )}
              </tbody>
            </table>
          </div>
        </div>
      );
    }

    let classCard = <span />;
    if (svcClass) {
      const tags = svcClass.spec.tags.reduce<string>((joined, tag) => {
        return `${joined} ${tag},`;
      }, "");
      const { spec } = svcClass;
      const { externalMetadata } = spec;
      const name = externalMetadata ? externalMetadata.displayName : spec.externalName;
      const description = externalMetadata ? externalMetadata.longDescription : spec.description;
      const imageUrl = externalMetadata && externalMetadata.imageUrl;

      classCard = (
        <Card
          key={svcClass.metadata.uid}
          header={name}
          icon={imageUrl}
          body={description}
          button={<span />}
          notes={
            <span style={{ fontSize: "small" }}>
              <strong>Tags:</strong> {tags}
            </span>
          }
        />
      );
    }

    let planCard = <span />;
    if (svcPlan) {
      const { spec } = svcPlan;
      const { externalMetadata } = spec;
      const name = externalMetadata ? externalMetadata.displayName : spec.externalName;
      const description =
        externalMetadata && externalMetadata.bullets
          ? externalMetadata.bullets
          : [spec.description];
      const free = svcPlan.spec.free ? <span>Free ✓</span> : undefined;
      const bullets = (
        <div>
          <ul>{description.map(bullet => <li key={bullet}>{bullet}</li>)}</ul>
        </div>
      );

      planCard = (
        <Card
          key={svcPlan.spec.externalID}
          header={name}
          body={bullets}
          notes={free}
          button={<span />}
        />
      );
    }

    return (
      <div className="InstanceView container">
        {instance && (
          <div className="found">
            <h1>
              {instance.metadata.namespace}/{instance.metadata.name}
            </h1>
            <h2>About</h2>
            <div>{body}</div>
            <DeprovisionButton deprovision={deprovision} instance={instance} />
            <h3>Spec</h3>
            <CardContainer>
              {classCard}
              {planCard}
            </CardContainer>
            <h2>Bindings</h2>
            <AddBindingButton
              bindingName={instance.metadata.name + "-binding"}
              instanceRefName={instance.metadata.name}
              namespace={instance.metadata.namespace}
              addBinding={this.addBinding}
            />
            <br />
            <BindingList bindings={bindings} addBinding={this.addBinding} getCatalog={getCatalog} />
          </div>
        )}
      </div>
    );
  }

  private addBinding = async (bindingName: string, instanceName: string, namespace: string) => {
    const binding = await ServiceBinding.create(bindingName, instanceName, namespace);
    await this.props.getCatalog();
    return binding;
  };
}
