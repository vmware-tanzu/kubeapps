import * as React from "react";
import * as Modal from "react-modal";
import { Redirect, Route, RouteComponentProps, RouteProps } from "react-router";

type IRouteComponentPropsAndRouteProps = RouteProps & RouteComponentProps<any>;

interface IPrivateRouteProps extends IRouteComponentPropsAndRouteProps {
  authenticated: boolean;
  sessionExpired: boolean;
}

class PrivateRoute extends React.Component<IPrivateRouteProps> {
  public render() {
    const { authenticated, component: Component, ...rest } = this.props;
    return <Route {...rest} render={this.renderRouteIfAuthenticated} />;
  }

  public renderRouteIfAuthenticated = (props: RouteComponentProps<any>) => {
    const { sessionExpired, authenticated, component: Component } = this.props;
    if (authenticated && Component) {
      return <Component {...props} />;
    }
    if (sessionExpired) {
      return (
        <Modal className="centered-modal" isOpen={true}>
          <div>
            <div className="margin-b-normal">
              Your session has expired or the connection has been lost, please reload the page.
            </div>
            <div className="flex text-c">
              <button className="button" onClick={this.reload}>
                Reload
              </button>
            </div>
          </div>
        </Modal>
      );
    }
    return <Redirect to={{ pathname: "/login", state: { from: props.location } }} />;
  };

  private reload() {
    window.location.reload();
  }
}

export default PrivateRoute;
