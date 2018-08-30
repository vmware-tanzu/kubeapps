import * as React from "react";
import { Link } from "react-router-dom";

import { IAppOverview, IAppState } from "../../shared/types";
import { escapeRegExp } from "../../shared/utils";
import { CardGrid } from "../Card";
import { MessageAlert, UnexpectedErrorAlert } from "../ErrorAlert";
import PageHeader from "../PageHeader";
import SearchFilter from "../SearchFilter";
import AppListItem from "./AppListItem";

interface IAppListProps {
  apps: IAppState;
  fetchApps: (ns: string, all: boolean) => Promise<void>;
  namespace: string;
  pushSearchFilter: (filter: string) => any;
  filter: string;
  toggleListAll: () => Promise<void>;
}

interface IAppListState {
  filter: string;
  listAll: boolean;
}

class AppList extends React.Component<IAppListProps, IAppListState> {
  public state: IAppListState = { filter: "", listAll: false };
  public componentDidMount() {
    const { fetchApps, filter, namespace, apps } = this.props;
    fetchApps(namespace, apps.listAll);
    this.setState({ filter });
  }

  public componentWillReceiveProps(nextProps: IAppListProps) {
    const { apps: { error, listAll }, fetchApps, filter, namespace } = this.props;
    // refetch if new namespace or error removed due to location change
    if (nextProps.namespace !== namespace || (error && !nextProps.apps.error)) {
      fetchApps(nextProps.namespace, listAll);
    }
    if (nextProps.filter !== filter) {
      this.setState({ filter: nextProps.filter });
    }
    if (nextProps.apps.listAll !== listAll) {
      this.setState({ listAll: nextProps.apps.listAll });
    }
  }

  public render() {
    const { pushSearchFilter, apps: { error, isFetching, listOverview, listAll } } = this.props;
    if (!listOverview) {
      return <div>Loading</div>;
    }
    return (
      <section className="AppList">
        <PageHeader>
          <div className="col-7">
            <div className="row">
              <h1>Applications</h1>
              {listOverview.length > 0 && (
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
          {listOverview.length > 0 && (
            <div className="col-5">
              <div className="text-r">
                <label className="checkbox margin-r-big">
                  <input type="checkbox" checked={listAll} onChange={this.toggleListAll} />
                  <span>List all</span>
                </label>
                <Link to="/charts">
                  <button className="button button-accent">Deploy App</button>
                </Link>
              </div>
            </div>
          )}
        </PageHeader>
        <main>
          {isFetching ? (
            <div>Loading</div>
          ) : error ? (
            this.renderError(error)
          ) : (
            this.appListItems(listOverview)
          )}
        </main>
      </section>
    );
  }

  public appListItems(items: IAppState["listOverview"]) {
    if (items) {
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
    this.props.fetchApps(this.props.namespace, !this.props.apps.listAll);
    this.props.toggleListAll();
  };

  private renderError(error: Error) {
    return <UnexpectedErrorAlert />;
  }

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
