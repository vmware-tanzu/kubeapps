import { ClarityIcons, infoCircleIcon } from "@clr/core/icon-shapes";
import { Location } from "history";
import * as React from "react";
import { useEffect, useState } from "react";
import { Redirect } from "react-router";
import { CdsIcon } from "../Clarity/clarity";

import LoadingWrapper from "../../components/LoadingWrapper";
import "./LoginForm.v2.css";
import OAuthLogin from "./OauthLogin";
import TokenLogin from "./TokenLogin";

ClarityIcons.addIcons(infoCircleIcon);

export interface ILoginFormProps {
  cluster: string;
  authenticated: boolean;
  authenticating: boolean;
  authenticationError: string | undefined;
  oauthLoginURI: string;
  authenticate: (cluster: string, token: string) => any;
  checkCookieAuthentication: (cluster: string) => void;
  appVersion: string;
  location: Location;
}

function LoginForm(props: ILoginFormProps) {
  const [token, setToken] = useState("");
  const { oauthLoginURI, checkCookieAuthentication } = props;
  useEffect(() => {
    if (oauthLoginURI) {
      checkCookieAuthentication(props.cluster);
    }
  }, [oauthLoginURI, checkCookieAuthentication]);

  if (props.authenticating) {
    return <LoadingWrapper />;
  }
  if (props.authenticated) {
    const { from } = (props.location.state as any) || { from: { pathname: "/" } };
    return <Redirect to={from} />;
  }

  const handleSubmit = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    return token && (await props.authenticate(props.cluster, token));
  };

  const handleTokenChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setToken(e.target.value);
  };

  return (
    <div className="login-wrapper">
      <form className="login clr-form" onSubmit={handleSubmit}>
        {props.oauthLoginURI ? (
          <OAuthLogin
            authenticationError={props.authenticationError}
            oauthLoginURI={props.oauthLoginURI}
          />
        ) : (
          <TokenLogin
            authenticationError={props.authenticationError}
            token={token}
            handleTokenChange={handleTokenChange}
          />
        )}
        <div className="login-moreinfo">
          <a
            href={`https://github.com/kubeapps/kubeapps/blob/${props.appVersion}/docs/user/access-control.md`}
            target="_blank"
            rel="noopener noreferrer"
          >
            <CdsIcon shape="info-circle" />
            More Info
          </a>
        </div>
      </form>
    </div>
  );
}

export default LoginForm;
