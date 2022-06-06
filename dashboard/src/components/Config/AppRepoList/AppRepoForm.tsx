// Copyright 2019-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import {
  CdsAccordion,
  CdsAccordionContent,
  CdsAccordionHeader,
  CdsAccordionPanel,
} from "@cds/react/accordion";
import { CdsButton } from "@cds/react/button";
import { CdsCheckbox } from "@cds/react/checkbox";
import { CdsControlMessage, CdsFormGroup } from "@cds/react/forms";
import { CdsInput } from "@cds/react/input";
import { CdsRadio, CdsRadioGroup } from "@cds/react/radio";
import { CdsTextarea } from "@cds/react/textarea";
import actions from "actions";
import Alert from "components/js/Alert";
import {
  PackageRepositoryAuth_PackageRepositoryAuthType,
  PackageRepositoryReference,
} from "gen/kubeappsapis/core/packages/v1alpha1/repositories";
import { Plugin } from "gen/kubeappsapis/core/plugins/v1alpha1/plugins";
import { RepositoryCustomDetails } from "gen/kubeappsapis/plugins/helm/packages/v1alpha1/helm";
import { useEffect, useRef, useState } from "react";
import { useDispatch, useSelector } from "react-redux";
import { Action } from "redux";
import { ThunkDispatch } from "redux-thunk";
import { toFilterRule, toParams } from "shared/jq";
import { PackageRepositoriesService } from "shared/PackageRepositoriesService";
import Secret from "shared/Secret";
import { IAppRepositoryFilter, ISecret, IStoreState } from "shared/types";
import { getPluginByName, getPluginPackageName, PluginNames } from "shared/utils";
import AppRepoAddDockerCreds from "./AppRepoAddDockerCreds";
import "./AppRepoForm.css";
interface IAppRepoFormProps {
  onSubmit: (
    name: string,
    plugin: Plugin,
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
    passCredentials: boolean,
    authMethod: PackageRepositoryAuth_PackageRepositoryAuthType,
    interval: number,
    username: string,
    password: string,
    filter?: IAppRepositoryFilter,
  ) => Promise<boolean>;
  onAfterInstall?: () => void;
  namespace: string;
  kubeappsNamespace: string;
  packageRepoRef?: PackageRepositoryReference;
}

// temporary enum for the type of package repository storage
enum RepositoryStorageTypes {
  PACKAGE_REPOSITORY_STORAGE_HELM = "helm",
  PACKAGE_REPOSITORY_STORAGE_OCI = "oci",
  PACKAGE_REPOSITORY_STORAGE_CARVEL_INLINE = "inline",
  PACKAGE_REPOSITORY_STORAGE_CARVEL_IMAGE = "image",
  PACKAGE_REPOSITORY_STORAGE_CARVEL_IMGPKGBUNDLE = "imgpkgBundle",
  PACKAGE_REPOSITORY_STORAGE_CARVEL_HTTP = "http",
  PACKAGE_REPOSITORY_STORAGE_CARVEL_GIT = "git",
}

export function AppRepoForm(props: IAppRepoFormProps) {
  const { onSubmit, onAfterInstall, namespace, kubeappsNamespace, packageRepoRef } = props;
  const isInstallingRef = useRef(false);
  const dispatch: ThunkDispatch<IStoreState, null, Action> = useDispatch();

  const {
    repos: {
      repo,
      errors: { create: createError, update: updateError, validate: validationError },
      validating,
    },
    config: { appVersion },
    clusters: { currentCluster },
  } = useSelector((state: IStoreState) => state);

  const [authMethod, setAuthMethod] = useState(
    PackageRepositoryAuth_PackageRepositoryAuthType.PACKAGE_REPOSITORY_AUTH_TYPE_UNSPECIFIED,
  );
  const [user, setUser] = useState("");
  const [password, setPassword] = useState("");
  const [authHeader, setAuthHeader] = useState("");
  const [token, setToken] = useState("");
  const [name, setName] = useState("");
  const [description, setDescription] = useState("");
  const [url, setURL] = useState("");
  const [customCA, setCustomCA] = useState("");
  const [syncJobPodTemplate, setSyncJobTemplate] = useState("");
  const [type, setType] = useState("");
  const [plugin, setPlugin] = useState(getPluginByName(PluginNames.PACKAGES_HELM));
  const [ociRepositories, setOCIRepositories] = useState("");
  const [skipTLS, setSkipTLS] = useState(!!repo?.tlsConfig?.insecureSkipVerify);
  const [passCredentials, setPassCredentials] = useState(!!repo?.auth?.passCredentials);
  const [interval, setInterval] = useState(3600);
  const [filterNames, setFilterNames] = useState("");
  const [filterRegex, setFilterRegex] = useState(false);
  const [filterExclude, setFilterExclude] = useState(false);
  const [secret, setSecret] = useState<ISecret>();
  const [selectedImagePullSecret, setSelectedImagePullSecret] = useState("");
  const [imagePullSecrets, setImagePullSecrets] = useState<string[]>([]);
  const [validated, setValidated] = useState(undefined as undefined | boolean);

  const [accordion, setAccordion] = useState([true, false, false, false]);

  const toggleAccordion = (section: number) => {
    const items = [...accordion];
    items[section] = !items[section];
    setAccordion(items);
  };

  useEffect(() => {
    if (packageRepoRef) {
      dispatch(actions.repos.fetchRepo(packageRepoRef));
    }
  }, [dispatch, packageRepoRef]);

  useEffect(() => {
    fetchImagePullSecrets(currentCluster, namespace);
  }, [dispatch, namespace, currentCluster]);

  async function fetchImagePullSecrets(cluster: string, repoNamespace: string) {
    setImagePullSecrets(await Secret.getDockerConfigSecretNames(cluster, repoNamespace));
  }

  useEffect(() => {
    // Select the pull secrets if they are already selected in the existing repo
    imagePullSecrets.forEach(secretName => {
      if (repo?.auth?.secretRef?.key === secretName) {
        setSelectedImagePullSecret(secretName);
      }
    });
  }, [imagePullSecrets, repo]);

  useEffect(() => {
    if (repo) {
      setName(repo.name);
      setURL(repo.url);
      setType(repo.type);
      setPlugin(repo.packageRepoRef?.plugin || ({ name: "", version: "" } as Plugin));
      setDescription(repo.description);
      setSkipTLS(!!repo.tlsConfig?.insecureSkipVerify);
      setPassCredentials(!!repo.auth?.passCredentials);
      setInterval(repo.interval);
      const repositoryCustomDetails = repo.customDetail as Partial<RepositoryCustomDetails>;
      // setSyncJobTemplate(
      //   repositoryCustomDetails?.syncJobPodTemplate
      //     ? yaml.dump(repositoryCustomDetails?.syncJobPodTemplate)
      //     : "",
      // );
      setOCIRepositories(repositoryCustomDetails?.ociRepositories?.join(", ") || "");
      if (repositoryCustomDetails?.filterRule?.jq) {
        const { names, regex, exclude } = toParams(repositoryCustomDetails.filterRule!);
        setFilterRegex(regex);
        setFilterExclude(exclude);
        setFilterNames(names);
      }

      if (repo?.tlsConfig?.certAuthority || repo?.auth?.header) {
        fetchRepoSecret(currentCluster, repo.packageRepoRef?.context?.namespace || "", repo.name);
      }
    }
  }, [repo, namespace, currentCluster, dispatch]);

  async function fetchRepoSecret(cluster: string, repoNamespace: string, repoName: string) {
    setSecret(await PackageRepositoriesService.getSecretForRepo(cluster, repoNamespace, repoName));
  }

  useEffect(() => {
    if (secret) {
      if (secret.data["ca.crt"]) {
        setCustomCA(Buffer.from(secret.data["ca.crt"], "base64")?.toString());
      }
      if (secret.data.authorizationHeader) {
        if (authHeader?.startsWith("Basic")) {
          const userPass = Buffer.from(authHeader?.split(" ")[1], "base64")?.toString()?.split(":");
          setUser(userPass[0]);
          setPassword(userPass[1]);
          setAuthMethod(
            PackageRepositoryAuth_PackageRepositoryAuthType.PACKAGE_REPOSITORY_AUTH_TYPE_BASIC_AUTH,
          );
        } else if (authHeader?.startsWith("Bearer")) {
          setToken(authHeader?.split(" ")[1]);
          setAuthMethod(
            PackageRepositoryAuth_PackageRepositoryAuthType.PACKAGE_REPOSITORY_AUTH_TYPE_BEARER,
          );
        } else {
          setAuthMethod(
            PackageRepositoryAuth_PackageRepositoryAuthType.PACKAGE_REPOSITORY_AUTH_TYPE_CUSTOM,
          );
          setAuthHeader(Buffer.from(secret.data.authorizationHeader, "base64")?.toString());
        }
      }
      if (secret.data[".dockerconfigjson"]) {
        setAuthMethod(
          PackageRepositoryAuth_PackageRepositoryAuthType.PACKAGE_REPOSITORY_AUTH_TYPE_DOCKER_CONFIG_JSON,
        );
      }
    }
  }, [secret, authHeader]);

  const handleInstallClick = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    install();
  };

  const install = async () => {
    if (isInstallingRef.current) {
      // Another installation is ongoing
      return;
    }
    isInstallingRef.current = true;
    let finalHeader = "";
    let dockerRegCreds = "";
    switch (authMethod) {
      case PackageRepositoryAuth_PackageRepositoryAuthType.PACKAGE_REPOSITORY_AUTH_TYPE_CUSTOM:
        finalHeader = authHeader;
        break;
      case PackageRepositoryAuth_PackageRepositoryAuthType.PACKAGE_REPOSITORY_AUTH_TYPE_BEARER:
        finalHeader = `Bearer ${token}`;
        break;
      case PackageRepositoryAuth_PackageRepositoryAuthType.PACKAGE_REPOSITORY_AUTH_TYPE_DOCKER_CONFIG_JSON:
        dockerRegCreds = selectedImagePullSecret;
    }
    const ociRepoList = ociRepositories.length
      ? ociRepositories?.split(",").map(r => r.trim())
      : [];
    let finalURL = url;
    // If the scheme is not specified, assume HTTPS. This is common for OCI registries
    // unless using the kapp plugin, which explicitly should not include https:// protocol prefix
    if (plugin?.name !== PluginNames.PACKAGES_KAPP && !url?.startsWith("http")) {
      finalURL = `https://${url}`;
    }
    // If the validation already failed and we try to reinstall,
    // skip validation and force install
    const force = validated === false;
    let currentlyValidated = validated;
    // Validation feature is only available in the Helm plugin for now
    if (plugin?.name === PluginNames.PACKAGES_HELM && !validated && !force) {
      currentlyValidated = await dispatch(
        actions.repos.validateRepo(
          name,
          plugin,
          namespace,
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
          passCredentials,
          authMethod,
          interval,
          user,
          password,
        ),
      );
      setValidated(currentlyValidated);
      // If using any other plugin, force the validation to pass
    } else if (plugin?.name !== PluginNames.PACKAGES_HELM) {
      currentlyValidated = true;
      setValidated(currentlyValidated);
    }
    let filter: IAppRepositoryFilter | undefined;
    if (type === RepositoryStorageTypes.PACKAGE_REPOSITORY_STORAGE_HELM && filterNames !== "") {
      filter = toFilterRule(filterNames, filterRegex, filterExclude);
    }
    if (currentlyValidated || force) {
      const success = await onSubmit(
        name,
        plugin,
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
        passCredentials,
        authMethod,
        interval,
        user,
        password,
        filter,
      );
      if (success && onAfterInstall) {
        onAfterInstall();
      }
    }
    isInstallingRef.current = false;
  };

  const handleNameChange = (e: React.ChangeEvent<HTMLInputElement>) => setName(e.target.value);
  const handleDescriptionChange = (e: React.ChangeEvent<HTMLInputElement>) =>
    setDescription(e.target.value);
  const handleIntervalChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setInterval(Number(e.target.value));
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
    setAuthMethod(PackageRepositoryAuth_PackageRepositoryAuthType[e.target.value]);
    setValidated(undefined);
  };
  const handleTypeRadioButtonChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setType(e.target.value);
    setValidated(undefined);
  };
  const handlePluginRadioButtonChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setPlugin(getPluginByName(e.target.value));
    if (!type) {
      // if no type, suggest one per plugin
      switch (getPluginByName(e.target.value)?.name) {
        case PluginNames.PACKAGES_HELM:
        case PluginNames.PACKAGES_FLUX:
          setType(RepositoryStorageTypes.PACKAGE_REPOSITORY_STORAGE_HELM);
          break;
        case PluginNames.PACKAGES_KAPP:
          setType(RepositoryStorageTypes.PACKAGE_REPOSITORY_STORAGE_CARVEL_IMGPKGBUNDLE);
          break;
      }
    }
    // TODO(agamez): workaround until Flux plugin also supports OCI artifacts
    if (getPluginByName(e.target.value)?.name === PluginNames.PACKAGES_FLUX) {
      setType(RepositoryStorageTypes.PACKAGE_REPOSITORY_STORAGE_HELM);
    }

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
  const handleSkipTLSChange = (_e: React.ChangeEvent<HTMLInputElement>) => {
    setSkipTLS(!skipTLS);
    setValidated(undefined);
  };
  const handlePassCredentialsChange = (_e: React.ChangeEvent<HTMLInputElement>) => {
    setPassCredentials(!passCredentials);
    setValidated(undefined);
  };
  const handleFilterNames = (e: React.ChangeEvent<HTMLTextAreaElement>) => {
    setFilterNames(e.target.value);
  };
  const handleFilterRegex = (_e: React.ChangeEvent<HTMLInputElement>) => {
    setFilterRegex(!filterRegex);
  };
  const handleFilterExclude = (_e: React.ChangeEvent<HTMLInputElement>) => {
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
    } catch (e: any) {
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
    <>
      <form onSubmit={handleInstallClick}>
        <CdsAccordion>
          <CdsAccordionPanel expanded={accordion[0]}>
            <CdsAccordionHeader onClick={() => toggleAccordion(0)}>
              Basic information
            </CdsAccordionHeader>
            <CdsAccordionContent>
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
                    disabled={repo?.name ? true : false}
                  />
                </CdsInput>
                <CdsInput>
                  <label> URL </label>
                  <input
                    id="kubeapps-repo-url"
                    type="text"
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
                {/* TODO(agamez): these plugin selectors should be loaded
                based on the current plugins that are loaded in the cluster */}
                <CdsRadioGroup layout="vertical">
                  <label>Packaging Format:</label>
                  <CdsControlMessage>Select the plugin to use.</CdsControlMessage>
                  <CdsRadio>
                    <label>{getPluginPackageName(PluginNames.PACKAGES_HELM)}</label>
                    <input
                      id="kubeapps-plugin-helm"
                      type="radio"
                      name="plugin"
                      value={PluginNames.PACKAGES_HELM}
                      checked={plugin?.name === PluginNames.PACKAGES_HELM}
                      onChange={handlePluginRadioButtonChange}
                      disabled={repo.packageRepoRef?.plugin ? true : false}
                    />
                  </CdsRadio>
                  <CdsRadio>
                    <label>{getPluginPackageName(PluginNames.PACKAGES_FLUX)}</label>
                    <input
                      id="kubeapps-plugin-fluxv2"
                      type="radio"
                      name="plugin"
                      value={PluginNames.PACKAGES_FLUX}
                      checked={plugin?.name === PluginNames.PACKAGES_FLUX}
                      onChange={handlePluginRadioButtonChange}
                      disabled={repo.packageRepoRef?.plugin ? true : false}
                    />
                  </CdsRadio>
                  <CdsRadio>
                    <label>{getPluginPackageName(PluginNames.PACKAGES_KAPP)}</label>
                    <input
                      id="kubeapps-plugin-kappcontroller"
                      type="radio"
                      name="plugin"
                      value={PluginNames.PACKAGES_KAPP}
                      checked={plugin?.name === PluginNames.PACKAGES_KAPP}
                      onChange={handlePluginRadioButtonChange}
                      disabled={repo.packageRepoRef?.plugin ? true : false}
                    />
                  </CdsRadio>
                </CdsRadioGroup>
                <CdsRadioGroup layout="vertical">
                  <label>Package Storage Type</label>
                  <CdsControlMessage>Select the package storage type.</CdsControlMessage>
                  {plugin?.name !== PluginNames.PACKAGES_KAPP ? (
                    <>
                      <CdsRadio>
                        <label>Helm Repository</label>
                        <input
                          id="kubeapps-repo-type-helm"
                          type="radio"
                          name="type"
                          value={RepositoryStorageTypes.PACKAGE_REPOSITORY_STORAGE_HELM}
                          checked={type === RepositoryStorageTypes.PACKAGE_REPOSITORY_STORAGE_HELM}
                          disabled={!!repo?.type}
                          onChange={handleTypeRadioButtonChange}
                        />
                      </CdsRadio>
                      <CdsRadio>
                        <label>OCI Registry</label>
                        <input
                          id="kubeapps-repo-type-oci"
                          type="radio"
                          name="type"
                          // TODO(agamez): workaround until Flux plugin also supports OCI artifacts
                          disabled={plugin?.name === PluginNames.PACKAGES_FLUX || !!repo?.type}
                          value={RepositoryStorageTypes.PACKAGE_REPOSITORY_STORAGE_OCI}
                          checked={type === RepositoryStorageTypes.PACKAGE_REPOSITORY_STORAGE_OCI}
                          onChange={handleTypeRadioButtonChange}
                        />
                      </CdsRadio>
                    </>
                  ) : (
                    <>
                      <CdsRadio>
                        <label>Imgpkg Bundle</label>
                        <input
                          id="kubeapps-repo-type-imgpkgbundle"
                          type="radio"
                          name="type"
                          disabled={!!repo?.type}
                          value={
                            RepositoryStorageTypes.PACKAGE_REPOSITORY_STORAGE_CARVEL_IMGPKGBUNDLE
                          }
                          checked={
                            type ===
                            RepositoryStorageTypes.PACKAGE_REPOSITORY_STORAGE_CARVEL_IMGPKGBUNDLE
                          }
                          onChange={handleTypeRadioButtonChange}
                        />
                      </CdsRadio>
                      <CdsRadio>
                        <label>Inline</label>
                        <input
                          id="kubeapps-repo-type-inline"
                          type="radio"
                          name="type"
                          disabled={!!repo?.type}
                          value={RepositoryStorageTypes.PACKAGE_REPOSITORY_STORAGE_CARVEL_INLINE}
                          checked={
                            type === RepositoryStorageTypes.PACKAGE_REPOSITORY_STORAGE_CARVEL_INLINE
                          }
                          onChange={handleTypeRadioButtonChange}
                        />
                      </CdsRadio>
                      <CdsRadio>
                        <label>Image</label>
                        <input
                          id="kubeapps-repo-type-image"
                          type="radio"
                          name="type"
                          disabled={!!repo?.type}
                          value={RepositoryStorageTypes.PACKAGE_REPOSITORY_STORAGE_CARVEL_IMAGE}
                          checked={
                            type === RepositoryStorageTypes.PACKAGE_REPOSITORY_STORAGE_CARVEL_IMAGE
                          }
                          onChange={handleTypeRadioButtonChange}
                        />
                      </CdsRadio>
                      <CdsRadio>
                        <label>HTTP</label>
                        <input
                          id="kubeapps-repo-type-http"
                          type="radio"
                          name="type"
                          disabled={!!repo?.type}
                          value={RepositoryStorageTypes.PACKAGE_REPOSITORY_STORAGE_CARVEL_HTTP}
                          checked={
                            type === RepositoryStorageTypes.PACKAGE_REPOSITORY_STORAGE_CARVEL_HTTP
                          }
                          onChange={handleTypeRadioButtonChange}
                        />
                      </CdsRadio>
                      <CdsRadio>
                        <label>Git</label>
                        <input
                          id="kubeapps-repo-type-git"
                          type="radio"
                          name="type"
                          disabled={!!repo?.type}
                          value={RepositoryStorageTypes.PACKAGE_REPOSITORY_STORAGE_CARVEL_GIT}
                          checked={
                            type === RepositoryStorageTypes.PACKAGE_REPOSITORY_STORAGE_CARVEL_GIT
                          }
                          onChange={handleTypeRadioButtonChange}
                        />
                      </CdsRadio>
                    </>
                  )}
                </CdsRadioGroup>
              </CdsFormGroup>
            </CdsAccordionContent>
          </CdsAccordionPanel>

          <CdsAccordionPanel expanded={accordion[1]}>
            <CdsAccordionHeader onClick={() => toggleAccordion(1)}>
              Authentication
            </CdsAccordionHeader>
            <CdsAccordionContent>
              <CdsFormGroup layout="vertical">
                <div cds-layout="grid gap:lg">
                  <CdsRadioGroup cds-layout="col@xs:4">
                    <label>Repository Authorization</label>
                    <CdsRadio>
                      <label>None (Public)</label>
                      <input
                        id="kubeapps-repo-auth-method-none"
                        type="radio"
                        name="auth"
                        value={
                          PackageRepositoryAuth_PackageRepositoryAuthType[
                            PackageRepositoryAuth_PackageRepositoryAuthType
                              .PACKAGE_REPOSITORY_AUTH_TYPE_UNSPECIFIED
                          ]
                        }
                        checked={
                          authMethod ===
                          PackageRepositoryAuth_PackageRepositoryAuthType.PACKAGE_REPOSITORY_AUTH_TYPE_UNSPECIFIED
                        }
                        onChange={handleAuthRadioButtonChange}
                      />
                    </CdsRadio>
                    <CdsRadio>
                      <label>Basic Auth</label>
                      <input
                        id="kubeapps-repo-auth-method-basic"
                        type="radio"
                        name="auth"
                        value={
                          PackageRepositoryAuth_PackageRepositoryAuthType[
                            PackageRepositoryAuth_PackageRepositoryAuthType
                              .PACKAGE_REPOSITORY_AUTH_TYPE_BASIC_AUTH
                          ]
                        }
                        checked={
                          authMethod ===
                          PackageRepositoryAuth_PackageRepositoryAuthType.PACKAGE_REPOSITORY_AUTH_TYPE_BASIC_AUTH
                        }
                        onChange={handleAuthRadioButtonChange}
                      />
                    </CdsRadio>
                    <CdsRadio>
                      <label>Bearer Token</label>
                      <input
                        id="kubeapps-repo-auth-method-bearer"
                        type="radio"
                        name="auth"
                        value={
                          PackageRepositoryAuth_PackageRepositoryAuthType[
                            PackageRepositoryAuth_PackageRepositoryAuthType
                              .PACKAGE_REPOSITORY_AUTH_TYPE_BEARER
                          ]
                        }
                        checked={
                          authMethod ===
                          PackageRepositoryAuth_PackageRepositoryAuthType.PACKAGE_REPOSITORY_AUTH_TYPE_BEARER
                        }
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
                          value={
                            PackageRepositoryAuth_PackageRepositoryAuthType[
                              PackageRepositoryAuth_PackageRepositoryAuthType
                                .PACKAGE_REPOSITORY_AUTH_TYPE_DOCKER_CONFIG_JSON
                            ]
                          }
                          checked={
                            authMethod ===
                            PackageRepositoryAuth_PackageRepositoryAuthType.PACKAGE_REPOSITORY_AUTH_TYPE_DOCKER_CONFIG_JSON
                          }
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
                        value={
                          PackageRepositoryAuth_PackageRepositoryAuthType[
                            PackageRepositoryAuth_PackageRepositoryAuthType
                              .PACKAGE_REPOSITORY_AUTH_TYPE_CUSTOM
                          ]
                        }
                        checked={
                          authMethod ===
                          PackageRepositoryAuth_PackageRepositoryAuthType.PACKAGE_REPOSITORY_AUTH_TYPE_CUSTOM
                        }
                        onChange={handleAuthRadioButtonChange}
                      />
                    </CdsRadio>
                  </CdsRadioGroup>
                  <div cds-layout="col@xs:8">
                    <div
                      hidden={
                        authMethod !==
                        PackageRepositoryAuth_PackageRepositoryAuthType.PACKAGE_REPOSITORY_AUTH_TYPE_BASIC_AUTH
                      }
                    >
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
                    <div
                      hidden={
                        authMethod !==
                        PackageRepositoryAuth_PackageRepositoryAuthType.PACKAGE_REPOSITORY_AUTH_TYPE_BEARER
                      }
                    >
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
                    <div
                      hidden={
                        authMethod !==
                        PackageRepositoryAuth_PackageRepositoryAuthType.PACKAGE_REPOSITORY_AUTH_TYPE_CUSTOM
                      }
                    >
                      <CdsInput>
                        <label>Raw Authorization Header</label>
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
                  required={
                    authMethod ===
                    PackageRepositoryAuth_PackageRepositoryAuthType.PACKAGE_REPOSITORY_AUTH_TYPE_DOCKER_CONFIG_JSON
                  }
                />
              </CdsFormGroup>
            </CdsAccordionContent>
          </CdsAccordionPanel>

          <CdsAccordionPanel
            expanded={accordion[2]}
            hidden={
              type !== RepositoryStorageTypes.PACKAGE_REPOSITORY_STORAGE_OCI &&
              plugin?.name !== PluginNames.PACKAGES_HELM
            }
          >
            <CdsAccordionHeader onClick={() => toggleAccordion(2)}>Filtering</CdsAccordionHeader>
            <CdsAccordionContent>
              <CdsFormGroup layout="vertical">
                {type === RepositoryStorageTypes.PACKAGE_REPOSITORY_STORAGE_OCI && (
                  <CdsTextarea>
                    <label htmlFor="kubeapps-oci-repositories">
                      List of Repositories (required)
                    </label>
                    <CdsControlMessage>
                      Include a list of comma-separated OCI repositories that will be available in
                      Kubeapps.
                    </CdsControlMessage>
                    <textarea
                      id="kubeapps-oci-repositories"
                      className="cds-textarea-fix"
                      placeholder={"nginx, jenkins"}
                      value={ociRepositories}
                      onChange={handleOCIRepositoriesChange}
                    />
                  </CdsTextarea>
                )}
                {/* TODO(agamez): workaround until Flux plugin also supports OCI artifacts */}
                {plugin?.name === PluginNames.PACKAGES_HELM && (
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
                      <CdsControlMessage>
                        Exclude packages matching the given filter
                      </CdsControlMessage>
                      <input
                        type="checkbox"
                        onChange={handleFilterExclude}
                        checked={filterExclude}
                      />
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
              </CdsFormGroup>
            </CdsAccordionContent>
          </CdsAccordionPanel>

          <CdsAccordionPanel expanded={accordion[3]}>
            <CdsAccordionHeader onClick={() => toggleAccordion(3)}>Advanced</CdsAccordionHeader>
            <CdsAccordionContent>
              <CdsFormGroup layout="vertical">
                <CdsInput>
                  <label>Synchronization Interval</label>
                  <input
                    id="kubeapps-repo-interval"
                    type="number"
                    placeholder="Synchronization interval in seconds"
                    value={interval}
                    onChange={handleIntervalChange}
                    required={false}
                  />
                </CdsInput>
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
                <CdsCheckbox className="ca-skip-tls">
                  <label className="clr-control-label">
                    Pass Credentials to 3rd party URLs (Icon and Tarball files)
                  </label>
                  <input
                    id="kubeapps-repo-pass-credentials"
                    type="checkbox"
                    checked={passCredentials}
                    onChange={handlePassCredentialsChange}
                  />
                </CdsCheckbox>

                <CdsTextarea layout="vertical">
                  <label>Custom Sync Job Template (optional)</label>
                  <CdsControlMessage>
                    It's possible to modify the default sync job. When doing this, the
                    pre-validation is not supported. More info{" "}
                    <a
                      target="_blank"
                      rel="noopener noreferrer"
                      href={`https://github.com/vmware-tanzu/kubeapps/blob/${appVersion}/site/content/docs/latest/howto/private-app-repository.md#modifying-the-synchronization-job`}
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
            </CdsAccordionContent>
          </CdsAccordionPanel>
        </CdsAccordion>

        {namespace === kubeappsNamespace && (
          <p>
            <strong>NOTE:</strong> This Package Repository will be created in the "
            {kubeappsNamespace}" global namespace. Consequently, its packages will be available for
            installation in every namespace and cluster.
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
          <CdsButton type="submit" disabled={validating}>
            {validating
              ? "Validating..."
              : `${repo.name ? `Update '${repo.name}'` : "Install"} Repo ${
                  validated === false ? "(force)" : ""
                }`}
          </CdsButton>
        </div>
      </form>
    </>
  );
}
