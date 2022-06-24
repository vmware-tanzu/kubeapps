// Copyright 2020-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { CdsButton } from "@cds/react/button";
import { waitFor } from "@testing-library/react";
import actions from "actions";
import Alert from "components/js/Alert";
import {
  PackageRepositoryAuth_PackageRepositoryAuthType,
  PackageRepositoryDetail,
  PackageRepositoryReference,
} from "gen/kubeappsapis/core/packages/v1alpha1/repositories";
import { Plugin } from "gen/kubeappsapis/core/plugins/v1alpha1/plugins";
import { RepositoryCustomDetails } from "gen/kubeappsapis/plugins/helm/packages/v1alpha1/helm";
import { act } from "react-dom/test-utils";
import * as ReactRedux from "react-redux";
import { IPackageRepositoryState } from "reducers/repos";
import { defaultStore, getStore, mountWrapper } from "shared/specs/mountWrapper";
import { IPkgRepoFormData, IStoreState } from "shared/types";
import { PluginNames } from "shared/utils";
import { AppRepoForm, RepositoryStorageTypes } from "./AppRepoForm";

const defaultProps = {
  onSubmit: jest.fn(),
  namespace: "default",
  cluster: "default",
  kubeappsNamespace: "kubeapps",
};

const defaultState = {
  repos: { repo: {} } as Partial<IPackageRepositoryState>,
} as IStoreState;

const pkgRepoFormData = {
  plugin: { name: PluginNames.PACKAGES_HELM, version: "v1alpha1" } as Plugin,
  authHeader: "",
  authMethod:
    PackageRepositoryAuth_PackageRepositoryAuthType.PACKAGE_REPOSITORY_AUTH_TYPE_UNSPECIFIED,
  basicAuth: {
    password: "",
    username: "",
  },
  customCA: "",
  customDetails: {
    dockerRegistrySecrets: [],
    ociRepositories: [],
    performValidation: true,
    filterRules: [],
  },
  description: "",
  dockerRegCreds: {
    password: "",
    username: "",
    email: "",
    server: "",
  },
  interval: "",
  name: "",
  passCredentials: false,
  secretAuthName: "",
  secretTLSName: "",
  skipTLS: false,
  type: RepositoryStorageTypes.PACKAGE_REPOSITORY_STORAGE_HELM,
  url: "",
  opaqueCreds: {
    data: {},
  },
  sshCreds: {
    knownHosts: "",
    privateKey: "",
  },
  tlsCertKey: {
    cert: "",
    key: "",
  },
} as IPkgRepoFormData;

let spyOnUseDispatch: jest.SpyInstance;
const kubeaActions = { ...actions.kube };
beforeEach(() => {
  actions.repos = {
    ...actions.repos,
  };
  const mockDispatch = jest.fn(r => r);
  spyOnUseDispatch = jest.spyOn(ReactRedux, "useDispatch").mockReturnValue(mockDispatch);
});

afterEach(() => {
  actions.kube = { ...kubeaActions };
  spyOnUseDispatch.mockRestore();
});

it("disables the submit button while fetching", async () => {
  let wrapper: any;
  await act(async () => {
    wrapper = mountWrapper(
      getStore({ repos: { validating: true } }),
      <AppRepoForm {...defaultProps} />,
    );
  });
  expect(
    wrapper
      .find(CdsButton)
      .filterWhere((b: any) => b.html().includes("Validating"))
      .prop("disabled"),
  ).toBe(true);
});

it("submit button can not be fired more than once", async () => {
  const onSubmit = jest.fn().mockReturnValue(true);
  const onAfterInstall = jest.fn().mockReturnValue(true);
  let wrapper: any;
  await act(async () => {
    wrapper = mountWrapper(
      defaultStore,
      <AppRepoForm {...defaultProps} onSubmit={onSubmit} onAfterInstall={onAfterInstall} />,
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

it("should show a validation error", async () => {
  let wrapper: any;
  await act(async () => {
    wrapper = mountWrapper(
      getStore({ repos: { errors: { validate: new Error("Boom!") } } }),
      <AppRepoForm {...defaultProps} />,
    );
  });
  expect(wrapper.find(Alert).text()).toContain("Boom!");
});

it("shows an error updating a repo", async () => {
  let wrapper: any;
  await act(async () => {
    wrapper = mountWrapper(
      getStore({ repos: { errors: { update: new Error("boom!") } } }),
      <AppRepoForm {...defaultProps} />,
    );
  });
  expect(wrapper.find(Alert)).toIncludeText("boom!");
});

it("should call the install method when the validation success", async () => {
  const install = jest.fn().mockReturnValue(true);
  actions.repos = {
    ...actions.repos,
  };

  let wrapper: any;
  await act(async () => {
    wrapper = mountWrapper(defaultStore, <AppRepoForm {...defaultProps} onSubmit={install} />);
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
    wrapper = mountWrapper(defaultStore, <AppRepoForm {...defaultProps} onSubmit={install} />);
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
    customDetails: {
      ociRepositories: ["apache", "jenkins"],
      dockerRegistrySecrets: [""],
      filterRule: undefined,
      performValidation: true,
    },
    interval: undefined,
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
    wrapper = mountWrapper(defaultStore, <AppRepoForm {...defaultProps} onSubmit={install} />);
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
    customDetails: {
      ociRepositories: [],
      dockerRegistrySecrets: [""],
      filterRule: undefined,
      performValidation: true,
    },
    skipTLS: true,
    interval: undefined,
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
    wrapper = mountWrapper(defaultStore, <AppRepoForm {...defaultProps} onSubmit={install} />);
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
    customDetails: {
      ociRepositories: [],
      dockerRegistrySecrets: [""],
      filterRule: undefined,
      performValidation: true,
    },
    passCredentials: true,
    interval: undefined,
    description: undefined,
  } as unknown as IPkgRepoFormData);
});

describe("when using a filter", () => {
  it("should call the install method with a filter", async () => {
    const install = jest.fn().mockReturnValue(true);
    let wrapper: any;
    await act(async () => {
      wrapper = mountWrapper(defaultStore, <AppRepoForm {...defaultProps} onSubmit={install} />);
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
      customDetails: {
        ociRepositories: [],
        dockerRegistrySecrets: [""],
        filterRule: {
          jq: ".name == $var0 or .name == $var1",
          variables: { $var0: "nginx", $var1: "wordpress" },
        },
        performValidation: true,
      },
      interval: undefined,
      description: undefined,
    } as unknown as IPkgRepoFormData);
  });

  it("should call the install method with a filter excluding a regex", async () => {
    const install = jest.fn().mockReturnValue(true);
    let wrapper: any;
    await act(async () => {
      wrapper = mountWrapper(defaultStore, <AppRepoForm {...defaultProps} onSubmit={install} />);
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
      customDetails: {
        ociRepositories: [],
        dockerRegistrySecrets: [""],
        filterRule: {
          jq: ".name | test($var) | not",
          variables: { $var: "nginx" },
        },
        performValidation: true,
      },
      interval: undefined,
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
      wrapper = mountWrapper(defaultStore, <AppRepoForm {...defaultProps} onSubmit={install} />);
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
      customDetails: {
        ociRepositories: ["apache", "jenkins"],
        dockerRegistrySecrets: [""],
        filterRule: undefined,
        performValidation: true,
      },
      interval: undefined,
      description: undefined,
    } as unknown as IPkgRepoFormData);
  });
});

it("should call the install method with a description", async () => {
  const install = jest.fn().mockReturnValue(true);
  let wrapper: any;
  await act(async () => {
    wrapper = mountWrapper(defaultStore, <AppRepoForm {...defaultProps} onSubmit={install} />);
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
    customDetails: {
      ociRepositories: [],
      dockerRegistrySecrets: [""],
      filterRule: undefined,
      performValidation: true,
    },
    interval: undefined,
    description: "description test",
  } as unknown as IPkgRepoFormData);
});

it("should not show the list of OCI repositories if using a Helm repo (default)", async () => {
  let wrapper: any;
  await act(async () => {
    wrapper = mountWrapper(defaultStore, <AppRepoForm {...defaultProps} />);
  });
  wrapper.find("#kubeapps-plugin-helm").simulate("change");
  wrapper.find("#kubeapps-repo-name").simulate("change", { target: { value: "helm-repo" } });
  wrapper.find("#kubeapps-repo-url").simulate("change", { target: { value: "helm.repo" } });
  wrapper.find("#kubeapps-repo-type-helm").simulate("change");
  expect(wrapper.find("#kubeapps-oci-repositories")).not.toExist();
});

describe("when the repository info is already populated", () => {
  const packageRepoRef = {
    identifier: "helm-repo",
    context: { cluster: defaultProps.cluster, namespace: defaultProps.namespace },
    plugin: { name: PluginNames.PACKAGES_HELM, version: "v1alpha1" },
  } as PackageRepositoryReference;
  const repo = {
    name: "",
    description: "",
    interval: "",
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
        getStore({ ...defaultState, repos: { ...defaultState.repos, repo: testRepo } }),
        <AppRepoForm {...defaultProps} packageRepoRef={packageRepoRef} />,
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
        getStore({ ...defaultState, repos: { ...defaultState.repos, repo: testRepo } }),
        <AppRepoForm {...defaultProps} packageRepoRef={packageRepoRef} />,
      );
    });
    await waitFor(() => {
      wrapper.update();
      expect(wrapper.find("#kubeapps-repo-url").prop("value")).toBe("http://repo");
    });
  });

  it("should parse the existing filter (simple)", async () => {
    const testRepo = {
      ...repo,
      type: "helm",
      customDetail: {
        filterRule: {
          jq: ".name == $var0 or .name == $var1",
          variables: { $var0: "nginx", $var1: "wordpress" },
        },
      } as Partial<RepositoryCustomDetails>,
    } as PackageRepositoryDetail;

    let wrapper: any;
    await act(async () => {
      wrapper = mountWrapper(
        getStore({ ...defaultState, repos: { ...defaultState.repos, repo: testRepo } }),
        <AppRepoForm {...defaultProps} packageRepoRef={packageRepoRef} />,
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
    const testRepo = {
      ...repo,
      type: "helm",
      customDetail: {
        filterRule: { jq: ".name | test($var) | not", variables: { $var: "nginx" } },
      } as Partial<RepositoryCustomDetails>,
    } as PackageRepositoryDetail;

    let wrapper: any;
    await act(async () => {
      wrapper = mountWrapper(
        getStore({ ...defaultState, repos: { ...defaultState.repos, repo: testRepo } }),
        <AppRepoForm {...defaultProps} packageRepoRef={packageRepoRef} />,
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
        getStore({ ...defaultState, repos: { ...defaultState.repos, repo: testRepo } }),
        <AppRepoForm {...defaultProps} packageRepoRef={packageRepoRef} />,
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
        getStore({ ...defaultState, repos: { ...defaultState.repos, repo: testRepo } }),
        <AppRepoForm {...defaultProps} packageRepoRef={packageRepoRef} />,
      );
    });
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
        getStore({ ...defaultState, repos: { ...defaultState.repos, repo: testRepo } }),
        <AppRepoForm {...defaultProps} packageRepoRef={packageRepoRef} />,
      );
    });
    expect(wrapper.find("#kubeapps-repo-pass-credentials")).toBeChecked();
  });

  describe("when there is a kubeapps-handled secret associated to the repo", () => {
    it("should parse the existing CA cert", async () => {
      const testRepo = {
        ...repo,
        tlsConfig: { certAuthority: "fooCA" },
      } as PackageRepositoryDetail;
      let wrapper: any;
      await act(async () => {
        wrapper = mountWrapper(
          getStore({ ...defaultState, repos: { ...defaultState.repos, repo: testRepo } }),
          <AppRepoForm {...defaultProps} packageRepoRef={packageRepoRef} />,
        );
      });
      wrapper.update();
      expect(wrapper.find("#kubeapps-repo-custom-ca").prop("value")).toBe("fooCA");
    });

    it("should parse the existing auth header", async () => {
      const testRepo = {
        ...repo,
        auth: {
          type: PackageRepositoryAuth_PackageRepositoryAuthType.PACKAGE_REPOSITORY_AUTH_TYPE_AUTHORIZATION_HEADER,
          header: "fooHeader",
        },
      } as PackageRepositoryDetail;
      let wrapper: any;
      await act(async () => {
        wrapper = mountWrapper(
          getStore({ ...defaultState, repos: { ...defaultState.repos, repo: testRepo } }),
          <AppRepoForm {...defaultProps} packageRepoRef={packageRepoRef} />,
        );
      });
      await waitFor(() => {
        wrapper.update();
        expect(wrapper.find("#kubeapps-repo-auth-method-custom")).toBeChecked();
      });
      expect(wrapper.find("#kubeapps-repo-custom-header").prop("value")).toBe("fooHeader");
    });

    it("should parse the existing basic auth", async () => {
      const testRepo = {
        ...repo,
        auth: {
          type: PackageRepositoryAuth_PackageRepositoryAuthType.PACKAGE_REPOSITORY_AUTH_TYPE_BASIC_AUTH,
          usernamePassword: { username: "foo", password: "bar" },
        },
      } as PackageRepositoryDetail;
      let wrapper: any;
      await act(async () => {
        wrapper = mountWrapper(
          getStore({ ...defaultState, repos: { ...defaultState.repos, repo: testRepo } }),
          <AppRepoForm {...defaultProps} packageRepoRef={packageRepoRef} />,
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
      const testRepo = {
        ...repo,
        auth: {
          type: PackageRepositoryAuth_PackageRepositoryAuthType.PACKAGE_REPOSITORY_AUTH_TYPE_BEARER,
          header: "Bearer fooToken",
        },
      } as PackageRepositoryDetail;
      let wrapper: any;
      await act(async () => {
        wrapper = mountWrapper(
          getStore({ ...defaultState, repos: { ...defaultState.repos, repo: testRepo } }),
          <AppRepoForm {...defaultProps} packageRepoRef={packageRepoRef} />,
        );
      });
      await waitFor(() => {
        wrapper.update();
        expect(wrapper.find("#kubeapps-repo-auth-method-bearer")).toBeChecked();
      });
      expect(wrapper.find("#kubeapps-repo-token").prop("value")).toBe("Bearer fooToken");
    });

    it("should select a docker secret as auth mechanism", async () => {
      const testRepo = {
        ...repo,
        auth: {
          type: PackageRepositoryAuth_PackageRepositoryAuthType.PACKAGE_REPOSITORY_AUTH_TYPE_DOCKER_CONFIG_JSON,
          dockerCreds: {
            email: "foo@foo.foo",
            password: "bar",
            server: "foobar",
            username: "foo",
          },
        },
      } as PackageRepositoryDetail;
      let wrapper: any;
      await act(async () => {
        wrapper = mountWrapper(
          getStore({ ...defaultState, repos: { ...defaultState.repos, repo: testRepo } }),
          <AppRepoForm {...defaultProps} packageRepoRef={packageRepoRef} />,
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
      const testRepo = {
        ...repo,
        auth: {
          type: PackageRepositoryAuth_PackageRepositoryAuthType.PACKAGE_REPOSITORY_AUTH_TYPE_OPAQUE,
          opaqueCreds: {
            data: {},
          },
        },
      } as PackageRepositoryDetail;
      let wrapper: any;
      await act(async () => {
        wrapper = mountWrapper(
          getStore({ ...defaultState, repos: { ...defaultState.repos, repo: testRepo } }),
          <AppRepoForm {...defaultProps} packageRepoRef={packageRepoRef} />,
        );
      });
      await waitFor(() => {
        wrapper.update();
        expect(wrapper.find("#kubeapps-repo-auth-method-opaque")).toBeChecked();
      });
      expect(wrapper.find("#kubeapps-repo-opaque-data").prop("value")).toBe("{}");
    });

    it("should select a ssh as auth mechanism", async () => {
      const testRepo = {
        ...repo,
        auth: {
          type: PackageRepositoryAuth_PackageRepositoryAuthType.PACKAGE_REPOSITORY_AUTH_TYPE_SSH,
          sshCreds: {
            knownHosts: "foo",
            privateKey: "bar",
          },
        },
      } as PackageRepositoryDetail;
      let wrapper: any;
      await act(async () => {
        wrapper = mountWrapper(
          getStore({ ...defaultState, repos: { ...defaultState.repos, repo: testRepo } }),
          <AppRepoForm {...defaultProps} packageRepoRef={packageRepoRef} />,
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
      const testRepo = {
        ...repo,
        auth: {
          type: PackageRepositoryAuth_PackageRepositoryAuthType.PACKAGE_REPOSITORY_AUTH_TYPE_TLS,
          tlsCertKey: {
            cert: "foo",
            key: "bar",
          },
        },
      } as PackageRepositoryDetail;
      let wrapper: any;
      await act(async () => {
        wrapper = mountWrapper(
          getStore({ ...defaultState, repos: { ...defaultState.repos, repo: testRepo } }),
          <AppRepoForm {...defaultProps} packageRepoRef={packageRepoRef} />,
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
