import * as React from "react";
import { Link } from "react-router-dom";

import { hapi } from "../../shared/hapi/release";
import { IAppState } from "../../shared/types";
import { escapeRegExp } from "../../shared/utils";
import { CardGrid } from "../Card";
import { MessageAlert, UnexpectedErrorAlert } from "../ErrorAlert";
import PageHeader from "../PageHeader";
import SearchFilter from "../SearchFilter";
import AppListItem from "./AppListItem";

interface IAppListProps {
  apps: IAppState;
  fetchApps: (ns: string) => Promise<void>;
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
    const { fetchApps, filter, namespace } = this.props;
    fetchApps(namespace);
    this.setState({ filter });
  }

  public componentWillReceiveProps(nextProps: IAppListProps) {
    const { apps: { error }, fetchApps, filter, namespace } = this.props;
    // refetch if new namespace or error removed due to location change
    if (nextProps.namespace !== namespace || (error && !nextProps.apps.error)) {
      fetchApps(nextProps.namespace);
    }
    if (nextProps.filter !== filter) {
      this.setState({ filter: nextProps.filter });
    }
  }

  public render() {
    const { pushSearchFilter, apps: { error, isFetching, items } } = this.props;

    return (
      <section className="AppList">
        <PageHeader>
          <div className="col-8">
            <div className="row">
              <h1>Applications</h1>
              {items.length > 0 && (
                <SearchFilter
                  className="margin-l-big"
                  placeholder="search apps..."
                  onChange={this.handleFilterQueryChange}
                  value={this.state.filter}
                  onSubmit={pushSearchFilter}
                />
              )}
            </div>
          </div>
          {items.length > 0 && (
            <div className="col-4 text-r align-center">
              <Link to="/charts">
                <button className="button button-accent">Deploy App</button>
              </Link>
            </div>
          )}
        </PageHeader>
        <main>
          {isFetching ? (
            <div>Loading</div>
          ) : error ? (
            this.renderError(error)
          ) : (
            this.appListItems(items)
          )}
        </main>
      </section>
    );
  }

  public appListItems(items: IAppState["items"]) {
    if (items.length === 0) {
      return (
        <MessageAlert header="Supercharge your Kubernetes cluster">
          <div>
            <p className="margin-v-normal">
              Deploy applications on your Kubernetes cluster with a single click.
            </p>
            <div className="padding-b-normal">
              <Link className="button button-accent" to="/charts">
                Deploy App
              </Link>
            </div>
          </div>
        </MessageAlert>
      );
    } else {
      return (
        <div>
          <CardGrid>
            {this.filteredApps(items, this.state.filter).map(r => {
              return <AppListItem key={r.name} app={r} />;
            })}
          </CardGrid>
        </div>
      );
    }
  }

  private renderError(error: Error) {
    return <UnexpectedErrorAlert />;
  }

  private filteredApps(apps: hapi.release.Release[], filter: string) {
    return apps.filter(a => new RegExp(escapeRegExp(filter), "i").test(a.name));
  }

  private handleFilterQueryChange = (filter: string) => {
    this.setState({
      filter,
    });
  };
}

export default AppList;
