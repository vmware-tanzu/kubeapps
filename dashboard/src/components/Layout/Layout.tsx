import * as React from "react";

import actions from "actions";
import Header from "components/Header";
import ErrorBoundaryContainer from "containers/ErrorBoundaryContainer";
import { useDispatch } from "react-redux";
import { Action } from "redux";
import { ThunkDispatch } from "redux-thunk";
import { IStoreState } from "shared/types";
import Clarity from "./Clarity";

import "./Layout.css";

function Layout({ children }: any) {
  const dispatch: ThunkDispatch<IStoreState, null, Action> = useDispatch();
  const logout = () => dispatch(actions.auth.logout());
  return (
    <section className="layout">
      <Clarity />
      <Header />
      <main>
        <div className="container kubeapps-main-container">
          <div className="content-area">
            <ErrorBoundaryContainer logout={logout}>{children}</ErrorBoundaryContainer>
          </div>
        </div>
      </main>
    </section>
  );
}

export default Layout;
