import * as React from "react";
import { Redirect, Route, RouteComponentProps, RouteProps } from "react-router";

type IRouteComponentPropsAndRouteProps = RouteProps & RouteComponentProps<any>;

interface IPrivateRouteProps extends IRouteComponentPropsAndRouteProps {
  authenticated: boolean;
}

class PrivateRoute extends React.Component<IPrivateRouteProps> {
  public render() {
    const { authenticated, component: Component, ...rest } = this.props;
    return <Route {...rest} render={this.renderRouteIfAuthenticated} />;
  }

  public renderRouteIfAuthenticated = (props: RouteComponentProps<any>) => {
    const { authenticated, component: Component } = this.props;
    return authenticated && Component ? (
      <Component {...props} />
    ) : (
      <Redirect to={{ pathname: "/login", state: { from: props.location } }} />
    );
  };
}

export default PrivateRoute;
