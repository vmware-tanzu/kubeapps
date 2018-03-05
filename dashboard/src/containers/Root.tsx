import createHistory from "history/createBrowserHistory";
import * as React from "react";
import { Provider } from "react-redux";
import { Route, RouteComponentProps } from "react-router";
import { ConnectedRouter } from "react-router-redux";
import { ClassViewContainer } from "./ClassView";

import Layout from "../components/Layout";
import configureStore from "../store";
import AppList from "./AppListContainer";
import AppNew from "./AppNewContainer";
import AppView from "./AppViewContainer";
import ChartList from "./ChartListContainer";
import ChartView from "./ChartViewContainer";
import ClassListContainer from "./ClassListContainer";
import InstanceListViewContainer from "./InstanceListViewContainer";
import InstanceView from "./InstanceView";

import RepoListContainer from "./RepoListContainer";
import ServiceCatalogContainer from "./ServiceCatalogContainer";

const history = createHistory();
const store = configureStore(history);

class Root extends React.Component {
  public static exactRoutes: {
    [route: string]: React.ComponentType<RouteComponentProps<any>> | React.ComponentType<any>;
  } = {
    "/": AppList,
    "/apps/:namespace/:releaseName": AppView,
    "/apps/new/:repo/:id/versions/:version": AppNew,
    "/charts": ChartList,
    "/charts/:repo": ChartList,
    "/charts/:repo/:id": ChartView,
    "/charts/:repo/:id/versions/:version": ChartView,
    "/config/brokers": ServiceCatalogContainer,
    "/config/repos": RepoListContainer,
    "/services/brokers/:brokerName/classes/:className": ClassViewContainer,
    "/services/brokers/:brokerName/instances/:namespace/:instanceName": InstanceView,
    "/services/classes": ClassListContainer,
    "/services/instances": InstanceListViewContainer,
  };

  public render() {
    return (
      <Provider store={store}>
        <ConnectedRouter history={history}>
          <Layout>
            <section className="routes">
              {Object.keys(Root.exactRoutes).map(route => (
                <Route key={route} exact={true} path={route} component={Root.exactRoutes[route]} />
              ))}
            </section>
          </Layout>
        </ConnectedRouter>
      </Provider>
    );
  }
}

export default Root;
