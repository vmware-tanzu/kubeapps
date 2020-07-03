import * as React from "react";
import { CdsButton } from "../Clarity/clarity";

interface ILoginFormProps {
  token: string;
  authenticationError: string | undefined;
  handleTokenChange: (e: React.ChangeEvent<HTMLInputElement>) => void;
}

function TokenLogin(props: ILoginFormProps) {
  return (
    <>
      <section className="title">
        <h3 className="welcome">Welcome to</h3>
        Kubeapps
        <h5 className="hint">
          Your cluster operator should provide you with a Kubernetes API token.
        </h5>
      </section>
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
      </div>
      <CdsButton id="login-submit-button" status="primary">
        Submit
      </CdsButton>
    </>
  );
}

export default TokenLogin;
