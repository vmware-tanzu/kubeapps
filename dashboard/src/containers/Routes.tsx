import * as React from "react";
import { Redirect, Route, RouteComponentProps, Switch } from "react-router";
import NotFound from "../components/NotFound";
import AppList from "./AppListContainer";
import AppNew from "./AppNewContainer";
import AppUpgrade from "./AppUpgradeContainer";
import AppView from "./AppViewContainer";
import ChartList from "./ChartListContainer";
import ChartView from "./ChartViewContainer";
import ClassListContainer from "./ClassListContainer";
import { ClassViewContainer } from "./ClassView";
import FunctionListContainer from "./FunctionListContainer";
import FunctionViewContainer from "./FunctionViewContainer";
import InstanceListViewContainer from "./InstanceListViewContainer";
import InstanceView from "./InstanceView";
import LoginFormContainer from "./LoginFormContainer";
import PrivateRouteContainer from "./PrivateRouteContainer";
import RepoListContainer from "./RepoListContainer";
import ServiceCatalogContainer from "./ServiceCatalogContainer";

interface IChartInfoProps {
  namespace: string;
}

class Routes extends React.Component<IChartInfoProps> {
  public static exactRoutes: {
    [route: string]: React.ComponentType<RouteComponentProps<any>> | React.ComponentType<any>;
  } = {
    // IMPORTANT: These routes should be added to the NGINX configuration
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
      <Switch>
        <Route exact={true} path="/" render={this.rootNamespacedRedirect} />
        <Route exact={true} path="/login" component={LoginFormContainer} />
        {Object.keys(Routes.exactRoutes).map(route => (
          <PrivateRouteContainer
            key={route}
            exact={true}
            path={route}
            component={Routes.exactRoutes[route]}
          />
        ))}
        <Route path="*" component={NotFound} />
      </Switch>
    );
  }
  public rootNamespacedRedirect = (props: any) => {
    return <Redirect to={`/apps/ns/${this.props.namespace}`} />;
  };
}

export default Routes;
