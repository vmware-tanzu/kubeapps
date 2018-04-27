import * as React from "react";
import { Link } from "react-router-dom";

import { IClusterServiceClass } from "../../shared/ClusterServiceClass";
import { IServiceBroker, IServicePlan } from "../../shared/ServiceCatalog";
import { IServiceInstance } from "../../shared/ServiceInstance";
import { ForbiddenError, IRBACRole } from "../../shared/types";
import {
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
              <ServiceBrokersNotFoundAlert />
            )}
          </div>
        ) : (
          <ServiceCatalogNotInstalledAlert />
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
