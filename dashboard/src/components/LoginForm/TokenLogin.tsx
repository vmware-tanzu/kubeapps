import * as React from "react";
import { CdsButton } from "../Clarity/clarity";

interface ILoginFormProps {
  token: string;
  authenticationError: string | undefined;
  handleTokenChange: (e: React.ChangeEvent<HTMLInputElement>) => void;
}

function TokenLogin(props: ILoginFormProps) {
  return (
    <section className="title" aria-labelledby="login-title" aria-describedby="login-desc">
      <h3 id="login-title" className="welcome">
        Welcome to <span>Kubeapps</span>
      </h3>
      <p id="login-desc" className="hint">
        Your cluster operator should provide you with a Kubernetes API token.
      </p>
      <div className="login-group">
        <div className="clr-form-control">
          <label htmlFor="token" className="clr-control-label">
            Token
          </label>
          <div className="clr-control-container">
            <div className="clr-input-wrapper">
              <input
                type="password"
                id="token"
                placeholder="Paste token here"
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
            There was an error connecting to the Kubernetes API. Please check that your token is
            valid.
          </div>
        )}
        <div className="login-submit-button">
          <CdsButton id="login-submit-button" status="primary">
            Submit
          </CdsButton>
        </div>
      </div>
    </section>
  );
}

export default TokenLogin;
