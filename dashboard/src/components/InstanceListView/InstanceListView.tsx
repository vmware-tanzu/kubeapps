import * as React from "react";
import { Link } from "react-router-dom";

import { IServiceInstance } from "../../shared/ServiceInstance";
import { IRBACRole } from "../../shared/types";
import { escapeRegExp } from "../../shared/utils";

import { IServiceCatalogState } from "reducers/catalog";
import {
  ErrorSelector,
  MessageAlert,
  ServiceBrokersNotFoundAlert,
  ServiceCatalogNotInstalledAlert,
} from "../ErrorAlert";
import LoadingWrapper from "../LoadingWrapper";
import PageHeader from "../PageHeader";
import SearchFilter from "../SearchFilter";
import { InstanceCardList } from "./InstanceCardList";

export interface InstanceListViewProps {
  brokers: IServiceCatalogState["brokers"];
  classes: IServiceCatalogState["classes"];
  error: Error | undefined;
  filter: string;
  getBrokers: () => Promise<any>;
  getClasses: () => Promise<any>;
  getInstances: (ns: string) => Promise<any>;
  checkCatalogInstalled: () => Promise<any>;
  instances: IServiceCatalogState["instances"];
  pushSearchFilter: (filter: string) => any;
  isServiceCatalogInstalled: boolean;
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
];

class InstanceListView extends React.PureComponent<InstanceListViewProps, InstanceListViewState> {
  public state: InstanceListViewState = { filter: "" };
  public async componentDidMount() {
    this.props.checkCatalogInstalled();
    this.props.getBrokers();
    this.props.getClasses();
    this.props.getInstances(this.props.namespace);
    this.setState({ filter: this.props.filter });
  }

  public componentWillReceiveProps(nextProps: InstanceListViewProps) {
    const { error, filter, getInstances, isServiceCatalogInstalled, namespace } = this.props;
    // refetch if new namespace or error removed due to location change
    if (
      isServiceCatalogInstalled &&
      (nextProps.namespace !== namespace || (error && !nextProps.error))
    ) {
      getInstances(nextProps.namespace);
    }
    if (nextProps.filter !== filter) {
      this.setState({ filter: nextProps.filter });
    }
  }

  public render() {
    const {
      error,
      isServiceCatalogInstalled,
      brokers,
      instances,
      classes,
      pushSearchFilter,
    } = this.props;
    const loaded = !brokers.isFetching && !instances.isFetching && !classes.isFetching;
    return (
      <section className="InstanceList">
        <PageHeader>
          <div className="col-8">
            <div className="row collapse-b-phone-land">
              <h1>Service Instances</h1>
              <SearchFilter
                className="margin-l-big "
                placeholder="search instances..."
                onChange={this.handleFilterQueryChange}
                value={this.state.filter}
                onSubmit={pushSearchFilter}
              />
            </div>
          </div>
          <div className="col-4 text-r align-center">
            <Link to="/services/classes">
              <button className="button button-accent">Deploy Service Instance</button>
            </Link>
          </div>
        </PageHeader>
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
        <LoadingWrapper loaded={loaded}>
          <main>
            {error ? (
              <ErrorSelector
                error={error}
                action="list"
                defaultRequiredRBACRoles={{ list: RequiredRBACRoles }}
                resource="Service Brokers, Classes and Instances"
              />
            ) : (
              (!isServiceCatalogInstalled && <ServiceCatalogNotInstalledAlert />) ||
              (brokers.list.length === 0 && <ServiceBrokersNotFoundAlert />) ||
              (instances.list.length > 0 ? (
                <InstanceCardList
                  instances={this.filteredServiceInstances(instances.list, this.state.filter)}
                  classes={classes.list}
                />
              ) : (
                <MessageAlert header="Provision External Services from the Kubernetes Service Catalog">
                  <div>
                    <p className="margin-v-normal">
                      Kubeapps lets you browse, provision and manage external services provided by
                      the Service Brokers configured in your cluster.
                    </p>
                    <div className="padding-t-normal padding-b-normal">
                      <Link to="/services/classes">
                        <button className="button button-accent">Deploy Service Instance</button>
                      </Link>
                    </div>
                  </div>
                </MessageAlert>
              ))
            )}
          </main>
        </LoadingWrapper>
      </section>
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

export default InstanceListView;
