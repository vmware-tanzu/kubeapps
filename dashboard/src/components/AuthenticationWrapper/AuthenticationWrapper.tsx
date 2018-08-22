import { Location } from "history";
import * as React from "react";
import { RouteComponentProps, RouteProps } from "react-router";

type IRouteComponentPropsAndRouteProps = RouteProps & RouteComponentProps<any>;

interface IAuthenticationWrapperProps extends IRouteComponentPropsAndRouteProps {
  authenticated: boolean;
  onLoad: (location: Location) => any;
}

class AuthenticationWrapper extends React.Component<IAuthenticationWrapperProps> {
  public componentDidMount() {
    this.props.onLoad(this.props.location);
  }

  public render() {
    return this.props.authenticated ? this.props.children : <div>Loading</div>;
  }
}

export default AuthenticationWrapper;
