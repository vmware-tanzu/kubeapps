// Copyright 2020-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { CdsButton } from "@cds/react/button";
import { useIntl } from "react-intl";

interface ILoginFormProps {
  authenticationError: string | undefined;
  oauthLoginURI: string;
}

function OAuthLogin(props: ILoginFormProps) {
  const intl = useIntl();
  return (
    <section className="title" aria-labelledby="login-title" aria-describedby="login-desc">
      <h3 id="login-title" className="welcome">
        {intl.formatMessage({ id: "login-title-welcome", defaultMessage: "Welcome to" })}
        <span>{intl.formatMessage({ id: "Kubeapps", defaultMessage: "Kubeapps" })}</span>
      </h3>
      <p id="login-desc" className="hint">
        {intl.formatMessage({
          id: "login-desc-oidc",
          defaultMessage: "Your cluster operator has enabled login via an authentication provider.",
        })}
      </p>
      <div className="login-group">
        {props.authenticationError && (
          <div className="error active">
            {intl.formatMessage({
              id: "error-login-token",
              defaultMessage:
                "There was an error connecting to the Kubernetes API. Please check that your token is valid.",
            })}
          </div>
        )}
        <a href={props.oauthLoginURI} className="login-submit-button">
          <CdsButton id="login-submit-button" status="primary">
            {intl.formatMessage({ id: "login-oidc", defaultMessage: "Login via OIDC Provider" })}
          </CdsButton>
        </a>
      </div>
    </section>
  );
}

export default OAuthLogin;
