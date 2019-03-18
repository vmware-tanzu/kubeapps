import * as React from "react";
import * as Modal from "react-modal";
import { Redirect, Route, RouteComponentProps, RouteProps } from "react-router";
import { Auth } from "shared/Auth";

type IRouteComponentPropsAndRouteProps = RouteProps & RouteComponentProps<any>;

interface IPrivateRouteProps extends IRouteComponentPropsAndRouteProps {
  authenticated: boolean;
}

interface IPrivateRouteState {
  previouslyAuthenticated: boolean;
}

class PrivateRoute extends React.Component<IPrivateRouteProps, IPrivateRouteState> {
  public state: IPrivateRouteState = {
    previouslyAuthenticated: false,
  };

  public render() {
    const { authenticated, component: Component, ...rest } = this.props;
    return <Route {...rest} render={this.renderRouteIfAuthenticated} />;
  }

  public renderRouteIfAuthenticated = (props: RouteComponentProps<any>) => {
    const { authenticated, component: Component } = this.props;
    if (authenticated && Component) {
      if (!this.state.previouslyAuthenticated) {
        this.setState({ previouslyAuthenticated: true });
      }
      return <Component {...props} />;
    }
    if (!authenticated && this.state.previouslyAuthenticated && Auth.usingOIDCToken()) {
      return (
        <Modal
          style={{
            content: {
              bottom: "auto",
              left: "50%",
              marginRight: "-50%",
              right: "auto",
              top: "50%",
              transform: "translate(-50%, -50%)",
            },
          }}
          isOpen={true}
        >
          <div>
            <div className="margin-b-normal">
              Your session has expired, please reload to refresh your credentials
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
