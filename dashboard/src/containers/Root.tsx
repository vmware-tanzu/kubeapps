import createHistory from "history/createBrowserHistory";
import * as React from "react";
import { Provider } from "react-redux";
import { Redirect, Route, RouteComponentProps } from "react-router";
import { ConnectedRouter } from "react-router-redux";

import Layout from "../components/Layout";
import configureStore from "../store";
import AppList from "./AppListContainer";
import AppNew from "./AppNewContainer";
import AppUpgrade from "./AppUpgradeContainer";
import AppView from "./AppViewContainer";
import ChartList from "./ChartListContainer";
import ChartView from "./ChartViewContainer";
import ClassListContainer from "./ClassListContainer";
import { ClassViewContainer } from "./ClassView";
import ConfigLoaderContainer from "./ConfigLoaderContainer";
import FunctionListContainer from "./FunctionListContainer";
import FunctionViewContainer from "./FunctionViewContainer";
import HeaderContainer from "./HeaderContainer";
import InstanceListViewContainer from "./InstanceListViewContainer";
import InstanceView from "./InstanceView";
import LoginFormContainer from "./LoginFormContainer";
import PrivateRouteContainer from "./PrivateRouteContainer";
import RepoListContainer from "./RepoListContainer";
import ServiceCatalogContainer from "./ServiceCatalogContainer";

const history = createHistory();
const store = configureStore(history);

class Root extends React.Component {
  public static exactRoutes: {
    [route: string]: React.ComponentType<RouteComponentProps<any>> | React.ComponentType<any>;
  } = {
    "/apps/ns/:namespace": AppList,
    "/apps/ns/:namespace/:releaseName": AppView,
    "/apps/ns/:namespace/new/:repo/:id/versions/:version": AppNew,
    "/apps/ns/:namespace/upgrade/:releaseName": AppUpgrade,
    "/charts": ChartList,
    "/charts/:repo": ChartList,
    "/charts/:repo/:id": ChartView,
    "/charts/:repo/:id/versions/:version": ChartView,
    "/config/brokers": ServiceCatalogContainer,
    "/config/repos": RepoListContainer,
    "/functions/ns/:namespace": FunctionListContainer,
    "/functions/ns/:namespace/:name": FunctionViewContainer,
    "/services/brokers/:brokerName/classes/:className": ClassViewContainer,
    "/services/brokers/:brokerName/instances/ns/:namespace/:instanceName": InstanceView,
    "/services/classes": ClassListContainer,
    "/services/instances/ns/:namespace": InstanceListViewContainer,
  };

  public render() {
    return (
      <Provider store={store}>
        <ConfigLoaderContainer>
          <ConnectedRouter history={history}>
            <Layout headerComponent={HeaderContainer}>
              <Route exact={true} path="/" render={this.rootNamespacedRedirect} />
              <Route exact={true} path="/login" component={LoginFormContainer} />
              {Object.keys(Root.exactRoutes).map(route => (
                <PrivateRouteContainer
                  key={route}
                  exact={true}
                  path={route}
                  component={Root.exactRoutes[route]}
                />
              ))}
            </Layout>
          </ConnectedRouter>
        </ConfigLoaderContainer>
      </Provider>
    );
  }

  public rootNamespacedRedirect = (props: any) => {
    const { namespace } = store.getState();
    return <Redirect to={`/apps/ns/${namespace.current}`} />;
  };
}

export default Root;
