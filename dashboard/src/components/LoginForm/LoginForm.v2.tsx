import { ClarityIcons, infoCircleIcon } from "@clr/core/icon-shapes";
import { Location } from "history";
import * as React from "react";
import { Redirect } from "react-router";
import { CdsButton } from "../Clarity/clarity";
import { CdsIcon } from "../Clarity/clarity";

import LoadingWrapper from "../../components/LoadingWrapper";
import "./LoginForm.v2.css";

ClarityIcons.addIcons(infoCircleIcon);

interface ILoginFormProps {
  authenticated: boolean;
  authenticating: boolean;
  authenticationError: string | undefined;
  oauthLoginURI: string;
  authenticate: (token: string) => any;
  checkCookieAuthentication: () => void;
  location: Location;
}

interface ILoginFormState {
  token: string;
}

class LoginForm extends React.Component<ILoginFormProps, ILoginFormState> {
  public state: ILoginFormState = { token: "" };

  public componentDidMount() {
    if (this.props.oauthLoginURI) {
      this.props.checkCookieAuthentication();
    }
  }

  public render() {
    if (this.props.authenticating) {
      return <LoadingWrapper />;
    }
    if (this.props.authenticated) {
      const { from } = (this.props.location.state as any) || { from: { pathname: "/" } };
      return <Redirect to={from} />;
    }

    return (
      <div className="login-wrapper">
        <form className="login clr-form" onSubmit={this.handleSubmit}>
          {this.props.oauthLoginURI ? this.oauthLogin() : this.tokenLogin()}
          <div className="login-moreinfo">
            <a
              href="https://github.com/kubeapps/kubeapps/blob/master/docs/user/access-control.md"
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

  private handleSubmit = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    const { token } = this.state;
    return token && (await this.props.authenticate(token));
  };

  private handleTokenChange = (e: React.FormEvent<HTMLInputElement>) => {
    this.setState({ token: e.currentTarget.value });
  };

  private oauthLogin = () => {
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
          {this.props.authenticationError && (
            <div className="error active">
              There was an error connecting to the Kubernetes API. Please check that your token is
              valid.
            </div>
          )}
        </div>
        <a href={this.props.oauthLoginURI} className="login-oauth-button">
          <CdsButton id="login-submit-button" status="primary">
            Login via OIDC Provider
          </CdsButton>
        </a>
      </>
    );
  };

  private tokenLogin = () => {
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
                  onChange={this.handleTokenChange}
                  value={this.state.token}
                />
              </div>
            </div>
          </div>
          {this.props.authenticationError && (
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
  };
}

export default LoginForm;
