import * as React from "react";
import { Redirect } from "react-router";
import Hint from "../../../components/Hint";

interface IAppRepoFormProps {
  message?: string;
  redirectTo?: string;
  install: (
    name: string,
    url: string,
    authHeader: string,
    customCA: string,
    syncJobPodTemplate: string,
  ) => Promise<boolean>;
  onAfterInstall?: () => Promise<any>;
}

interface IAppRepoFormState {
  name: string;
  url: string;
  authMethod: string;
  user: string;
  password: string;
  authHeader: string;
  token: string;
  customCA: string;
  syncJobPodTemplate: string;
}

const AUTH_METHOD_NONE = "none";
const AUTH_METHOD_BASIC = "basic";
const AUTH_METHOD_BEARER = "bearer";
const AUTH_METHOD_CUSTOM = "custom";

export class AppRepoForm extends React.Component<IAppRepoFormProps, IAppRepoFormState> {
  public state = {
    authMethod: AUTH_METHOD_NONE,
    user: "",
    password: "",
    authHeader: "",
    token: "",
    name: "",
    url: "",
    customCA: "",
    syncJobPodTemplate: "",
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
              <label htmlFor="kubeapps-repo-name">Name:</label>
              <input
                type="text"
                id="kubeapps-repo-name"
                placeholder="example"
                value={this.state.name}
                onChange={this.handleNameChange}
                required={true}
                pattern="[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*"
                title="Use lower case alphanumeric characters, '-' or '.'"
              />
            </div>
            <div>
              <label htmlFor="kubeapps-repo-url">URL:</label>
              <input
                type="url"
                id="kubeapps-repo-url"
                placeholder="https://charts.example.com/stable"
                value={this.state.url}
                onChange={this.handleURLChange}
                required={true}
              />
            </div>
            <div>
              <span>Authorization (optional):</span>
              <div className="row">
                <div className="col-2">
                  <label className="margin-l-big" htmlFor="kubeapps-repo-auth-method-none">
                    <input
                      type="radio"
                      id="kubeapps-repo-auth-method-none"
                      name="auth"
                      value={AUTH_METHOD_NONE}
                      defaultChecked={true}
                      onChange={this.handleAuthRadioButtonChange}
                    />
                    None
                    <br />
                  </label>
                  <label htmlFor="kubeapps-repo-auth-method-basic">
                    <input
                      type="radio"
                      id="kubeapps-repo-auth-method-basic"
                      name="auth"
                      value={AUTH_METHOD_BASIC}
                      onChange={this.handleAuthRadioButtonChange}
                    />
                    Basic Auth
                    <br />
                  </label>
                  <label htmlFor="kubeapps-repo-auth-method-bearer">
                    <input
                      type="radio"
                      id="kubeapps-repo-auth-method-bearer"
                      name="auth"
                      value={AUTH_METHOD_BEARER}
                      onChange={this.handleAuthRadioButtonChange}
                    />
                    Bearer Token
                    <br />
                  </label>
                  <label htmlFor="kubeapps-repo-auth-method-custom">
                    <input
                      type="radio"
                      id="kubeapps-repo-auth-method-custom"
                      name="auth"
                      value={AUTH_METHOD_CUSTOM}
                      onChange={this.handleAuthRadioButtonChange}
                    />
                    Custom
                    <br />
                  </label>
                </div>
                <div className="col-10" aria-live="polite">
                  <div
                    hidden={this.state.authMethod !== AUTH_METHOD_BASIC}
                    className="secondary-input"
                  >
                    <label htmlFor="kubeapps-repo-username">Username</label>
                    <input
                      type="text"
                      id="kubeapps-repo-username"
                      value={this.state.user}
                      onChange={this.handleUserChange}
                      placeholder="Username"
                    />
                    <label htmlFor="kubeapps-repo-password">Password</label>
                    <input
                      type="password"
                      id="kubeapps-repo-password"
                      value={this.state.password}
                      onChange={this.handlePasswordChange}
                      placeholder="Password"
                    />
                  </div>
                  <div
                    hidden={this.state.authMethod !== AUTH_METHOD_BEARER}
                    className="secondary-input"
                  >
                    <label htmlFor="kubeapps-repo-token">Token</label>
                    <input
                      id="kubeapps-repo-token"
                      type="text"
                      value={this.state.token}
                      onChange={this.handleAuthTokenChange}
                    />
                  </div>
                  <div
                    hidden={this.state.authMethod !== AUTH_METHOD_CUSTOM}
                    className="secondary-input"
                  >
                    <label htmlFor="kubeapps-repo-custom-header">
                      Complete Authorization Header
                    </label>
                    <input
                      type="text"
                      id="kubeapps-repo-custom-header"
                      placeholder="Bearer xrxNcWghpRLdcPHFgVRM73rr4N7qjvjm"
                      value={this.state.authHeader}
                      onChange={this.handleAuthHeaderChange}
                    />
                  </div>
                </div>
              </div>
            </div>
            <div className="margin-t-big">
              <label>
                <span>Custom CA Certificate (optional):</span>
                <pre className="CodeContainer">
                  <textarea
                    className="Code"
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
            <div style={{ marginBottom: "1em" }}>
              <label htmlFor="syncJobPodTemplate">Custom Sync Job Template (optional)</label>
              <Hint reactTooltipOpts={{ delayHide: 1000 }} id="syncJobHelp">
                <span>
                  It's possible to modify the default sync job.
                  <br />
                  More info{" "}
                  <a
                    target="_blank"
                    href="https://github.com/kubeapps/kubeapps/blob/master/docs/user/private-app-repository.md#modifying-the-synchronization-job"
                  >
                    here
                  </a>
                </span>
              </Hint>
              <pre className="CodeContainer">
                <textarea
                  id="syncJobPodTemplate"
                  className="Code"
                  rows={4}
                  placeholder={
                    "spec:\n" +
                    "  containers:\n" +
                    "  - env:\n" +
                    "    - name: FOO\n" +
                    "      value: BAR\n"
                  }
                  value={this.state.syncJobPodTemplate}
                  onChange={this.handleSyncJobPodTemplateChange}
                />
              </pre>
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
      authMethod,
      token,
      user,
      password,
      customCA,
      syncJobPodTemplate,
    } = this.state;
    e.preventDefault();
    let finalHeader = "";
    switch (authMethod) {
      case AUTH_METHOD_CUSTOM:
        finalHeader = authHeader;
        break;
      case AUTH_METHOD_BASIC:
        finalHeader = `Basic ${btoa(`${user}:${password}`)}`;
        break;
      case AUTH_METHOD_BEARER:
        finalHeader = `Bearer ${token}`;
        break;
    }
    const installed = await install(name, url, finalHeader, customCA, syncJobPodTemplate);
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
    this.setState({ authMethod: e.target.value });
  };

  private handleUserChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    this.setState({ user: e.target.value });
  };

  private handlePasswordChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    this.setState({ password: e.target.value });
  };

  private handleSyncJobPodTemplateChange = (e: React.ChangeEvent<HTMLTextAreaElement>) => {
    this.setState({ syncJobPodTemplate: e.target.value });
  };
}
