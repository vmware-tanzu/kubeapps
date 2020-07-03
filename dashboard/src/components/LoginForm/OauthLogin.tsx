import * as React from "react";
import { CdsButton } from "../Clarity/clarity";

interface ILoginFormProps {
  authenticationError: string | undefined;
  oauthLoginURI: string;
}

function OAuthLogin(props: ILoginFormProps) {
  return (
    <>
      <section className="title">
        <h3 className="welcome">Welcome to</h3>
        Kubeapps
        <h5 className="hint">
          Your cluster operator has enabled login via an authentication provider.
        </h5>
      </section>
      <div className="login-group">
        {props.authenticationError && (
          <div className="error active">
            There was an error connecting to the Kubernetes API. Please check that your token is
            valid.
          </div>
        )}
      </div>
      <a href={props.oauthLoginURI} className="login-oauth-button">
        <CdsButton id="login-submit-button" status="primary">
          Login via OIDC Provider
        </CdsButton>
      </a>
    </>
  );
}

export default OAuthLogin;
