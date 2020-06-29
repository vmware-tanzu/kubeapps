import * as yaml from "js-yaml";
import * as React from "react";
import { Redirect } from "react-router";
import Hint from "../../../components/Hint";
import { definedNamespaces } from "../../../shared/Namespace";
import { IAppRepository, ISecret } from "../../../shared/types";
import { UnexpectedErrorAlert } from "../../ErrorAlert";
import AppRepoAddDockerCreds from "./AppRepoAddDockerCreds";

interface IAppRepoFormProps {
  message?: string;
  redirectTo?: string;
  onSubmit: (
    name: string,
    url: string,
    authHeader: string,
    customCA: string,
    syncJobPodTemplate: string,
    registrySecrets: string[],
  ) => Promise<boolean>;
  validate: (url: string, authHeader: string, customCA: string) => Promise<any>;
  onAfterInstall?: () => Promise<any>;
  validating: boolean;
  validationError?: Error;
  imagePullSecrets: ISecret[];
  namespace: string;
  kubeappsNamespace: string;
  fetchImagePullSecrets: (namespace: string) => void;
  createDockerRegistrySecret: (
    name: string,
    user: string,
    password: string,
    email: string,
    server: string,
    namespace: string,
  ) => Promise<boolean>;
  repo?: IAppRepository;
  secret?: ISecret;
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
  selectedImagePullSecrets: { [key: string]: boolean };
  validated?: boolean;
}

const AUTH_METHOD_NONE = "none";
const AUTH_METHOD_BASIC = "basic";
const AUTH_METHOD_BEARER = "bearer";
const AUTH_METHOD_CUSTOM = "custom";

export class AppRepoForm extends React.Component<IAppRepoFormProps, IAppRepoFormState> {
  public state: IAppRepoFormState = {
    authMethod: AUTH_METHOD_NONE,
    user: "",
    password: "",
    authHeader: "",
    token: "",
    name: "",
    url: "",
    customCA: "",
    syncJobPodTemplate: "",
    selectedImagePullSecrets: {},
  };

  public componentDidMount() {
    if (this.props.repo) {
      const name = this.props.repo.metadata.name;
      const url = this.props.repo.spec?.url || "";
      const syncJobPodTemplate = this.props.repo.spec?.syncJobPodTemplate
        ? yaml.safeDump(this.props.repo.spec?.syncJobPodTemplate)
        : "";
      let customCA = "";
      let authHeader = "";
      let token = "";
      let user = "";
      let password = "";
      let authMethod = AUTH_METHOD_NONE;
      if (this.props.secret) {
        if (this.props.secret.data["ca.crt"]) {
          customCA = atob(this.props.secret.data["ca.crt"]);
        }
        if (this.props.secret.data.authorizationHeader) {
          authMethod = AUTH_METHOD_CUSTOM;
          authHeader = atob(this.props.secret.data.authorizationHeader);
          if (authHeader.startsWith("Basic")) {
            const userPass = atob(authHeader.split(" ")[1]).split(":");
            user = userPass[0];
            password = userPass[1];
            authMethod = AUTH_METHOD_BASIC;
          } else if (authHeader.startsWith("Bearer")) {
            token = authHeader.split(" ")[1];
            authMethod = AUTH_METHOD_BEARER;
          }
        }
      }
      this.setState({
        name,
        url,
        syncJobPodTemplate,
        customCA,
        authHeader,
        user,
        password,
        token,
        authMethod,
      });
    }

    this.parseSecrets(this.props.imagePullSecrets, this.props.repo);
  }

  public componentDidUpdate(prevProps: IAppRepoFormProps) {
    if (prevProps.imagePullSecrets !== this.props.imagePullSecrets) {
      this.parseSecrets(this.props.imagePullSecrets);
    }
  }

  public render() {
    const {
      repo,
      imagePullSecrets,
      namespace,
      kubeappsNamespace,
      createDockerRegistrySecret,
      fetchImagePullSecrets,
    } = this.props;
    const { authMethod, selectedImagePullSecrets } = this.state;
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
                disabled={this.props.repo?.metadata.name ? true : false}
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
              <p className="margin-b-small">Repository Authorization (optional):</p>
              <span className="AppRepoInputDescription">
                Introduce the credentials to access the Chart repository if authentication is
                enabled.
              </span>
              <div className="row">
                <div className="col-2">
                  <label className="margin-l-big" htmlFor="kubeapps-repo-auth-method-none">
                    <input
                      type="radio"
                      id="kubeapps-repo-auth-method-none"
                      name="auth"
                      value={AUTH_METHOD_NONE}
                      checked={authMethod === AUTH_METHOD_NONE}
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
                      checked={authMethod === AUTH_METHOD_BASIC}
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
                      checked={authMethod === AUTH_METHOD_BEARER}
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
                      checked={authMethod === AUTH_METHOD_CUSTOM}
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
            {/* Only when using a namespace different than the Kubeapps namespace (Global)
              the repository can be associated with Docker Registry Credentials since
              the pull secret won't be available in all namespaces */
            namespace !== kubeappsNamespace && (
              <div>
                <p className="margin-b-small">Associate Docker Registry Credentials (optional):</p>
                <span className="AppRepoInputDescription">
                  Select existing secret(s) to access a private Docker registry and pull images from
                  it. Note that this functionality is supported for Kubeapps with Helm3 only, more
                  info{" "}
                  <a
                    href="https://github.com/kubeapps/kubeapps/blob/master/docs/user/private-app-repository.md"
                    target="_blank"
                    rel="noopener noreferrer"
                  >
                    here
                  </a>
                  .
                  <AppRepoAddDockerCreds
                    imagePullSecrets={imagePullSecrets}
                    togglePullSecret={this.togglePullSecret}
                    selectedImagePullSecrets={selectedImagePullSecrets}
                    createDockerRegistrySecret={createDockerRegistrySecret}
                    namespace={namespace}
                    fetchImagePullSecrets={fetchImagePullSecrets}
                  />
                </span>
              </div>
            )}
            <div className="margin-t-big">
              <label>
                <span>Custom CA Certificate (optional):</span>
                <pre className="CodeContainer">
                  <textarea
                    className="Code"
                    rows={4}
                    placeholder={"-----BEGIN CERTIFICATE-----\n...\n-----END CERTIFICATE-----"}
                    value={this.state.customCA}
                    onChange={this.handleCustomCAChange}
                  />
                </pre>
              </label>
            </div>
            <div>
              <label htmlFor="syncJobPodTemplate">Custom Sync Job Template (optional)</label>
              <Hint reactTooltipOpts={{ delayHide: 1000 }} id="syncJobHelp">
                <span>
                  It's possible to modify the default sync job. More info{" "}
                  <a
                    target="_blank"
                    rel="noopener noreferrer"
                    href="https://github.com/kubeapps/kubeapps/blob/master/docs/user/private-app-repository.md#modifying-the-synchronization-job"
                  >
                    here
                  </a>
                  <br />
                  When modifying the default sync job, the pre-validation is not supported.
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
            {(namespace === kubeappsNamespace || namespace === definedNamespaces.all) && (
              <div className="margin-b-normal">
                <strong>NOTE:</strong> This App Repository will be created in the "
                {kubeappsNamespace}" namespace and charts will be available in all namespaces for
                installation.
              </div>
            )}
            {this.props.validationError && this.parseValidationError(this.props.validationError)}
            <div>
              <button
                className="button button-primary"
                type="submit"
                disabled={this.props.validating}
              >
                {this.props.validating
                  ? "Validating..."
                  : `${repo ? "Update" : "Install"} Repo ${
                      this.state.validated === false ? "(force)" : ""
                    }`}
              </button>
            </div>
            {this.props.redirectTo && <Redirect to={this.props.redirectTo} />}
          </div>
        </div>
      </form>
    );
  }

  private handleInstallClick = async (e: React.FormEvent<HTMLFormElement>) => {
    const { onSubmit, onAfterInstall, validate } = this.props;
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
      selectedImagePullSecrets,
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
    let validated = this.state.validated;
    // If the validation already failed and we try to reinstall,
    // skip validation and force install
    const force = validated === false;
    if (!validated && !force) {
      validated = await validate(url, finalHeader, customCA);
      this.setState({ validated });
    }
    if (validated || force) {
      const imagePullSecrets = Object.keys(selectedImagePullSecrets).filter(
        s => selectedImagePullSecrets[s],
      );
      const success = await onSubmit(
        name,
        url,
        finalHeader,
        customCA,
        syncJobPodTemplate,
        imagePullSecrets,
      );
      if (success && onAfterInstall) {
        await onAfterInstall();
      }
    }
  };

  private handleNameChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    this.setState({ name: e.target.value });
  };

  private handleURLChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    this.setState({ url: e.target.value, validated: undefined });
  };
  private handleAuthHeaderChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    this.setState({ authHeader: e.target.value, validated: undefined });
  };
  private handleAuthTokenChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    this.setState({ token: e.target.value, validated: undefined });
  };
  private handleCustomCAChange = (e: React.ChangeEvent<HTMLTextAreaElement>) => {
    this.setState({ customCA: e.target.value, validated: undefined });
  };
  private handleAuthRadioButtonChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    this.setState({ authMethod: e.target.value });
  };

  private handleUserChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    this.setState({ user: e.target.value, validated: undefined });
  };

  private handlePasswordChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    this.setState({ password: e.target.value, validated: undefined });
  };

  private handleSyncJobPodTemplateChange = (e: React.ChangeEvent<HTMLTextAreaElement>) => {
    this.setState({ syncJobPodTemplate: e.target.value, validated: undefined });
  };

  private togglePullSecret = (imagePullSecret: string) => {
    return () => {
      const { selectedImagePullSecrets } = this.state;
      this.setState({
        selectedImagePullSecrets: {
          ...selectedImagePullSecrets,
          [imagePullSecret]: !selectedImagePullSecrets[imagePullSecret],
        },
      });
    };
  };

  // Select the pull secrets based on the current status and if they are already
  // selected in the existing repo info
  private parseSecrets = (secrets: ISecret[], repo?: IAppRepository) => {
    const selectedImagePullSecrets = this.state.selectedImagePullSecrets;
    secrets.forEach(secret => {
      let selected = false;
      // If it has been already selected
      if (selectedImagePullSecrets[secret.metadata.name]) {
        selected = true;
      }
      // Or if it's already selected in the existing repo
      if (repo?.spec?.dockerRegistrySecrets?.some(s => s === secret.metadata.name)) {
        selected = true;
      }
      selectedImagePullSecrets[secret.metadata.name] = selected;
    });
    this.setState({ selectedImagePullSecrets });
  };

  private parseValidationError = (error: Error) => {
    let message = error.message;
    try {
      const parsedMessage = JSON.parse(message);
      if (parsedMessage.code && parsedMessage.message) {
        message = `Code: ${parsedMessage.code}. Message: ${parsedMessage.message}`;
      }
    } catch (e) {
      // Not a json message
    }
    return <UnexpectedErrorAlert title="Validation Failed. Got:" text={message} raw={true} />;
  };
}
