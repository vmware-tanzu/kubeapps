import * as React from "react";
import { Link } from "react-router-dom";

import { IServiceBroker, IServicePlan } from "shared/ServiceCatalog";
import { IServiceInstance } from "shared/ServiceInstance";
import { ForbiddenError, IRBACRole } from "shared/types";
import { escapeRegExp } from "shared/utils";

import { IServiceCatalogState } from "reducers/catalog";
import {
  MessageAlert,
  PermissionsErrorAlert,
  ServiceBrokersNotFoundAlert,
  ServiceCatalogNotInstalledAlert,
  UnexpectedErrorAlert,
} from "../ErrorAlert";
import PageHeader from "../PageHeader";
import SearchFilter from "../SearchFilter";
import ServiceInstanceCardList from "./ServiceInstanceCardList";

export interface IServiceInstanceListProps {
  brokers: IServiceBroker[];
  classes: IServiceCatalogState["classes"];
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

interface IServiceInstanceListState {
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

class ServiceInstanceList extends React.PureComponent<
  IServiceInstanceListProps,
  IServiceInstanceListState
> {
  public state: IServiceInstanceListState = { filter: "" };
  public async componentDidMount() {
    this.props.checkCatalogInstalled();
    this.props.getCatalog(this.props.namespace);
    this.setState({ filter: this.props.filter });
  }

  public componentWillReceiveProps(nextProps: IServiceInstanceListProps) {
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
      <section className="ServiceInstanceList">
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
          <MessageAlert level="warning">
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
                    // TODO: Check isFetching
                    <ServiceInstanceCardList
                      instances={this.filteredServiceInstances(instances, this.state.filter)}
                      classes={classes.list}
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

  // TODO: Replace with ErrorSelector
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

export default ServiceInstanceList;
