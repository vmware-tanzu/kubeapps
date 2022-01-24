// Copyright 2018-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { CdsButton } from "@cds/react/button";
import { CdsModal, CdsModalActions, CdsModalContent } from "@cds/react/modal";
import React from "react";
import { Redirect, Route, RouteComponentProps, RouteProps } from "react-router-dom";
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
      return (
        <>
          {" "}
          {sessionExpired ? (
            <CdsModal closable={false} onCloseChange={refreshPage}>
              <CdsModalContent>
                {" "}
                <p>
                  Your session has expired or the connection has been lost, please reload the page.
                </p>
              </CdsModalContent>
              <CdsModalActions>
                <CdsButton onClick={refreshPage} type="button">
                  Reload
                </CdsButton>
              </CdsModalActions>
            </CdsModal>
          ) : (
            <Redirect to={{ pathname: "/login", state: { from: props.location } }} />
          )}
        </>
      );
    }
  };
}

export default PrivateRoute;
