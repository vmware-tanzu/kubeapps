import * as React from "react";
import { Link } from "react-router-dom";

import { IClusterServiceClass } from "../../shared/ClusterServiceClass";
import { IServiceBroker, IServicePlan } from "../../shared/ServiceCatalog";
import { IServiceInstance } from "../../shared/ServiceInstance";
import { ForbiddenError, IRBACRole } from "../../shared/types";
import {
  MessageAlert,
  PermissionsErrorAlert,
  ServiceBrokersNotFoundAlert,
  ServiceCatalogNotInstalledAlert,
  UnexpectedErrorAlert,
} from "../ErrorAlert";
import { InstanceCardList } from "./InstanceCardList";

export interface InstanceListViewProps {
  brokers: IServiceBroker[];
  classes: IClusterServiceClass[];
  error: Error;
  getCatalog: (ns: string) => Promise<any>;
  checkCatalogInstalled: () => Promise<any>;
  instances: IServiceInstance[];
  plans: IServicePlan[];
  isInstalled: boolean;
  namespace: string;
}

const RequiredRBACRoles: IRBACRole[] = [
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
  // TODO: these 2 roles should not be required for this view and should be decoupled.
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
];

export class InstanceListView extends React.PureComponent<InstanceListViewProps> {
  public async componentDidMount() {
    this.props.checkCatalogInstalled();
    this.props.getCatalog(this.props.namespace);
  }

  public componentWillReceiveProps(nextProps: InstanceListViewProps) {
    const { error, getCatalog, isInstalled, namespace } = this.props;
    // refetch if new namespace or error removed due to location change
    if (isInstalled && (nextProps.namespace !== namespace || (error && !nextProps.error))) {
      getCatalog(nextProps.namespace);
    }
  }

  public render() {
    const { error, isInstalled, brokers, instances, classes } = this.props;

    return (
      <section className="InstanceList">
        <header className="InstanceList__header">
          <div className="row padding-t-big collapse-b-phone-land">
            <div className="col-8">
              <h1 className="margin-v-reset">Service Instances</h1>
            </div>
            {instances.length > 0 && (
              <div className="col-4 text-r align-center">
                <Link to="/services/classes">
                  <button className="button button-accent">Deploy Service Instance</button>
                </Link>
              </div>
            )}
          </div>
          <hr />
        </header>
        <main>
          {isInstalled ? (
            <div>
              {error ? (
                this.renderError()
              ) : brokers.length > 0 ? (
                <div>
                  {instances.length > 0 ? (
                    <InstanceCardList instances={instances} classes={classes} />
                  ) : (
                    <MessageAlert header="Provision External Services from the Kubernetes Service Catalog">
                      <div>
                        <p className="margin-v-normal">
                          Kubeapps lets you browse, provision and manage external services provided
                          by the Service Brokers configured in your cluster.
                        </p>
                        <div className="padding-t-normal padding-b-normal">
                          <Link to="/services/classes">
                            <button className="button button-accent">
                              Deploy Service Instance
                            </button>
                          </Link>
                        </div>
                      </div>
                    </MessageAlert>
                  )}
                </div>
              ) : (
                <ServiceBrokersNotFoundAlert />
              )}
            </div>
          ) : (
            <ServiceCatalogNotInstalledAlert />
          )}
        </main>
      </section>
    );
  }

  private renderError() {
    const { error, namespace } = this.props;
    return error instanceof ForbiddenError ? (
      <PermissionsErrorAlert
        action="list Service Instances"
        namespace={namespace}
        roles={RequiredRBACRoles}
      />
    ) : (
      <UnexpectedErrorAlert />
    );
  }
}
