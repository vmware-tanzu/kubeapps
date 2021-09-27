import AppList from "components/AppList/AppList";
import AppUpgrade from "components/AppUpgrade";
import AppView from "components/AppView";
import Catalog from "components/Catalog/Catalog";
import ChartView from "components/ChartView";
import AppRepoList from "components/Config/AppRepoList";
import DeploymentForm from "components/DeploymentForm";
import LoadingWrapper from "components/LoadingWrapper";
import React from "react";
import { Redirect, Route, RouteComponentProps, RouteProps, Switch } from "react-router";
import { app } from "shared/url";
import ApiDocs from "../../components/ApiDocs";
import NotFound from "../../components/NotFound";
// TODO(andresmgot): Containers should be no longer needed, replace them when possible
import LoginFormContainer from "../../containers/LoginFormContainer";
import OperatorInstanceCreateContainer from "../../containers/OperatorInstanceCreateContainer";
import OperatorInstanceUpdateContainer from "../../containers/OperatorInstanceUpdateContainer";
import OperatorInstanceViewContainer from "../../containers/OperatorInstanceViewContainer";
import OperatorNewContainer from "../../containers/OperatorNewContainer";
import OperatorsListContainer from "../../containers/OperatorsListContainer";
import OperatorViewContainer from "../../containers/OperatorViewContainer";
import PrivateRouteContainer from "../../containers/PrivateRouteContainer";

type IRouteComponentPropsAndRouteProps = RouteProps & RouteComponentProps<any>;

const privateRoutes = {
  "/c/:cluster/ns/:namespace/apps": AppList,
  "/c/:cluster/ns/:namespace/apps/:pluginName/:pluginVersion/:releaseName": AppView,
  "/c/:cluster/ns/:namespace/apps/:pluginName/:pluginVersion/:releaseName/upgrade": AppUpgrade,
  "/c/:cluster/ns/:namespace/apps/new/:repo/:pluginName/:pluginVersion/:id/versions/:version":
    DeploymentForm,
  "/c/:cluster/ns/:namespace/apps/new/:repo/:pluginName/:pluginVersion/:id/versions/":
    DeploymentForm,
  "/c/:cluster/ns/:namespace/apps/new-from-:global(global)/:repo/:pluginName/:pluginVersion/:id/versions/:version":
    DeploymentForm,
  "/c/:cluster/ns/:namespace/apps/new-from-:global(global)/:repo/:pluginName/:pluginVersion/:id/versions/":
    DeploymentForm,
  "/c/:cluster/ns/:namespace/catalog": Catalog,
  "/c/:cluster/ns/:namespace/catalog/:repo": Catalog,
  "/c/:cluster/ns/:namespace/charts/:repo/:pluginName/:pluginVersion/:id": ChartView,
  "/c/:cluster/ns/:namespace/charts/:repo/:pluginName/:pluginVersion/:id/versions/:version":
    ChartView,
  "/c/:cluster/ns/:namespace/:global(global)-charts/:repo/:pluginName/:pluginVersion/:id":
    ChartView,
  "/c/:cluster/ns/:namespace/:global(global)-charts/:repo/:pluginName/:pluginVersion/:id/versions/:version":
    ChartView,
  "/c/:cluster/ns/:namespace/operators": OperatorsListContainer,
  "/c/:cluster/ns/:namespace/operators/:operator": OperatorViewContainer,
  "/c/:cluster/ns/:namespace/operators/new/:operator": OperatorNewContainer,
  "/c/:cluster/ns/:namespace/operators-instances/new/:csv/:crd": OperatorInstanceCreateContainer,
  "/c/:cluster/ns/:namespace/operators-instances/:csv/:crd/:instanceName":
    OperatorInstanceViewContainer,
  "/c/:cluster/ns/:namespace/operators-instances/:csv/:crd/:instanceName/update":
    OperatorInstanceUpdateContainer,
  "/c/:cluster/ns/:namespace/config/repos": AppRepoList,
  "/docs": ApiDocs,
} as const;

// Public routes that don't require authentication
const routes = {
  "/login": LoginFormContainer,
} as const;

interface IRoutesProps extends IRouteComponentPropsAndRouteProps {
  cluster: string;
  currentNamespace: string;
  authenticated: boolean;
}

class Routes extends React.Component<IRoutesProps> {
  public render() {
    return (
      <Switch>
        <Route exact={true} path="/" render={this.rootNamespacedRedirect} />
        {Object.entries(routes).map(([route, component]) => (
          <Route key={route} exact={true} path={route} component={component} />
        ))}
        {Object.entries(privateRoutes).map(([route, component]) => (
          <PrivateRouteContainer key={route} exact={true} path={route} component={component} />
        ))}
        {/* If the route doesn't match any expected path redirect to a 404 page  */}
        <Route component={NotFound} />
      </Switch>
    );
  }
  private rootNamespacedRedirect = () => {
    if (this.props.authenticated) {
      if (!this.props.cluster || !this.props.currentNamespace) {
        return <LoadingWrapper className="margin-t-xxl" loadingText="Fetching Cluster Info..." />;
      }
      return (
        <Redirect
          to={{ pathname: app.apps.list(this.props.cluster, this.props.currentNamespace) }}
        />
      );
    }
    // There is not a default namespace, redirect to login page
    return <Redirect to={{ pathname: "/login" }} />;
  };
}

export default Routes;
