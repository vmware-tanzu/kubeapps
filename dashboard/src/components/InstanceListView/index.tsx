import * as React from "react";
import { Link } from "react-router-dom";

import { IClusterServiceClass } from "../../shared/ClusterServiceClass";
import { IServiceBroker, IServicePlan } from "../../shared/ServiceCatalog";
import { IServiceInstance } from "../../shared/ServiceInstance";
import { ForbiddenError, IRBACRole } from "../../shared/types";
import { escapeRegExp } from "../../shared/utils";

import {
  MessageAlert,
  PermissionsErrorAlert,
  ServiceBrokersNotFoundAlert,
  ServiceCatalogNotInstalledAlert,
  UnexpectedErrorAlert,
} from "../ErrorAlert";
import PageHeader from "../PageHeader";
import SearchFilter from "../SearchFilter";
import { InstanceCardList } from "./InstanceCardList";

export interface InstanceListViewProps {
  brokers: IServiceBroker[];
  classes: IClusterServiceClass[];
  error: Error;
  filter: string;
  getCatalog: (ns: string) => Promise<any>;
  checkCatalogInstalled: () => Promise<any>;
  instances: IServiceInstance[];
  plans: IServicePlan[];
  pushSearchFilter: (filter: string) => any;
  isInstalled: boolean;
  namespace: string;
}

export interface InstanceListViewState {
  filter: string;
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

export class InstanceListView extends React.PureComponent<
  InstanceListViewProps,
  InstanceListViewState
> {
  public state: InstanceListViewState = { filter: "" };
  public async componentDidMount() {
    this.props.checkCatalogInstalled();
    this.props.getCatalog(this.props.namespace);
    this.setState({ filter: this.props.filter });
  }

  public componentWillReceiveProps(nextProps: InstanceListViewProps) {
    const { error, filter, getCatalog, isInstalled, namespace } = this.props;
    // refetch if new namespace or error removed due to location change
    if (isInstalled && (nextProps.namespace !== namespace || (error && !nextProps.error))) {
      getCatalog(nextProps.namespace);
    }
    if (nextProps.filter !== filter) {
      this.setState({ filter: nextProps.filter });
    }
  }

  public render() {
    const { error, isInstalled, brokers, instances, classes, pushSearchFilter } = this.props;

    return (
      <section className="InstanceList">
        <PageHeader>
          <div className="col-8">
            <div className="row collapse-b-phone-land">
              <h1>Service Instances</h1>
              {instances.length > 0 && (
                <SearchFilter
                  className="margin-l-big "
                  placeholder="search instances..."
                  onChange={this.handleFilterQueryChange}
                  value={this.state.filter}
                  onSubmit={pushSearchFilter}
                />
              )}
            </div>
          </div>
          {instances.length > 0 && (
            <div className="col-4 text-r align-center">
              <Link to="/services/classes">
                <button className="button button-accent">Deploy Service Instance</button>
              </Link>
            </div>
          )}
        </PageHeader>
        <main>
          <MessageAlert type="warning">
            <div>
              Service Catalog integration is under heavy development. If you find an issue please
              report it{" "}
              <a target="_blank" href="https://github.com/kubeapps/kubeapps/issues">
                {" "}
                here.
              </a>
            </div>
          </MessageAlert>
          {isInstalled ? (
            <div>
              {error ? (
                this.renderError()
              ) : brokers.length > 0 ? (
                <div>
                  {instances.length > 0 ? (
                    <InstanceCardList
                      instances={this.filteredServiceInstances(instances, this.state.filter)}
                      classes={classes}
                    />
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

  private filteredServiceInstances(instances: IServiceInstance[], filter: string) {
    return instances.filter(i => new RegExp(escapeRegExp(filter), "i").test(i.metadata.name));
  }

  private handleFilterQueryChange = (filter: string) => {
    this.setState({
      filter,
    });
  };
}
