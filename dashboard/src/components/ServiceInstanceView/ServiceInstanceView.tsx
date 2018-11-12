import * as React from "react";

import { IServiceCatalogState } from "reducers/catalog";
import { IClusterServiceClass } from "../../shared/ClusterServiceClass";
import { IServicePlan } from "../../shared/ServiceCatalog";
import { IServiceInstance } from "../../shared/ServiceInstance";
import { IRBACRole, NotFoundError } from "../../shared/types";
import BindingList from "../BindingList";
import Card, { CardContent, CardGrid, CardIcon } from "../Card";
import { ErrorSelector } from "../ErrorAlert";
import LoadingWrapper from "../LoadingWrapper";
import PageHeader from "../PageHeader";
import AddBindingButton from "./AddBindingButton";
import DeprovisionButton from "./DeprovisionButton";

interface IServiceInstanceViewProps {
  errors: {
    fetch?: Error;
    create?: Error;
    delete?: Error;
    deprovision?: Error;
  };
  instances: IServiceCatalogState["instances"];
  bindingsWithSecrets: IServiceCatalogState["bindingsWithSecrets"];
  name: string;
  namespace: string;
  classes: IServiceCatalogState["classes"];
  plans: IServiceCatalogState["plans"];
  getInstances: (ns: string) => Promise<any>;
  getClasses: () => Promise<any>;
  getPlans: () => Promise<any>;
  getBindings: (ns: string) => Promise<any>;
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
  list: [
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

class ServiceInstanceView extends React.Component<IServiceInstanceViewProps> {
  public componentDidMount() {
    this.props.getInstances(this.props.namespace);
    this.props.getBindings(this.props.namespace);
    this.props.getClasses();
    this.props.getPlans();
  }

  public componentWillReceiveProps(nextProps: IServiceInstanceViewProps) {
    const { getInstances, getBindings, namespace } = this.props;
    if (nextProps.namespace !== namespace) {
      getInstances(nextProps.namespace);
      getBindings(nextProps.namespace);
    }
  }

  public render() {
    const {
      name,
      namespace,
      instances,
      bindingsWithSecrets,
      classes,
      plans,
      deprovision,
    } = this.props;

    let body = <span />;
    let planCard = <span />;
    let classCard = <span />;
    let bindingSection = <span />;

    let instance: IServiceInstance | undefined;
    let svcPlan: IServicePlan | undefined;

    const loaded =
      !instances.isFetching &&
      !classes.isFetching &&
      !plans.isFetching &&
      !bindingsWithSecrets.isFetching;

    if (loaded && instances.list.length > 0) {
      instance = instances.list.find(
        i => i.metadata.name === name && i.metadata.namespace === namespace,
      );
      if (!instance) {
        return (
          <ErrorSelector
            error={new NotFoundError(`Instance ${name} not found in ${namespace}`)}
            resource={`Instance ${name}`}
            action="get"
            namespace={namespace}
          />
        );
      }
      body = this.renderInstance(instance);

      const svcClass =
        instance &&
        classes.list.find(
          c =>
            !!instance &&
            !!instance.spec.clusterServiceClassRef &&
            c.metadata.name === instance.spec.clusterServiceClassRef.name,
        );
      if (svcClass) {
        classCard = this.renderSVCClass(svcClass);
      }

      svcPlan =
        instance &&
        plans.list.find(
          p =>
            !!instance &&
            !!instance.spec.clusterServicePlanRef &&
            p.metadata.name === instance.spec.clusterServicePlanRef.name,
        );
      if (svcPlan) {
        planCard = this.renderSVCPlan(svcPlan);
      }

      if (svcClass && svcClass.spec.bindable) {
        const bindings =
          instance &&
          bindingsWithSecrets.list.filter(
            b =>
              b.binding.spec.instanceRef.name === name &&
              b.binding.metadata.namespace === namespace,
          );
        bindingSection = (
          <div>
            <AddBindingButton
              bindingSchema={svcPlan && svcPlan.spec.serviceBindingCreateParameterSchema}
              instanceRefName={instance.metadata.name}
              namespace={instance.metadata.namespace}
              addBinding={this.props.addBinding}
              onAddBinding={this.onAddBinding}
              error={this.props.errors.create}
            />
            <br />
            {this.props.errors.delete && (
              <ErrorSelector
                error={this.props.errors.delete}
                resource="Binding"
                action="delete"
                defaultRequiredRBACRoles={RequiredRBACRoles}
              />
            )}
            <BindingList bindingsWithSecrets={bindings} removeBinding={this.props.removeBinding} />
          </div>
        );
      } else {
        bindingSection = <p>This instance cannot be bound to applications.</p>;
      }
    }

    return (
      <div className="container">
        <PageHeader>
          <h1>{name}</h1>
        </PageHeader>
        <main>
          <LoadingWrapper loaded={loaded}>
            {this.props.errors.fetch && (
              <ErrorSelector
                error={this.props.errors.fetch}
                resource={`Instance ${name}`}
                action="list"
                defaultRequiredRBACRoles={RequiredRBACRoles}
              />
            )}
            {this.props.errors.deprovision && (
              <ErrorSelector
                error={this.props.errors.deprovision}
                resource={`Instance ${name}`}
                action="deprovision"
                defaultRequiredRBACRoles={RequiredRBACRoles}
              />
            )}
            {instance && (
              <div className="found">
                <h2>About</h2>
                <div>{body}</div>
                <DeprovisionButton deprovision={deprovision} instance={instance} />
                <h3>Spec</h3>
                <CardGrid>
                  {classCard}
                  {planCard}
                </CardGrid>
                <h2>Bindings</h2>
                {bindingSection}
              </div>
            )}
          </LoadingWrapper>
        </main>
      </div>
    );
  }

  private renderInstance(instance: IServiceInstance) {
    const conditions = [...instance.status.conditions];
    return (
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

  private renderSVCClass(svcClass: IClusterServiceClass) {
    const { spec } = svcClass;
    const { externalMetadata } = spec;
    const svcName = externalMetadata ? externalMetadata.displayName : spec.externalName;
    const description = externalMetadata ? externalMetadata.longDescription : spec.description;
    const imageUrl = externalMetadata && externalMetadata.imageUrl;

    return (
      <Card key={svcClass.metadata.uid} responsive={true} responsiveColumns={2}>
        <CardIcon icon={imageUrl} />
        <CardContent>
          <h5>{svcName}</h5>
          <p className="margin-b-reset">{description}</p>
        </CardContent>
      </Card>
    );
  }

  private renderSVCPlan(svcPlan: IServicePlan) {
    const { spec } = svcPlan;
    const { externalMetadata } = spec;
    const planName = externalMetadata ? externalMetadata.displayName : spec.externalName;
    const description =
      externalMetadata && externalMetadata.bullets ? externalMetadata.bullets : [spec.description];
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

    return (
      <Card key={svcPlan.spec.externalID} responsive={true} responsiveColumns={2}>
        <CardContent>
          <h5>{planName}</h5>
          <p className="type-small margin-reset margin-b-big type-color-light-blue">{free}</p>
          {bullets}
        </CardContent>
      </Card>
    );
  }

  private onAddBinding = () => {
    const { namespace, getBindings } = this.props;
    getBindings(namespace);
  };
}

export default ServiceInstanceView;
