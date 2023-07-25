// Copyright 2018-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { CdsButton } from "@cds/react/button";
import { CdsModal, CdsModalActions, CdsModalContent } from "@cds/react/modal";
import { ReactElement } from "react";
import { Redirect, useLocation } from "react-router-dom";
import "./RequireAuthentication.css";
import { useSelector } from "react-redux";
import { IStoreState } from "shared/types";

interface IRequireAuthenticationProps {
  children: ReactElement;
}

export function RequireAuthentication({ children }: IRequireAuthenticationProps): ReactElement {
  const { authenticated, sessionExpired } = useSelector((state: IStoreState) => state.auth);
  const refreshPage = () => {
    window.location.reload();
  };
  const location = useLocation();
  if (authenticated && children) {
    return children;
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
          <Redirect to={{ pathname: "/login", state: { from: location.pathname } }} />
        )}
      </>
    );
  }
}

export default RequireAuthentication;
