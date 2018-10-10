import * as React from "react";
import { Redirect, Route, RouteComponentProps, RouteProps, Switch } from "react-router";

import NotFound from "../../components/NotFound";
import AppList from "../../containers/AppListContainer";
import AppNew from "../../containers/AppNewContainer";
import AppUpgrade from "../../containers/AppUpgradeContainer";
import AppView from "../../containers/AppViewContainer";
import ChartList from "../../containers/ChartListContainer";
import ChartView from "../../containers/ChartViewContainer";
import ClassViewContainer from "../../containers/ClassViewContainer";
import InstanceListViewContainer from "../../containers/InstanceListViewContainer";
import InstanceView from "../../containers/InstanceViewContainer";
import LoginFormContainer from "../../containers/LoginFormContainer";
import PrivateRouteContainer from "../../containers/PrivateRouteContainer";
import RepoListContainer from "../../containers/RepoListContainer";
import ServiceCatalogContainer from "../../containers/ServiceCatalogContainer";
import ServiceClassListContainer from "../../containers/ServiceClassListContainer";

type IRouteComponentPropsAndRouteProps = RouteProps & RouteComponentProps<any>;

const privateRoutes: {
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
  "/services/brokers/:brokerName/classes/:className": ClassViewContainer,
  "/services/brokers/:brokerName/instances/ns/:namespace/:instanceName": InstanceView,
  "/services/classes": ServiceClassListContainer,
  "/services/instances/ns/:namespace": InstanceListViewContainer,
};

// Public routes that don't require authentication
const routes: {
  [route: string]: React.ComponentType<RouteComponentProps<any>> | React.ComponentType<any>;
} = {
  "/login": LoginFormContainer,
};

interface IRoutesProps extends IRouteComponentPropsAndRouteProps {
  namespace: string;
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
  private rootNamespacedRedirect = (props: any) => {
    return <Redirect to={`/apps/ns/${this.props.namespace}`} />;
  };
}

export default Routes;
