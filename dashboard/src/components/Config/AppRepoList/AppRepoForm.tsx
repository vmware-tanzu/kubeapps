import * as React from "react";
import { Redirect } from "react-router";

interface IAppRepoFormProps {
  message?: string;
  redirectTo?: string;
  install: (name: string, url: string, authHeader: string, customCA: string) => Promise<boolean>;
  onAfterInstall?: () => Promise<any>;
}

interface IAppRepoFormState {
  useRawAuthHeader: boolean;
  useBasicAuth: boolean;
  useBearerToken: boolean;
  user: string;
  password: string;
  authHeader: string;
  token: string;
  name: string;
  url: string;
  customCA: string;
}

export class AppRepoForm extends React.Component<IAppRepoFormProps, IAppRepoFormState> {
  public state = {
    useRawAuthHeader: false,
    useBasicAuth: false,
    useBearerToken: false,
    user: "",
    password: "",
    authHeader: "",
    token: "",
    name: "",
    url: "",
    customCA: "",
  };

  public render() {
    return (
      <form className="container padding-b-bigger" onSubmit={this.handleInstallClick}>
        <div className="row">
          <div className="col-12">
            <div>
              <h2>Add an App Repository</h2>
            </div>
            <div>
              <label>
                <span>Name:</span>
                <input
                  type="text"
                  placeholder="example"
                  value={this.state.name}
                  onChange={this.handleNameChange}
                  required={true}
                  pattern="[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*"
                  title="Use lower case alphanumeric characters, '-' or '.'"
                />
              </label>
            </div>
            <div>
              <label>
                <span>URL:</span>
                <input
                  type="url"
                  placeholder="https://charts.example.com/stable"
                  value={this.state.url}
                  onChange={this.handleURLChange}
                  required={true}
                />
              </label>
            </div>
            <div>
              <label>
                <span>Authorization (optional):</span>
                <div className="row">
                  <div className="col-3">
                    <label className="margin-l-big">
                      <input
                        type="radio"
                        name="auth"
                        value="None"
                        defaultChecked={true}
                        onChange={this.handleAuthRadioButtonChange}
                      />
                      None
                      <br />
                    </label>
                    <label>
                      <input
                        type="radio"
                        name="auth"
                        value="Basic"
                        onChange={this.handleAuthRadioButtonChange}
                      />
                      Basic Auth
                      <br />
                    </label>
                    <label>
                      <input
                        type="radio"
                        name="auth"
                        value="Bearer"
                        onChange={this.handleAuthRadioButtonChange}
                      />
                      Bearer Token
                      <br />
                    </label>
                    <label>
                      <input
                        type="radio"
                        name="auth"
                        value="HTTP Header"
                        onChange={this.handleAuthRadioButtonChange}
                      />
                      Custom
                      <br />
                    </label>
                  </div>
                  <div className="col-9">
                    <div hidden={!this.state.useBasicAuth}>
                      <span>User</span>
                      <br />
                      <input
                        type="text"
                        value={this.state.user}
                        onChange={this.handleUserChange}
                        placeholder="Username"
                      />
                      <span>Password</span>
                      <br />
                      <input
                        type="password"
                        value={this.state.password}
                        onChange={this.handlePasswordChange}
                        placeholder="Password"
                      />
                    </div>
                    <div hidden={!this.state.useBearerToken}>
                      Token <br />
                      <input
                        type="text"
                        value={this.state.token}
                        onChange={this.handleAuthTokenChange}
                      />
                    </div>
                    <div hidden={!this.state.useRawAuthHeader}>
                      Complete Authorization Header <br />
                      <input
                        type="text"
                        placeholder="Bearer xrxNcWghpRLdcPHFgVRM73rr4N7qjvjm"
                        value={this.state.authHeader}
                        onChange={this.handleAuthHeaderChange}
                      />
                    </div>
                  </div>
                </div>
              </label>
            </div>
            <div className="margin-t-big">
              <label>
                <span>Custom CA Certificate (optional):</span>
                <pre className="CertContainer">
                  <textarea
                    className="CertContent"
                    rows={4}
                    placeholder={
                      "-----BEGIN CERTIFICATE-----\n" + "...\n" + "-----END CERTIFICATE-----"
                    }
                    value={this.state.customCA}
                    onChange={this.handleCustomCAChange}
                  />
                </pre>
              </label>
            </div>
            <div>
              <button className="button button-primary" type="submit">
                Install Repo
              </button>
            </div>
            {this.props.redirectTo && <Redirect to={this.props.redirectTo} />}
          </div>
        </div>
      </form>
    );
  }

  private handleInstallClick = async (e: React.FormEvent<HTMLFormElement>) => {
    const { install, onAfterInstall } = this.props;
    const {
      name,
      url,
      authHeader,
      useRawAuthHeader,
      useBasicAuth,
      useBearerToken,
      token,
      user,
      password,
      customCA,
    } = this.state;
    e.preventDefault();
    let finalHeader = "";
    switch (true) {
      case useRawAuthHeader:
        finalHeader = authHeader;
        break;
      case useBasicAuth:
        finalHeader = `Basic ${btoa(`${user}:${password}`)}`;
        break;
      case useBearerToken:
        finalHeader = `Bearer ${token}`;
        break;
    }
    const installed = await install(name, url, finalHeader, customCA);
    if (installed && onAfterInstall) {
      await onAfterInstall();
    }
  };

  private handleNameChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    this.setState({ name: e.target.value });
  };

  private handleURLChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    this.setState({ url: e.target.value });
  };
  private handleAuthHeaderChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    this.setState({ authHeader: e.target.value });
  };
  private handleAuthTokenChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    this.setState({ token: e.target.value });
  };
  private handleCustomCAChange = (e: React.ChangeEvent<HTMLTextAreaElement>) => {
    this.setState({ customCA: e.target.value });
  };
  private handleAuthRadioButtonChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    switch (e.target.value) {
      case "Basic":
        this.setState({ useBasicAuth: true, useRawAuthHeader: false, useBearerToken: false });
        break;
      case "HTTP Header":
        this.setState({ useBasicAuth: false, useRawAuthHeader: true, useBearerToken: false });
        break;
      case "Bearer":
        this.setState({ useBasicAuth: false, useRawAuthHeader: false, useBearerToken: true });
        break;
      default:
        this.setState({ useBasicAuth: false, useRawAuthHeader: false, useBearerToken: false });
    }
  };

  private handleUserChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    this.setState({ user: e.target.value });
  };

  private handlePasswordChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    this.setState({ password: e.target.value });
  };
}
