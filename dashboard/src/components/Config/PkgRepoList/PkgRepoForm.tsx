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
import { CdsToggle, CdsToggleGroup } from "@cds/react/toggle";
import actions from "actions";
import Alert from "components/js/Alert";
import {
  DockerCredentials,
  PackageRepositoryAuth_PackageRepositoryAuthType,
  PackageRepositoryReference,
  UsernamePassword,
} from "gen/kubeappsapis/core/packages/v1alpha1/repositories";
import { Plugin } from "gen/kubeappsapis/core/plugins/v1alpha1/plugins";
import { HelmPackageRepositoryCustomDetail } from "gen/kubeappsapis/plugins/helm/packages/v1alpha1/helm";
import { useEffect, useRef, useState } from "react";
import { useDispatch, useSelector } from "react-redux";
import { Action } from "redux";
import { ThunkDispatch } from "redux-thunk";
import { IConfig } from "shared/Config";
import { toFilterRule, toParams } from "shared/jq";
import {
  IPkgRepoFormData,
  IPkgRepositoryFilter,
  IStoreState,
  PluginNames,
  RepositoryStorageTypes,
} from "shared/types";
import {
  getGlobalNamespaceOrNamespace,
  getPluginByName,
  getPluginPackageName,
  getSupportedPackageRepositoryAuthTypes,
  isGlobalNamespace,
  k8sObjectNameRegex,
} from "shared/utils";
import "./PkgRepoForm.css";

export interface IPkgRepoFormProps {
  onSubmit: (data: IPkgRepoFormData) => Promise<boolean>;
  onAfterInstall?: () => void;
  namespace: string;
  helmGlobalNamespace: string;
  carvelGlobalNamespace: string;
  packageRepoRef?: PackageRepositoryReference;
}

export function PkgRepoForm(props: IPkgRepoFormProps) {
  const {
    onSubmit,
    onAfterInstall,
    namespace,
    helmGlobalNamespace,
    carvelGlobalNamespace,
    packageRepoRef: selectedPkgRepo,
  } = props;
  const isInstallingRef = useRef(false);
  const dispatch: ThunkDispatch<IStoreState, null, Action> = useDispatch();

  const {
    repos: {
      repoDetail: repo,
      isFetching,
      errors: { create: createError, update: updateError, delete: deleteError, fetch: fetchError },
    },
    clusters: { currentCluster },
    config: { configuredPlugins },
  } = useSelector((state: IStoreState) => state);

  const currentNsConfig = {
    helmGlobalNamespace: helmGlobalNamespace,
    carvelGlobalNamespace: carvelGlobalNamespace,
  } as IConfig;

  const configuredPluginsNames = configuredPlugins.map(plugin => plugin.name);

  const initialInterval = "10m";

  // -- Auth-related variables --

  // Auth type of the package repository
  const [authMethod, setAuthMethod] = useState(
    PackageRepositoryAuth_PackageRepositoryAuthType.PACKAGE_REPOSITORY_AUTH_TYPE_UNSPECIFIED,
  );
  // Auth type of the registry (for Helm-based repos)
  const [helmPSAuthMethod, setHelmPsAuthMethod] = useState(
    PackageRepositoryAuth_PackageRepositoryAuthType.PACKAGE_REPOSITORY_AUTH_TYPE_UNSPECIFIED,
  );

  // PACKAGE_REPOSITORY_AUTH_TYPE_AUTHORIZATION_HEADER
  const [authCustomHeader, setAuthCustomHeader] = useState("");

  // PACKAGE_REPOSITORY_AUTH_TYPE_BASIC_AUTH
  const [basicPassword, setBasicPassword] = useState("");
  const [basicUser, setBasicUser] = useState("");

  // Repository scope
  const [isNamespaceScoped, setIsNamespaceScoped] = useState(true);

  // PACKAGE_REPOSITORY_AUTH_TYPE_BEARER
  const [bearerToken, setBearerToken] = useState("");

  // PACKAGE_REPOSITORY_AUTH_TYPE_DOCKER_CONFIG_JSON
  const [secretEmail, setSecretEmail] = useState("");
  const [secretUser, setSecretUser] = useState("");
  const [secretPassword, setSecretPassword] = useState("");
  const [secretServer, setSecretServer] = useState("");

  // Registry pullsecrets
  const [useSameAuthCreds, setUseSameAuthCreds] = useState(false);
  const [pullSecretEmail, setPullSecretEmail] = useState("");
  const [pullSecretUser, setPullSecretUser] = useState("");
  const [pullSecretPassword, setPullSecretPassword] = useState("");
  const [pullSecretServer, setPullSecretServer] = useState("");

  // PACKAGE_REPOSITORY_AUTH_TYPE_SSH
  const [sshKnownHosts, setSshKnownHosts] = useState("");
  const [sshPrivateKey, setSshPrivateKey] = useState("");

  // PACKAGE_REPOSITORY_AUTH_TYPE_TLS
  const [tlsAuthCert, setTlsAuthCert] = useState("");
  const [tlsAuthKey, setTlsAuthKey] = useState("");

  // PACKAGE_REPOSITORY_AUTH_TYPE_OPAQUE
  const [opaqueData, setOpaqueData] = useState("");

  // User-managed secrets
  const [secretAuthName, setSecretAuthName] = useState("");
  const [secretPSName, setSecretPSName] = useState("");
  const [secretTLSName, setSecretTLSName] = useState("");

  // rest of the package repo form variables

  // -- Basic info variables --
  const [name, setName] = useState("");
  const [url, setURL] = useState("");
  const [description, setDescription] = useState("");
  const [plugin, setPlugin] = useState({} as Plugin);
  const [type, setType] = useState("");

  // -- Filtering  variables --
  const [ociRepositories, setOCIRepositories] = useState("");
  const [filterNames, setFilterNames] = useState("");
  const [filterExclude, setFilterExclude] = useState(false);
  const [filterRegex, setFilterRegex] = useState(false);

  // -- Advanced  variables --
  const [syncInterval, setSyncInterval] = useState(initialInterval);
  const [performValidation, setPerformValidation] = useState(true);
  const [customCA, setCustomCA] = useState("");
  const [skipTLS, setSkipTLS] = useState(!!repo?.tlsConfig?.insecureSkipVerify);
  const [passCredentials, setPassCredentials] = useState(!!repo?.auth?.passCredentials);

  // -- Kubeapps/User-managed secrets  variables --
  const [isUserManagedSecret, setIsUserManagedSecret] = useState(false);
  const [isUserManagedPSSecret, setIsUserManagedPSSecret] = useState(false);
  const [isUserManagedCASecret, setIsUserManagedCASecret] = useState(false);

  // initial state (collapsed or not) of each accordion tab
  const [accordion, setAccordion] = useState([true, false, false, false]);

  const toggleAccordion = (section: number) => {
    const items = [...accordion];
    items[section] = !items[section];
    setAccordion(items);
  };

  useEffect(() => {
    if (selectedPkgRepo) {
      dispatch(actions.repos.fetchRepoDetail(selectedPkgRepo));
    }
  }, [dispatch, selectedPkgRepo]);

  useEffect(() => {
    if (repo) {
      // populate state properties from the incoming repo
      setName(repo.name);
      setURL(repo.url);
      setType(repo.type);
      setPlugin(repo.packageRepoRef?.plugin || ({ name: "", version: "" } as Plugin));
      setIsNamespaceScoped(repo.namespaceScoped);
      setDescription(repo.description);
      setSkipTLS(!!repo.tlsConfig?.insecureSkipVerify);
      setPassCredentials(!!repo.auth?.passCredentials);
      setSyncInterval(repo.interval);
      setCustomCA(repo.tlsConfig?.certAuthority || "");
      setAuthCustomHeader(repo.auth?.header || "");
      setBearerToken(repo.auth?.header || "");
      setBasicPassword(repo.auth?.usernamePassword?.password || "");
      setBasicUser(repo.auth?.usernamePassword?.username || "");
      setSecretEmail(repo.auth?.dockerCreds?.email || "");
      setSecretPassword(repo.auth?.dockerCreds?.password || "");
      setSecretServer(repo.auth?.dockerCreds?.server || "");
      setSecretUser(repo.auth?.dockerCreds?.username || "");
      setSshKnownHosts(repo.auth?.sshCreds?.knownHosts || "");
      setSshPrivateKey(repo.auth?.sshCreds?.privateKey || "");
      setTlsAuthCert(repo.auth?.tlsCertKey?.cert || "");
      setTlsAuthKey(repo.auth?.tlsCertKey?.key || "");
      setOpaqueData(JSON.stringify(repo.auth?.opaqueCreds?.data) || "");
      setAuthMethod(
        repo.auth?.type ||
          PackageRepositoryAuth_PackageRepositoryAuthType.PACKAGE_REPOSITORY_AUTH_TYPE_UNSPECIFIED,
      );
      setSecretAuthName(repo.auth?.secretRef?.name || "");
      setSecretTLSName(repo.tlsConfig?.secretRef?.name || "");
      setIsUserManagedSecret(!!repo.auth?.secretRef?.name);

      // setting custom details for the Helm plugin
      if (repo.packageRepoRef?.plugin?.name === PluginNames.PACKAGES_HELM) {
        const helmPackageRepositoryCustomDetail =
          repo.customDetail as Partial<HelmPackageRepositoryCustomDetail>;
        setOCIRepositories(helmPackageRepositoryCustomDetail?.ociRepositories?.join(", ") || "");
        setPerformValidation(helmPackageRepositoryCustomDetail?.performValidation || false);
        if (helmPackageRepositoryCustomDetail?.filterRule?.jq) {
          const { names, regex, exclude } = toParams(helmPackageRepositoryCustomDetail.filterRule!);
          setFilterRegex(regex);
          setFilterExclude(exclude);
          setFilterNames(names);
        }
        if (
          helmPackageRepositoryCustomDetail?.imagesPullSecret?.secretRef ||
          helmPackageRepositoryCustomDetail?.imagesPullSecret?.credentials
        ) {
          setUseSameAuthCreds(false);
          setIsUserManagedPSSecret(
            !!helmPackageRepositoryCustomDetail?.imagesPullSecret?.secretRef,
          );
          setHelmPsAuthMethod(
            PackageRepositoryAuth_PackageRepositoryAuthType.PACKAGE_REPOSITORY_AUTH_TYPE_DOCKER_CONFIG_JSON,
          );
          setSecretPSName(helmPackageRepositoryCustomDetail?.imagesPullSecret?.secretRef || "");
          setPullSecretEmail(
            helmPackageRepositoryCustomDetail?.imagesPullSecret?.credentials?.email || "",
          );
          setPullSecretUser(
            helmPackageRepositoryCustomDetail?.imagesPullSecret?.credentials?.username || "",
          );
          setPullSecretPassword(
            helmPackageRepositoryCustomDetail?.imagesPullSecret?.credentials?.password || "",
          );
          setPullSecretServer(
            helmPackageRepositoryCustomDetail?.imagesPullSecret?.credentials?.server || "",
          );
        }
      }
    }
  }, [repo, namespace, currentCluster, dispatch]);

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

    // send the proper header depending on the auth method
    let finalHeader = "";
    switch (authMethod) {
      case PackageRepositoryAuth_PackageRepositoryAuthType.PACKAGE_REPOSITORY_AUTH_TYPE_AUTHORIZATION_HEADER:
        finalHeader = authCustomHeader;
        break;
      case PackageRepositoryAuth_PackageRepositoryAuthType.PACKAGE_REPOSITORY_AUTH_TYPE_BEARER:
        finalHeader = `Bearer ${bearerToken}`;
        break;
    }

    // create an array from the (trimmed) comma separated string
    const ociRepoList = ociRepositories.length
      ? ociRepositories?.split(",").map(r => r.trim())
      : [];

    // In the Helm plugin, if the scheme is not specified, assume HTTPS (also for OCI registries)
    // Other plugins don't allow passing a scheme (eg. carvel) and others require a different one (eg. flux: oci://)
    let finalURL = url;
    if (plugin?.name === PluginNames.PACKAGES_HELM && !url?.startsWith("http")) {
      finalURL = `https://${url}`;
    }

    // build the IPkgRepositoryFilter object based on the filter names plus the regex and exclude options
    let filter: IPkgRepositoryFilter | undefined;
    if (type === RepositoryStorageTypes.PACKAGE_REPOSITORY_STORAGE_HELM && filterNames !== "") {
      filter = toFilterRule(filterNames, filterRegex, filterExclude);
    }

    // build the custom details object based for each plugin
    const helmCustomDetail = {
      ociRepositories: ociRepoList,
      performValidation,
      filterRule: filter,
      imagesPullSecret: {
        // if using the same credentials toggle is set, use the repo auth's creds instead
        secretRef: isUserManagedPSSecret ? (useSameAuthCreds ? secretAuthName : secretPSName) : "",
        credentials: !isUserManagedPSSecret
          ? {
              email: useSameAuthCreds ? secretEmail : pullSecretEmail,
              username: useSameAuthCreds ? secretUser : pullSecretUser,
              password: useSameAuthCreds ? secretPassword : pullSecretPassword,
              server: useSameAuthCreds ? secretServer : pullSecretServer,
            }
          : undefined,
      },
    } as HelmPackageRepositoryCustomDetail;

    //TODO(agamez): add support for kapp's custom details
    // const kappCustomDetail = undefined as unknown as KappControllerPackageRepositoryCustomDetail;

    // build request object for the install call
    const request = {
      authHeader: !isUserManagedSecret ? finalHeader : "",
      authMethod,
      basicAuth: {
        password: !isUserManagedSecret ? basicPassword : "",
        username: !isUserManagedSecret ? basicUser : "",
      } as UsernamePassword,
      customCA: !isUserManagedCASecret ? customCA : "",
      description,
      dockerRegCreds: {
        username: !isUserManagedSecret ? secretUser : "",
        email: !isUserManagedSecret ? secretEmail : "",
        password: !isUserManagedSecret ? secretPassword : "",
        server: !isUserManagedSecret ? secretServer : "",
      } as DockerCredentials,
      interval: syncInterval,
      name,
      passCredentials,
      plugin,
      secretAuthName: isUserManagedSecret ? secretAuthName : "",
      secretTLSName: isUserManagedCASecret ? secretTLSName : "",
      skipTLS,
      type,
      url: finalURL,
      opaqueCreds: {
        data: !isUserManagedSecret && opaqueData ? JSON.parse(opaqueData) : {},
      },
      sshCreds: {
        knownHosts: !isUserManagedSecret ? sshKnownHosts : "",
        privateKey: !isUserManagedSecret ? sshPrivateKey : "",
      },
      tlsCertKey: {
        cert: !isUserManagedSecret ? tlsAuthCert : "",
        key: !isUserManagedSecret ? tlsAuthKey : "",
      },
      namespace: isNamespaceScoped
        ? namespace
        : getGlobalNamespaceOrNamespace(namespace, plugin?.name, currentNsConfig),
      isNamespaceScoped,
    } as IPkgRepoFormData;

    // enrich the request object with the corresponding plugin's custom details
    switch (plugin?.name) {
      case PluginNames.PACKAGES_HELM:
        request.customDetail = helmCustomDetail;
        break;
      //TODO(agamez): add it once other PRs get merged
      // case PluginNames.PACKAGES_KAPP:
      //   request.customDetail = kappCustomDetail;
      //   break;
      default:
        break;
    }

    const success = await onSubmit(request);
    if (success && onAfterInstall) {
      onAfterInstall();
    }
    isInstallingRef.current = false;
  };

  const handleNameChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setName(e.target.value);
  };
  const handleDescriptionChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setDescription(e.target.value);
  };
  const handleIntervalChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setSyncInterval(e.target.value);
  };
  const handlePerformValidationChange = (_e: React.ChangeEvent<HTMLInputElement>) => {
    setPerformValidation(!performValidation);
  };
  const handleURLChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setURL(e.target.value);
  };
  const handleAuthCustomHeaderChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setAuthCustomHeader(e.target.value);
  };
  const handleAuthBearerTokenChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setBearerToken(e.target.value);
  };
  const handleCustomCAChange = (e: React.ChangeEvent<HTMLTextAreaElement>) => {
    setCustomCA(e.target.value);
  };
  const handleAuthRadioButtonChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setAuthMethod(PackageRepositoryAuth_PackageRepositoryAuthType[e.target.value]);

    // reset the pull secret copy from auth
    setUseSameAuthCreds(false);

    // if the user selects the docker config, suggest also setting the pull secret
    if (
      PackageRepositoryAuth_PackageRepositoryAuthType[e.target.value] ===
      PackageRepositoryAuth_PackageRepositoryAuthType.PACKAGE_REPOSITORY_AUTH_TYPE_DOCKER_CONFIG_JSON
    ) {
      setHelmPsAuthMethod(
        PackageRepositoryAuth_PackageRepositoryAuthType.PACKAGE_REPOSITORY_AUTH_TYPE_DOCKER_CONFIG_JSON,
      );
      // if user hasn't set any field yet, suggest using the same pullsecret
      if (
        !secretPSName &&
        !pullSecretServer &&
        !pullSecretEmail &&
        !pullSecretUser &&
        !pullSecretPassword
      ) {
        setUseSameAuthCreds(true);
      }
    }
  };
  const handleImgPSChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setHelmPsAuthMethod(PackageRepositoryAuth_PackageRepositoryAuthType[e.target.value]);
  };
  const handleTypeRadioButtonChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const newType = e.target.value;
    setType(newType);

    // if a unsupported auth method is selected, reset it to the default one
    if (!getSupportedPackageRepositoryAuthTypes(plugin, newType).includes(authMethod)) {
      setAuthMethod(
        PackageRepositoryAuth_PackageRepositoryAuthType.PACKAGE_REPOSITORY_AUTH_TYPE_UNSPECIFIED,
      );
    }
  };
  const handlePluginRadioButtonChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const newPlugin = getPluginByName(e.target.value);
    setPlugin(newPlugin);

    // if a unsupported auth method is selected, reset it to the default one
    if (!getSupportedPackageRepositoryAuthTypes(newPlugin, type).includes(authMethod)) {
      setAuthMethod(
        PackageRepositoryAuth_PackageRepositoryAuthType.PACKAGE_REPOSITORY_AUTH_TYPE_UNSPECIFIED,
      );
    }

    // set some default values based on the selected plugin
    switch (getPluginByName(e.target.value)?.name) {
      case PluginNames.PACKAGES_HELM: {
        if (!type) {
          setType(RepositoryStorageTypes.PACKAGE_REPOSITORY_STORAGE_HELM);
        }
        // helm plugin doesn't allow setting sync interval
        setSyncInterval("");

        // update the isNampespaced field based on the plugin
        setIsNamespaceScoped(
          !isGlobalNamespace(namespace, PluginNames.PACKAGES_HELM, currentNsConfig),
        );
        break;
      }
      case PluginNames.PACKAGES_FLUX: {
        if (!type) {
          setType(RepositoryStorageTypes.PACKAGE_REPOSITORY_STORAGE_HELM);
        }
        // set the sync interval to the user's value or default value
        setSyncInterval(syncInterval || initialInterval);

        // update the isNampespaced field based on the plugin
        setIsNamespaceScoped(
          !isGlobalNamespace(namespace, PluginNames.PACKAGES_FLUX, currentNsConfig),
        );
        break;
      }
      case PluginNames.PACKAGES_KAPP:
        if (!type) {
          setType(RepositoryStorageTypes.PACKAGE_REPOSITORY_STORAGE_CARVEL_IMGPKGBUNDLE);
        }
        // set the sync interval to the user's value or default value
        setSyncInterval(syncInterval || initialInterval);

        // carvel plugin doesn't allow description
        setDescription("");

        // update the isNampespaced field based on the plugin
        setIsNamespaceScoped(
          !isGlobalNamespace(namespace, PluginNames.PACKAGES_KAPP, currentNsConfig),
        );
        break;
    }
  };
  const handleRepoScopeChange = (_e: React.ChangeEvent<HTMLInputElement>) => {
    setIsNamespaceScoped(!isNamespaceScoped);
  };
  const handleBasicUserChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setBasicUser(e.target.value);
  };
  const handleBasicPasswordChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setBasicPassword(e.target.value);
  };
  const handleOCIRepositoriesChange = (e: React.ChangeEvent<HTMLTextAreaElement>) => {
    setOCIRepositories(e.target.value);
  };
  const handleSkipTLSChange = (_e: React.ChangeEvent<HTMLInputElement>) => {
    setSkipTLS(!skipTLS);
  };
  const handlePassCredentialsChange = (_e: React.ChangeEvent<HTMLInputElement>) => {
    setPassCredentials(!passCredentials);
  };
  const handleFilterNamesChange = (e: React.ChangeEvent<HTMLTextAreaElement>) => {
    setFilterNames(e.target.value);
  };
  const handleFilterRegexChange = (_e: React.ChangeEvent<HTMLInputElement>) => {
    setFilterRegex(!filterRegex);
  };
  const handleFilterExcludeChange = (_e: React.ChangeEvent<HTMLInputElement>) => {
    setFilterExclude(!filterExclude);
  };
  const handleAuthSecretUserChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setSecretUser(e.target.value);
  };
  const handleAuthSecretPasswordChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setSecretPassword(e.target.value);
  };
  const handleAuthSecretEmailChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setSecretEmail(e.target.value);
  };
  const handleAuthSecretServerChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setSecretServer(e.target.value);
  };
  const handleImgPSUserChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setPullSecretUser(e.target.value);
  };
  const handleImgPSPasswordChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setPullSecretPassword(e.target.value);
  };
  const handleImgPSEmailChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setPullSecretEmail(e.target.value);
  };
  const handleImgPSServerChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setPullSecretServer(e.target.value);
  };
  const handleSshKnownHostsChange = (e: React.ChangeEvent<HTMLTextAreaElement>) => {
    setSshKnownHosts(e.target.value);
  };
  const handleSshPrivateKeyChange = (e: React.ChangeEvent<HTMLTextAreaElement>) => {
    setSshPrivateKey(e.target.value);
  };
  const handleTlsAuthCertChange = (e: React.ChangeEvent<HTMLTextAreaElement>) => {
    setTlsAuthCert(e.target.value);
  };
  const handleTlsAuthKeyChange = (e: React.ChangeEvent<HTMLTextAreaElement>) => {
    setTlsAuthKey(e.target.value);
  };
  const handleOpaqueDataChange = (e: React.ChangeEvent<HTMLTextAreaElement>) => {
    setOpaqueData(e.target.value);
  };
  const handleSecretAuthNameChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setSecretAuthName(e.target.value);
  };
  const setSecretPSNameChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setSecretPSName(e.target.value);
  };
  const handleSecretTLSNameChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setSecretTLSName(e.target.value);
  };
  const handleIsUserManagedSecretChange = (_e: React.ChangeEvent<HTMLInputElement>) => {
    setIsUserManagedSecret(!isUserManagedSecret);
  };
  const handleIsUserManagedPSSecretChange = (_e: React.ChangeEvent<HTMLInputElement>) => {
    setIsUserManagedPSSecret(!isUserManagedPSSecret);
  };
  const handleIsUserManagedCASecretChange = (_e: React.ChangeEvent<HTMLInputElement>) => {
    setIsUserManagedCASecret(!isUserManagedCASecret);
  };
  const handelUseSameAuthCredsChange = (_e: React.ChangeEvent<HTMLInputElement>) => {
    setUseSameAuthCreds(!useSameAuthCreds);
  };

  const userManagedSecretText = "Use an existing secret";
  const kubeappsManagedSecretText = "Provide the secret values";

  const isUserManagedSecretToggle = (section: string) => (
    <>
      <CdsToggleGroup className="flex-v-center">
        <CdsToggle>
          <label htmlFor={`${section}-isUserManagedSecretToggle`}>
            {isUserManagedSecret ? userManagedSecretText : kubeappsManagedSecretText}
          </label>
          <input
            id={`${section}-isUserManagedSecretToggle`}
            type="checkbox"
            onChange={handleIsUserManagedSecretChange}
            checked={isUserManagedSecret}
            disabled={!!repo.auth?.type}
          />
        </CdsToggle>
      </CdsToggleGroup>
    </>
  );

  const secretNameInput = (authType: string) => (
    <>
      <CdsInput>
        <label htmlFor={`kubeapps-repo-auth-secret-name-${authType}`}>Secret Name</label>
        <input
          id={`kubeapps-repo-auth-secret-name-${authType}`}
          type="text"
          placeholder="my-secret-name"
          value={secretAuthName || ""}
          onChange={handleSecretAuthNameChange}
          required={
            isUserManagedSecret &&
            authMethod !==
              PackageRepositoryAuth_PackageRepositoryAuthType.PACKAGE_REPOSITORY_AUTH_TYPE_UNSPECIFIED
          }
          pattern={k8sObjectNameRegex}
          title="Use lower case alphanumeric characters, '-' or '.'"
        />
      </CdsInput>
      <br />
      <CdsControlMessage>
        Name of the{" "}
        <a
          href="https://kubernetes.io/docs/concepts/configuration/secret/"
          target="_blank"
          rel="noopener noreferrer"
        >
          Kubernetes Secret
        </a>{" "}
        object holding the authentication data. Please ensure that this secret's type is the
        expected one by the selected authentication method.
      </CdsControlMessage>
    </>
  );

  return (
    <>
      <form onSubmit={handleInstallClick}>
        <CdsAccordion>
          <CdsAccordionPanel id="panel-basic" expanded={accordion[0]}>
            <CdsAccordionHeader onClick={() => toggleAccordion(0)}>
              Basic information
            </CdsAccordionHeader>
            <CdsAccordionContent>
              <CdsFormGroup layout="vertical">
                <CdsInput>
                  <label htmlFor="kubeapps-repo-name">Name</label>
                  <input
                    id="kubeapps-repo-name"
                    type="text"
                    placeholder="example"
                    value={name || ""}
                    onChange={handleNameChange}
                    required={true}
                    pattern={k8sObjectNameRegex}
                    title="Use lower case alphanumeric characters, '-' or '.'"
                    disabled={!!repo.name}
                  />
                </CdsInput>
                <CdsInput>
                  <label htmlFor="kubeapps-repo-url"> URL </label>
                  <input
                    id="kubeapps-repo-url"
                    type="text"
                    placeholder="https://charts.example.com/stable"
                    value={url || ""}
                    onChange={handleURLChange}
                    required={true}
                  />
                </CdsInput>
                <CdsInput>
                  <label htmlFor="kubeapps-repo-description"> Description (optional)</label>
                  <input
                    id="kubeapps-repo-description"
                    type="text"
                    placeholder="Description of the repository"
                    value={description || ""}
                    onChange={handleDescriptionChange}
                    disabled={(
                      [PluginNames.PACKAGES_FLUX, PluginNames.PACKAGES_KAPP] as string[]
                    ).includes(plugin?.name)}
                  />
                </CdsInput>
                {/* TODO(agamez): these plugin selectors should be loaded
                based on the current plugins that are loaded in the cluster */}
                <CdsRadioGroup layout="vertical">
                  {/* eslint-disable-next-line jsx-a11y/label-has-associated-control */}
                  <label>Packaging Format:</label>
                  <CdsControlMessage>Select the plugin to use.</CdsControlMessage>
                  <CdsRadio>
                    <label htmlFor="kubeapps-plugin-helm">
                      {getPluginPackageName(PluginNames.PACKAGES_HELM, true)}
                    </label>
                    <input
                      id="kubeapps-plugin-helm"
                      type="radio"
                      name="plugin"
                      value={PluginNames.PACKAGES_HELM || ""}
                      checked={plugin?.name === PluginNames.PACKAGES_HELM}
                      onChange={handlePluginRadioButtonChange}
                      disabled={
                        !!repo.packageRepoRef?.plugin ||
                        !configuredPluginsNames.includes(PluginNames.PACKAGES_HELM)
                      }
                      required={true}
                    />
                  </CdsRadio>
                  <CdsRadio>
                    <label htmlFor="kubeapps-plugin-fluxv2">
                      {getPluginPackageName(PluginNames.PACKAGES_FLUX, true)}
                    </label>
                    <input
                      id="kubeapps-plugin-fluxv2"
                      type="radio"
                      name="plugin"
                      value={PluginNames.PACKAGES_FLUX || ""}
                      checked={plugin?.name === PluginNames.PACKAGES_FLUX}
                      onChange={handlePluginRadioButtonChange}
                      disabled={
                        !!repo.packageRepoRef?.plugin ||
                        !configuredPluginsNames.includes(PluginNames.PACKAGES_FLUX)
                      }
                      required={true}
                    />
                  </CdsRadio>
                  <CdsRadio>
                    <label htmlFor="kubeapps-plugin-kappcontroller">
                      {getPluginPackageName(PluginNames.PACKAGES_KAPP, true)}
                    </label>
                    <input
                      id="kubeapps-plugin-kappcontroller"
                      type="radio"
                      name="plugin"
                      value={PluginNames.PACKAGES_KAPP || ""}
                      checked={plugin?.name === PluginNames.PACKAGES_KAPP}
                      onChange={handlePluginRadioButtonChange}
                      disabled={
                        !!repo.packageRepoRef?.plugin ||
                        !configuredPluginsNames.includes(PluginNames.PACKAGES_KAPP)
                      }
                      required={true}
                    />
                  </CdsRadio>
                </CdsRadioGroup>
                {plugin?.name && (
                  <>
                    <CdsRadioGroup layout="vertical">
                      {/* eslint-disable-next-line jsx-a11y/label-has-associated-control */}
                      <label>Scope:</label>
                      <CdsControlMessage>
                        If the repository is global, its packages will be available in every
                        namespace of the cluster. If it is namespaced-scoped, its packages will be
                        available only in the namespace where the repository has been installed.
                      </CdsControlMessage>

                      {repo.namespaceScoped ||
                        (isGlobalNamespace(namespace, plugin?.name, currentNsConfig) && (
                          <>
                            <CdsControlMessage>
                              <br />
                              The "{repo.packageRepoRef?.context?.namespace || namespace}" namespace
                              is already a global one for the "{getPluginPackageName(plugin)}"
                              plugin.
                            </CdsControlMessage>
                          </>
                        ))}

                      <CdsRadio>
                        <label htmlFor="kubeapps-scope-namespaced">
                          {`Namespaced repository (using the "${
                            repo.packageRepoRef?.context?.namespace || namespace
                          }" namespace)`}
                        </label>
                        <input
                          id="kubeapps-scope-namespaced"
                          type="radio"
                          name="scope"
                          checked={repo.namespaceScoped || isNamespaceScoped}
                          onChange={handleRepoScopeChange}
                          disabled={
                            !!repo.name ||
                            isGlobalNamespace(namespace, plugin?.name, currentNsConfig)
                          }
                          required={true}
                        />
                      </CdsRadio>
                      <CdsRadio>
                        <label htmlFor="kubeapps-scope-global">
                          {`Global repository (using the "${getGlobalNamespaceOrNamespace(
                            repo.packageRepoRef?.context?.namespace || namespace,
                            plugin?.name,
                            currentNsConfig,
                          )}" namespace)`}
                        </label>
                        <input
                          id="kubeapps-scope-global"
                          type="radio"
                          name="scope"
                          checked={!repo.namespaceScoped && !isNamespaceScoped}
                          onChange={handleRepoScopeChange}
                          disabled={
                            !!repo.name ||
                            isGlobalNamespace(namespace, plugin?.name, currentNsConfig)
                          }
                          required={true}
                        />
                      </CdsRadio>
                    </CdsRadioGroup>
                    <CdsRadioGroup layout="vertical">
                      {/* eslint-disable-next-line jsx-a11y/label-has-associated-control */}
                      <label>Package Storage Type</label>
                      <CdsControlMessage>Select the package storage type.</CdsControlMessage>
                      {(
                        [PluginNames.PACKAGES_HELM, PluginNames.PACKAGES_FLUX] as string[]
                      ).includes(plugin?.name) && (
                        <>
                          <CdsRadio>
                            <label htmlFor="kubeapps-repo-type-helm">Helm Repository</label>
                            <input
                              id="kubeapps-repo-type-helm"
                              type="radio"
                              name="type"
                              value={RepositoryStorageTypes.PACKAGE_REPOSITORY_STORAGE_HELM || ""}
                              checked={
                                type === RepositoryStorageTypes.PACKAGE_REPOSITORY_STORAGE_HELM
                              }
                              disabled={!!repo?.type}
                              onChange={handleTypeRadioButtonChange}
                              required={(
                                [PluginNames.PACKAGES_HELM, PluginNames.PACKAGES_FLUX] as string[]
                              ).includes(plugin?.name)}
                            />
                          </CdsRadio>
                          <CdsRadio>
                            <label htmlFor="kubeapps-repo-type-oci">OCI Registry</label>
                            <input
                              id="kubeapps-repo-type-oci"
                              type="radio"
                              name="type"
                              disabled={!!repo?.type}
                              value={RepositoryStorageTypes.PACKAGE_REPOSITORY_STORAGE_OCI || ""}
                              checked={
                                type === RepositoryStorageTypes.PACKAGE_REPOSITORY_STORAGE_OCI
                              }
                              onChange={handleTypeRadioButtonChange}
                              required={(
                                [PluginNames.PACKAGES_HELM, PluginNames.PACKAGES_FLUX] as string[]
                              ).includes(plugin?.name)}
                            />
                          </CdsRadio>
                        </>
                      )}
                      {plugin?.name === PluginNames.PACKAGES_KAPP && (
                        <>
                          <CdsRadio>
                            <label htmlFor="kubeapps-repo-type-imgpkgbundle">Imgpkg Bundle</label>
                            <input
                              id="kubeapps-repo-type-imgpkgbundle"
                              type="radio"
                              name="type"
                              disabled={!!repo?.type}
                              value={
                                RepositoryStorageTypes.PACKAGE_REPOSITORY_STORAGE_CARVEL_IMGPKGBUNDLE ||
                                ""
                              }
                              checked={
                                type ===
                                RepositoryStorageTypes.PACKAGE_REPOSITORY_STORAGE_CARVEL_IMGPKGBUNDLE
                              }
                              onChange={handleTypeRadioButtonChange}
                              required={plugin?.name === PluginNames.PACKAGES_KAPP}
                            />
                          </CdsRadio>
                          <CdsRadio>
                            <label htmlFor="kubeapps-repo-type-inline">Inline</label>
                            <input
                              id="kubeapps-repo-type-inline"
                              type="radio"
                              name="type"
                              // TODO(agamez): enable it back once the API supports it
                              disabled={true}
                              // disabled={!!repo?.type}
                              value={
                                RepositoryStorageTypes.PACKAGE_REPOSITORY_STORAGE_CARVEL_INLINE ||
                                ""
                              }
                              checked={
                                type ===
                                RepositoryStorageTypes.PACKAGE_REPOSITORY_STORAGE_CARVEL_INLINE
                              }
                              onChange={handleTypeRadioButtonChange}
                              required={plugin?.name === PluginNames.PACKAGES_KAPP}
                            />
                          </CdsRadio>
                          <CdsRadio>
                            <label htmlFor="kubeapps-repo-type-image">Image</label>
                            <input
                              id="kubeapps-repo-type-image"
                              type="radio"
                              name="type"
                              disabled={!!repo?.type}
                              value={
                                RepositoryStorageTypes.PACKAGE_REPOSITORY_STORAGE_CARVEL_IMAGE || ""
                              }
                              checked={
                                type ===
                                RepositoryStorageTypes.PACKAGE_REPOSITORY_STORAGE_CARVEL_IMAGE
                              }
                              onChange={handleTypeRadioButtonChange}
                              required={plugin?.name === PluginNames.PACKAGES_KAPP}
                            />
                          </CdsRadio>
                          <CdsRadio>
                            <label htmlFor="kubeapps-repo-type-http">HTTP</label>
                            <input
                              id="kubeapps-repo-type-http"
                              type="radio"
                              name="type"
                              disabled={!!repo?.type}
                              value={
                                RepositoryStorageTypes.PACKAGE_REPOSITORY_STORAGE_CARVEL_HTTP || ""
                              }
                              checked={
                                type ===
                                RepositoryStorageTypes.PACKAGE_REPOSITORY_STORAGE_CARVEL_HTTP
                              }
                              onChange={handleTypeRadioButtonChange}
                              required={plugin?.name === PluginNames.PACKAGES_KAPP}
                            />
                          </CdsRadio>
                          <CdsRadio>
                            <label htmlFor="kubeapps-repo-type-git">Git</label>
                            <input
                              id="kubeapps-repo-type-git"
                              type="radio"
                              name="type"
                              disabled={!!repo?.type}
                              value={
                                RepositoryStorageTypes.PACKAGE_REPOSITORY_STORAGE_CARVEL_GIT || ""
                              }
                              checked={
                                type ===
                                RepositoryStorageTypes.PACKAGE_REPOSITORY_STORAGE_CARVEL_GIT
                              }
                              onChange={handleTypeRadioButtonChange}
                              required={plugin?.name === PluginNames.PACKAGES_KAPP}
                            />
                          </CdsRadio>
                        </>
                      )}
                    </CdsRadioGroup>
                  </>
                )}
              </CdsFormGroup>
            </CdsAccordionContent>
          </CdsAccordionPanel>

          <CdsAccordionPanel id="panel-auth" expanded={accordion[1]}>
            <CdsAccordionHeader onClick={() => toggleAccordion(1)}>
              Authentication
            </CdsAccordionHeader>
            <CdsAccordionContent>
              <CdsFormGroup layout="vertical">
                <div cds-layout="grid gap:lg">
                  {/* Begin authentication selection */}
                  <CdsRadioGroup cds-layout="col@xs:4">
                    {/* eslint-disable-next-line jsx-a11y/label-has-associated-control */}
                    <label>Repository Authorization</label>
                    <CdsRadio>
                      <label htmlFor="kubeapps-repo-auth-method-none">None (Public)</label>
                      <input
                        id="kubeapps-repo-auth-method-none"
                        type="radio"
                        name="auth"
                        value={
                          PackageRepositoryAuth_PackageRepositoryAuthType[
                            PackageRepositoryAuth_PackageRepositoryAuthType
                              .PACKAGE_REPOSITORY_AUTH_TYPE_UNSPECIFIED
                          ] || ""
                        }
                        checked={
                          authMethod ===
                          PackageRepositoryAuth_PackageRepositoryAuthType.PACKAGE_REPOSITORY_AUTH_TYPE_UNSPECIFIED
                        }
                        onChange={handleAuthRadioButtonChange}
                        disabled={!!repo.auth?.type}
                      />
                    </CdsRadio>
                    <CdsRadio>
                      <label htmlFor="kubeapps-repo-auth-method-basic">Basic Auth</label>
                      <input
                        id="kubeapps-repo-auth-method-basic"
                        type="radio"
                        name="auth"
                        value={
                          PackageRepositoryAuth_PackageRepositoryAuthType[
                            PackageRepositoryAuth_PackageRepositoryAuthType
                              .PACKAGE_REPOSITORY_AUTH_TYPE_BASIC_AUTH
                          ] || ""
                        }
                        checked={
                          authMethod ===
                          PackageRepositoryAuth_PackageRepositoryAuthType.PACKAGE_REPOSITORY_AUTH_TYPE_BASIC_AUTH
                        }
                        onChange={handleAuthRadioButtonChange}
                        disabled={
                          !!repo.auth?.type ||
                          !getSupportedPackageRepositoryAuthTypes(plugin, type).includes(
                            PackageRepositoryAuth_PackageRepositoryAuthType.PACKAGE_REPOSITORY_AUTH_TYPE_BASIC_AUTH,
                          )
                        }
                      />
                    </CdsRadio>
                    <CdsRadio>
                      <label htmlFor="kubeapps-repo-auth-method-bearer">Bearer Token</label>
                      <input
                        id="kubeapps-repo-auth-method-bearer"
                        type="radio"
                        name="auth"
                        value={
                          PackageRepositoryAuth_PackageRepositoryAuthType[
                            PackageRepositoryAuth_PackageRepositoryAuthType
                              .PACKAGE_REPOSITORY_AUTH_TYPE_BEARER
                          ] || ""
                        }
                        checked={
                          authMethod ===
                          PackageRepositoryAuth_PackageRepositoryAuthType.PACKAGE_REPOSITORY_AUTH_TYPE_BEARER
                        }
                        onChange={handleAuthRadioButtonChange}
                        disabled={
                          !!repo.auth?.type ||
                          !getSupportedPackageRepositoryAuthTypes(plugin, type).includes(
                            PackageRepositoryAuth_PackageRepositoryAuthType.PACKAGE_REPOSITORY_AUTH_TYPE_BEARER,
                          )
                        }
                      />
                    </CdsRadio>
                    <CdsRadio>
                      <label htmlFor="kubeapps-repo-auth-method-registry">
                        Docker Registry Credentials
                      </label>
                      <input
                        id="kubeapps-repo-auth-method-registry"
                        type="radio"
                        name="auth"
                        value={
                          PackageRepositoryAuth_PackageRepositoryAuthType[
                            PackageRepositoryAuth_PackageRepositoryAuthType
                              .PACKAGE_REPOSITORY_AUTH_TYPE_DOCKER_CONFIG_JSON
                          ] || ""
                        }
                        checked={
                          authMethod ===
                          PackageRepositoryAuth_PackageRepositoryAuthType.PACKAGE_REPOSITORY_AUTH_TYPE_DOCKER_CONFIG_JSON
                        }
                        onChange={handleAuthRadioButtonChange}
                        disabled={
                          !!repo.auth?.type ||
                          !getSupportedPackageRepositoryAuthTypes(plugin, type).includes(
                            PackageRepositoryAuth_PackageRepositoryAuthType.PACKAGE_REPOSITORY_AUTH_TYPE_DOCKER_CONFIG_JSON,
                          )
                        }
                      />
                    </CdsRadio>
                    <CdsRadio>
                      <label htmlFor="kubeapps-repo-auth-method-custom">
                        Custom Authorization Header
                      </label>
                      <input
                        id="kubeapps-repo-auth-method-custom"
                        type="radio"
                        name="auth"
                        value={
                          PackageRepositoryAuth_PackageRepositoryAuthType[
                            PackageRepositoryAuth_PackageRepositoryAuthType
                              .PACKAGE_REPOSITORY_AUTH_TYPE_AUTHORIZATION_HEADER
                          ] || ""
                        }
                        checked={
                          authMethod ===
                          PackageRepositoryAuth_PackageRepositoryAuthType.PACKAGE_REPOSITORY_AUTH_TYPE_AUTHORIZATION_HEADER
                        }
                        onChange={handleAuthRadioButtonChange}
                        disabled={
                          !!repo.auth?.type ||
                          !getSupportedPackageRepositoryAuthTypes(plugin, type).includes(
                            PackageRepositoryAuth_PackageRepositoryAuthType.PACKAGE_REPOSITORY_AUTH_TYPE_AUTHORIZATION_HEADER,
                          )
                        }
                      />
                    </CdsRadio>
                    <CdsRadio>
                      <label htmlFor="kubeapps-repo-auth-method-ssh">
                        SSH-based Authentication
                      </label>
                      <input
                        id="kubeapps-repo-auth-method-ssh"
                        type="radio"
                        name="auth"
                        value={
                          PackageRepositoryAuth_PackageRepositoryAuthType[
                            PackageRepositoryAuth_PackageRepositoryAuthType
                              .PACKAGE_REPOSITORY_AUTH_TYPE_SSH
                          ] || ""
                        }
                        checked={
                          authMethod ===
                          PackageRepositoryAuth_PackageRepositoryAuthType.PACKAGE_REPOSITORY_AUTH_TYPE_SSH
                        }
                        onChange={handleAuthRadioButtonChange}
                        disabled={
                          !!repo.auth?.type ||
                          !getSupportedPackageRepositoryAuthTypes(plugin, type).includes(
                            PackageRepositoryAuth_PackageRepositoryAuthType.PACKAGE_REPOSITORY_AUTH_TYPE_SSH,
                          )
                        }
                      />
                    </CdsRadio>
                    <CdsRadio>
                      <label htmlFor="kubeapps-repo-auth-method-tls">
                        TLS-based Authentication
                      </label>
                      <input
                        id="kubeapps-repo-auth-method-tls"
                        type="radio"
                        name="auth"
                        value={
                          PackageRepositoryAuth_PackageRepositoryAuthType[
                            PackageRepositoryAuth_PackageRepositoryAuthType
                              .PACKAGE_REPOSITORY_AUTH_TYPE_TLS
                          ] || ""
                        }
                        checked={
                          authMethod ===
                          PackageRepositoryAuth_PackageRepositoryAuthType.PACKAGE_REPOSITORY_AUTH_TYPE_TLS
                        }
                        onChange={handleAuthRadioButtonChange}
                        disabled={
                          !!repo.auth?.type ||
                          !getSupportedPackageRepositoryAuthTypes(plugin, type).includes(
                            PackageRepositoryAuth_PackageRepositoryAuthType.PACKAGE_REPOSITORY_AUTH_TYPE_TLS,
                          )
                        }
                      />
                    </CdsRadio>
                    <CdsRadio>
                      <label htmlFor="kubeapps-repo-auth-method-opaque">
                        Opaque-based Authentication
                      </label>
                      <input
                        id="kubeapps-repo-auth-method-opaque"
                        type="radio"
                        name="auth"
                        value={
                          PackageRepositoryAuth_PackageRepositoryAuthType[
                            PackageRepositoryAuth_PackageRepositoryAuthType
                              .PACKAGE_REPOSITORY_AUTH_TYPE_OPAQUE
                          ] || ""
                        }
                        checked={
                          authMethod ===
                          PackageRepositoryAuth_PackageRepositoryAuthType.PACKAGE_REPOSITORY_AUTH_TYPE_OPAQUE
                        }
                        onChange={handleAuthRadioButtonChange}
                        disabled={
                          !!repo.auth?.type ||
                          !getSupportedPackageRepositoryAuthTypes(plugin, type).includes(
                            PackageRepositoryAuth_PackageRepositoryAuthType.PACKAGE_REPOSITORY_AUTH_TYPE_OPAQUE,
                          )
                        }
                      />
                    </CdsRadio>
                  </CdsRadioGroup>
                  {/* End authentication selection */}

                  {/* Begin authentication details */}
                  <div cds-layout="col@xs:8">
                    {/* Begin basic authentication */}
                    <div
                      id="kubeapps-repo-auth-details-basic"
                      hidden={
                        authMethod !==
                        PackageRepositoryAuth_PackageRepositoryAuthType.PACKAGE_REPOSITORY_AUTH_TYPE_BASIC_AUTH
                      }
                    >
                      {isUserManagedSecretToggle("basic")}
                      <br />
                      {isUserManagedSecret ? (
                        secretNameInput("basic")
                      ) : (
                        <>
                          <CdsInput>
                            <label htmlFor="kubeapps-repo-username">Username</label>
                            <input
                              id="kubeapps-repo-username"
                              type="text"
                              value={basicUser || ""}
                              onChange={handleBasicUserChange}
                              placeholder="username"
                              required={
                                authMethod ===
                                PackageRepositoryAuth_PackageRepositoryAuthType.PACKAGE_REPOSITORY_AUTH_TYPE_BASIC_AUTH
                              }
                              disabled={!!repo.auth?.type}
                            />
                          </CdsInput>
                          <br />
                          <CdsInput>
                            <label htmlFor="kubeapps-repo-password">Password</label>
                            <input
                              id="kubeapps-repo-password"
                              type="password"
                              value={basicPassword || ""}
                              onChange={handleBasicPasswordChange}
                              placeholder="password"
                              required={
                                authMethod ===
                                PackageRepositoryAuth_PackageRepositoryAuthType.PACKAGE_REPOSITORY_AUTH_TYPE_BASIC_AUTH
                              }
                              disabled={!!repo.auth?.type}
                            />
                          </CdsInput>
                        </>
                      )}
                    </div>
                    {/* End basic authentication */}

                    {/* Begin http bearer authentication */}
                    <div
                      id="kubeapps-repo-auth-details-bearer"
                      hidden={
                        authMethod !==
                        PackageRepositoryAuth_PackageRepositoryAuthType.PACKAGE_REPOSITORY_AUTH_TYPE_BEARER
                      }
                    >
                      {isUserManagedSecretToggle("bearer")}
                      <br />
                      {isUserManagedSecret ? (
                        secretNameInput("bearer")
                      ) : (
                        <>
                          <CdsInput>
                            <label htmlFor="kubeapps-repo-token">Token</label>
                            <input
                              type="text"
                              value={bearerToken || ""}
                              onChange={handleAuthBearerTokenChange}
                              id="kubeapps-repo-token"
                              placeholder="xrxNcWghpRLdcPHFgVRM73rr4N7qjvjm"
                              required={
                                authMethod ===
                                PackageRepositoryAuth_PackageRepositoryAuthType.PACKAGE_REPOSITORY_AUTH_TYPE_BEARER
                              }
                              disabled={!!repo.auth?.type}
                            />
                          </CdsInput>
                        </>
                      )}
                    </div>
                    {/* End http bearer authentication */}

                    {/* Begin docker creds authentication */}
                    <div
                      id="kubeapps-repo-auth-details-docker"
                      hidden={
                        authMethod !==
                        PackageRepositoryAuth_PackageRepositoryAuthType.PACKAGE_REPOSITORY_AUTH_TYPE_DOCKER_CONFIG_JSON
                      }
                    >
                      {isUserManagedSecretToggle("docker")}
                      <br />
                      {isUserManagedSecret ? (
                        secretNameInput("docker")
                      ) : (
                        <>
                          <CdsInput className="margin-t-sm">
                            <label htmlFor="kubeapps-docker-cred-server">Server</label>
                            <input
                              id="kubeapps-docker-cred-server"
                              value={secretServer || ""}
                              onChange={handleAuthSecretServerChange}
                              placeholder="https://index.docker.io/v1/"
                              required={
                                authMethod ===
                                PackageRepositoryAuth_PackageRepositoryAuthType.PACKAGE_REPOSITORY_AUTH_TYPE_DOCKER_CONFIG_JSON
                              }
                              disabled={!!repo.auth?.type}
                            />
                          </CdsInput>
                          <br />
                          <CdsInput className="margin-t-sm">
                            <label htmlFor="kubeapps-docker-cred-username">Username</label>
                            <input
                              id="kubeapps-docker-cred-username"
                              value={secretUser || ""}
                              onChange={handleAuthSecretUserChange}
                              placeholder="Username"
                              required={
                                authMethod ===
                                PackageRepositoryAuth_PackageRepositoryAuthType.PACKAGE_REPOSITORY_AUTH_TYPE_DOCKER_CONFIG_JSON
                              }
                              disabled={!!repo.auth?.type}
                            />
                          </CdsInput>
                          <br />
                          <CdsInput className="margin-t-sm">
                            <label htmlFor="kubeapps-docker-cred-password">Password</label>
                            <input
                              id="kubeapps-docker-cred-password"
                              type="password"
                              value={secretPassword || ""}
                              onChange={handleAuthSecretPasswordChange}
                              placeholder="Password"
                              required={
                                authMethod ===
                                PackageRepositoryAuth_PackageRepositoryAuthType.PACKAGE_REPOSITORY_AUTH_TYPE_DOCKER_CONFIG_JSON
                              }
                              disabled={!!repo.auth?.type}
                            />
                          </CdsInput>
                          <br />
                          <CdsInput className="margin-t-sm">
                            <label htmlFor="kubeapps-docker-cred-email">Email</label>
                            <input
                              id="kubeapps-docker-cred-email"
                              value={secretEmail || ""}
                              onChange={handleAuthSecretEmailChange}
                              placeholder="user@example.com"
                              disabled={!!repo.auth?.type}
                            />
                          </CdsInput>
                        </>
                      )}
                    </div>
                    {/* End docker creds authentication */}

                    {/* Begin HTTP custom authentication */}
                    <div
                      id="kubeapps-repo-auth-details-custom"
                      hidden={
                        authMethod !==
                        PackageRepositoryAuth_PackageRepositoryAuthType.PACKAGE_REPOSITORY_AUTH_TYPE_AUTHORIZATION_HEADER
                      }
                    >
                      {isUserManagedSecretToggle("custom")}
                      <br />
                      {isUserManagedSecret ? (
                        secretNameInput("custom")
                      ) : (
                        <>
                          <CdsInput>
                            <label htmlFor="kubeapps-repo-custom-header">
                              Raw Authorization Header
                            </label>
                            <input
                              id="kubeapps-repo-custom-header"
                              type="text"
                              placeholder="MyAuth xrxNcWghpRLdcPHFgVRM73rr4N7qjvjm"
                              value={authCustomHeader || ""}
                              onChange={handleAuthCustomHeaderChange}
                              required={
                                authMethod ===
                                PackageRepositoryAuth_PackageRepositoryAuthType.PACKAGE_REPOSITORY_AUTH_TYPE_AUTHORIZATION_HEADER
                              }
                              disabled={!!repo.auth?.type}
                            />
                          </CdsInput>
                        </>
                      )}
                    </div>
                    {/* End HTTP custom authentication */}

                    {/* Begin SSH authentication */}
                    <div
                      id="kubeapps-repo-auth-details-ssh"
                      hidden={
                        authMethod !==
                        PackageRepositoryAuth_PackageRepositoryAuthType.PACKAGE_REPOSITORY_AUTH_TYPE_SSH
                      }
                    >
                      {isUserManagedSecretToggle("ssh")}
                      <br />
                      {isUserManagedSecret ? (
                        secretNameInput("ssh")
                      ) : (
                        <>
                          <CdsTextarea>
                            <label htmlFor="kubeapps-repo-ssh-knownhosts">
                              Raw SSH Known Hosts
                            </label>
                            <textarea
                              id="kubeapps-repo-ssh-knownhosts"
                              className="cds-textarea-fix"
                              placeholder="github.com ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIOMqqnkVzrm0SdG6UOoqKLsabgH5C9okWi0dh2l9GKJl"
                              value={sshKnownHosts || ""}
                              onChange={handleSshKnownHostsChange}
                              required={
                                authMethod ===
                                PackageRepositoryAuth_PackageRepositoryAuthType.PACKAGE_REPOSITORY_AUTH_TYPE_SSH
                              }
                              disabled={!!repo.auth?.type}
                            />
                          </CdsTextarea>
                          <br />
                          <CdsTextarea>
                            <label htmlFor="kubeapps-repo-ssh-privatekey">
                              Raw SSH Private Key
                            </label>
                            <textarea
                              id="kubeapps-repo-ssh-privatekey"
                              className="cds-textarea-fix"
                              placeholder={
                                "-----BEGIN CERTIFICATE-----\n...\n-----END CERTIFICATE-----"
                              }
                              value={sshPrivateKey || ""}
                              onChange={handleSshPrivateKeyChange}
                              required={
                                authMethod ===
                                PackageRepositoryAuth_PackageRepositoryAuthType.PACKAGE_REPOSITORY_AUTH_TYPE_SSH
                              }
                              disabled={!!repo.auth?.type}
                            />
                          </CdsTextarea>
                        </>
                      )}
                    </div>
                    {/* End SSH authentication */}

                    {/* Begin TLS authentication */}
                    <div
                      id="kubeapps-repo-auth-details-tls"
                      hidden={
                        authMethod !==
                        PackageRepositoryAuth_PackageRepositoryAuthType.PACKAGE_REPOSITORY_AUTH_TYPE_TLS
                      }
                    >
                      {isUserManagedSecretToggle("tls")}
                      <br />
                      {isUserManagedSecret ? (
                        secretNameInput("tls")
                      ) : (
                        <>
                          <CdsTextarea>
                            <label htmlFor="kubeapps-repo-tls-cert">Raw TLS Cert</label>
                            <textarea
                              id="kubeapps-repo-tls-cert"
                              className="cds-textarea-fix"
                              placeholder={
                                "-----BEGIN CERTIFICATE-----\n...\n-----END CERTIFICATE-----"
                              }
                              value={tlsAuthCert || ""}
                              onChange={handleTlsAuthCertChange}
                              required={
                                authMethod ===
                                PackageRepositoryAuth_PackageRepositoryAuthType.PACKAGE_REPOSITORY_AUTH_TYPE_TLS
                              }
                              disabled={!!repo.auth?.type}
                            />
                          </CdsTextarea>
                          <br />
                          <CdsTextarea>
                            <label htmlFor="kubeapps-repo-tls-key">Raw TLS Key</label>
                            <textarea
                              id="kubeapps-repo-tls-key"
                              className="cds-textarea-fix"
                              placeholder={
                                "-----BEGIN CERTIFICATE-----\n...\n-----END CERTIFICATE-----"
                              }
                              value={tlsAuthKey || ""}
                              onChange={handleTlsAuthKeyChange}
                              required={
                                authMethod ===
                                PackageRepositoryAuth_PackageRepositoryAuthType.PACKAGE_REPOSITORY_AUTH_TYPE_TLS
                              }
                              disabled={!!repo.auth?.type}
                            />
                          </CdsTextarea>
                        </>
                      )}
                    </div>
                    {/* End TLS authentication */}

                    {/* Begin opaque authentication */}
                    <div
                      id="kubeapps-repo-auth-details-opaque"
                      hidden={
                        authMethod !==
                        PackageRepositoryAuth_PackageRepositoryAuthType.PACKAGE_REPOSITORY_AUTH_TYPE_OPAQUE
                      }
                    >
                      {isUserManagedSecretToggle("opaque")}
                      <br />
                      {isUserManagedSecret ? (
                        secretNameInput("opaque")
                      ) : (
                        <>
                          <CdsTextarea>
                            <label htmlFor="kubeapps-repo-opaque-data">
                              Raw Opaque Data (JSON)
                            </label>
                            <textarea
                              id="kubeapps-repo-opaque-data"
                              className="cds-textarea-fix"
                              placeholder={'{\n  "username": "admin",\n  "password": "admin"\n}'}
                              value={opaqueData || ""}
                              onChange={handleOpaqueDataChange}
                              required={
                                authMethod ===
                                PackageRepositoryAuth_PackageRepositoryAuthType.PACKAGE_REPOSITORY_AUTH_TYPE_OPAQUE
                              }
                              disabled={!!repo.auth?.type}
                            />
                          </CdsTextarea>
                        </>
                      )}
                    </div>
                    {/* End opaque authentication */}
                  </div>
                  {/* End authentication details */}
                </div>
                {plugin?.name === PluginNames.PACKAGES_HELM && (
                  <div cds-layout="grid gap:lg">
                    {/* Begin imagePullSecrets selection */}
                    <CdsRadioGroup cds-layout="col@xs:4">
                      {/* eslint-disable-next-line jsx-a11y/label-has-associated-control */}
                      <label>Container Registry Credentials</label>
                      <CdsRadio>
                        <label htmlFor="kubeapps-repo-pullsecret-method-none">None (Public)</label>
                        <input
                          id="kubeapps-repo-pullsecret-method-none"
                          type="radio"
                          name="auth"
                          value={
                            PackageRepositoryAuth_PackageRepositoryAuthType[
                              PackageRepositoryAuth_PackageRepositoryAuthType
                                .PACKAGE_REPOSITORY_AUTH_TYPE_UNSPECIFIED
                            ] || ""
                          }
                          checked={
                            helmPSAuthMethod ===
                            PackageRepositoryAuth_PackageRepositoryAuthType.PACKAGE_REPOSITORY_AUTH_TYPE_UNSPECIFIED
                          }
                          onChange={handleImgPSChange}
                          disabled={
                            !!(repo?.customDetail as Partial<HelmPackageRepositoryCustomDetail>)
                              ?.imagesPullSecret?.credentials ||
                            !!(repo?.customDetail as Partial<HelmPackageRepositoryCustomDetail>)
                              ?.imagesPullSecret?.secretRef
                          }
                          required={
                            authMethod ===
                            PackageRepositoryAuth_PackageRepositoryAuthType.PACKAGE_REPOSITORY_AUTH_TYPE_DOCKER_CONFIG_JSON
                          }
                        />
                      </CdsRadio>
                      {/* TODO(agamez): for a better UX, we might want to allow copying the values from the existing auth credentials */}
                      <CdsRadio>
                        <label htmlFor="kubeapps-repo-pullsecret-method-registry">
                          Docker Registry Credentials
                        </label>
                        <input
                          id="kubeapps-repo-pullsecret-method-registry"
                          type="radio"
                          name="auth"
                          value={
                            PackageRepositoryAuth_PackageRepositoryAuthType[
                              PackageRepositoryAuth_PackageRepositoryAuthType
                                .PACKAGE_REPOSITORY_AUTH_TYPE_DOCKER_CONFIG_JSON
                            ] || ""
                          }
                          checked={
                            helmPSAuthMethod ===
                            PackageRepositoryAuth_PackageRepositoryAuthType.PACKAGE_REPOSITORY_AUTH_TYPE_DOCKER_CONFIG_JSON
                          }
                          onChange={handleImgPSChange}
                          disabled={
                            !!(repo?.customDetail as Partial<HelmPackageRepositoryCustomDetail>)
                              ?.imagesPullSecret?.credentials ||
                            !!(repo?.customDetail as Partial<HelmPackageRepositoryCustomDetail>)
                              ?.imagesPullSecret?.secretRef
                          }
                          required={
                            authMethod ===
                            PackageRepositoryAuth_PackageRepositoryAuthType.PACKAGE_REPOSITORY_AUTH_TYPE_DOCKER_CONFIG_JSON
                          }
                        />
                      </CdsRadio>
                    </CdsRadioGroup>
                    {/* End imagePullSecrets selection */}
                    {/* Begin imagePullSecrets details */}
                    <div cds-layout="col@xs:8">
                      {/* Begin docker creds authentication */}
                      <div
                        id="kubeapps-repo-imagePullSecrets-details-docker"
                        hidden={
                          helmPSAuthMethod !==
                          PackageRepositoryAuth_PackageRepositoryAuthType.PACKAGE_REPOSITORY_AUTH_TYPE_DOCKER_CONFIG_JSON
                        }
                      >
                        {authMethod ===
                          PackageRepositoryAuth_PackageRepositoryAuthType.PACKAGE_REPOSITORY_AUTH_TYPE_DOCKER_CONFIG_JSON && (
                          <CdsToggleGroup className="flex-v-center">
                            <CdsToggle>
                              {/* eslint-disable-next-line jsx-a11y/label-has-associated-control */}
                              <label>
                                {useSameAuthCreds
                                  ? "Use the same auth credentials as the container registry credentials"
                                  : "Manually set the container registry credentials"}
                              </label>
                              <input
                                type="checkbox"
                                onChange={handelUseSameAuthCredsChange}
                                checked={useSameAuthCreds}
                                disabled={
                                  !!(
                                    repo?.customDetail as Partial<HelmPackageRepositoryCustomDetail>
                                  )?.imagesPullSecret?.credentials ||
                                  !!(
                                    repo?.customDetail as Partial<HelmPackageRepositoryCustomDetail>
                                  )?.imagesPullSecret?.secretRef
                                }
                              />
                            </CdsToggle>
                          </CdsToggleGroup>
                        )}
                        <br />
                        {!useSameAuthCreds && (
                          <>
                            <CdsToggleGroup className="flex-v-center">
                              <CdsToggle>
                                <label htmlFor="kubeapps-repo-imagePullSecrets-details-docker-userManaged">
                                  {isUserManagedPSSecret
                                    ? userManagedSecretText
                                    : kubeappsManagedSecretText}
                                </label>
                                <input
                                  id="kubeapps-repo-imagePullSecrets-details-docker-userManaged"
                                  type="checkbox"
                                  onChange={handleIsUserManagedPSSecretChange}
                                  checked={isUserManagedPSSecret}
                                  disabled={
                                    !!(
                                      repo?.customDetail as Partial<HelmPackageRepositoryCustomDetail>
                                    )?.imagesPullSecret?.credentials
                                  }
                                />
                              </CdsToggle>
                            </CdsToggleGroup>
                            <br />
                            {isUserManagedPSSecret ? (
                              <>
                                <CdsInput>
                                  <label htmlFor={`kubeapps-repo-auth-secret-name-pullsecret`}>
                                    Registry Secret Name
                                  </label>
                                  <input
                                    id="kubeapps-repo-auth-secret-name-pullsecret"
                                    type="text"
                                    placeholder="my-registry-secret-name"
                                    value={secretPSName || ""}
                                    onChange={setSecretPSNameChange}
                                    required={
                                      isUserManagedPSSecret &&
                                      helmPSAuthMethod !==
                                        PackageRepositoryAuth_PackageRepositoryAuthType.PACKAGE_REPOSITORY_AUTH_TYPE_UNSPECIFIED
                                    }
                                    pattern={k8sObjectNameRegex}
                                    title="Use lower case alphanumeric characters, '-' or '.'"
                                    disabled={
                                      !!(
                                        repo?.customDetail as Partial<HelmPackageRepositoryCustomDetail>
                                      )?.imagesPullSecret?.secretRef
                                    }
                                  />
                                </CdsInput>
                                <br />
                                <CdsControlMessage>
                                  Name of the{" "}
                                  <a
                                    href="https://kubernetes.io/docs/concepts/configuration/secret/"
                                    target="_blank"
                                    rel="noopener noreferrer"
                                  >
                                    Kubernetes Secret
                                  </a>{" "}
                                  object holding the registry secret data. Please ensure that this
                                  secret's type is <i>dockerconfigjson</i>.
                                </CdsControlMessage>
                              </>
                            ) : (
                              <>
                                <CdsInput className="margin-t-sm">
                                  <label htmlFor="kubeapps-imagePullSecrets-cred-server">
                                    Server
                                  </label>
                                  <input
                                    id="kubeapps-imagePullSecrets-cred-server"
                                    value={pullSecretServer || ""}
                                    onChange={handleImgPSServerChange}
                                    placeholder="https://index.docker.io/v1/"
                                    required={
                                      !isUserManagedPSSecret &&
                                      helmPSAuthMethod !==
                                        PackageRepositoryAuth_PackageRepositoryAuthType.PACKAGE_REPOSITORY_AUTH_TYPE_UNSPECIFIED
                                    }
                                    disabled={
                                      !!(
                                        repo?.customDetail as Partial<HelmPackageRepositoryCustomDetail>
                                      )?.imagesPullSecret?.credentials
                                    }
                                  />
                                </CdsInput>
                                <br />
                                <CdsInput className="margin-t-sm">
                                  <label htmlFor="kubeapps-imagePullSecrets-cred-username">
                                    Username
                                  </label>
                                  <input
                                    id="kubeapps-imagePullSecrets-cred-username"
                                    value={pullSecretUser || ""}
                                    onChange={handleImgPSUserChange}
                                    placeholder="Username"
                                    required={
                                      !isUserManagedPSSecret &&
                                      helmPSAuthMethod !==
                                        PackageRepositoryAuth_PackageRepositoryAuthType.PACKAGE_REPOSITORY_AUTH_TYPE_UNSPECIFIED
                                    }
                                    disabled={
                                      !!(
                                        repo?.customDetail as Partial<HelmPackageRepositoryCustomDetail>
                                      )?.imagesPullSecret?.credentials
                                    }
                                  />
                                </CdsInput>
                                <br />
                                <CdsInput className="margin-t-sm">
                                  <label htmlFor="kubeapps-imagePullSecrets-cred-password">
                                    Password
                                  </label>
                                  <input
                                    id="kubeapps-imagePullSecrets-cred-password"
                                    type="password"
                                    value={pullSecretPassword || ""}
                                    onChange={handleImgPSPasswordChange}
                                    placeholder="Password"
                                    required={
                                      !isUserManagedPSSecret &&
                                      helmPSAuthMethod !==
                                        PackageRepositoryAuth_PackageRepositoryAuthType.PACKAGE_REPOSITORY_AUTH_TYPE_UNSPECIFIED
                                    }
                                    disabled={
                                      !!(
                                        repo?.customDetail as Partial<HelmPackageRepositoryCustomDetail>
                                      )?.imagesPullSecret?.credentials
                                    }
                                  />
                                </CdsInput>
                                <br />
                                <CdsInput className="margin-t-sm">
                                  <label htmlFor="kubeapps-imagePullSecrets-cred-email">
                                    Email
                                  </label>
                                  <input
                                    id="kubeapps-imagePullSecrets-cred-email"
                                    value={pullSecretEmail || ""}
                                    onChange={handleImgPSEmailChange}
                                    placeholder="user@example.com"
                                    disabled={
                                      !!(
                                        repo?.customDetail as Partial<HelmPackageRepositoryCustomDetail>
                                      )?.imagesPullSecret?.credentials
                                    }
                                  />
                                </CdsInput>
                              </>
                            )}
                          </>
                        )}
                      </div>
                      {/* End docker creds authentication */}
                    </div>
                    {/* End imagePullSecrets details */}
                  </div>
                )}
              </CdsFormGroup>
            </CdsAccordionContent>
          </CdsAccordionPanel>

          <CdsAccordionPanel
            id="panel-filtering"
            expanded={accordion[2]}
            hidden={
              type !== RepositoryStorageTypes.PACKAGE_REPOSITORY_STORAGE_OCI ||
              plugin?.name !== PluginNames.PACKAGES_HELM
            }
          >
            <CdsAccordionHeader onClick={() => toggleAccordion(2)}>Filtering</CdsAccordionHeader>
            <CdsAccordionContent>
              <CdsFormGroup layout="vertical">
                {plugin?.name === PluginNames.PACKAGES_HELM && (
                  <>
                    {type === RepositoryStorageTypes.PACKAGE_REPOSITORY_STORAGE_OCI && (
                      <CdsTextarea>
                        <label htmlFor="kubeapps-oci-repositories">
                          List of Repositories (required)
                        </label>
                        <CdsControlMessage>
                          Include a list of comma-separated OCI repositories that will be available
                          in Kubeapps.
                        </CdsControlMessage>
                        <textarea
                          id="kubeapps-oci-repositories"
                          className="cds-textarea-fix"
                          placeholder={"nginx, jenkins"}
                          value={ociRepositories || ""}
                          onChange={handleOCIRepositoriesChange}
                          required={type === RepositoryStorageTypes.PACKAGE_REPOSITORY_STORAGE_OCI}
                        />
                      </CdsTextarea>
                    )}
                    <CdsTextarea>
                      <label htmlFor="kubeapps-repo-filter-repositories">
                        Filter Applications (optional)
                      </label>
                      <CdsControlMessage>
                        Comma-separated list of applications to be included or excluded (all will be
                        included by default).
                      </CdsControlMessage>
                      <textarea
                        className="cds-textarea-fix"
                        id="kubeapps-repo-filter-repositories"
                        placeholder={"nginx, jenkins"}
                        value={filterNames || ""}
                        onChange={handleFilterNamesChange}
                      />
                    </CdsTextarea>
                    <CdsCheckbox className="reduced-margin-top">
                      <label htmlFor="kubeapps-repo-filter-exclude">Exclude Packages</label>
                      <CdsControlMessage>
                        Exclude packages matching the given filter
                      </CdsControlMessage>
                      <input
                        id="kubeapps-repo-filter-exclude"
                        type="checkbox"
                        onChange={handleFilterExcludeChange}
                        checked={filterExclude}
                      />
                    </CdsCheckbox>
                    <CdsCheckbox className="reduced-margin-top">
                      <label htmlFor="kubeapps-repo-filter-regex">Regular Expression</label>
                      <CdsControlMessage>
                        Mark this box to treat the filter as a regular expression
                      </CdsControlMessage>
                      <input
                        id="kubeapps-repo-filter-regex"
                        type="checkbox"
                        onChange={handleFilterRegexChange}
                        checked={filterRegex}
                      />
                    </CdsCheckbox>
                  </>
                )}
              </CdsFormGroup>
            </CdsAccordionContent>
          </CdsAccordionPanel>
          <CdsAccordionPanel id="panel-advanced" expanded={accordion[3]}>
            <CdsAccordionHeader onClick={() => toggleAccordion(3)}>Advanced</CdsAccordionHeader>
            <CdsAccordionContent>
              <CdsFormGroup layout="vertical">
                {([PluginNames.PACKAGES_FLUX, PluginNames.PACKAGES_KAPP] as string[]).includes(
                  plugin?.name,
                ) && (
                  <CdsInput>
                    <label htmlFor="kubeapps-repo-interval">Synchronization Interval</label>
                    <input
                      id="kubeapps-repo-interval"
                      type="text"
                      placeholder={initialInterval}
                      value={syncInterval || ""}
                      onChange={handleIntervalChange}
                    />
                    <CdsControlMessage>
                      Time to wait between each synchronization of the repository. Expressed as a{" "}
                      <a
                        href={"https://pkg.go.dev/time#ParseDuration"}
                        target="_blank"
                        rel="noopener noreferrer"
                      >
                        Golang duration
                      </a>{" "}
                      according to the Kubernetes specification.
                    </CdsControlMessage>
                  </CdsInput>
                )}
                {plugin?.name === PluginNames.PACKAGES_HELM && (
                  <CdsCheckbox>
                    <label htmlFor="kubeapps-repo-performvalidation">Perform Validation</label>
                    <CdsControlMessage>
                      Ensure that a connection can be established with the repository before adding
                      it.
                    </CdsControlMessage>
                    <input
                      id="kubeapps-repo-performvalidation"
                      type="checkbox"
                      onChange={handlePerformValidationChange}
                      checked={performValidation}
                    />
                  </CdsCheckbox>
                )}

                {([PluginNames.PACKAGES_HELM, PluginNames.PACKAGES_FLUX] as string[]).includes(
                  plugin?.name,
                ) && (
                  <>
                    <CdsToggleGroup>
                      <CdsToggle>
                        <label htmlFor="kubeapps-repo-usermanagedsecret">
                          {isUserManagedCASecret
                            ? userManagedSecretText
                            : kubeappsManagedSecretText}
                        </label>
                        <input
                          id="kubeapps-repo-usermanagedsecret"
                          type="checkbox"
                          onChange={handleIsUserManagedCASecretChange}
                          checked={isUserManagedCASecret}
                          disabled={skipTLS}
                        />
                      </CdsToggle>
                    </CdsToggleGroup>
                    {isUserManagedCASecret ? (
                      <>
                        <CdsInput>
                          <label htmlFor="kubeapps-repo-secret-ca">
                            Custom CA Secret Name (optional)
                          </label>
                          <input
                            id="kubeapps-repo-secret-ca"
                            type="text"
                            placeholder="my-ca-secret"
                            pattern={k8sObjectNameRegex}
                            title="Use lower case alphanumeric characters, '-' or '.'"
                            value={secretTLSName || ""}
                            disabled={skipTLS}
                            onChange={handleSecretTLSNameChange}
                          />
                        </CdsInput>
                        <br />
                        <CdsControlMessage>
                          Name of the{" "}
                          <a
                            href="https://kubernetes.io/docs/concepts/configuration/secret/"
                            target="_blank"
                            rel="noopener noreferrer"
                          >
                            Kubernetes Secret
                          </a>{" "}
                          object holding the TLS Certificate Authority data.
                        </CdsControlMessage>
                      </>
                    ) : (
                      <>
                        <CdsTextarea layout="vertical">
                          <label htmlFor="kubeapps-repo-custom-ca">
                            Custom CA Certificate (optional)
                          </label>
                          <textarea
                            id="kubeapps-repo-custom-ca"
                            placeholder={
                              "-----BEGIN CERTIFICATE-----\n...\n-----END CERTIFICATE-----"
                            }
                            className="cds-textarea-fix"
                            value={customCA || ""}
                            disabled={skipTLS}
                            onChange={handleCustomCAChange}
                          />
                          <CdsControlMessage>
                            Custom Certificate Authority (CA) to use when connecting to the
                            repository.
                          </CdsControlMessage>
                        </CdsTextarea>
                      </>
                    )}
                  </>
                )}

                {([PluginNames.PACKAGES_HELM, PluginNames.PACKAGES_FLUX] as string[]).includes(
                  plugin?.name,
                ) && (
                  <CdsCheckbox className="reduced-margin-top">
                    <label htmlFor="kubeapps-repo-skip-tls">Skip TLS Verification</label>
                    <input
                      id="kubeapps-repo-skip-tls"
                      type="checkbox"
                      checked={skipTLS}
                      onChange={handleSkipTLSChange}
                    />
                    <CdsControlMessage>
                      If enabled, the TLS certificate will not be verified (potentially insecure).
                    </CdsControlMessage>
                  </CdsCheckbox>
                )}

                {([PluginNames.PACKAGES_HELM, PluginNames.PACKAGES_FLUX] as string[]).includes(
                  plugin?.name,
                ) && (
                  <CdsCheckbox className="reduced-margin-top">
                    <label htmlFor="kubeapps-repo-pass-credentials">
                      Pass Credentials to 3rd party URLs
                    </label>
                    <input
                      id="kubeapps-repo-pass-credentials"
                      type="checkbox"
                      checked={passCredentials}
                      onChange={handlePassCredentialsChange}
                    />
                    <CdsControlMessage>
                      If enabled, the same credentials will be sent to those URLs for fetching the
                      icon and the tarball files (potentially insecure).
                    </CdsControlMessage>
                  </CdsCheckbox>
                )}
              </CdsFormGroup>
            </CdsAccordionContent>
          </CdsAccordionPanel>
        </CdsAccordion>

        {repo &&
          !repo.namespaceScoped &&
          isGlobalNamespace(namespace, plugin?.name, currentNsConfig) && (
            <p>
              <strong>NOTE:</strong> This Package Repository is assigned to the "{namespace}" global
              namespace. Consequently, its packages will be available for installation in every
              namespace and cluster.
            </p>
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
        {deleteError && (
          <Alert theme="danger">
            An error occurred while deleting the repository: {deleteError.message}
          </Alert>
        )}
        {fetchError && (
          <Alert theme="danger">
            An error occurred while fetching the repository: {fetchError.message}
          </Alert>
        )}
        <div className="margin-t-xl">
          <CdsButton type="submit" disabled={isFetching}>
            {isFetching
              ? "Loading..."
              : `${repo.name ? `Update '${repo.name}'` : "Install"} Repository`}
          </CdsButton>
        </div>
      </form>
    </>
  );
}
