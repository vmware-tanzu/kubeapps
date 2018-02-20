import * as React from "react";
import { Link } from "react-router-dom";

import { IAppState } from "../../shared/types";
import { CardGrid } from "../Card";
import AppListItem from "./AppListItem";

interface IAppListProps {
  apps: IAppState;
  fetchApps: () => Promise<void>;
}

class AppList extends React.Component<IAppListProps> {
  public componentDidMount() {
    const { fetchApps } = this.props;
    fetchApps();
  }

  public render() {
    const { isFetching, items } = this.props.apps;
    return (
      <section className="AppList">
        <header className="AppList__header">
          <div className="row padding-t-big collapse-b-phone-land">
            <div className="col-8">
              <h1 className="margin-v-reset">Applications</h1>
            </div>
            <div className="col-4 text-r align-center">
              <Link to={`/charts`}>
                <button className="button button-accent">Deploy New App</button>
              </Link>
            </div>
          </div>
          <hr />
        </header>
        <main>{isFetching ? <div>Loading</div> : this.appListItems(items)}</main>
      </section>
    );
  }

  public appListItems(items: IAppState["items"]) {
    if (items.length === 0) {
      return (
        <div>
          <div>No Apps installed</div>
          <div className="padding-normal">
            <Link className="button button-primary" to="/charts">
              deploy one
            </Link>
          </div>
        </div>
      );
    } else {
      return (
        <CardGrid>
          {items.map(r => {
            return <AppListItem key={r.data.name} app={r} />;
          })}
        </CardGrid>
      );
    }
  }
}

export default AppList;
