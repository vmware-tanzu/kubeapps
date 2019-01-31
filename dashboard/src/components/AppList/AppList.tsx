import * as React from "react";
import { Link } from "react-router-dom";

import { IAppOverview, IAppState } from "../../shared/types";
import { escapeRegExp } from "../../shared/utils";
import { CardGrid } from "../Card";
import { ErrorSelector, MessageAlert } from "../ErrorAlert";
import LoadingWrapper from "../LoadingWrapper";
import PageHeader from "../PageHeader";
import SearchFilter from "../SearchFilter";
import AppListItem from "./AppListItem";

interface IAppListProps {
  apps: IAppState;
  fetchAppsWithUpdateInfo: (ns: string, all: boolean) => void;
  namespace: string;
  pushSearchFilter: (filter: string) => any;
  filter: string;
}

interface IAppListState {
  filter: string;
}

class AppList extends React.Component<IAppListProps, IAppListState> {
  public state: IAppListState = { filter: "" };
  public componentDidMount() {
    const { fetchAppsWithUpdateInfo, filter, namespace, apps } = this.props;
    fetchAppsWithUpdateInfo(namespace, apps.listingAll);
    this.setState({ filter });
  }

  public componentWillReceiveProps(nextProps: IAppListProps) {
    const {
      apps: { error, listingAll },
      fetchAppsWithUpdateInfo,
      filter,
      namespace,
    } = this.props;
    // refetch if new namespace or error removed due to location change
    if (nextProps.namespace !== namespace || (error && !nextProps.apps.error)) {
      fetchAppsWithUpdateInfo(nextProps.namespace, listingAll);
    }
    if (nextProps.filter !== filter) {
      this.setState({ filter: nextProps.filter });
    }
  }

  public render() {
    const {
      apps: { error, isFetching, listOverview },
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
              <Link to="/catalog">
                <button className="deploy-button button button-accent">Deploy App</button>
              </Link>
            )}
          </div>
        </PageHeader>
        <main>
          <LoadingWrapper loaded={!isFetching}>
            {error ? (
              <ErrorSelector
                error={error}
                action="list"
                resource="Applications"
                namespace={this.props.namespace}
              />
            ) : (
              this.appListItems(listOverview)
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

  public appListItems(items: IAppState["listOverview"]) {
    if (items) {
      const filteredItems = this.filteredApps(items, this.state.filter);
      if (filteredItems.length === 0) {
        return (
          <MessageAlert header="Supercharge your Kubernetes cluster">
            <div>
              <p className="margin-v-normal">
                Deploy applications on your Kubernetes cluster with a single click.
              </p>
            </div>
          </MessageAlert>
        );
      } else {
        return (
          <div>
            <CardGrid>
              {filteredItems.map(r => {
                return <AppListItem key={r.releaseName} app={r} />;
              })}
            </CardGrid>
          </div>
        );
      }
    }
    return <div />;
  }

  private toggleListAll = () => {
    this.props.fetchAppsWithUpdateInfo(this.props.namespace, !this.props.apps.listingAll);
  };

  private filteredApps(apps: IAppOverview[], filter: string) {
    return apps.filter(a => new RegExp(escapeRegExp(filter), "i").test(a.releaseName));
  }

  private handleFilterQueryChange = (filter: string) => {
    this.setState({
      filter,
    });
  };
}

export default AppList;
