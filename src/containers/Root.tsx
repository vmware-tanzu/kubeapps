import createHistory from "history/createBrowserHistory";
import * as React from "react";
import { Provider } from "react-redux";
import { Route, RouteComponentProps } from "react-router";
import { ConnectedRouter } from "react-router-redux";

import Dashboard from "../components/Dashboard";
import Layout from "../components/Layout";
import configureStore from "../store";
import ChartList from "./ChartListContainer";
import ChartView from "./ChartViewContainer";

import BrokerView from "./BrokerView";
import ClassListContainer from "./ClassListContainer";
import { ClassViewContainer } from "./ClassView";
import RepoListContainer from "./RepoListContainer";
import ServiceCatalogContainer from "./ServiceCatalogContainer";

const history = createHistory();
const store = configureStore(history);

class Root extends React.Component {
  public static exactRoutes: {
    [route: string]: React.ComponentType<RouteComponentProps<any>> | React.ComponentType<any>;
  } = {
    "/": Dashboard,
    "/charts": ChartList,
    "/charts/:repo": ChartList,
    "/charts/:repo/:id": ChartView,
    "/charts/:repo/:id/versions/:version": ChartView,
    "/repos": RepoListContainer,
    "/services": ServiceCatalogContainer,
    "/services/brokers/:brokerName/classes": ClassListContainer,
    "/services/brokers/:brokerName/classes/:className": ClassViewContainer,
    "/services/brokers/:name": BrokerView,
  };

  public render() {
    return (
      <Provider store={store}>
        <ConnectedRouter history={history}>
          <Layout>
            <section className="routes">
              {Object.keys(Root.exactRoutes).map(route => (
                <Route exact={true} path={route} component={Root.exactRoutes[route]} />
              ))}
            </section>
          </Layout>
        </ConnectedRouter>
      </Provider>
    );
  }
}

export default Root;
