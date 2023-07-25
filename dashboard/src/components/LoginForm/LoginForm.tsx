// Copyright 2018-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { CdsIcon } from "@cds/react/icon";
import qs from "qs";
import { useEffect, useState } from "react";
import { useIntl } from "react-intl";
import { useDispatch, useSelector } from "react-redux";
import * as ReactRouter from "react-router-dom";
import { IStoreState } from "shared/types";
import LoadingWrapper from "../../components/LoadingWrapper";
import "./LoginForm.css";
import OAuthLogin from "./OauthLogin";
import TokenLogin from "./TokenLogin";
import actions from "actions";
import { ThunkDispatch } from "redux-thunk";
import { Action } from "typesafe-actions";

function LoginForm() {
  const intl = useIntl();
  const [token, setToken] = useState("");
  const [cookieChecked, setCookieChecked] = useState(false);
  const [queryParamTokenAttempted, setQueryParamTokenAttempted] = useState(false);

  const dispatch: ThunkDispatch<IStoreState, null, Action> = useDispatch();

  const location = ReactRouter.useLocation();
  const queryParamToken =
    qs.parse(location.search, { ignoreQueryPrefix: true }).token?.toString() || "";

  const {
    config: { appVersion, authProxyEnabled, oauthLoginURI, authProxySkipLoginPage },
    clusters: { currentCluster: cluster },
    auth: { authenticated, authenticating, authenticationError },
  } = useSelector((state: IStoreState) => state);

  useEffect(() => {
    if (authProxyEnabled) {
      dispatch(actions.auth.checkCookieAuthentication(cluster)).then(() => setCookieChecked(true));
    } else {
      setCookieChecked(true);
    }
  }, [dispatch, authProxyEnabled, cluster]);

  useEffect(() => {
    // In token auth, if not yet authenticated, if the token is passed in the query param,
    // use it straight away; if it fails, stop don't retry
    if (!oauthLoginURI && !authenticated && !queryParamTokenAttempted && queryParamToken !== "") {
      setQueryParamTokenAttempted(true);
      dispatch(actions.auth.authenticate(cluster, queryParamToken, false));
    }
  }, [cluster, authenticated, dispatch, oauthLoginURI, queryParamToken, queryParamTokenAttempted]);

  if (authenticating || !cookieChecked) {
    return (
      <LoadingWrapper
        className="margin-t-xxl"
        loadingText={<h2>Welcome To Kubeapps</h2>}
        loaded={false}
      />
    );
  }
  if (authenticated) {
    // TODO(minelson): I don't think this redirect has been working for a while. Nothing
    // populates this location prop with the from attribute (from the history package) other
    // than a test.
    const { from } = (location.state as any) || { from: { pathname: "/" } };
    return <ReactRouter.Redirect to={from} />;
  }

  const handleSubmit = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    return token && (await dispatch(actions.auth.authenticate(cluster, token, false)));
  };

  const handleTokenChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setToken(e.target.value);
  };

  if (oauthLoginURI && authProxySkipLoginPage) {
    // If the oauth login page should be skipped, simply redirect to the login URI.
    window.location.replace(oauthLoginURI);
  }
  return (
    <div className="login-wrapper">
      <form className="login clr-form" onSubmit={handleSubmit}>
        {oauthLoginURI ? (
          <OAuthLogin authenticationError={authenticationError} oauthLoginURI={oauthLoginURI} />
        ) : (
          <TokenLogin
            authenticationError={authenticationError}
            token={token}
            handleTokenChange={handleTokenChange}
          />
        )}
        <div className="login-moreinfo">
          <a
            href={`https://github.com/vmware-tanzu/kubeapps/blob/${appVersion}/site/content/docs/latest/howto/access-control.md`}
            target="_blank"
            rel="noopener noreferrer"
          >
            <CdsIcon shape="info-circle" />
            {intl.formatMessage({ id: "more-info", defaultMessage: "More Info" })}
          </a>
        </div>
      </form>
    </div>
  );
}

export default LoginForm;
