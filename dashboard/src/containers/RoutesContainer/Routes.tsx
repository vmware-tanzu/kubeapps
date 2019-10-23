import * as React from "react";
import { Redirect, Route, RouteComponentProps, RouteProps, Switch } from "react-router";

import NotFound from "../../components/NotFound";
import AppListContainer from "../../containers/AppListContainer";
import AppNewContainer from "../../containers/AppNewContainer";
import AppUpgradeContainer from "../../containers/AppUpgradeContainer";
import AppViewContainer from "../../containers/AppViewContainer";
import CatalogContainer from "../../containers/CatalogContainer";
import ChartViewContainer from "../../containers/ChartViewContainer";
import LoginFormContainer from "../../containers/LoginFormContainer";
import PrivateRouteContainer from "../../containers/PrivateRouteContainer";
import RepoListContainer from "../../containers/RepoListContainer";
import ServiceBrokerListContainer from "../../containers/ServiceBrokerListContainer";
import ServiceClassListContainer from "../../containers/ServiceClassListContainer";
import ServiceClassViewContainer from "../../containers/ServiceClassViewContainer";
import ServiceInstanceListContainer from "../../containers/ServiceInstanceListContainer";
import ServiceInstanceViewContainer from "../../containers/ServiceInstanceViewContainer";

type IRouteComponentPropsAndRouteProps = RouteProps & RouteComponentProps<any>;

const privateRoutes = {
  "/apps/ns/:namespace": AppListContainer,
  "/apps/ns/:namespace/:releaseName": AppViewContainer,
  "/apps/ns/:namespace/new/:repo/:id/versions/:version": AppNewContainer,
  "/apps/ns/:namespace/upgrade/:releaseName": AppUpgradeContainer,
  "/catalog": CatalogContainer,
  "/catalog/:repo": CatalogContainer,
  "/charts/:repo/:id": ChartViewContainer,
  "/charts/:repo/:id/versions/:version": ChartViewContainer,
  "/config/brokers": ServiceBrokerListContainer,
  "/config/repos": RepoListContainer,
  "/services/brokers/:brokerName/classes/:className": ServiceClassViewContainer,
  "/services/brokers/:brokerName/instances/ns/:namespace/:instanceName": ServiceInstanceViewContainer,
  "/services/classes": ServiceClassListContainer,
  "/services/instances/ns/:namespace": ServiceInstanceListContainer,
} as const;

// Public routes that don't require authentication
const routes = {
  "/login": LoginFormContainer,
} as const;

interface IRoutesProps extends IRouteComponentPropsAndRouteProps {
  namespace: string;
  authenticated: boolean;
}

class Routes extends React.Component<IRoutesProps> {
  public render() {
    return (
      <Switch>
        <Route exact={true} path="/" render={this.rootNamespacedRedirect} />
        {Object.keys(routes).map(route => (
          <Route key={route} exact={true} path={route} component={routes[route]} />
        ))}
        {Object.keys(privateRoutes).map(route => (
          <PrivateRouteContainer
            key={route}
            exact={true}
            path={route}
            component={privateRoutes[route]}
          />
        ))}
        {/* If the route doesn't match any expected path redirect to a 404 page  */}
        <Route component={NotFound} />
      </Switch>
    );
  }
  private rootNamespacedRedirect = () => {
    if (this.props.namespace && this.props.authenticated) {
      return <Redirect to={`/apps/ns/${this.props.namespace}`} />;
    }
    // There is not a default namespace, redirect to login page
    return <Redirect to={"/login"} />;
  };
}

export default Routes;
