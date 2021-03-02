import { CdsButton } from "@cds/react/button";
import Modal from "components/Modal/Modal";
import * as React from "react";
import { Redirect, Route, RouteComponentProps, RouteProps } from "react-router";

import "./PrivateRoute.css";

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
    const refreshPage = () => {
      window.location.reload();
    };
    if (authenticated && Component) {
      return <Component {...props} />;
    } else {
      return sessionExpired ? (
        <Modal
          showModal={sessionExpired}
          hideCloseButton={true}
          onModalClose={refreshPage}
          footer={
            <CdsButton onClick={refreshPage} type="button">
              Reload
            </CdsButton>
          }
        >
          <p>Your session has expired or the connection has been lost, please reload the page.</p>
        </Modal>
      ) : (
        <Redirect to={{ pathname: "/login", state: { from: props.location } }} />
      );
    }
  };
}

export default PrivateRoute;
