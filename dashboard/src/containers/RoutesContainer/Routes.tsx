import * as React from "react";
import { Redirect, Route, RouteComponentProps, RouteProps, Switch } from "react-router";
import { IFeatureFlags } from "shared/Config";
import NotFound from "../../components/NotFound";
import AppListContainer from "../../containers/AppListContainer";
import AppNewContainer from "../../containers/AppNewContainer";
import AppUpgradeContainer from "../../containers/AppUpgradeContainer";
import AppViewContainer from "../../containers/AppViewContainer";
import CatalogContainer from "../../containers/CatalogContainer";
import ChartViewContainer from "../../containers/ChartViewContainer";
import LoginFormContainer from "../../containers/LoginFormContainer";
import OperatorInstanceCreateContainer from "../../containers/OperatorInstanceCreateContainer";
import OperatorInstanceUpdateContainer from "../../containers/OperatorInstanceUpdateContainer";
import OperatorInstanceViewContainer from "../../containers/OperatorInstanceViewContainer";
import OperatorNewContainer from "../../containers/OperatorNewContainer";
import OperatorsListContainer from "../../containers/OperatorsListContainer";
import OperatorViewContainer from "../../containers/OperatorViewContainer";
import PrivateRouteContainer from "../../containers/PrivateRouteContainer";
import RepoListContainer from "../../containers/RepoListContainer";
import ServiceBrokerListContainer from "../../containers/ServiceBrokerListContainer";
import ServiceClassListContainer from "../../containers/ServiceClassListContainer";
import ServiceClassViewContainer from "../../containers/ServiceClassViewContainer";
import ServiceInstanceListContainer from "../../containers/ServiceInstanceListContainer";
import ServiceInstanceViewContainer from "../../containers/ServiceInstanceViewContainer";
import { app } from "../../shared/url";

type IRouteComponentPropsAndRouteProps = RouteProps & RouteComponentProps<any>;

const privateRoutes = {
  "/c/:cluster/ns/:namespace/apps": AppListContainer,
  "/c/:cluster/ns/:namespace/apps/:releaseName": AppViewContainer,
  "/c/:cluster/ns/:namespace/apps/:releaseName/upgrade": AppUpgradeContainer,
  "/c/:cluster/ns/:namespace/apps/new/:repo/:id/versions/:version": AppNewContainer,
  "/c/:cluster/ns/:namespace/apps/new-from-:global(global)/:repo/:id/versions/:version": AppNewContainer,
  "/c/:cluster/ns/:namespace/catalog": CatalogContainer,
  "/c/:cluster/ns/:namespace/catalog/:repo": CatalogContainer,
  "/c/:cluster/ns/:namespace/charts/:repo/:id": ChartViewContainer,
  "/c/:cluster/ns/:namespace/:global(global)-charts/:repo/:id": ChartViewContainer,
  "/c/:cluster/ns/:namespace/charts/:repo/:id/versions/:version": ChartViewContainer,
  "/c/:cluster/ns/:namespace/:global(global)-charts/:repo/:id/versions/:version": ChartViewContainer,
  "/c/:cluster/config/brokers": ServiceBrokerListContainer,
  "/services/brokers/:brokerName/classes/:className": ServiceClassViewContainer,
  "/services/brokers/:brokerName/instances/ns/:namespace/:instanceName": ServiceInstanceViewContainer,
  "/services/classes": ServiceClassListContainer,
  "/ns/:namespace/services/instances": ServiceInstanceListContainer,
} as const;

// Public routes that don't require authentication
const routes = {
  "/login": LoginFormContainer,
} as const;

interface IRoutesProps extends IRouteComponentPropsAndRouteProps {
  namespace: string;
  cluster: string;
  authenticated: boolean;
  featureFlags: IFeatureFlags;
}

class Routes extends React.Component<IRoutesProps> {
  public static defaultProps = {
    featureFlags: { operators: false, ui: "hex" },
  };
  public render() {
    const reposPath = "/c/:cluster/ns/:namespace/config/repos";
    if (this.props.featureFlags.operators) {
      // Add routes related to operators
      Object.assign(privateRoutes, {
        "/c/:cluster/ns/:namespace/operators": OperatorsListContainer,
        "/c/:cluster/ns/:namespace/operators/:operator": OperatorViewContainer,
        "/c/:cluster/ns/:namespace/operators/new/:operator": OperatorNewContainer,
        "/c/:cluster/ns/:namespace/operators-instances/new/:csv/:crd": OperatorInstanceCreateContainer,
        "/c/:cluster/ns/:namespace/operators-instances/:csv/:crd/:instanceName": OperatorInstanceViewContainer,
        "/c/:cluster/ns/:namespace/operators-instances/:csv/:crd/:instanceName/update": OperatorInstanceUpdateContainer,
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
    if (this.props.cluster && this.props.namespace && this.props.authenticated) {
      return <Redirect to={app.apps.list(this.props.cluster, this.props.namespace)} />;
    }
    // There is not a default namespace, redirect to login page
    return <Redirect to={"/login"} />;
  };
}

export default Routes;
