// Copyright 2020-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { CdsButton } from "@cds/react/button";
import { useIntl } from "react-intl";

interface ILoginFormProps {
  token: string;
  authenticationError: string | undefined;
  handleTokenChange: (e: React.ChangeEvent<HTMLInputElement>) => void;
}

function TokenLogin(props: ILoginFormProps) {
  const intl = useIntl();
  return (
    <section className="title" aria-labelledby="login-title" aria-describedby="login-desc">
      <h3 id="login-title" className="welcome">
        {intl.formatMessage({ id: "login-title-welcome", defaultMessage: "Welcome to" })}
        <span>{intl.formatMessage({ id: "Kubeapps", defaultMessage: "Kubeapps" })}</span>
      </h3>
      <p id="login-desc" className="hint">
        {intl.formatMessage({
          id: "login-desc-token",
          defaultMessage: "Your cluster operator should provide you with a Kubernetes API token.",
        })}
      </p>
      <div className="login-group">
        <div className="clr-form-control">
          <label htmlFor="token" className="clr-control-label">
            {intl.formatMessage({ id: "Token", defaultMessage: "Token" })}
          </label>
          <div className="clr-control-container">
            <div className="clr-input-wrapper">
              <input
                type="password"
                id="token"
                name="token"
                placeholder={intl.formatMessage({
                  id: "paste-token-here",
                  defaultMessage: "Paste token here",
                })}
                className="clr-input"
                required={true}
                onChange={props.handleTokenChange}
                value={props.token}
              />
            </div>
          </div>
        </div>
        {props.authenticationError && (
          <div className="error active ">
            {intl.formatMessage({
              id: "error-login-token",
              defaultMessage:
                "There was an error connecting to the Kubernetes API. Please check that your token is valid.",
            })}
          </div>
        )}
        <div className="login-submit-button">
          <CdsButton id="login-submit-button" status="primary">
            {intl.formatMessage({ id: "Submit", defaultMessage: "Submit" })}
          </CdsButton>
        </div>
      </div>
    </section>
  );
}

export default TokenLogin;
