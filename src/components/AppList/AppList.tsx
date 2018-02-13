import * as React from "react";
import { Link } from "react-router-dom";

import { IAppState } from "../../shared/types";
import AppListItem from "./AppListItem";

interface IAppListProps {
  apps: IAppState;
  fetchApps: () => Promise<{}>;
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
          <h1>Applications</h1>
          <hr />
        </header>
        <main className="text-c">
          {isFetching ? <div>Loading</div> : this.chartListItems(items)}
        </main>
      </section>
    );
  }

  public chartListItems(items: IAppState["items"]) {
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
        <div>
          <div className="padding-normal">
            <Link className="button button-primary" to="/charts">
              deploy another one
            </Link>
          </div>
          <div>
            {items.map(r => {
              return <AppListItem key={r.data.name} app={r} />;
            })}
          </div>
        </div>
      );
    }
  }
}

export default AppList;
