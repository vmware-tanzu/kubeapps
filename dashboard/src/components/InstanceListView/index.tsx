import * as React from "react";
import { Link } from "react-router-dom";

import { IClusterServiceClass } from "../../shared/ClusterServiceClass";
import { IServiceBroker, IServicePlan } from "../../shared/ServiceCatalog";
import { IServiceInstance } from "../../shared/ServiceInstance";
import { ForbiddenError, IRBACRole } from "../../shared/types";
import { NotFoundErrorAlert, PermissionsErrorAlert, UnexpectedErrorAlert } from "../ErrorAlert";
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
      <div className="InstanceList">
        <h1 className="margin-b-reset">Service Instances</h1>
        <hr />
        {isInstalled ? (
          <div>
            {error ? (
              this.renderError()
            ) : brokers.length > 0 ? (
              <div>
                <div className="row align-center">
                  <div className="col-8">
                    <p>Service instances from your brokers:</p>
                  </div>
                  <div className="col-4 text-r">
                    <Link to={`/services/classes`}>
                      <button className="button button-accent">Provision New Service</button>
                    </Link>
                  </div>
                </div>
                {instances.length > 0 ? (
                  <InstanceCardList instances={instances} classes={classes} />
                ) : (
                  <div>No service instances are found.</div>
                )}
              </div>
            ) : (
              <NotFoundErrorAlert header="No Service Brokers installed.">
                <p>
                  Ask an administrator to install a compatible{" "}
                  <a href="https://github.com/osbkit/brokerlist" target="_blank">
                    Service Broker
                  </a>{" "}
                  to browse, provision and manage external services within Kubeapps.
                </p>
              </NotFoundErrorAlert>
            )}
          </div>
        ) : (
          <NotFoundErrorAlert header="Service Catalog not installed.">
            <div>
              <p>
                Ask an administrator to install the{" "}
                <a href="https://github.com/kubernetes-incubator/service-catalog" target="_blank">
                  Kubernetes Service Catalog
                </a>{" "}
                to browse, provision and manage external services within Kubeapps.
              </p>
              <Link className="button button-primary button-small" to={`/charts/svc-cat/catalog`}>
                Install Catalog
              </Link>
            </div>
          </NotFoundErrorAlert>
        )}
      </div>
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
