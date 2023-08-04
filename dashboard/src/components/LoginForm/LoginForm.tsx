// Copyright 2018-2023 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { CdsIcon } from "@cds/react/icon";
import actions from "actions";
import LoadingWrapper from "components/LoadingWrapper";
import qs from "qs";
import { useEffect, useState } from "react";
import { useIntl } from "react-intl";
import { useDispatch, useSelector } from "react-redux";
import * as ReactRouter from "react-router-dom";
import { ThunkDispatch } from "redux-thunk";
import { IStoreState } from "shared/types";
import { Action } from "typesafe-actions";
import "./LoginForm.css";
import OAuthLogin from "./OauthLogin";
import TokenLogin from "./TokenLogin";

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

  const oAuthEnabledAndConfigured = authProxyEnabled && oauthLoginURI !== "";

  if (oAuthEnabledAndConfigured && authProxySkipLoginPage) {
    // If the oauth login page should be skipped, simply redirect to the login URI.
    window.location.replace(oauthLoginURI);
  }

  useEffect(() => {
    if (oAuthEnabledAndConfigured) {
      dispatch(actions.auth.checkCookieAuthentication(cluster)).then(() => setCookieChecked(true));
    } else {
      setCookieChecked(true);
    }
  }, [dispatch, oAuthEnabledAndConfigured, cluster]);

  useEffect(() => {
    // In token auth, if not yet authenticated, if the token is passed in the query param,
    // use it straight away; if it fails, stop don't retry
    if (
      !oAuthEnabledAndConfigured &&
      !authenticated &&
      !queryParamTokenAttempted &&
      queryParamToken !== ""
    ) {
      setQueryParamTokenAttempted(true);
      dispatch(actions.auth.authenticate(cluster, queryParamToken, false));
    }
  }, [
    cluster,
    authenticated,
    dispatch,
    oAuthEnabledAndConfigured,
    queryParamToken,
    queryParamTokenAttempted,
  ]);

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
    const { from } = (location.state as any) || { from: { pathname: "/" } };
    return <ReactRouter.Navigate to={from} />;
  }

  const handleSubmit = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    return token && (await dispatch(actions.auth.authenticate(cluster, token, false)));
  };

  const handleTokenChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setToken(e.target.value);
  };

  return (
    <div className="login-wrapper">
      <form className="login clr-form" onSubmit={handleSubmit}>
        {oAuthEnabledAndConfigured ? (
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
