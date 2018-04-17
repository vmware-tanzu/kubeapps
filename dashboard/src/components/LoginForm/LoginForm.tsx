import { Location } from "history";
import * as React from "react";
import { Lock, X } from "react-feather";
import { Redirect } from "react-router";

import "./LoginForm.css";

interface ILoginFormProps {
  authenticated: boolean;
  authenticate: (token: string) => any;
  location: Location;
}

interface ILoginFormState {
  authenticating: boolean;
  token: string;
  error?: string;
}

class LoginForm extends React.Component<ILoginFormProps, ILoginFormState> {
  public state: ILoginFormState = { token: "", authenticating: false };
  public render() {
    if (this.props.authenticated) {
      const { from } = this.props.location.state || { from: { pathname: "/" } };
      return <Redirect to={from} />;
    }
    return (
      <section className="LoginForm">
        <div className="LoginForm__container padding-v-bigger bg-skew">
          <div className="container container-tiny">
            {this.state.error && (
              <div className="alert alert-error margin-c" role="alert">
                There was an error connecting to the Kubernetes API. Please check that your token is
                valid.
                <button className="alert__close" onClick={this.handleAlertClose}>
                  <X />
                </button>
              </div>
            )}
          </div>
          <div className="bg-skew__pattern bg-skew__pattern-dark type-color-reverse">
            <div className="container">
              <h2>
                <Lock /> Login
              </h2>
              <p>
                Your cluster operator should provide you with a Kubernetes API token.{" "}
                <a href="#">Click here</a> for more info on how to create and use Bearer Tokens.
              </p>
              <div className="bg-skew__content">
                <form onSubmit={this.handleSubmit}>
                  <div>
                    <label htmlFor="token">Kubernetes API Token</label>
                    <input
                      name="token"
                      id="token"
                      type="password"
                      placeholder="Token"
                      required={true}
                      onChange={this.handleTokenChange}
                      value={this.state.token}
                    />
                  </div>
                  <p>
                    <button
                      type="submit"
                      className="button button-accent"
                      disabled={this.state.authenticating}
                    >
                      Login
                    </button>
                  </p>
                </form>
              </div>
            </div>
          </div>
        </div>
      </section>
    );
  }

  private handleSubmit = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    this.setState({ authenticating: true });
    const { token } = this.state;
    try {
      return token && (await this.props.authenticate(token));
    } catch (e) {
      this.setState({ error: e.toString(), token: "", authenticating: false });
    }
  };

  private handleTokenChange = (e: React.FormEvent<HTMLInputElement>) => {
    this.setState({ token: e.currentTarget.value });
  };

  private handleAlertClose = (e: React.FormEvent<HTMLButtonElement>) => {
    this.setState({ error: undefined });
  };
}

export default LoginForm;
