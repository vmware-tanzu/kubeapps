import * as React from "react";

import { IServiceCatalogState } from "reducers/catalog";
import { IClusterServiceClass } from "../../shared/ClusterServiceClass";
import { IServicePlan } from "../../shared/ServiceCatalog";
import { IServiceInstance } from "../../shared/ServiceInstance";
import { IRBACRole, NotFoundError } from "../../shared/types";
import BindingList from "../BindingList";
import { ErrorSelector, MessageAlert } from "../ErrorAlert";
import LoadingWrapper from "../LoadingWrapper";
import AddBindingButton from "./AddBindingButton";
import DeprovisionButton from "./DeprovisionButton";
import ServiceInstanceInfo from "./ServiceInstanceInfo";
import ServiceInstanceStatus from "./ServiceInstanceStatus";

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
    let bindingSection = <span />;

    let instance: IServiceInstance | undefined;
    let svcPlan: IServicePlan | undefined;
    let svcClass: IClusterServiceClass | undefined;
    let isPending: boolean = true;

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

      // Check if the instance is being provisioned or deprovisioned to disable
      // certain actions
      const status = instance.status.conditions[0];
      isPending = !!(status && status.reason && status.reason.match(/provisioning/i));

      // TODO(prydonius): We should probably show an error if the svcClass or
      // svcPlan cannot be found for some reason.
      svcClass =
        instance &&
        classes.list.find(
          c =>
            !!instance &&
            !!instance.spec.clusterServiceClassRef &&
            c.metadata.name === instance.spec.clusterServiceClassRef.name,
        );

      svcPlan =
        instance &&
        plans.list.find(
          p =>
            !!instance &&
            !!instance.spec.clusterServicePlanRef &&
            p.metadata.name === instance.spec.clusterServicePlanRef.name,
        );

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
            {this.props.errors.delete && (
              <ErrorSelector
                error={this.props.errors.delete}
                resource="Binding"
                action="delete"
                defaultRequiredRBACRoles={RequiredRBACRoles}
              />
            )}
            <BindingList bindingsWithSecrets={bindings} removeBinding={this.props.removeBinding} />
            <AddBindingButton
              disabled={isPending}
              bindingSchema={svcPlan && svcPlan.spec.serviceBindingCreateParameterSchema}
              instanceRefName={instance.metadata.name}
              namespace={instance.metadata.namespace}
              addBinding={this.props.addBinding}
              onAddBinding={this.onAddBinding}
              error={this.props.errors.create}
            />
          </div>
        );
      } else {
        bindingSection = <p>This instance cannot be bound to applications.</p>;
      }
    }

    return (
      <section className="ServiceInstanceView padding-b-big">
        <main>
          <LoadingWrapper loaded={loaded}>
            <div className="container">
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
                <div className="row collapse-b-tablet">
                  <div className="col-12">
                    <MessageAlert level="warning">
                      <div>
                        <div>Refresh the page to update the status of this Service Instance.</div>
                        Service Catalog integration is under heavy development. If you find an issue
                        please report it{" "}
                        <a target="_blank" href="https://github.com/kubeapps/kubeapps/issues">
                          {" "}
                          here
                        </a>
                        .
                      </div>
                    </MessageAlert>
                  </div>
                  <div className="col-3">
                    <ServiceInstanceInfo instance={instance} svcClass={svcClass} plan={svcPlan} />
                  </div>
                  <div className="col-9">
                    <div className="row padding-t-bigger">
                      <div className="col-4">
                        <ServiceInstanceStatus instance={instance} />
                      </div>
                      <div className="col-8 text-r">
                        <DeprovisionButton
                          deprovision={deprovision}
                          instance={instance}
                          disabled={isPending}
                        />
                      </div>
                    </div>
                    <div className="ServiceInstanceView__details">
                      <div>{body}</div>
                      <h2>Bindings</h2>
                      <hr />
                      {bindingSection}
                    </div>
                  </div>
                </div>
              )}
            </div>
          </LoadingWrapper>
        </main>
      </section>
    );
  }

  private renderInstance(instance: IServiceInstance) {
    const conditions = [...instance.status.conditions];
    return (
      <div>
        <div>
          <h2>Status</h2>
          <hr />
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

  private onAddBinding = () => {
    const { namespace, getBindings } = this.props;
    getBindings(namespace);
  };
}

export default ServiceInstanceView;
