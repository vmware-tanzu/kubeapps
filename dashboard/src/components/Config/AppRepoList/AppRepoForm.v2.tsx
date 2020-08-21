import actions from "actions";
import { CdsButton } from "components/Clarity/clarity";
import Alert from "components/js/Alert";
import Column from "components/js/Column";
import Row from "components/js/Row";
import * as yaml from "js-yaml";
import React, { useEffect, useState } from "react";
import { useDispatch, useSelector } from "react-redux";
import { Action } from "redux";
import { ThunkDispatch } from "redux-thunk";
import { definedNamespaces } from "../../../shared/Namespace";
import { IAppRepository, ISecret, IStoreState } from "../../../shared/types";
import AppRepoAddDockerCreds from "./AppRepoAddDockerCreds.v2";
import "./AppRepoForm.v2.css";

interface IAppRepoFormProps {
  onSubmit: (
    name: string,
    url: string,
    authHeader: string,
    customCA: string,
    syncJobPodTemplate: string,
    registrySecrets: string[],
  ) => Promise<boolean>;
  onAfterInstall?: () => void;
  namespace: string;
  kubeappsNamespace: string;
  repo?: IAppRepository;
  secret?: ISecret;
}

const AUTH_METHOD_NONE = "none";
const AUTH_METHOD_BASIC = "basic";
const AUTH_METHOD_BEARER = "bearer";
const AUTH_METHOD_CUSTOM = "custom";

export function AppRepoForm({
  onSubmit,
  onAfterInstall,
  namespace,
  kubeappsNamespace,
  repo,
  secret,
}: IAppRepoFormProps) {
  const dispatch: ThunkDispatch<IStoreState, null, Action> = useDispatch();

  const [authMethod, setAuthMethod] = useState(AUTH_METHOD_NONE);
  const [user, setUser] = useState("");
  const [password, setPassword] = useState("");
  const [authHeader, setAuthHeader] = useState("");
  const [token, setToken] = useState("");
  const [name, setName] = useState("");
  const [url, setURL] = useState("");
  const [customCA, setCustomCA] = useState("");
  const [syncJobPodTemplate, setSyncJobTemplate] = useState("");
  const [selectedImagePullSecrets, setSelectedImagePullSecrets] = useState(
    {} as { [key: string]: boolean },
  );
  const [validated, setValidated] = useState(undefined as undefined | boolean);

  const {
    imagePullSecrets,
    errors: { create: createError, update: updateError, validate: validationError },
    validating,
  } = useSelector((state: IStoreState) => state.repos);

  useEffect(() => {
    // Select the pull secrets if they are already selected in the existing repo
    imagePullSecrets.forEach(pullSecret => {
      const secretName = pullSecret.metadata.name;
      if (
        repo?.spec?.dockerRegistrySecrets?.some(s => s === secretName) &&
        !selectedImagePullSecrets[secretName]
      ) {
        setSelectedImagePullSecrets({
          ...selectedImagePullSecrets,
          [pullSecret.metadata.name]: true,
        });
      }
    });
  }, [imagePullSecrets, repo, selectedImagePullSecrets]);

  useEffect(() => {
    if (repo) {
      setName(repo.metadata.name);
      setURL(repo.spec?.url || "");
      setSyncJobTemplate(
        repo.spec?.syncJobPodTemplate ? yaml.safeDump(repo.spec?.syncJobPodTemplate) : "",
      );
      if (secret) {
        if (secret.data["ca.crt"]) {
          setCustomCA(atob(secret.data["ca.crt"]));
        }
        if (secret.data.authorizationHeader) {
          if (authHeader.startsWith("Basic")) {
            const userPass = atob(authHeader.split(" ")[1]).split(":");
            setUser(userPass[0]);
            setPassword(userPass[1]);
            setAuthMethod(AUTH_METHOD_BASIC);
          } else if (authHeader.startsWith("Bearer")) {
            setToken(authHeader.split(" ")[1]);
            setAuthMethod(AUTH_METHOD_BEARER);
          } else {
            setAuthMethod(AUTH_METHOD_CUSTOM);
            setAuthHeader(atob(secret.data.authorizationHeader));
          }
        }
      }
    }
  }, [repo, secret, authHeader]);

  const handleInstallClick = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    install();
  };

  const install = async () => {
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
    // If the validation already failed and we try to reinstall,
    // skip validation and force install
    const force = validated === false;
    let currentlyValidated = validated;
    if (!validated && !force) {
      currentlyValidated = await dispatch(actions.repos.validateRepo(url, finalHeader, customCA));
      setValidated(currentlyValidated);
    }
    if (currentlyValidated || force) {
      const imagePullSecretsNames = Object.keys(selectedImagePullSecrets).filter(
        s => selectedImagePullSecrets[s],
      );
      const success = await onSubmit(
        name,
        url,
        finalHeader,
        customCA,
        syncJobPodTemplate,
        imagePullSecretsNames,
      );
      if (success && onAfterInstall) {
        onAfterInstall();
      }
    }
  };

  const handleNameChange = (e: React.ChangeEvent<HTMLInputElement>) => setName(e.target.value);
  const handleURLChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setURL(e.target.value);
    setValidated(undefined);
  };
  const handleAuthHeaderChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setAuthHeader(e.target.value);
    setValidated(undefined);
  };
  const handleAuthTokenChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setToken(e.target.value);
    setValidated(undefined);
  };
  const handleCustomCAChange = (e: React.ChangeEvent<HTMLTextAreaElement>) => {
    setCustomCA(e.target.value);
    setValidated(undefined);
  };
  const handleAuthRadioButtonChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setAuthMethod(e.target.value);
    setValidated(undefined);
  };
  const handleUserChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setUser(e.target.value);
    setValidated(undefined);
  };
  const handlePasswordChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setPassword(e.target.value);
    setValidated(undefined);
  };
  const handleSyncJobPodTemplateChange = (e: React.ChangeEvent<HTMLTextAreaElement>) => {
    setSyncJobTemplate(e.target.value);
    setValidated(undefined);
  };

  const togglePullSecret = (imagePullSecret: string) => {
    return () => {
      setSelectedImagePullSecrets({
        ...selectedImagePullSecrets,
        [imagePullSecret]: !selectedImagePullSecrets[imagePullSecret],
      });
    };
  };

  const parseValidationError = (error: Error) => {
    let message = error.message;
    try {
      const parsedMessage = JSON.parse(message);
      if (parsedMessage.code && parsedMessage.message) {
        message = `Code: ${parsedMessage.code}. Message: ${parsedMessage.message}`;
      }
    } catch (e) {
      // Not a json message
    }
    return message;
  };

  return (
    <form onSubmit={handleInstallClick} className="app-repo-form">
      <h2>Add an App Repository</h2>
      <div className="clr-form-control">
        <label htmlFor="kubeapps-repo-name" className="clr-control-label">
          Name
        </label>
        <div className="clr-control-container">
          <div className="clr-input-wrapper">
            <input
              type="text"
              id="kubeapps-repo-name"
              className="clr-input"
              placeholder="example"
              value={name}
              onChange={handleNameChange}
              required={true}
              pattern="[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*"
              title="Use lower case alphanumeric characters, '-' or '.'"
              disabled={repo?.metadata.name ? true : false}
            />
          </div>
        </div>
      </div>
      <div className="clr-form-control">
        <label htmlFor="kubeapps-repo-url" className="clr-control-label">
          URL
        </label>
        <div className="clr-control-container">
          <div className="clr-input-wrapper">
            <input
              type="url"
              id="kubeapps-repo-url"
              placeholder="https://charts.example.com/stable"
              value={url}
              className="clr-input"
              onChange={handleURLChange}
              required={true}
            />
          </div>
        </div>
      </div>
      <div className="clr-form-control">
        <label className="clr-control-label">Repository Authorization (optional)</label>
        <span className="clr-form-description">
          Introduce the credentials to access the Chart repository if authentication is enabled.
        </span>
        <div className="clr-form-columns">
          <Row>
            <Column span={3}>
              <label
                htmlFor="kubeapps-repo-auth-method-none"
                className="clr-control-label clr-control-label-radio"
              >
                <input
                  type="radio"
                  id="kubeapps-repo-auth-method-none"
                  name="auth"
                  value={AUTH_METHOD_NONE}
                  checked={authMethod === AUTH_METHOD_NONE}
                  onChange={handleAuthRadioButtonChange}
                />
                None
                <br />
              </label>
              <label
                htmlFor="kubeapps-repo-auth-method-basic"
                className="clr-control-label clr-control-label-radio"
              >
                <input
                  type="radio"
                  id="kubeapps-repo-auth-method-basic"
                  name="auth"
                  checked={authMethod === AUTH_METHOD_BASIC}
                  value={AUTH_METHOD_BASIC}
                  onChange={handleAuthRadioButtonChange}
                />
                Basic Auth
                <br />
              </label>
              <label
                htmlFor="kubeapps-repo-auth-method-bearer"
                className="clr-control-label clr-control-label-radio"
              >
                <input
                  type="radio"
                  id="kubeapps-repo-auth-method-bearer"
                  name="auth"
                  value={AUTH_METHOD_BEARER}
                  checked={authMethod === AUTH_METHOD_BEARER}
                  onChange={handleAuthRadioButtonChange}
                />
                Bearer Token
                <br />
              </label>
              <label
                htmlFor="kubeapps-repo-auth-method-custom"
                className="clr-control-label clr-control-label-radio"
              >
                <input
                  type="radio"
                  id="kubeapps-repo-auth-method-custom"
                  name="auth"
                  value={AUTH_METHOD_CUSTOM}
                  checked={authMethod === AUTH_METHOD_CUSTOM}
                  onChange={handleAuthRadioButtonChange}
                />
                Custom
                <br />
              </label>
            </Column>
            <Column span={9}>
              <div className="column-valing-center clr-control-container">
                <div hidden={authMethod !== AUTH_METHOD_BASIC}>
                  <label className="clr-control-label" htmlFor="kubeapps-repo-username">
                    Username
                  </label>
                  <input
                    type="text"
                    id="kubeapps-repo-username"
                    className="clr-input"
                    value={user}
                    onChange={handleUserChange}
                    placeholder="Username"
                  />
                  <label className="clr-control-label" htmlFor="kubeapps-repo-password">
                    Password
                  </label>
                  <input
                    type="password"
                    id="kubeapps-repo-password"
                    className="clr-input"
                    value={password}
                    onChange={handlePasswordChange}
                    placeholder="Password"
                  />
                </div>
                <div hidden={authMethod !== AUTH_METHOD_BEARER}>
                  <label className="clr-control-label" htmlFor="kubeapps-repo-token">
                    Token
                  </label>
                  <input
                    id="kubeapps-repo-token"
                    className="clr-input"
                    type="text"
                    value={token}
                    onChange={handleAuthTokenChange}
                  />
                </div>
                <div hidden={authMethod !== AUTH_METHOD_CUSTOM}>
                  <label className="clr-control-label" htmlFor="kubeapps-repo-custom-header">
                    Complete Authorization Header
                  </label>
                  <input
                    type="text"
                    className="clr-input"
                    id="kubeapps-repo-custom-header"
                    placeholder="Bearer xrxNcWghpRLdcPHFgVRM73rr4N7qjvjm"
                    value={authHeader}
                    onChange={handleAuthHeaderChange}
                  />
                </div>
              </div>
            </Column>
          </Row>
        </div>
      </div>
      {/* Only when using a namespace different than the Kubeapps namespace (Global)
              the repository can be associated with Docker Registry Credentials since
              the pull secret won't be available in all namespaces */
      namespace !== kubeappsNamespace && (
        <div className="clr-form-control">
          <label className="clr-control-label">
            Associate Docker Registry Credentials (optional)
          </label>
          <span className="clr-form-description">
            Select existing secret(s) to access a private Docker registry and pull images from it.
            Note that this functionality is supported for Kubeapps with Helm3 only, more info{" "}
            <a
              href="https://github.com/kubeapps/kubeapps/blob/master/docs/user/private-app-repository.md"
              target="_blank"
              rel="noopener noreferrer"
            >
              here
            </a>
            .
          </span>
          <div className="clr-form-separator-sm">
            <AppRepoAddDockerCreds
              imagePullSecrets={imagePullSecrets}
              togglePullSecret={togglePullSecret}
              selectedImagePullSecrets={selectedImagePullSecrets}
              namespace={namespace}
            />
          </div>
        </div>
      )}
      <div className="clr-form-control">
        <label className="clr-control-label" htmlFor="kubeapps-repo-custom-ca">
          Custom CA Certificate (optional)
        </label>
        <div className="clr-control-container">
          <div className="clr-textarea-wrapper">
            <textarea
              id="kubeapps-repo-custom-ca"
              rows={4}
              className="clr-textarea"
              placeholder={"-----BEGIN CERTIFICATE-----\n...\n-----END CERTIFICATE-----"}
              value={customCA}
              onChange={handleCustomCAChange}
            />
          </div>
        </div>
      </div>
      <div className="clr-form-control">
        <label className="clr-control-label" htmlFor="kubeapps-repo-sync-job-tpl">
          Custom Sync Job Template (optional)
        </label>
        <span className="clr-form-description">
          It's possible to modify the default sync job. More info{" "}
          <a
            target="_blank"
            rel="noopener noreferrer"
            href="https://github.com/kubeapps/kubeapps/blob/master/docs/user/private-app-repository.md#modifying-the-synchronization-job"
          >
            here
          </a>
          . When modifying the default sync job, the pre-validation is not supported.
        </span>
        <div className="clr-control-container">
          <div className="clr-textarea-wrapper">
            <textarea
              id="kubeapps-repo-sync-job-tpl"
              rows={5}
              className="clr-textarea"
              placeholder={
                "spec:\n" +
                "  containers:\n" +
                "  - env:\n" +
                "    - name: FOO\n" +
                "      value: BAR\n"
              }
              value={syncJobPodTemplate}
              onChange={handleSyncJobPodTemplateChange}
            />
          </div>
        </div>
      </div>
      {(namespace === kubeappsNamespace || namespace === definedNamespaces.all) && (
        <div className="clr-form-description clr-form-separator">
          <strong>NOTE:</strong> This App Repository will be created in the "{kubeappsNamespace}"
          namespace and charts will be available in all namespaces for installation.
        </div>
      )}
      {validationError && (
        <Alert theme="danger">
          Validation Failed. Got: {parseValidationError(validationError)}
        </Alert>
      )}
      {createError && (
        <Alert theme="danger">
          An error occurred while creating the repository: {createError.message}
        </Alert>
      )}
      {updateError && (
        <Alert theme="danger">
          An error occurred while updating the repository: {updateError.message}
        </Alert>
      )}
      <div className="clr-form-separator">
        <CdsButton disabled={validating} onClick={install}>
          {validating
            ? "Validating..."
            : `${repo ? "Update" : "Install"} Repo ${validated === false ? "(force)" : ""}`}
        </CdsButton>
      </div>
    </form>
  );
}
