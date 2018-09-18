import * as React from "react";
import { Redirect, Route, RouteComponentProps, RouteProps, Switch } from "react-router";

import PrivateRouteContainer from "../../containers/PrivateRouteContainer";
import NotFound from "../NotFound";

type IRouteComponentPropsAndRouteProps = RouteProps & RouteComponentProps<any>;

interface IRoutesProps extends IRouteComponentPropsAndRouteProps {
  namespace: string;
  privateRoutes: {
    [route: string]: React.ComponentType<RouteComponentProps<any>> | React.ComponentType<any>;
  };
  routes: {
    [route: string]: React.ComponentType<RouteComponentProps<any>> | React.ComponentType<any>;
  };
}

class Routes extends React.Component<IRoutesProps> {
  public render() {
    return (
      <Switch>
        <Route exact={true} path="/" render={this.rootNamespacedRedirect} />
        {Object.keys(this.props.routes).map(route => (
          <Route key={route} exact={true} path={route} component={this.props.routes[route]} />
        ))}
        {Object.keys(this.props.privateRoutes).map(route => (
          <PrivateRouteContainer
            key={route}
            exact={true}
            path={route}
            component={this.props.privateRoutes[route]}
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
