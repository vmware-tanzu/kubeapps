import { Location } from "history";
import * as React from "react";
import { Lock } from "react-feather";
import { Redirect } from "react-router";

import "./LoginForm.css";

interface ILoginFormProps {
  authenticated: boolean;
  authenticate: (token: string) => any;
  location: Location;
}

interface ILoginFormState {
  token?: string;
}

class LoginForm extends React.Component<ILoginFormProps, ILoginFormState> {
  public render() {
    if (this.props.authenticated) {
      const { from } = this.props.location.state || { from: { pathname: "/" } };
      return <Redirect to={from} />;
    }
    return (
      <section className="LoginForm">
        <div className="LoginForm__container padding-v-bigger bg-skew type-color-reverse">
          <div className="bg-skew__pattern bg-skew__pattern-dark">
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
                    />
                  </div>
                  <p>
                    <a className="button button-accent">Login</a>
                  </p>
                </form>
              </div>
            </div>
          </div>
        </div>
      </section>
    );
  }

  private handleSubmit = (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    const { token } = this.state;
    return token && this.props.authenticate(token);
  };

  private handleTokenChange = (e: React.FormEvent<HTMLInputElement>) => {
    this.setState({ token: e.currentTarget.value });
  };
}

export default LoginForm;
