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
import OperatorInstanceCreateContainer from "../../containers/OperatorInstanceCreateContainer";
import OperatorInstanceViewContainer from "../../containers/OperatorInstanceViewContainer";
import OperatorsListContainer from "../../containers/OperatorsListContainer";
import OperatorViewContainer from "../../containers/OperatorViewContainer";
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
  featureFlags: {
    reposPerNamespace: boolean;
    operators: boolean;
  };
}

class Routes extends React.Component<IRoutesProps> {
  public static defaultProps = {
    featureFlags: { reposPerNamespace: false, operators: false },
  };
  public render() {
    // The path used for AppRepository list depends on a feature flag.
    // TODO(mnelson, #1256) Remove when feature becomes default.
    const reposPath = this.props.featureFlags.reposPerNamespace
      ? "/config/ns/:namespace/repos"
      : "/config/repos";
    if (this.props.featureFlags.operators) {
      // Add routes related to operators
      Object.assign(privateRoutes, {
        "/operators/ns/:namespace": OperatorsListContainer,
        "/operators/ns/:namespace/:operator": OperatorViewContainer,
        "/operators-instances/ns/:namespace/:instanceName": OperatorInstanceViewContainer,
        "/operators-instances/ns/:namespace/new/:csv/:crd": OperatorInstanceCreateContainer,
      });
    }
    return (
      <Switch>
        <Route exact={true} path="/" render={this.rootNamespacedRedirect} />
        {Object.entries(routes).map(([route, component]) => (
          <Route key={route} exact={true} path={route} component={component} />
        ))}
        {Object.entries(privateRoutes).map(([route, component]) => (
          <PrivateRouteContainer key={route} exact={true} path={route} component={component} />
        ))}
        <PrivateRouteContainer
          key={reposPath}
          exact={true}
          path={reposPath}
          component={RepoListContainer}
        />
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
