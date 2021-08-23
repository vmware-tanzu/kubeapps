import actions from "actions";
import AlertGroup from "components/AlertGroup";
import Header from "components/Header";
import ErrorBoundaryContainer from "containers/ErrorBoundaryContainer";
import React from "react";
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
    clusters,
  } = useSelector((state: IStoreState) => state);

  React.useEffect(() => {
    if (authenticated && clusters.currentCluster) {
      dispatch(actions.kube.getResourceKinds(clusters.currentCluster));
    }
  }, [dispatch, authenticated, clusters.currentCluster]);

  return (
    <section className="layout">
      <Clarity />
      <Header />
      <main>
        <div className="container kubeapps-main-container">
          <div className="content-area">
            <ErrorBoundaryContainer logout={logout}>
              {kindsError && (
                <div className="margin-t-sm">
                  <AlertGroup status="warning" closable={true} size="sm">
                    Unable to retrieve API info: {kindsError.message}
                  </AlertGroup>
                </div>
              )}
              {children}
            </ErrorBoundaryContainer>
          </div>
        </div>
      </main>
    </section>
  );
}

export default Layout;
