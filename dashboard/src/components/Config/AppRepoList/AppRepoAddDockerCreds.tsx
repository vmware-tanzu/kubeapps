import * as React from "react";

import { ISecret } from "../../../shared/types";

interface IAppRepoFormProps {
  imagePullSecrets: ISecret[];
  togglePullSecret: (imagePullSecret: string) => () => void;
  selectedImagePullSecrets: { [key: string]: boolean };
  namespace: string;
  createDockerRegistrySecret: (
    name: string,
    user: string,
    password: string,
    email: string,
    server: string,
    namespace: string,
  ) => Promise<boolean>;
  fetchImagePullSecrets: (namespace: string) => void;
}

interface IAppRepoFormState {
  user: string;
  password: string;
  email: string;
  server: string;
  secretName: string;
  showSecretSubForm: boolean;
  creating: boolean;
}

export class AppRepoAddDockerCreds extends React.Component<IAppRepoFormProps, IAppRepoFormState> {
  public state: IAppRepoFormState = {
    secretName: "",
    user: "",
    password: "",
    email: "",
    server: "",
    showSecretSubForm: false,
    creating: false,
  };

  public render() {
    const { imagePullSecrets, togglePullSecret, selectedImagePullSecrets } = this.props;
    const { showSecretSubForm } = this.state;
    return (
      <div className="margin-l-big margin-t-normal">
        {imagePullSecrets.length > 0 ? (
          imagePullSecrets.map(secret => {
            return (
              <div key={secret.metadata.name}>
                <label
                  className="checkbox"
                  key={secret.metadata.name}
                  onChange={togglePullSecret(secret.metadata.name)}
                >
                  <input type="checkbox" checked={selectedImagePullSecrets[secret.metadata.name]} />
                  <span>{secret.metadata.name}</span>
                </label>
              </div>
            );
          })
        ) : (
          <div className="margin-b-small">No existing credentials found.</div>
        )}
        {this.state.showSecretSubForm && (
          <div className="secondary-input margin-t-big">
            <div className="row">
              <div className="col-1 margin-t-normal">
                <label htmlFor="kubeapps-docker-cred-secret-name">Secret Name</label>
              </div>
              <div className="col-11">
                <input
                  id="kubeapps-docker-cred-secret-name"
                  value={this.state.secretName}
                  onChange={this.handleSecretNameChange}
                  placeholder="Secret"
                  required={true}
                />
              </div>
            </div>
            <div className="row">
              <div className="col-1 margin-t-normal">
                <label htmlFor="kubeapps-docker-cred-server">Server</label>
              </div>
              <div className="col-11">
                <input
                  id="kubeapps-docker-cred-server"
                  value={this.state.server}
                  onChange={this.handleServerChange}
                  placeholder="https://index.docker.io/v1/"
                  required={true}
                />
              </div>
            </div>
            <div className="row">
              <div className="col-1 margin-t-normal">
                <label htmlFor="kubeapps-docker-cred-username">Username</label>
              </div>
              <div className="col-11">
                <input
                  id="kubeapps-docker-cred-username"
                  value={this.state.user}
                  onChange={this.handleUserChange}
                  placeholder="Username"
                  required={true}
                />
              </div>
            </div>
            <div className="row">
              <div className="col-1 margin-t-normal">
                <label htmlFor="kubeapps-docker-cred-password">Password</label>
              </div>
              <div className="col-11">
                <input
                  type="password"
                  id="kubeapps-docker-cred-password"
                  value={this.state.password}
                  onChange={this.handlePasswordChange}
                  placeholder="Password"
                  required={true}
                />
              </div>
            </div>
            <div className="row">
              <div className="col-1 margin-t-normal">
                <label htmlFor="kubeapps-docker-cred-email">Email</label>
              </div>
              <div className="col-11">
                <input
                  id="kubeapps-docker-cred-email"
                  value={this.state.email}
                  onChange={this.handleEmailChange}
                  placeholder="user@example.com"
                  required={true}
                />
              </div>
            </div>
            <div>
              <button
                className="button button-primary"
                type="button"
                disabled={this.state.creating}
                onClick={this.handleInstallClick}
              >
                {this.state.creating ? "Creating..." : "Submit"}
              </button>
              <button onClick={this.toggleCredSubForm} type="button" className="button">
                Cancel
              </button>
            </div>
          </div>
        )}
        {!showSecretSubForm && (
          <button onClick={this.toggleCredSubForm} className="button margin-t-normal" type="button">
            Add new credentials
          </button>
        )}
      </div>
    );
  }

  private handleUserChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    this.setState({ user: e.target.value });
  };

  private handleSecretNameChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    this.setState({ secretName: e.target.value });
  };

  private handlePasswordChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    this.setState({ password: e.target.value });
  };

  private handleEmailChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    this.setState({ email: e.target.value });
  };

  private handleServerChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    this.setState({ server: e.target.value });
  };

  private toggleCredSubForm = () => {
    this.setState({ showSecretSubForm: !this.state.showSecretSubForm });
  };

  private handleInstallClick = async () => {
    const { fetchImagePullSecrets, namespace } = this.props;
    const { secretName, user, password, email, server } = this.state;
    const success = await this.props.createDockerRegistrySecret(
      secretName,
      user,
      password,
      email,
      server,
      namespace,
    );
    if (success) {
      // re-fetch secrets
      fetchImagePullSecrets(namespace);
      this.setState({
        secretName: "",
        user: "",
        password: "",
        email: "",
        server: "",
        showSecretSubForm: false,
        creating: false,
      });
    }
  };
}

export default AppRepoAddDockerCreds;
