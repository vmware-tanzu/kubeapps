import * as React from "react";
import { Link } from "react-router-dom";

import { IAppOverview, IAppState, IClusterServiceVersion, IResource } from "../../shared/types";
import { escapeRegExp } from "../../shared/utils";
import { CardGrid } from "../Card";
import { ErrorSelector, MessageAlert } from "../ErrorAlert";
import LoadingWrapper from "../LoadingWrapper";
import PageHeader from "../PageHeader";
import SearchFilter from "../SearchFilter";
import AppListItem from "./AppListItem";
import CustomResourceListItem from "./CustomResourceListItem";

interface IAppListProps {
  apps: IAppState;
  fetchAppsWithUpdateInfo: (ns: string, all: boolean) => void;
  namespace: string;
  pushSearchFilter: (filter: string) => any;
  filter: string;
  getCustomResources: (ns: string) => void;
  customResources: IResource[];
  isFetchingResources: boolean;
  csvs: IClusterServiceVersion[];
  featureFlags: { operators: boolean };
}

interface IAppListState {
  filter: string;
}

class AppList extends React.Component<IAppListProps, IAppListState> {
  public state: IAppListState = { filter: "" };
  public componentDidMount() {
    const { fetchAppsWithUpdateInfo, filter, namespace, apps, getCustomResources } = this.props;
    fetchAppsWithUpdateInfo(namespace, apps.listingAll);
    if (this.props.featureFlags.operators) {
      getCustomResources(namespace);
    }
    this.setState({ filter });
  }

  public componentDidUpdate(prevProps: IAppListProps) {
    const {
      apps: { error, listingAll },
      fetchAppsWithUpdateInfo,
      getCustomResources,
      filter,
      namespace,
    } = this.props;
    // refetch if new namespace or error removed due to location change
    if (prevProps.namespace !== namespace || (!error && prevProps.apps.error)) {
      fetchAppsWithUpdateInfo(namespace, listingAll);
      if (this.props.featureFlags.operators) {
        getCustomResources(namespace);
      }
    }
    if (prevProps.filter !== filter) {
      this.setState({ filter });
    }
  }

  public render() {
    const {
      apps: { error, isFetching },
      isFetchingResources,
      namespace,
    } = this.props;
    return (
      <section className="AppList">
        <PageHeader>
          <div className="col-9">
            <div className="row">
              <h1>Applications</h1>
              {!error && this.appListControls()}
            </div>
          </div>
          <div className="col-3 text-r align-center">
            {!error && (
              <Link to={`/ns/${namespace}/catalog`}>
                <button className="deploy-button button button-accent">Deploy App</button>
              </Link>
            )}
          </div>
        </PageHeader>
        <main>
          <LoadingWrapper loaded={!isFetching && !isFetchingResources}>
            {error ? (
              <ErrorSelector
                error={error}
                action="list"
                resource="Applications"
                namespace={this.props.namespace}
              />
            ) : (
              this.appListItems()
            )}
          </LoadingWrapper>
        </main>
      </section>
    );
  }

  public appListControls() {
    const {
      pushSearchFilter,
      apps: { listingAll },
    } = this.props;
    return (
      <React.Fragment>
        <SearchFilter
          key="searchFilter"
          className="margin-l-big"
          placeholder="search apps..."
          onChange={this.handleFilterQueryChange}
          value={this.state.filter}
          onSubmit={pushSearchFilter}
        />
        <label className="checkbox margin-r-big margin-l-big margin-t-big" key="listall">
          <input type="checkbox" checked={listingAll} onChange={this.toggleListAll} />
          <span>Show deleted apps</span>
        </label>
      </React.Fragment>
    );
  }

  public appListItems() {
    const {
      apps: { listOverview },
      customResources,
    } = this.props;
    const filteredReleases = this.filteredReleases(listOverview || [], this.state.filter);
    const filteredCRs = this.filteredCRs(customResources, this.state.filter);
    if (filteredReleases.length === 0 && filteredCRs.length === 0) {
      return (
        <MessageAlert header="Supercharge your Kubernetes cluster">
          <div>
            <p className="margin-v-normal">
              Deploy applications on your Kubernetes cluster with a single click.
            </p>
          </div>
        </MessageAlert>
      );
    }
    return (
      <div>
        <CardGrid>
          {filteredReleases.map(r => {
            return <AppListItem key={r.releaseName} app={r} />;
          })}
          {filteredCRs.map(r => {
            const csv = this.props.csvs.find(c =>
              c.spec.customresourcedefinitions.owned.some(crd => crd.kind === r.kind),
            );
            return <CustomResourceListItem key={r.metadata.name} resource={r} csv={csv!} />;
          })}
        </CardGrid>
      </div>
    );
  }

  private toggleListAll = () => {
    this.props.fetchAppsWithUpdateInfo(this.props.namespace, !this.props.apps.listingAll);
  };

  private filteredReleases(apps: IAppOverview[], filter: string) {
    return apps.filter(a => new RegExp(escapeRegExp(filter), "i").test(a.releaseName));
  }

  private filteredCRs(crs: IResource[], filter: string) {
    return crs.filter(cr => new RegExp(escapeRegExp(filter), "i").test(cr.metadata.name));
  }

  private handleFilterQueryChange = (filter: string) => {
    this.setState({
      filter,
    });
  };
}

export default AppList;
