import * as React from "react";

import { IClusterServiceClass } from "../../shared/ClusterServiceClass";
import { IServiceBindingWithSecret } from "../../shared/ServiceBinding";
import { IServicePlan } from "../../shared/ServiceCatalog";
import { IServiceInstance } from "../../shared/ServiceInstance";
import { ForbiddenError, IRBACRole, NotFoundError } from "../../shared/types";
import BindingList from "../BindingList";
import Card, { CardContent, CardGrid, CardIcon } from "../Card";
import { NotFoundErrorAlert, PermissionsErrorAlert, UnexpectedErrorAlert } from "../ErrorAlert";
import AddBindingButton from "./AddBindingButton";
import DeprovisionButton from "./DeprovisionButton";

interface IInstanceViewProps {
  errors: {
    fetch?: Error;
    create?: Error;
    delete?: Error;
    deprovision?: Error;
  };
  instance: IServiceInstance | undefined;
  bindingsWithSecrets: IServiceBindingWithSecret[];
  name: string;
  namespace: string;
  svcClass: IClusterServiceClass | undefined;
  svcPlan: IServicePlan | undefined;
  getCatalog: (ns: string) => Promise<any>;
  deprovision: (instance: IServiceInstance) => Promise<boolean>;
  addBinding: (
    bindingName: string,
    instanceName: string,
    namespace: string,
    parameters: {},
  ) => Promise<boolean>;
  removeBinding: (name: string, ns: string) => Promise<boolean>;
}

const RequiredRBACRoles: { [s: string]: IRBACRole[] } = {
  delete: [
    {
      apiGroup: "servicecatalog.k8s.io",
      resource: "servicebindings",
      verbs: ["delete"],
    },
  ],
  deprovision: [
    {
      apiGroup: "servicecatalog.k8s.io",
      resource: "serviceinstances",
      verbs: ["delete"],
    },
  ],
  view: [
    // TODO: cleanup non-required roles
    {
      apiGroup: "servicecatalog.k8s.io",
      clusterWide: true,
      resource: "clusterservicebrokers",
      verbs: ["list"],
    },
    {
      apiGroup: "servicecatalog.k8s.io",
      clusterWide: true,
      resource: "clusterserviceclasses",
      verbs: ["list"],
    },
    {
      apiGroup: "servicecatalog.k8s.io",
      resource: "serviceinstances",
      verbs: ["list"],
    },
    {
      apiGroup: "servicecatalog.k8s.io",
      resource: "servicebindings",
      verbs: ["list"],
    },
    {
      apiGroup: "servicecatalog.k8s.io",
      clusterWide: true,
      resource: "clusterserviceplans",
      verbs: ["list"],
    },
  ],
};

class InstanceView extends React.Component<IInstanceViewProps> {
  public componentDidMount() {
    this.props.getCatalog(this.props.namespace);
  }

  public componentWillReceiveProps(nextProps: IInstanceViewProps) {
    const { getCatalog, namespace } = this.props;
    if (nextProps.namespace !== namespace) {
      getCatalog(nextProps.namespace);
    }
  }

  public render() {
    const { instance, bindingsWithSecrets, svcClass, svcPlan, deprovision } = this.props;

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
      const { spec } = svcClass;
      const { externalMetadata } = spec;
      const name = externalMetadata ? externalMetadata.displayName : spec.externalName;
      const description = externalMetadata ? externalMetadata.longDescription : spec.description;
      const imageUrl = externalMetadata && externalMetadata.imageUrl;

      classCard = (
        <Card key={svcClass.metadata.uid} responsive={true} responsiveColumns={2}>
          <CardIcon icon={imageUrl} />
          <CardContent>
            <h5>{name}</h5>
            <p className="margin-b-reset">{description}</p>
          </CardContent>
        </Card>
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
          <ul>
            {description.map(bullet => (
              <li key={bullet}>{bullet}</li>
            ))}
          </ul>
        </div>
      );

      planCard = (
        <Card key={svcPlan.spec.externalID} responsive={true} responsiveColumns={2}>
          <CardContent>
            <h5>{name}</h5>
            <p className="type-small margin-reset margin-b-big type-color-light-blue">{free}</p>
            {bullets}
          </CardContent>
        </Card>
      );
    }

    return (
      <div className="InstanceView container">
        {this.props.errors.fetch && this.renderError(this.props.errors.fetch)}
        {this.props.errors.deprovision &&
          this.renderError(this.props.errors.deprovision, "deprovision")}
        {instance && (
          <div className="found">
            <h1>
              {instance.metadata.namespace}/{instance.metadata.name}
            </h1>
            <h2>About</h2>
            <div>{body}</div>
            <DeprovisionButton deprovision={deprovision} instance={instance} />
            <h3>Spec</h3>
            <CardGrid>
              {classCard}
              {planCard}
            </CardGrid>
            <h2>Bindings</h2>
            <AddBindingButton
              bindingSchema={svcPlan && svcPlan.spec.serviceBindingCreateParameterSchema}
              instanceRefName={instance.metadata.name}
              namespace={instance.metadata.namespace}
              addBinding={this.props.addBinding}
              onAddBinding={this.onAddBinding}
              error={this.props.errors.create}
            />
            <br />
            {this.props.errors.delete &&
              this.renderError(this.props.errors.delete, "delete", "Binding")}
            <BindingList
              bindingsWithSecrets={bindingsWithSecrets}
              removeBinding={this.props.removeBinding}
            />
          </div>
        )}
      </div>
    );
  }

  private renderError(error: Error, action: string = "view", resource: string = "Instance") {
    const { namespace, name } = this.props;
    switch (error.constructor) {
      case ForbiddenError:
        return (
          <PermissionsErrorAlert
            namespace={namespace}
            roles={RequiredRBACRoles[action]}
            action={`${action} Service ${resource} "${name}"`}
          />
        );
      case NotFoundError:
        return (
          <NotFoundErrorAlert resource={`Service ${resource} "${name}"`} namespace={namespace} />
        );
      default:
        return <UnexpectedErrorAlert />;
    }
  }

  private onAddBinding = () => {
    const { namespace, getCatalog } = this.props;
    getCatalog(namespace);
  };
}

export default InstanceView;
