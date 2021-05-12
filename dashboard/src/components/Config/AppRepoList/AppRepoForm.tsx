import { CdsButton } from "@cds/react/button";
import { CdsCheckbox } from "@cds/react/checkbox";
import { CdsControlMessage, CdsFormGroup } from "@cds/react/forms";

import { CdsInput } from "@cds/react/input";
import { CdsRadio, CdsRadioGroup } from "@cds/react/radio";
import { CdsTextarea } from "@cds/react/textarea";
import actions from "actions";
import Alert from "components/js/Alert";
import * as yaml from "js-yaml";
import { useEffect, useState } from "react";
import { useDispatch, useSelector } from "react-redux";
import { Action } from "redux";
import { ThunkDispatch } from "redux-thunk";
import { toFilterRule, toParams } from "shared/jq";
import { IAppRepository, IAppRepositoryFilter, ISecret, IStoreState } from "../../../shared/types";
import AppRepoAddDockerCreds from "./AppRepoAddDockerCreds";
import "./AppRepoForm.css";

interface IAppRepoFormProps {
  onSubmit: (
    name: string,
    url: string,
    type: string,
    description: string,
    authHeader: string,
    dockerRegCreds: string,
    customCA: string,
    syncJobPodTemplate: string,
    registrySecrets: string[],
    ociRepositories: string[],
    skipTLS: boolean,
    filter?: IAppRepositoryFilter,
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
const AUTH_METHOD_REGISTRY_SECRET = "registry";

const TYPE_HELM = "helm";
const TYPE_OCI = "oci";

export function AppRepoForm(props: IAppRepoFormProps) {
  const { onSubmit, onAfterInstall, namespace, kubeappsNamespace, repo, secret } = props;
  const dispatch: ThunkDispatch<IStoreState, null, Action> = useDispatch();

  const [authMethod, setAuthMethod] = useState(AUTH_METHOD_NONE);
  const [user, setUser] = useState("");
  const [password, setPassword] = useState("");
  const [authHeader, setAuthHeader] = useState("");
  const [token, setToken] = useState("");
  const [name, setName] = useState("");
  const [description, setDescription] = useState("");
  const [url, setURL] = useState("");
  const [customCA, setCustomCA] = useState("");
  const [syncJobPodTemplate, setSyncJobTemplate] = useState("");
  const [type, setType] = useState(TYPE_HELM);
  const [ociRepositories, setOCIRepositories] = useState("");
  const [skipTLS, setSkipTLS] = useState(!!repo?.spec?.tlsInsecureSkipVerify);
  const [filterNames, setFilterNames] = useState("");
  const [filterRegex, setFilterRegex] = useState(false);
  const [filterExclude, setFilterExclude] = useState(false);
  const [selectedImagePullSecret, setSelectedImagePullSecret] = useState("");
  const [validated, setValidated] = useState(undefined as undefined | boolean);

  useEffect(() => {
    dispatch(actions.repos.fetchImagePullSecrets(namespace));
  }, [dispatch, namespace]);

  const {
    repos: {
      imagePullSecrets,
      errors: { create: createError, update: updateError, validate: validationError },
      validating,
    },
    config: { appVersion },
  } = useSelector((state: IStoreState) => state);

  useEffect(() => {
    // Select the pull secrets if they are already selected in the existing repo
    imagePullSecrets.forEach(pullSecret => {
      const secretName = pullSecret.metadata.name;
      if (repo?.spec?.dockerRegistrySecrets?.some(s => s === secretName)) {
        setSelectedImagePullSecret(secretName);
      }
    });
  }, [imagePullSecrets, repo, selectedImagePullSecret]);

  useEffect(() => {
    if (repo) {
      setName(repo.metadata.name);
      setURL(repo.spec?.url || "");
      setType(repo.spec?.type || "");
      setDescription(repo.spec?.description || "");
      setSyncJobTemplate(
        repo.spec?.syncJobPodTemplate ? yaml.dump(repo.spec?.syncJobPodTemplate) : "",
      );
      setOCIRepositories(repo.spec?.ociRepositories?.join(", ") || "");
      setSkipTLS(!!repo.spec?.tlsInsecureSkipVerify);
      if (repo.spec?.filterRule?.jq) {
        const { names, regex, exclude } = toParams(repo.spec.filterRule);
        setFilterRegex(regex);
        setFilterExclude(exclude);
        setFilterNames(names);
      }
      if (repo.spec?.auth?.customCA || repo.spec?.auth?.header) {
        const secrets = [];
        if (repo.spec?.auth?.customCA) {
          secrets.push(repo.spec.auth.customCA.secretKeyRef.name);
        }
        if (repo.spec?.auth?.header && !secrets.includes(repo.spec.auth.header.secretKeyRef.name)) {
          secrets.push(repo.spec.auth.header.secretKeyRef.name);
        }
        secrets.forEach(s => dispatch(actions.repos.fetchRepoSecret(namespace, s)));
      }
    }
  }, [repo, namespace, dispatch]);

  useEffect(() => {
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
      if (secret.data[".dockerconfigjson"]) {
        setAuthMethod(AUTH_METHOD_REGISTRY_SECRET);
      }
    }
  }, [secret, authHeader]);

  const handleInstallClick = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    install();
  };

  const install = async () => {
    let finalHeader = "";
    let dockerRegCreds = "";
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
      case AUTH_METHOD_REGISTRY_SECRET:
        dockerRegCreds = selectedImagePullSecret;
    }
    const ociRepoList = ociRepositories.length ? ociRepositories.split(",").map(r => r.trim()) : [];
    // If the scheme is not specified, assume HTTPS. This is common for OCI registries
    const finalURL = url.startsWith("http") ? url : `https://${url}`;
    // If the validation already failed and we try to reinstall,
    // skip validation and force install
    const force = validated === false;
    let currentlyValidated = validated;
    if (!validated && !force) {
      currentlyValidated = await dispatch(
        actions.repos.validateRepo(
          finalURL,
          type,
          description,
          finalHeader,
          dockerRegCreds,
          customCA,
          ociRepoList,
          skipTLS,
        ),
      );
      setValidated(currentlyValidated);
    }
    let filter: IAppRepositoryFilter | undefined;
    if (type === TYPE_HELM && filterNames !== "") {
      filter = toFilterRule(filterNames, filterRegex, filterExclude);
    }
    if (currentlyValidated || force) {
      const success = await onSubmit(
        name,
        finalURL,
        type,
        description,
        finalHeader,
        dockerRegCreds,
        customCA,
        syncJobPodTemplate,
        selectedImagePullSecret.length ? [selectedImagePullSecret] : [],
        ociRepoList,
        skipTLS,
        filter,
      );
      if (success && onAfterInstall) {
        onAfterInstall();
      }
    }
  };

  const handleNameChange = (e: React.ChangeEvent<HTMLInputElement>) => setName(e.target.value);
  const handleDescriptionChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setDescription(e.target.value);
    setValidated(undefined);
  };
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
  const handleTypeRadioButtonChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setType(e.target.value);
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
  const handleOCIRepositoriesChange = (e: React.ChangeEvent<HTMLTextAreaElement>) => {
    setOCIRepositories(e.target.value);
    setValidated(undefined);
  };
  const handleSkipTLSChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setSkipTLS(!skipTLS);
    setValidated(undefined);
  };
  const handleFilterNames = (e: React.ChangeEvent<HTMLTextAreaElement>) => {
    setFilterNames(e.target.value);
  };
  const handleFilterRegex = (e: React.ChangeEvent<HTMLInputElement>) => {
    setFilterRegex(!filterRegex);
  };
  const handleFilterExclude = (e: React.ChangeEvent<HTMLInputElement>) => {
    setFilterExclude(!filterExclude);
  };

  const selectPullSecret = (imagePullSecret: string) => {
    setSelectedImagePullSecret(imagePullSecret);
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

  /* Only when using a namespace different than the Kubeapps namespace (Global)
    the repository can be associated with Docker Registry Credentials since
    the pull secret won't be available in all namespaces */
  const shouldEnableDockerRegistryCreds = namespace !== kubeappsNamespace;

  /* eslint-disable jsx-a11y/label-has-associated-control */
  return (
    <form onSubmit={handleInstallClick}>
      <CdsFormGroup layout="vertical">
        <CdsInput>
          <label>Name</label>
          <input
            id="kubeapps-repo-name"
            type="text"
            placeholder="example"
            value={name}
            onChange={handleNameChange}
            required={true}
            pattern="[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*"
            title="Use lower case alphanumeric characters, '-' or '.'"
            disabled={repo?.metadata.name ? true : false}
          />
        </CdsInput>
        <CdsInput>
          <label> URL </label>
          <input
            id="kubeapps-repo-url"
            type="url"
            placeholder="https://charts.example.com/stable"
            value={url}
            onChange={handleURLChange}
            required={true}
          />
        </CdsInput>
        <CdsInput>
          <label> Description (optional)</label>
          <input
            id="kubeapps-repo-description"
            type="text"
            placeholder="Description of the repository"
            value={description}
            onChange={handleDescriptionChange}
            required={false}
          />
        </CdsInput>

        <CdsRadioGroup layout="vertical">
          <label>Repository Type</label>
          <CdsControlMessage>Select the chart storage type.</CdsControlMessage>
          <CdsRadio>
            <label>Helm Repository</label>
            <input
              id="kubeapps-repo-type-helm"
              type="radio"
              name="type"
              value={TYPE_HELM}
              checked={type === TYPE_HELM}
              onChange={handleTypeRadioButtonChange}
            />
          </CdsRadio>
          <CdsRadio>
            <label>OCI Registry</label>
            <input
              id="kubeapps-repo-type-oci"
              type="radio"
              name="type"
              value={TYPE_OCI}
              checked={type === TYPE_OCI}
              onChange={handleTypeRadioButtonChange}
            />
          </CdsRadio>
        </CdsRadioGroup>

        <div cds-layout="grid gap:lg">
          <CdsRadioGroup cds-layout="col@xs:4">
            <label>Repository Authorization</label>
            <CdsRadio>
              <label>None (Public)</label>
              <input
                id="kubeapps-repo-auth-method-none"
                type="radio"
                name="auth"
                value={AUTH_METHOD_NONE}
                checked={authMethod === AUTH_METHOD_NONE}
                onChange={handleAuthRadioButtonChange}
              />
            </CdsRadio>
            <CdsRadio>
              <label>Basic Auth</label>
              <input
                id="kubeapps-repo-auth-method-basic"
                type="radio"
                name="auth"
                checked={authMethod === AUTH_METHOD_BASIC}
                value={AUTH_METHOD_BASIC}
                onChange={handleAuthRadioButtonChange}
              />
            </CdsRadio>
            <CdsRadio>
              <label>Bearer Token</label>
              <input
                id="kubeapps-repo-auth-method-bearer"
                type="radio"
                name="auth"
                value={AUTH_METHOD_BEARER}
                checked={authMethod === AUTH_METHOD_BEARER}
                onChange={handleAuthRadioButtonChange}
              />
            </CdsRadio>
            {shouldEnableDockerRegistryCreds && (
              <CdsRadio>
                <label>Use Docker Registry Credentials</label>
                <input
                  id="kubeapps-repo-auth-method-registry"
                  type="radio"
                  name="auth"
                  value={AUTH_METHOD_REGISTRY_SECRET}
                  checked={authMethod === AUTH_METHOD_REGISTRY_SECRET}
                  onChange={handleAuthRadioButtonChange}
                />
              </CdsRadio>
            )}
            <CdsRadio>
              <label>Custom</label>
              <input
                id="kubeapps-repo-auth-method-custom"
                type="radio"
                name="auth"
                value={AUTH_METHOD_CUSTOM}
                checked={authMethod === AUTH_METHOD_CUSTOM}
                onChange={handleAuthRadioButtonChange}
              />
            </CdsRadio>
          </CdsRadioGroup>
          <div cds-layout="col@xs:8">
            <div hidden={authMethod !== AUTH_METHOD_BASIC}>
              <CdsInput>
                <label>Username</label>
                <input
                  id="kubeapps-repo-username"
                  type="text"
                  value={user}
                  onChange={handleUserChange}
                  placeholder="Username"
                />
              </CdsInput>
              <br />
              <CdsInput>
                <label>Password</label>
                <input
                  id="kubeapps-repo-password"
                  type="password"
                  value={password}
                  onChange={handlePasswordChange}
                  placeholder="Password"
                />
              </CdsInput>
            </div>
            <div hidden={authMethod !== AUTH_METHOD_BEARER}>
              <CdsInput>
                <label>Token</label>
                <input
                  type="text"
                  value={token}
                  onChange={handleAuthTokenChange}
                  id="kubeapps-repo-token"
                />
              </CdsInput>
            </div>
            <div hidden={authMethod !== AUTH_METHOD_CUSTOM}>
              <CdsInput>
                <label>Complete Authorization Header</label>
                <input
                  id="kubeapps-repo-custom-header"
                  type="text"
                  placeholder="Bearer xrxNcWghpRLdcPHFgVRM73rr4N7qjvjm"
                  value={authHeader}
                  onChange={handleAuthHeaderChange}
                />
              </CdsInput>
            </div>
          </div>
        </div>

        <AppRepoAddDockerCreds
          imagePullSecrets={imagePullSecrets}
          selectPullSecret={selectPullSecret}
          selectedImagePullSecret={selectedImagePullSecret}
          namespace={namespace}
          appVersion={appVersion}
          disabled={!shouldEnableDockerRegistryCreds}
          required={authMethod === AUTH_METHOD_REGISTRY_SECRET}
        />

        {type === TYPE_OCI ? (
          <CdsTextarea>
            <label htmlFor="kubeapps-oci-repositories">List of Repositories</label>
            <CdsControlMessage>
              Include a list of comma-separated repositories that will be available in Kubeapps.
            </CdsControlMessage>
            <textarea
              id="kubeapps-oci-repositories"
              className="cds-textarea-fix"
              placeholder={"nginx, jenkins"}
              value={ociRepositories}
              onChange={handleOCIRepositoriesChange}
            />
          </CdsTextarea>
        ) : (
          <>
            <CdsTextarea>
              <label>Filter Applications (optional)</label>
              <CdsControlMessage>
                Comma-separated list of applications to be included or excluded (all will be
                included by default).
              </CdsControlMessage>
              <textarea
                className="cds-textarea-fix"
                placeholder={"nginx, jenkins"}
                value={filterNames}
                onChange={handleFilterNames}
              />
            </CdsTextarea>
            <CdsCheckbox className="ca-skip-tls">
              <label>Exclude Packages</label>
              <CdsControlMessage>Exclude packages matching the given filter</CdsControlMessage>
              <input type="checkbox" onChange={handleFilterExclude} checked={filterExclude} />
            </CdsCheckbox>
            <CdsCheckbox className="ca-skip-tls">
              <label>Regular Expression</label>
              <CdsControlMessage>
                Mark this box to treat the filter as a regular expression
              </CdsControlMessage>
              <input type="checkbox" onChange={handleFilterRegex} checked={filterRegex} />
            </CdsCheckbox>
          </>
        )}

        <CdsTextarea layout="vertical">
          <label>Custom CA Certificate (optional)</label>
          <textarea
            id="kubeapps-repo-custom-ca"
            placeholder={"-----BEGIN CERTIFICATE-----\n...\n-----END CERTIFICATE-----"}
            className="cds-textarea-fix"
            value={customCA}
            disabled={skipTLS}
            onChange={handleCustomCAChange}
          />
        </CdsTextarea>
        <CdsCheckbox className="ca-skip-tls">
          <label className="clr-control-label">Skip TLS Verification</label>
          <input
            id="kubeapps-repo-skip-tls"
            type="checkbox"
            checked={skipTLS}
            onChange={handleSkipTLSChange}
          />
        </CdsCheckbox>

        <CdsTextarea layout="vertical">
          <label>Custom Sync Job Template (optional)</label>
          <CdsControlMessage>
            It's possible to modify the default sync job. When doing this, the pre-validation is not
            supported. More info{" "}
            <a
              target="_blank"
              rel="noopener noreferrer"
              href={`https://github.com/kubeapps/kubeapps/blob/${appVersion}/docs/user/private-app-repository.md#modifying-the-synchronization-job`}
            >
              here
            </a>
            .
          </CdsControlMessage>
          <textarea
            id="kubeapps-repo-sync-job-tpl"
            rows={5}
            className="cds-textarea-fix"
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
        </CdsTextarea>
      </CdsFormGroup>

      {namespace === kubeappsNamespace && (
        <p>
          <strong>NOTE:</strong> This App Repository will be created in the "{kubeappsNamespace}"
          namespace and charts will be available in all namespaces for installation.
        </p>
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
      <div className="margin-t-xl">
        <CdsButton disabled={validating} onClick={install}>
          {validating
            ? "Validating..."
            : `${repo ? "Update" : "Install"} Repo ${validated === false ? "(force)" : ""}`}
        </CdsButton>
      </div>
    </form>
  );
}
