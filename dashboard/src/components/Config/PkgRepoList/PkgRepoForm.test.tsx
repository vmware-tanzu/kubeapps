// Copyright 2020-2023 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { Any } from "@bufbuild/protobuf";
import { CdsButton } from "@cds/react/button";
import { act, waitFor } from "@testing-library/react";
import actions from "actions";
import AlertGroup from "components/AlertGroup";
import {
  DockerCredentials,
  OpaqueCredentials,
  PackageRepositoryAuth_PackageRepositoryAuthType,
  PackageRepositoryDetail,
  PackageRepositoryReference,
  PackageRepositorySummary,
  PackageRepositoryTlsConfig,
  SshCredentials,
  TlsCertKey,
  UsernamePassword,
} from "gen/kubeappsapis/core/packages/v1alpha1/repositories_pb";
import { Plugin } from "gen/kubeappsapis/core/plugins/v1alpha1/plugins_pb";
import { FluxPackageRepositoryCustomDetail } from "gen/kubeappsapis/plugins/fluxv2/packages/v1alpha1/fluxv2_pb";
import {
  HelmPackageRepositoryCustomDetail,
  RepositoryFilterRule,
} from "gen/kubeappsapis/plugins/helm/packages/v1alpha1/helm_pb";
import * as ReactRedux from "react-redux";
import { IPackageRepositoryState } from "reducers/repos";
import { defaultStore, getStore, mountWrapper } from "shared/specs/mountWrapper";
import { IPkgRepoFormData, IStoreState, PluginNames, RepositoryStorageTypes } from "shared/types";
import { IPkgRepoFormProps, PkgRepoForm } from "./PkgRepoForm";

const defaultProps = {
  onSubmit: jest.fn(),
  namespace: "default",
  kubeappsNamespace: "kubeapps",
  helmGlobalNamespace: "kubeapps",
  carvelGlobalNamespace: "carvel-global",
  packageRepoRef: new PackageRepositoryReference({
    identifier: "test",
    context: { cluster: "default", namespace: "default" },
  }),
} as IPkgRepoFormProps;

const defaultState = {
  repos: {
    isFetching: false,
    reposSummaries: [] as PackageRepositorySummary[],
    repoDetail: {} as PackageRepositoryDetail,
    errors: {},
  } as IPackageRepositoryState,
} as IStoreState;

const pkgRepoFormData = {
  plugin: { name: PluginNames.PACKAGES_HELM, version: "v1alpha1" } as Plugin,
  authHeader: "",
  authMethod: PackageRepositoryAuth_PackageRepositoryAuthType.UNSPECIFIED,
  isUserManaged: false,
  basicAuth: new UsernamePassword({
    password: "",
    username: "",
  }),
  customCA: "",
  customDetail: {
    ociRepositories: [],
    performValidation: true,
    filterRules: undefined,
    nodeSelector: {},
    tolerations: [],
  },
  description: "",
  dockerRegCreds: new DockerCredentials({
    password: "",
    username: "",
    email: "",
    server: "",
  }),
  interval: "10m",
  name: "",
  passCredentials: false,
  secretAuthName: "",
  secretTLSName: "",
  skipTLS: false,
  type: RepositoryStorageTypes.PACKAGE_REPOSITORY_STORAGE_HELM,
  url: "",
  opaqueCreds: new OpaqueCredentials({
    data: {},
  }),
  sshCreds: new SshCredentials({
    knownHosts: "",
    privateKey: "",
  }),
  tlsCertKey: new TlsCertKey({
    cert: "",
    key: "",
  }),
  namespace: "default",
  isNamespaceScoped: true,
} as IPkgRepoFormData;

let spyOnUseDispatch: jest.SpyInstance;
const kubeActions = { ...actions.kube };
beforeEach(() => {
  actions.repos = {
    ...actions.repos,
  };
  const mockDispatch = jest.fn(r => r);
  spyOnUseDispatch = jest.spyOn(ReactRedux, "useDispatch").mockReturnValue(mockDispatch);
});

afterEach(() => {
  actions.kube = { ...kubeActions };
  spyOnUseDispatch.mockRestore();
});

it("disables the submit button while loading", async () => {
  let wrapper: any;
  await act(async () => {
    wrapper = mountWrapper(
      getStore({
        ...defaultState,
        repos: { ...defaultState.repos, isFetching: true } as IPackageRepositoryState,
      } as Partial<IStoreState>),
      <PkgRepoForm {...defaultProps} />,
    );
  });
  await waitFor(() => {
    wrapper.update();
    expect(
      wrapper
        .find(CdsButton)
        .filterWhere((b: any) => b.html().includes("Loading"))
        .prop("disabled"),
    ).toBe(true);
  });
});

it("submit button can not be fired more than once", async () => {
  const onSubmit = jest.fn().mockReturnValue(true);
  const onAfterInstall = jest.fn().mockReturnValue(true);
  let wrapper: any;
  await act(async () => {
    wrapper = mountWrapper(
      defaultStore,
      <PkgRepoForm {...defaultProps} onSubmit={onSubmit} onAfterInstall={onAfterInstall} />,
    );
  });
  const installButton = wrapper
    .find(CdsButton)
    .filterWhere((b: any) => b.html().includes("Install Repo"));
  await act(async () => {
    Promise.all([
      installButton.simulate("submit"),
      installButton.simulate("submit"),
      installButton.simulate("submit"),
    ]);
  });
  wrapper.update();
  expect(onSubmit.mock.calls.length).toBe(1);
});

it("shows an error creating a repo", async () => {
  let wrapper: any;
  await act(async () => {
    wrapper = mountWrapper(
      getStore({
        repos: {
          errors: { create: new Error("boom!") },
        } as IPackageRepositoryState,
      } as Partial<IStoreState>),
      <PkgRepoForm {...defaultProps} />,
    );
  });
  wrapper.update();
  expect(wrapper.find(AlertGroup)).toIncludeText("boom!");
});

it("shows an error deleting a repo", async () => {
  let wrapper: any;
  await act(async () => {
    wrapper = mountWrapper(
      getStore({
        repos: {
          errors: { delete: new Error("boom!") },
        } as IPackageRepositoryState,
      } as Partial<IStoreState>),
      <PkgRepoForm {...defaultProps} />,
    );
  });
  wrapper.update();
  expect(wrapper.find(AlertGroup)).toIncludeText("boom!");
});

it("shows an error fetching a repo", async () => {
  let wrapper: any;
  await act(async () => {
    wrapper = mountWrapper(
      getStore({
        repos: {
          errors: { fetch: new Error("boom!") },
        } as IPackageRepositoryState,
      } as Partial<IStoreState>),
      <PkgRepoForm {...defaultProps} />,
    );
  });
  wrapper.update();
  expect(wrapper.find(AlertGroup)).toIncludeText("boom!");
});

it("shows an error updating a repo", async () => {
  let wrapper: any;
  await act(async () => {
    wrapper = mountWrapper(
      getStore({
        repos: {
          errors: { update: new Error("boom!") },
        } as IPackageRepositoryState,
      } as Partial<IStoreState>),
      <PkgRepoForm {...defaultProps} />,
    );
  });
  wrapper.update();
  expect(wrapper.find(AlertGroup)).toIncludeText("boom!");
});

it("disables unavailable plugins", async () => {
  let wrapper: any;
  await act(async () => {
    wrapper = mountWrapper(
      getStore({
        config: {
          configuredPlugins: [{ name: PluginNames.PACKAGES_HELM, version: "v1alpha1" }],
        },
      } as Partial<IStoreState>),
      <PkgRepoForm {...defaultProps} />,
    );
  });
  expect(wrapper.find("#kubeapps-plugin-helm").prop("disabled")).toBe(false);
  expect(wrapper.find("#kubeapps-plugin-fluxv2").prop("disabled")).toBe(true);
  expect(wrapper.find("#kubeapps-plugin-kappcontroller").prop("disabled")).toBe(true);
});

it("should call the install method", async () => {
  const install = jest.fn().mockReturnValue(true);
  actions.repos = {
    ...actions.repos,
  };

  let wrapper: any;
  await act(async () => {
    wrapper = mountWrapper(defaultStore, <PkgRepoForm {...defaultProps} onSubmit={install} />);
  });
  const form = wrapper.find("form");
  await act(async () => {
    await (form.prop("onSubmit") as (e: any) => Promise<any>)({ preventDefault: jest.fn() });
  });
  wrapper.update();
  expect(install).toHaveBeenCalled();
});

it("should call the install method with OCI information", async () => {
  const install = jest.fn().mockReturnValue(true);
  actions.repos = {
    ...actions.repos,
  };
  let wrapper: any;
  await act(async () => {
    wrapper = mountWrapper(defaultStore, <PkgRepoForm {...defaultProps} onSubmit={install} />);
  });
  wrapper.find("#kubeapps-plugin-helm").simulate("change");
  wrapper.find("#kubeapps-repo-name").simulate("change", { target: { value: "oci-repo" } });
  wrapper.find("#kubeapps-repo-url").simulate("change", { target: { value: "oci.repo" } });
  wrapper.find("#kubeapps-repo-type-oci").simulate("change");
  wrapper
    .find("#kubeapps-oci-repositories")
    .simulate("change", { target: { value: "apache, jenkins" } });
  const form = wrapper.find("form");
  await act(async () => {
    await (form.prop("onSubmit") as (e: any) => Promise<any>)({ preventDefault: jest.fn() });
  });
  wrapper.update();
  expect(install).toHaveBeenCalledWith({
    ...pkgRepoFormData,
    name: "oci-repo",
    type: "oci",
    url: "https://oci.repo",
    plugin: { name: PluginNames.PACKAGES_HELM, version: "v1alpha1" },
    customDetail: {
      ociRepositories: ["apache", "jenkins"],
      filterRule: undefined,
      performValidation: true,
      nodeSelector: {},
      tolerations: [],
    },
    interval: "10m",
    description: undefined,
  } as unknown as IPkgRepoFormData);
});

it("should call the install method with an OCI protocol if specified", async () => {
  const install = jest.fn().mockReturnValue(true);
  actions.repos = {
    ...actions.repos,
  };
  let wrapper: any;
  await act(async () => {
    wrapper = mountWrapper(defaultStore, <PkgRepoForm {...defaultProps} onSubmit={install} />);
  });
  wrapper.find("#kubeapps-plugin-helm").simulate("change");
  wrapper.find("#kubeapps-repo-name").simulate("change", { target: { value: "oci-repo" } });
  wrapper.find("#kubeapps-repo-url").simulate("change", { target: { value: "oci://oci.repo" } });
  wrapper.find("#kubeapps-repo-type-oci").simulate("change");
  wrapper
    .find("#kubeapps-oci-repositories")
    .simulate("change", { target: { value: "apache, jenkins" } });
  const form = wrapper.find("form");
  await act(async () => {
    await (form.prop("onSubmit") as (e: any) => Promise<any>)({ preventDefault: jest.fn() });
  });
  wrapper.update();
  expect(install).toHaveBeenCalledWith({
    ...pkgRepoFormData,
    name: "oci-repo",
    type: "oci",
    url: "oci://oci.repo",
    plugin: { name: PluginNames.PACKAGES_HELM, version: "v1alpha1" },
    customDetail: {
      ociRepositories: ["apache", "jenkins"],
      filterRule: undefined,
      performValidation: true,
      nodeSelector: {},
      tolerations: [],
    },
    interval: "10m",
    description: undefined,
  } as unknown as IPkgRepoFormData);
});

it("should call the install skipping TLS verification", async () => {
  const install = jest.fn().mockReturnValue(true);
  actions.repos = {
    ...actions.repos,
  };
  let wrapper: any;
  await act(async () => {
    wrapper = mountWrapper(defaultStore, <PkgRepoForm {...defaultProps} onSubmit={install} />);
  });
  wrapper.find("#kubeapps-plugin-helm").simulate("change");
  wrapper.find("#kubeapps-repo-name").simulate("change", { target: { value: "helm-repo" } });
  wrapper.find("#kubeapps-repo-url").simulate("change", { target: { value: "helm.repo" } });
  wrapper.find("#kubeapps-repo-type-helm").simulate("change");
  wrapper.find("#kubeapps-repo-skip-tls").simulate("change");
  const form = wrapper.find("form");
  await act(async () => {
    await (form.prop("onSubmit") as (e: any) => Promise<any>)({ preventDefault: jest.fn() });
  });
  wrapper.update();
  expect(install).toHaveBeenCalledWith({
    ...pkgRepoFormData,
    name: "helm-repo",
    type: "helm",
    url: "https://helm.repo",
    plugin: { name: PluginNames.PACKAGES_HELM, version: "v1alpha1" },
    customDetail: {
      ociRepositories: [],
      filterRule: undefined,
      performValidation: true,
      nodeSelector: {},
      tolerations: [],
    },
    skipTLS: true,
    interval: "10m",
    description: undefined,
  } as unknown as IPkgRepoFormData);
});

it("should call the install passing credentials", async () => {
  const install = jest.fn().mockReturnValue(true);
  actions.repos = {
    ...actions.repos,
  };
  let wrapper: any;
  await act(async () => {
    wrapper = mountWrapper(defaultStore, <PkgRepoForm {...defaultProps} onSubmit={install} />);
  });
  wrapper.find("#kubeapps-plugin-helm").simulate("change");
  wrapper.find("#kubeapps-repo-name").simulate("change", { target: { value: "helm-repo" } });
  wrapper.find("#kubeapps-repo-url").simulate("change", { target: { value: "helm.repo" } });
  wrapper.find("#kubeapps-repo-type-helm").simulate("change");
  wrapper.find("#kubeapps-repo-pass-credentials").simulate("change");
  const form = wrapper.find("form");
  await act(async () => {
    await (form.prop("onSubmit") as (e: any) => Promise<any>)({ preventDefault: jest.fn() });
  });
  wrapper.update();
  expect(install).toHaveBeenCalledWith({
    ...pkgRepoFormData,
    name: "helm-repo",
    type: "helm",
    url: "https://helm.repo",
    plugin: { name: PluginNames.PACKAGES_HELM, version: "v1alpha1" },
    customDetail: {
      ociRepositories: [],
      filterRule: undefined,
      performValidation: true,
      nodeSelector: {},
      tolerations: [],
    },
    passCredentials: true,
    interval: "10m",
    description: undefined,
  } as unknown as IPkgRepoFormData);
});

it("should call the install using the interval", async () => {
  const install = jest.fn().mockReturnValue(true);
  actions.repos = {
    ...actions.repos,
  };
  let wrapper: any;
  await act(async () => {
    wrapper = mountWrapper(defaultStore, <PkgRepoForm {...defaultProps} onSubmit={install} />);
  });
  wrapper.find("#kubeapps-plugin-helm").simulate("change");
  wrapper.find("#kubeapps-repo-name").simulate("change", { target: { value: "helm-repo" } });
  wrapper.find("#kubeapps-repo-url").simulate("change", { target: { value: "helm.repo" } });
  wrapper.find("#kubeapps-repo-type-helm").simulate("change");
  wrapper.find("#kubeapps-repo-interval").simulate("change", { target: { value: "15m" } });
  const form = wrapper.find("form");
  await act(async () => {
    await (form.prop("onSubmit") as (e: any) => Promise<any>)({ preventDefault: jest.fn() });
  });
  wrapper.update();
  expect(install).toHaveBeenCalledWith({
    ...pkgRepoFormData,
    name: "helm-repo",
    type: "helm",
    url: "https://helm.repo",
    description: undefined,
    plugin: { name: PluginNames.PACKAGES_HELM, version: "v1alpha1" },
    interval: "15m",
  } as unknown as IPkgRepoFormData);
});

describe("when using a filter", () => {
  it("should call the install method with a filter", async () => {
    const install = jest.fn().mockReturnValue(true);
    let wrapper: any;
    await act(async () => {
      wrapper = mountWrapper(defaultStore, <PkgRepoForm {...defaultProps} onSubmit={install} />);
    });
    wrapper.find("#kubeapps-plugin-helm").simulate("change");
    wrapper.find("#kubeapps-repo-name").simulate("change", { target: { value: "helm-repo" } });
    wrapper.find("#kubeapps-repo-url").simulate("change", { target: { value: "helm.repo" } });
    wrapper.find("#kubeapps-repo-type-helm").simulate("change");
    wrapper
      .find("#kubeapps-repo-filter-repositories")
      .simulate("change", { target: { value: "nginx, wordpress" } });
    const form = wrapper.find("form");
    await act(async () => {
      await (form.prop("onSubmit") as (e: any) => Promise<any>)({ preventDefault: jest.fn() });
    });
    wrapper.update();
    expect(install).toHaveBeenCalledWith({
      ...pkgRepoFormData,
      name: "helm-repo",
      type: "helm",
      url: "https://helm.repo",
      plugin: { name: PluginNames.PACKAGES_HELM, version: "v1alpha1" },
      customDetail: {
        ociRepositories: [],
        filterRule: {
          jq: ".name == $var0 or .name == $var1",
          variables: { $var0: "nginx", $var1: "wordpress" },
        },
        performValidation: true,
        nodeSelector: {},
        tolerations: [],
      },
      interval: "10m",
      description: undefined,
    } as unknown as IPkgRepoFormData);
  });

  it("should call the install method with a filter excluding a regex", async () => {
    const install = jest.fn().mockReturnValue(true);
    let wrapper: any;
    await act(async () => {
      wrapper = mountWrapper(defaultStore, <PkgRepoForm {...defaultProps} onSubmit={install} />);
    });
    wrapper.find("#kubeapps-plugin-helm").simulate("change");
    wrapper.find("#kubeapps-repo-name").simulate("change", { target: { value: "helm-repo" } });
    wrapper.find("#kubeapps-repo-url").simulate("change", { target: { value: "helm.repo" } });
    wrapper.find("#kubeapps-repo-type-helm").simulate("change");
    wrapper
      .find("#kubeapps-repo-filter-repositories")
      .simulate("change", { target: { value: "nginx" } });

    wrapper.find("#kubeapps-repo-filter-exclude").simulate("change");
    wrapper.find("#kubeapps-repo-filter-regex").simulate("change");
    const form = wrapper.find("form");
    await act(async () => {
      await (form.prop("onSubmit") as (e: any) => Promise<any>)({ preventDefault: jest.fn() });
    });
    wrapper.update();
    expect(install).toHaveBeenCalledWith({
      ...pkgRepoFormData,
      name: "helm-repo",
      type: "helm",
      url: "https://helm.repo",
      plugin: { name: PluginNames.PACKAGES_HELM, version: "v1alpha1" },
      customDetail: {
        ociRepositories: [],
        nodeSelector: {},
        filterRule: {
          jq: ".name | test($var) | not",
          variables: { $var: "nginx" },
        },
        performValidation: true,
        tolerations: [],
      },
      interval: "10m",
      description: undefined,
    } as unknown as IPkgRepoFormData);
  });

  it("ignore the filter for the OCI case", async () => {
    const install = jest.fn().mockReturnValue(true);
    actions.repos = {
      ...actions.repos,
    };
    let wrapper: any;
    await act(async () => {
      wrapper = mountWrapper(defaultStore, <PkgRepoForm {...defaultProps} onSubmit={install} />);
    });
    wrapper.find("#kubeapps-plugin-helm").simulate("change");
    wrapper.find("#kubeapps-repo-name").simulate("change", { target: { value: "oci-repo" } });
    wrapper.find("#kubeapps-repo-url").simulate("change", { target: { value: "oci.repo" } });
    wrapper.find("#kubeapps-repo-type-oci").simulate("change");
    wrapper
      .find("#kubeapps-oci-repositories")
      .simulate("change", { target: { value: "apache, jenkins" } });
    wrapper
      .find("#kubeapps-repo-filter-repositories")
      .simulate("change", { target: { value: "nginx, wordpress" } });
    const form = wrapper.find("form");
    await act(async () => {
      await (form.prop("onSubmit") as (e: any) => Promise<any>)({ preventDefault: jest.fn() });
    });
    wrapper.update();
    wrapper.find("#kubeapps-repo-type-oci").simulate("change");
    expect(install).toHaveBeenCalledWith({
      ...pkgRepoFormData,
      name: "oci-repo",
      type: "oci",
      url: "https://oci.repo",
      plugin: { name: PluginNames.PACKAGES_HELM, version: "v1alpha1" },
      customDetail: {
        ociRepositories: ["apache", "jenkins"],
        nodeSelector: {},
        performValidation: true,
        tolerations: [],
      },
      interval: "10m",
      description: undefined,
    } as unknown as IPkgRepoFormData);
  });
});

it("should call the install method with a description", async () => {
  const install = jest.fn().mockReturnValue(true);
  let wrapper: any;
  await act(async () => {
    wrapper = mountWrapper(defaultStore, <PkgRepoForm {...defaultProps} onSubmit={install} />);
  });
  wrapper.find("#kubeapps-plugin-helm").simulate("change");
  wrapper.find("#kubeapps-repo-name").simulate("change", { target: { value: "helm-repo" } });
  wrapper.find("#kubeapps-repo-url").simulate("change", { target: { value: "https://helm.repo" } });
  wrapper
    .find("#kubeapps-repo-description")
    .simulate("change", { target: { value: "description test" } });
  const form = wrapper.find("form");
  await act(async () => {
    await (form.prop("onSubmit") as (e: any) => Promise<any>)({ preventDefault: jest.fn() });
  });
  wrapper.update();
  expect(install).toHaveBeenCalledWith({
    ...pkgRepoFormData,
    name: "helm-repo",
    type: "helm",
    url: "https://helm.repo",
    customDetail: {
      ociRepositories: [],
      nodeSelector: {},
      filterRule: undefined,
      performValidation: true,
      tolerations: [],
    },
    interval: "10m",
    description: "description test",
  } as unknown as IPkgRepoFormData);
});

it("should not show the list of OCI repositories if using a Helm repo (default)", async () => {
  let wrapper: any;
  await act(async () => {
    wrapper = mountWrapper(defaultStore, <PkgRepoForm {...defaultProps} />);
  });
  wrapper.find("#kubeapps-plugin-helm").simulate("change");
  wrapper.find("#kubeapps-repo-name").simulate("change", { target: { value: "helm-repo" } });
  wrapper.find("#kubeapps-repo-url").simulate("change", { target: { value: "helm.repo" } });
  wrapper.find("#kubeapps-repo-type-helm").simulate("change");
  wrapper.update();
  expect(wrapper.find("#kubeapps-oci-repositories")).not.toExist();
});

describe("when the repository info is already populated", () => {
  const packageRepoRef = {
    identifier: "helm-repo",
    context: defaultProps.packageRepoRef?.context,
    plugin: { name: PluginNames.PACKAGES_HELM, version: "v1alpha1" },
  } as PackageRepositoryReference;
  const repo = {
    name: "",
    description: "",
    interval: "10m",
    packageRepoRef: packageRepoRef,
    namespaceScoped: false,
    type: "",
    url: "",
  } as PackageRepositoryDetail;

  it("should parse the existing name", async () => {
    const testRepo = {
      ...repo,
      name: "foo",
    } as PackageRepositoryDetail;
    let wrapper: any;
    await act(async () => {
      wrapper = mountWrapper(
        getStore({
          ...defaultState,
          repos: { ...defaultState.repos, repoDetail: testRepo } as IPackageRepositoryState,
        } as Partial<IStoreState>),
        <PkgRepoForm {...defaultProps} packageRepoRef={packageRepoRef} />,
      );
    });
    await waitFor(() => {
      wrapper.update();
      expect(wrapper.find("#kubeapps-repo-name").prop("value")).toBe("foo");
    });
    // It should also deactivate the name input if it's already been set
    expect(wrapper.find("#kubeapps-repo-name").prop("disabled")).toBe(true);
  });

  it("should parse the existing url", async () => {
    const testRepo = {
      ...repo,
      url: "http://repo",
    } as PackageRepositoryDetail;
    let wrapper: any;
    await act(async () => {
      wrapper = mountWrapper(
        getStore({
          ...defaultState,
          repos: { ...defaultState.repos, repoDetail: testRepo } as IPackageRepositoryState,
        } as Partial<IStoreState>),
        <PkgRepoForm {...defaultProps} packageRepoRef={packageRepoRef} />,
      );
    });
    await waitFor(() => {
      wrapper.update();
      expect(wrapper.find("#kubeapps-repo-url").prop("value")).toBe("http://repo");
    });
  });

  it("should parse the existing filter (simple)", async () => {
    const testRepo = new PackageRepositoryDetail({
      ...repo,
      type: "helm",
      customDetail: new Any({
        value: new HelmPackageRepositoryCustomDetail({
          filterRule: {
            jq: ".name == $var0 or .name == $var1",
            variables: { $var0: "nginx", $var1: "wordpress" },
          },
        }).toBinary(),
      }),
    });

    let wrapper: any;
    await act(async () => {
      wrapper = mountWrapper(
        getStore({
          ...defaultState,
          repos: { ...defaultState.repos, repoDetail: testRepo } as IPackageRepositoryState,
        } as Partial<IStoreState>),
        <PkgRepoForm {...defaultProps} packageRepoRef={packageRepoRef} />,
      );
    });
    await waitFor(() => {
      wrapper.update();
      expect(wrapper.find("#kubeapps-repo-filter-repositories").prop("value")).toBe(
        "nginx, wordpress",
      );
    });
    expect(wrapper.find("#kubeapps-repo-filter-exclude")).not.toBeChecked();
    expect(wrapper.find("#kubeapps-repo-filter-regex")).not.toBeChecked();
  });

  it("should parse the existing filter (negated regex)", async () => {
    const testRepo = new PackageRepositoryDetail({
      ...repo,
      type: "helm",
      customDetail: new Any({
        value: new HelmPackageRepositoryCustomDetail({
          filterRule: new RepositoryFilterRule({
            jq: ".name | test($var) | not",
            variables: { $var: "nginx" },
          }),
        }).toBinary(),
      }),
    });

    let wrapper: any;
    await act(async () => {
      wrapper = mountWrapper(
        getStore({
          ...defaultState,
          repos: { ...defaultState.repos, repoDetail: testRepo } as IPackageRepositoryState,
        } as Partial<IStoreState>),
        <PkgRepoForm {...defaultProps} packageRepoRef={packageRepoRef} />,
      );
    });
    await waitFor(() => {
      wrapper.update();
      expect(wrapper.find("#kubeapps-repo-filter-repositories").prop("value")).toBe("nginx");
    });
    expect(wrapper.find("#kubeapps-repo-filter-exclude")).toBeChecked();
    expect(wrapper.find("#kubeapps-repo-filter-regex")).toBeChecked();
  });

  it("should parse the existing type", async () => {
    const testRepo = {
      ...repo,
      type: "oci",
    } as PackageRepositoryDetail;
    let wrapper: any;
    await act(async () => {
      wrapper = mountWrapper(
        getStore({
          ...defaultState,
          repos: { ...defaultState.repos, repoDetail: testRepo } as IPackageRepositoryState,
        } as Partial<IStoreState>),
        <PkgRepoForm {...defaultProps} packageRepoRef={packageRepoRef} />,
      );
    });
    await waitFor(() => {
      wrapper.update();
      wrapper.find("#kubeapps-plugin-helm").simulate("change");
      expect(wrapper.find("#kubeapps-repo-type-oci")).toBeChecked();
    });
    expect(wrapper.find("#kubeapps-oci-repositories")).toExist();
  });

  it("should parse the existing skip tls config", async () => {
    const testRepo = {
      ...repo,
      tlsConfig: { insecureSkipVerify: true },
    } as PackageRepositoryDetail;
    let wrapper: any;
    await act(async () => {
      wrapper = mountWrapper(
        getStore({
          ...defaultState,
          repos: { ...defaultState.repos, repoDetail: testRepo } as IPackageRepositoryState,
        } as Partial<IStoreState>),
        <PkgRepoForm {...defaultProps} packageRepoRef={packageRepoRef} />,
      );
    });
    wrapper.update();
    expect(wrapper.find("#kubeapps-repo-skip-tls")).toBeChecked();
  });

  it("should parse the existing pass credentials config", async () => {
    const testRepo = {
      ...repo,
      auth: { passCredentials: true },
    } as PackageRepositoryDetail;
    let wrapper: any;
    await act(async () => {
      wrapper = mountWrapper(
        getStore({
          ...defaultState,
          repos: { ...defaultState.repos, repoDetail: testRepo } as IPackageRepositoryState,
        } as Partial<IStoreState>),
        <PkgRepoForm {...defaultProps} packageRepoRef={packageRepoRef} />,
      );
    });
    wrapper.update();
    expect(wrapper.find("#kubeapps-repo-pass-credentials")).toBeChecked();
  });

  describe("when there is a kubeapps-handled secret associated to the repo", () => {
    it("should parse the existing CA cert", async () => {
      const testRepo = new PackageRepositoryDetail({
        ...repo,
        tlsConfig: new PackageRepositoryTlsConfig({
          packageRepoTlsConfigOneOf: {
            case: "certAuthority",
            value: "fooCA",
          },
        }),
      });
      let wrapper: any;
      await act(async () => {
        wrapper = mountWrapper(
          getStore({
            ...defaultState,
            repos: { ...defaultState.repos, repoDetail: testRepo } as IPackageRepositoryState,
          } as Partial<IStoreState>),
          <PkgRepoForm {...defaultProps} packageRepoRef={packageRepoRef} />,
        );
      });
      wrapper.update();
      expect(wrapper.find("#kubeapps-repo-custom-ca").prop("value")).toBe("fooCA");
    });

    it("should parse the existing auth header", async () => {
      const testRepo = new PackageRepositoryDetail({
        ...repo,
        auth: {
          type: PackageRepositoryAuth_PackageRepositoryAuthType.AUTHORIZATION_HEADER,
          packageRepoAuthOneOf: {
            case: "header",
            value: "fooHeader",
          },
        },
      });
      let wrapper: any;
      await act(async () => {
        wrapper = mountWrapper(
          getStore({
            ...defaultState,
            repos: { ...defaultState.repos, repoDetail: testRepo } as IPackageRepositoryState,
          } as Partial<IStoreState>),
          <PkgRepoForm {...defaultProps} packageRepoRef={packageRepoRef} />,
        );
      });
      await waitFor(() => {
        wrapper.update();
        expect(wrapper.find("#kubeapps-repo-auth-method-custom")).toBeChecked();
      });
      expect(wrapper.find("#kubeapps-repo-custom-header").prop("value")).toBe("fooHeader");
    });

    it("should parse the existing basic auth", async () => {
      const testRepo = new PackageRepositoryDetail({
        ...repo,
        auth: {
          type: PackageRepositoryAuth_PackageRepositoryAuthType.BASIC_AUTH,
          packageRepoAuthOneOf: {
            case: "usernamePassword",
            value: new UsernamePassword({
              username: "foo",
              password: "bar",
            }),
          },
        },
      });
      let wrapper: any;
      await act(async () => {
        wrapper = mountWrapper(
          getStore({
            ...defaultState,
            repos: { ...defaultState.repos, repoDetail: testRepo } as IPackageRepositoryState,
          } as Partial<IStoreState>),
          <PkgRepoForm {...defaultProps} packageRepoRef={packageRepoRef} />,
        );
      });
      await waitFor(() => {
        wrapper.update();
        expect(wrapper.find("#kubeapps-repo-auth-method-basic")).toBeChecked();
      });
      expect(wrapper.find("#kubeapps-repo-username").prop("value")).toBe("foo");
      expect(wrapper.find("#kubeapps-repo-password").prop("value")).toBe("bar");
    });

    it("should parse a bearer token", async () => {
      const testRepo = new PackageRepositoryDetail({
        ...repo,
        auth: {
          type: PackageRepositoryAuth_PackageRepositoryAuthType.BEARER,
          packageRepoAuthOneOf: {
            case: "header",
            value: "Bearer fooToken",
          },
        },
      });
      let wrapper: any;
      await act(async () => {
        wrapper = mountWrapper(
          getStore({
            ...defaultState,
            repos: { ...defaultState.repos, repoDetail: testRepo } as IPackageRepositoryState,
          } as Partial<IStoreState>),
          <PkgRepoForm {...defaultProps} packageRepoRef={packageRepoRef} />,
        );
      });
      await waitFor(() => {
        wrapper.update();
        expect(wrapper.find("#kubeapps-repo-auth-method-bearer")).toBeChecked();
      });
      expect(wrapper.find("#kubeapps-repo-token").prop("value")).toBe("Bearer fooToken");
    });

    it("should select a docker secret as auth mechanism", async () => {
      const testRepo = new PackageRepositoryDetail({
        ...repo,
        auth: {
          type: PackageRepositoryAuth_PackageRepositoryAuthType.DOCKER_CONFIG_JSON,
          packageRepoAuthOneOf: {
            case: "dockerCreds",
            value: new DockerCredentials({
              email: "foo@foo.foo",
              password: "bar",
              server: "foobar",
              username: "foo",
            }),
          },
        },
      });
      let wrapper: any;
      await act(async () => {
        wrapper = mountWrapper(
          getStore({
            ...defaultState,
            repos: { ...defaultState.repos, repoDetail: testRepo } as IPackageRepositoryState,
          } as Partial<IStoreState>),
          <PkgRepoForm {...defaultProps} packageRepoRef={packageRepoRef} />,
        );
      });
      await waitFor(() => {
        wrapper.update();
        expect(wrapper.find("#kubeapps-repo-auth-method-registry")).toBeChecked();
      });
      expect(wrapper.find("#kubeapps-docker-cred-server").prop("value")).toBe("foobar");
      expect(wrapper.find("#kubeapps-docker-cred-username").prop("value")).toBe("foo");
      expect(wrapper.find("#kubeapps-docker-cred-password").prop("value")).toBe("bar");
      expect(wrapper.find("#kubeapps-docker-cred-email").prop("value")).toBe("foo@foo.foo");
    });

    it("should select a opaque as auth mechanism", async () => {
      const testRepo = new PackageRepositoryDetail({
        ...repo,
        auth: {
          type: PackageRepositoryAuth_PackageRepositoryAuthType.OPAQUE,
          packageRepoAuthOneOf: {
            case: "opaqueCreds",
            value: new OpaqueCredentials({
              data: {},
            }),
          },
        },
      });
      let wrapper: any;
      await act(async () => {
        wrapper = mountWrapper(
          getStore({
            ...defaultState,
            repos: { ...defaultState.repos, repoDetail: testRepo } as IPackageRepositoryState,
          } as Partial<IStoreState>),
          <PkgRepoForm {...defaultProps} packageRepoRef={packageRepoRef} />,
        );
      });
      await waitFor(() => {
        wrapper.update();
        expect(wrapper.find("#kubeapps-repo-auth-method-opaque")).toBeChecked();
      });
      expect(wrapper.find("#kubeapps-repo-opaque-data").prop("value")).toBe("{}");
    });

    it("should select a ssh as auth mechanism", async () => {
      const testRepo = new PackageRepositoryDetail({
        ...repo,
        auth: {
          type: PackageRepositoryAuth_PackageRepositoryAuthType.SSH,
          packageRepoAuthOneOf: {
            case: "sshCreds",
            value: new SshCredentials({
              knownHosts: "foo",
              privateKey: "bar",
            }),
          },
        },
      });
      let wrapper: any;
      await act(async () => {
        wrapper = mountWrapper(
          getStore({
            ...defaultState,
            repos: { ...defaultState.repos, repoDetail: testRepo } as IPackageRepositoryState,
          } as Partial<IStoreState>),
          <PkgRepoForm {...defaultProps} packageRepoRef={packageRepoRef} />,
        );
      });
      await waitFor(() => {
        wrapper.update();
        expect(wrapper.find("#kubeapps-repo-auth-method-ssh")).toBeChecked();
      });
      expect(wrapper.find("#kubeapps-repo-ssh-knownhosts").prop("value")).toBe("foo");
      expect(wrapper.find("#kubeapps-repo-ssh-privatekey").prop("value")).toBe("bar");
    });

    it("should select a tls as auth mechanism", async () => {
      const testRepo = new PackageRepositoryDetail({
        ...repo,
        auth: {
          type: PackageRepositoryAuth_PackageRepositoryAuthType.TLS,
          packageRepoAuthOneOf: {
            case: "tlsCertKey",
            value: new TlsCertKey({
              cert: "foo",
              key: "bar",
            }),
          },
        },
      });
      let wrapper: any;
      await act(async () => {
        wrapper = mountWrapper(
          getStore({
            ...defaultState,
            repos: { ...defaultState.repos, repoDetail: testRepo } as IPackageRepositoryState,
          } as Partial<IStoreState>),
          <PkgRepoForm {...defaultProps} packageRepoRef={packageRepoRef} />,
        );
      });
      await waitFor(() => {
        wrapper.update();
        expect(wrapper.find("#kubeapps-repo-auth-method-tls")).toBeChecked();
      });
      expect(wrapper.find("#kubeapps-repo-tls-cert").prop("value")).toBe("foo");
      expect(wrapper.find("#kubeapps-repo-tls-key").prop("value")).toBe("bar");
    });
  });
});

describe("auth provider selector for Flux repositories", () => {
  const packageRepoRef = {
    identifier: "flux-repo",
    context: defaultProps.packageRepoRef?.context,
    plugin: { name: PluginNames.PACKAGES_FLUX, version: "v1alpha1" },
  } as PackageRepositoryReference;
  const repo = {
    name: "",
    description: "",
    interval: "10m",
    packageRepoRef: packageRepoRef,
    namespaceScoped: false,
    type: "",
    url: "",
  } as PackageRepositoryDetail;

  it("repository auth provider should appear as a valid option for Flux and OCI", async () => {
    const testRepo = new PackageRepositoryDetail({
      ...repo,
      type: "oci",
      customDetail: new Any({
        value: new FluxPackageRepositoryCustomDetail({
          provider: "",
        }).toBinary(),
      }),
    });
    let wrapper: any;
    await act(async () => {
      wrapper = mountWrapper(
        getStore({
          ...defaultState,
          repos: { ...defaultState.repos, repoDetail: testRepo } as IPackageRepositoryState,
        } as Partial<IStoreState>),
        <PkgRepoForm {...defaultProps} packageRepoRef={packageRepoRef} />,
      );
    });
    await waitFor(() => {
      wrapper.update();
      expect(wrapper.find("#kubeapps-repo-auth-method-provider").first()).not.toBeDisabled();
    });
    expect(wrapper.find("#kubeapps-flux-auth-provider")).not.toExist();
  });

  it("auth provider dropdown should not appear for other repo types", async () => {
    const testRepo = {
      ...repo,
      type: "helm",
    } as PackageRepositoryDetail;
    let wrapper: any;
    await act(async () => {
      wrapper = mountWrapper(
        getStore({
          ...defaultState,
          repos: { ...defaultState.repos, repoDetail: testRepo } as IPackageRepositoryState,
        } as Partial<IStoreState>),
        <PkgRepoForm {...defaultProps} packageRepoRef={packageRepoRef} />,
      );
    });
    await waitFor(() => {
      wrapper.update();
      expect(wrapper.find("#kubeapps-repo-auth-method-provider")).toBeDisabled();
    });
  });

  it("auth provider dropdown should not appear for other plugins", async () => {
    const testRepo = {
      ...repo,
      type: "oci",
      packageRepoRef: {
        identifier: "helm-repo",
        context: defaultProps.packageRepoRef?.context,
        plugin: { name: PluginNames.PACKAGES_HELM, version: "v1alpha1" },
      } as PackageRepositoryReference,
    } as PackageRepositoryDetail;
    let wrapper: any;
    await act(async () => {
      wrapper = mountWrapper(
        getStore({
          ...defaultState,
          repos: { ...defaultState.repos, repoDetail: testRepo } as IPackageRepositoryState,
        } as Partial<IStoreState>),
        <PkgRepoForm {...defaultProps} packageRepoRef={packageRepoRef} />,
      );
    });
    await waitFor(() => {
      wrapper.update();
      expect(wrapper.find("#kubeapps-repo-auth-method-provider")).toBeDisabled();
    });
  });

  it("repository auth provider selected should show the subsequent options dropdown", async () => {
    const customDetail = new FluxPackageRepositoryCustomDetail({
      provider: "aws",
    });
    const testRepo = new PackageRepositoryDetail({
      ...repo,
      type: "oci",
      customDetail: {
        typeUrl: FluxPackageRepositoryCustomDetail.typeName,
        value: customDetail.toBinary(),
      } as Any,
    });
    let wrapper: any;
    await act(async () => {
      wrapper = mountWrapper(
        getStore({
          ...defaultState,
          repos: { ...defaultState.repos, repoDetail: testRepo } as IPackageRepositoryState,
        } as Partial<IStoreState>),
        <PkgRepoForm {...defaultProps} packageRepoRef={packageRepoRef} />,
      );
    });
    await waitFor(() => {
      wrapper.update();
      expect(wrapper.find("#kubeapps-repo-auth-method-provider")).not.toBeDisabled();
    });
    expect(wrapper.find("#kubeapps-flux-auth-provider")).toExist();
    expect(wrapper.find("#kubeapps-flux-auth-provider")).not.toBeDisabled();
    expect(wrapper.find("#kubeapps-flux-auth-provider select").props().value).toEqual("aws");
  });
});
