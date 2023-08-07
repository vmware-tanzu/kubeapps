// Copyright 2018-2023 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { CdsButton } from "@cds/react/button";
import actions from "actions";
import AlertGroup from "components/AlertGroup";
import Column from "components/Column";
import Header from "components/Header";
import React from "react";
import { ErrorBoundary, FallbackProps } from "react-error-boundary";
import { useDispatch, useSelector } from "react-redux";
import { Action } from "redux";
import { ThunkDispatch } from "redux-thunk";
import { IStoreState } from "shared/types";
import Clarity from "./Clarity";
import "./Layout.css";

function Layout({ children }: any) {
  const dispatch: ThunkDispatch<IStoreState, null, Action> = useDispatch();
  const logout = () => dispatch(actions.auth.logout());
  const {
    auth: { authenticated },
    kube: { kindsError },
    config: {
      featureFlags: { operators },
    },
    clusters,
  } = useSelector((state: IStoreState) => state);

  React.useEffect(() => {
    if (authenticated && clusters.currentCluster && operators) {
      dispatch(actions.kube.getResourceKinds(clusters.currentCluster));
    }
  }, [dispatch, authenticated, operators, clusters.currentCluster]);

  function fallbackRender({ error }: FallbackProps) {
    return (
      <Column>
        <AlertGroup
          status="danger"
          closable={false}
          alertActions={<CdsButton onClick={logout}>Log Out</CdsButton>}
        >
          An error occurred: {error.message}.
        </AlertGroup>
      </Column>
    );
  }

  return (
    <section className="layout">
      <Clarity />
      <Header />
      <main>
        <div className="container kubeapps-main-container">
          <div className="content-area">
            <ErrorBoundary fallbackRender={fallbackRender}>
              {kindsError && (
                <AlertGroup status="warning">
                  Unable to retrieve API info: {kindsError.message}.
                </AlertGroup>
              )}
              {children}
            </ErrorBoundary>
          </div>
        </div>
      </main>
    </section>
  );
}

export default Layout;
