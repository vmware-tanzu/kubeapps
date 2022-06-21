// Copyright 2020-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { CdsButton } from "@cds/react/button";
import actions from "actions";
import Alert from "components/js/Alert";
import { act } from "react-dom/test-utils";
import * as ReactRedux from "react-redux";
import { defaultStore, getStore, mountWrapper } from "shared/specs/mountWrapper";
import { IPkgRepoFormData } from "shared/types";
import { PackageRepositoryAuth_PackageRepositoryAuthType } from "gen/kubeappsapis/core/packages/v1alpha1/repositories";
import Secret from "shared/Secret";
import { AppRepoForm } from "./AppRepoForm";

const defaultProps = {
  onSubmit: jest.fn(),
  namespace: "default",
  kubeappsNamespace: "kubeapps",
};

const repoData = {
  plugin: undefined,
  name: undefined,
  type: undefined,
  url: undefined,
  authHeader: "",
  authMethod:
    PackageRepositoryAuth_PackageRepositoryAuthType.PACKAGE_REPOSITORY_AUTH_TYPE_UNSPECIFIED,
  basicAuth: {
    password: "",
    username: "",
  },
  customCA: "",
  customDetails: {
    dockerRegistrySecrets: [""],
    ociRepositories: [],
    performValidation: undefined,
    filterRules: undefined,
  },
  description: "",
  dockerRegCreds: {
    password: "",
    username: "",
    email: "",
    server: "",
  },
  interval: "10m",
  passCredentials: false,
  secretAuthName: "",
  secretTLSName: "",
  skipTLS: false,
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
} as unknown as IPkgRepoFormData;

let spyOnUseDispatch: jest.SpyInstance;
const kubeaActions = { ...actions.kube };
beforeEach(() => {
  actions.repos = {
    ...actions.repos,
    validateRepo: jest.fn().mockReturnValue(true),
  };
  const mockDispatch = jest.fn(r => r);
  spyOnUseDispatch = jest.spyOn(ReactRedux, "useDispatch").mockReturnValue(mockDispatch);
  Secret.getDockerConfigSecretNames = jest.fn(() => Promise.resolve([]));
});

afterEach(() => {
  actions.kube = { ...kubeaActions };
  spyOnUseDispatch.mockRestore();
});

// TODO(agamez): re-enable this test once we we add a dropdown to select the secret name
// eslint-disable-next-line jest/no-commented-out-tests
// it("fetches secret names", async () => {
//   await act(async () => {
//     mountWrapper(defaultStore, <AppRepoForm {...defaultProps} />);
//   });
//   expect(Secret.getDockerConfigSecretNames).toHaveBeenCalledWith(
//     "default-cluster",
//     defaultProps.namespace,
//   );
// });

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
  const validateRepo = jest.fn().mockReturnValue(true);
  const install = jest.fn().mockReturnValue(true);
  actions.repos = {
    ...actions.repos,
    validateRepo,
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
  const validateRepo = jest.fn().mockReturnValue(true);
  const install = jest.fn().mockReturnValue(true);
  actions.repos = {
    ...actions.repos,
    validateRepo,
  };
  let wrapper: any;
  await act(async () => {
    wrapper = mountWrapper(defaultStore, <AppRepoForm {...defaultProps} onSubmit={install} />);
  });
  wrapper.find("#kubeapps-repo-name").simulate("change", { target: { value: "my-oci-repo" } });
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
    ...repoData,
    customDetails: {
      ...repoData.customDetails,
      performValidation: true,
      ociRepositories: ["apache", "jenkins"],
    },
    name: "my-oci-repo",
    type: "oci",
    url: "https://oci.repo",
    plugin: { name: "helm.packages", version: "v1alpha1" },
  });
});

it("should call the install skipping TLS verification", async () => {
  const validateRepo = jest.fn().mockReturnValue(true);
  const install = jest.fn().mockReturnValue(true);
  actions.repos = {
    ...actions.repos,
    validateRepo,
  };
  let wrapper: any;
  await act(async () => {
    wrapper = mountWrapper(defaultStore, <AppRepoForm {...defaultProps} onSubmit={install} />);
  });
  wrapper.find("#kubeapps-repo-name").simulate("change", { target: { value: "my-repo" } });
  wrapper.find("#kubeapps-repo-url").simulate("change", { target: { value: "helm.repo" } });
  wrapper.find("#kubeapps-repo-type-helm").simulate("change");
  wrapper.find("#kubeapps-repo-skip-tls").simulate("change");
  const form = wrapper.find("form");
  await act(async () => {
    await (form.prop("onSubmit") as (e: any) => Promise<any>)({ preventDefault: jest.fn() });
  });
  wrapper.update();
  expect(install).toHaveBeenCalledWith({
    ...repoData,
    customDetails: {
      ...repoData.customDetails,
      performValidation: true,
    },
    name: "my-repo",
    type: "helm",
    url: "https://helm.repo",
    plugin: { name: "helm.packages", version: "v1alpha1" },
    skipTLS: true,
  });
});

it("should call the install passing credentials", async () => {
  const validateRepo = jest.fn().mockReturnValue(true);
  const install = jest.fn().mockReturnValue(true);
  actions.repos = {
    ...actions.repos,
    validateRepo,
  };
  let wrapper: any;
  await act(async () => {
    wrapper = mountWrapper(defaultStore, <AppRepoForm {...defaultProps} onSubmit={install} />);
  });
  wrapper.find("#kubeapps-repo-name").simulate("change", { target: { value: "my-repo" } });
  wrapper.find("#kubeapps-repo-url").simulate("change", { target: { value: "helm.repo" } });
  wrapper.find("#kubeapps-repo-type-helm").simulate("change");
  wrapper.find("#kubeapps-repo-pass-credentials").simulate("change");
  const form = wrapper.find("form");
  await act(async () => {
    await (form.prop("onSubmit") as (e: any) => Promise<any>)({ preventDefault: jest.fn() });
  });
  wrapper.update();
  expect(install).toHaveBeenCalledWith({
    ...repoData,
    customDetails: {
      ...repoData.customDetails,
      performValidation: true,
    },
    name: "my-repo",
    type: "helm",
    url: "https://helm.repo",
    plugin: { name: "helm.packages", version: "v1alpha1" },
    passCredentials: true,
  });
});

describe("when using a filter", () => {
  it("should call the install method with a filter", async () => {
    const install = jest.fn().mockReturnValue(true);
    let wrapper: any;
    await act(async () => {
      wrapper = mountWrapper(defaultStore, <AppRepoForm {...defaultProps} onSubmit={install} />);
    });
    wrapper.find("#kubeapps-repo-name").simulate("change", { target: { value: "my-repo" } });
    wrapper
      .find("#kubeapps-repo-url")
      .simulate("change", { target: { value: "https://helm.repo" } });
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
      ...repoData,
      customDetails: {
        ...repoData.customDetails,
        performValidation: true,
        filterRule: {
          jq: ".name == $var0 or .name == $var1",
          variables: { $var0: "nginx", $var1: "wordpress" },
        },
      },
      name: "my-repo",
      type: "helm",
      url: "https://helm.repo",
      plugin: { name: "helm.packages", version: "v1alpha1" },
    });
  });

  it("should call the install method with a filter excluding a regex", async () => {
    const install = jest.fn().mockReturnValue(true);
    let wrapper: any;
    await act(async () => {
      wrapper = mountWrapper(defaultStore, <AppRepoForm {...defaultProps} onSubmit={install} />);
    });
    wrapper.find("#kubeapps-repo-name").simulate("change", { target: { value: "my-repo" } });
    wrapper
      .find("#kubeapps-repo-url")
      .simulate("change", { target: { value: "https://helm.repo" } });
    wrapper.find("#kubeapps-repo-type-helm").simulate("change");
    wrapper
      .find("#kubeapps-repo-filter-repositories")
      .simulate("change", { target: { value: "nginx" } });
    wrapper.find('input[type="checkbox"]').at(0).simulate("change");
    wrapper.find('input[type="checkbox"]').at(1).simulate("change");
    const form = wrapper.find("form");
    await act(async () => {
      await (form.prop("onSubmit") as (e: any) => Promise<any>)({ preventDefault: jest.fn() });
    });
    wrapper.update();
    expect(install).toHaveBeenCalledWith({
      ...repoData,
      customDetails: {
        ...repoData.customDetails,
        performValidation: true,
        filterRule: { jq: ".name | test($var) | not", variables: { $var: "nginx" } },
      },
      name: "my-repo",
      type: "helm",
      url: "https://helm.repo",
      plugin: { name: "helm.packages", version: "v1alpha1" },
    });
  });

  it("ignore the filter for the OCI case", async () => {
    const install = jest.fn().mockReturnValue(true);
    let wrapper: any;
    await act(async () => {
      wrapper = mountWrapper(defaultStore, <AppRepoForm {...defaultProps} onSubmit={install} />);
    });
    wrapper.find("#kubeapps-repo-name").simulate("change", { target: { value: "my-repo" } });
    wrapper
      .find("#kubeapps-repo-url")
      .simulate("change", { target: { value: "https://oci.repo" } });
    wrapper
      .find("#kubeapps-repo-filter-repositories")
      .simulate("change", { target: { value: "nginx, wordpress" } });
    wrapper.find("#kubeapps-repo-type-oci").simulate("change");
    const form = wrapper.find("form");
    await act(async () => {
      await (form.prop("onSubmit") as (e: any) => Promise<any>)({ preventDefault: jest.fn() });
    });
    wrapper.update();
    expect(install).toHaveBeenCalledWith({
      ...repoData,
      customDetails: {
        ...repoData.customDetails,
        performValidation: true,
      },
      name: "my-repo",
      type: "oci",
      url: "https://oci.repo",
      plugin: { name: "helm.packages", version: "v1alpha1" },
    });
  });
});

describe("when using a description", () => {
  it("should call the install method with a description", async () => {
    const install = jest.fn().mockReturnValue(true);
    let wrapper: any;
    await act(async () => {
      wrapper = mountWrapper(defaultStore, <AppRepoForm {...defaultProps} onSubmit={install} />);
    });
    wrapper.find("#kubeapps-repo-name").simulate("change", { target: { value: "my-repo" } });
    wrapper.find("#kubeapps-repo-type-helm").simulate("change");
    wrapper
      .find("#kubeapps-repo-url")
      .simulate("change", { target: { value: "https://helm.repo" } });
    wrapper
      .find("#kubeapps-repo-description")
      .simulate("change", { target: { value: "description test" } });
    const form = wrapper.find("form");
    await act(async () => {
      await (form.prop("onSubmit") as (e: any) => Promise<any>)({ preventDefault: jest.fn() });
    });
    wrapper.update();
    expect(install).toHaveBeenCalledWith({
      ...repoData,
      customDetails: {
        ...repoData.customDetails,
        performValidation: true,
      },
      name: "my-repo",
      type: "helm",
      url: "https://helm.repo",
      description: "description test",
      plugin: { name: "helm.packages", version: "v1alpha1" },
    });
  });
});

it("should not show the list of OCI repositories if using a Helm repo (default)", async () => {
  let wrapper: any;
  await act(async () => {
    wrapper = mountWrapper(defaultStore, <AppRepoForm {...defaultProps} />);
  });
  wrapper.find("#kubeapps-repo-type-helm").simulate("change");
  expect(wrapper.find("#kubeapps-oci-repositories")).not.toExist();
});

// TODO(agamez): fix those tests in the upcoming PR
// describe("when the repository info is already populated", () => {
//   it("should parse the existing name", async () => {
//     const repo = { metadata: { name: "foo" } } as any;
//     let wrapper: any;
//     await act(async () => {
//       wrapper = await mountWrapper(
//         defaultStore,
//         <AppRepoForm {...defaultProps} packageRepoRef={repo} />,
//       );
//     });
//     await waitFor(() => {
//       wrapper.update();
//       expect(wrapper.find("#kubeapps-repo-name").prop("value")).toBe("foo");
//     });
//     // It should also deactivate the name input if it's already been set
//     expect(wrapper.find("#kubeapps-repo-name").prop("disabled")).toBe(true);
//   });

//   it("should parse the existing url", async () => {
//     const repo = { metadata: { name: "foo" }, spec: { url: "http://repo" } } as any;
//     let wrapper: any;
//     await act(async () => {
//       wrapper = mountWrapper(defaultStore, <AppRepoForm {...defaultProps} packageRepoRef={repo} />);
//     });
//     await waitFor(() => {
//       wrapper.update();
//       expect(wrapper.find("#kubeapps-repo-url").prop("value")).toBe("http://repo");
//     });
//   });

//   describe("when there is a secret associated to the repo", () => {
//     it("should parse the existing CA cert", async () => {
//       const repo = {
//         metadata: { name: "foo", namespace: "default" },
//         spec: { auth: { customCA: { secretKeyRef: { name: "bar" } } } },
//       } as any;
//       const secret = { data: { "ca.crt": "Zm9v" } } as any;
//       AppRepository.getSecretForRepo = jest.fn(() => secret);

//       let wrapper: any;
//       act(() => {
//         wrapper = mountWrapper(
//           defaultStore,
//           <AppRepoForm {...defaultProps} packageRepoRef={repo} />,
//         );
//       });

//       await waitFor(() => {
//         expect(AppRepository.getSecretForRepo).toHaveBeenCalledWith(
//           "default-cluster",
//           "default",
//           "foo",
//         );
//       });
//       wrapper.update();
//       expect(wrapper.find("#kubeapps-repo-custom-ca").prop("value")).toBe("foo");
//     });

//     it("should parse the existing auth header", async () => {
//       const repo = {
//         metadata: { name: "foo", namespace: "default" },
//         spec: { auth: { header: { secretKeyRef: { name: "bar" } } } },
//       } as any;
//       const secret = { data: { authorizationHeader: "Zm9v" } } as any;
//       AppRepository.getSecretForRepo = jest.fn(() => secret);

//       let wrapper: any;
//       act(() => {
//         wrapper = mountWrapper(
//           defaultStore,
//           <AppRepoForm {...defaultProps} packageRepoRef={repo} />,
//         );
//       });

//       await waitFor(() => {
//         expect(AppRepository.getSecretForRepo).toHaveBeenCalledWith(
//           "default-cluster",
//           "default",
//           "foo",
//         );
//       });
//       wrapper.update();
//       expect(wrapper.find("#kubeapps-repo-custom-header").prop("value")).toBe("foo");
//     });

//     it("should parse the existing basic auth", async () => {
//       const repo = {
//         metadata: { name: "foo", namespace: "default" },
//         spec: { auth: { header: { secretKeyRef: { name: "bar" } } } },
//       } as any;
//       const secret = { data: { authorizationHeader: "QmFzaWMgWm05dk9tSmhjZz09" } } as any;
//       AppRepository.getSecretForRepo = jest.fn(() => secret);

//       let wrapper: any;
//       act(() => {
//         wrapper = mountWrapper(
//           defaultStore,
//           <AppRepoForm {...defaultProps} packageRepoRef={repo} />,
//         );
//       });

//       await waitFor(() => {
//         wrapper.update();
//         expect(wrapper.find("#kubeapps-repo-username").prop("value")).toBe("foo");
//       });
//       expect(wrapper.find("#kubeapps-repo-password").prop("value")).toBe("bar");
//     });

//     it("should parse the existing type", async () => {
//       const repo = { metadata: { name: "foo" }, spec: { type: "oci" } } as any;
//       let wrapper: any;
//       await act(async () => {
//         wrapper = mountWrapper(
//           defaultStore,
//           <AppRepoForm {...defaultProps} packageRepoRef={repo} />,
//         );
//       });
//       await waitFor(() => {
//         wrapper.update();
//         expect(wrapper.find("#kubeapps-repo-type-oci")).toBeChecked();
//       });
//       expect(wrapper.find("#kubeapps-oci-repositories")).toExist();
//     });

//     it("should parse the existing skip tls config", async () => {
//       const repo = { metadata: { name: "foo" }, spec: { tlsInsecureSkipVerify: true } } as any;
//       let wrapper: any;
//       await act(async () => {
//         wrapper = mountWrapper(
//           defaultStore,
//           <AppRepoForm {...defaultProps} packageRepoRef={repo} />,
//         );
//       });
//       expect(wrapper.find("#kubeapps-repo-skip-tls")).toBeChecked();
//     });

//     it("should parse the existing pass credentials config", async () => {
//       const repo = { metadata: { name: "foo" }, spec: { passCredentials: true } } as any;
//       let wrapper: any;
//       await act(async () => {
//         wrapper = mountWrapper(
//           defaultStore,
//           <AppRepoForm {...defaultProps} packageRepoRef={repo} />,
//         );
//       });
//       expect(wrapper.find("#kubeapps-repo-pass-credentials")).toBeChecked();
//     });

//     it("should parse a bearer token", async () => {
//       const repo = {
//         metadata: { name: "foo", namespace: "default" },
//         spec: { auth: { header: { secretKeyRef: { name: "bar" } } } },
//       } as any;
//       const secret = { data: { authorizationHeader: "QmVhcmVyIGZvbw==" } } as any;
//       AppRepository.getSecretForRepo = jest.fn(() => secret);

//       let wrapper: any;
//       act(() => {
//         wrapper = mountWrapper(
//           defaultStore,
//           <AppRepoForm {...defaultProps} packageRepoRef={repo} />,
//         );
//       });

//       await waitFor(() => {
//         wrapper.update();
//         expect(wrapper.find("#kubeapps-repo-token").prop("value")).toBe("foo");
//       });
//     });

//     it("should select a docker secret as auth mechanism", async () => {
//       const repo = {
//         metadata: { name: "foo", namespace: "default" },
//         spec: { auth: { header: { secretKeyRef: { name: "bar" } } } },
//       } as any;
//       const secret = { data: { ".dockerconfigjson": "QmVhcmVyIGZvbw==" } } as any;
//       AppRepository.getSecretForRepo = jest.fn(() => secret);

//       let wrapper: any;
//       act(() => {
//         wrapper = mountWrapper(
//           defaultStore,
//           <AppRepoForm {...defaultProps} packageRepoRef={repo} />,
//         );
//       });

//       await waitFor(() => {
//         wrapper.update();
//         expect(wrapper.find("#kubeapps-repo-auth-method-registry")).toBeChecked();
//       });
//     });

//     it("should parse the existing filter (simple)", async () => {
//       const repo = {
//         metadata: { name: "foo" },
//         spec: {
//           type: "helm",
//           filterRule: {
//             jq: ".name == $var0 or .name == $var1",
//             variables: { $var0: "nginx", $var1: "wordpress" },
//           },
//         },
//       } as any;
//       let wrapper: any;
//       await act(async () => {
//         wrapper = mountWrapper(
//           defaultStore,
//           <AppRepoForm {...defaultProps} packageRepoRef={repo} />,
//         );
//       });
//       await waitFor(() => {
//         wrapper.update();
//         expect(wrapper.find("textarea").at(0).prop("value")).toBe("nginx, wordpress");
//       });
//       expect(wrapper.find('input[type="checkbox"]').at(0)).not.toBeChecked();
//       expect(wrapper.find('input[type="checkbox"]').at(1)).not.toBeChecked();
//     });

//     it("should parse the existing filter (negated regex)", async () => {
//       const repo = {
//         metadata: { name: "foo" },
//         spec: {
//           type: "helm",
//           filterRule: { jq: ".name | test($var) | not", variables: { $var: "nginx" } },
//         },
//       } as any;
//       let wrapper: any;
//       await act(async () => {
//         wrapper = mountWrapper(
//           defaultStore,
//           <AppRepoForm {...defaultProps} packageRepoRef={repo} />,
//         );
//       });
//       await waitFor(() => {
//         wrapper.update();
//         expect(wrapper.find("textarea").at(0).prop("value")).toBe("nginx");
//       });
//       expect(wrapper.find('input[type="checkbox"]').at(0)).toBeChecked();
//       expect(wrapper.find('input[type="checkbox"]').at(1)).toBeChecked();
//     });
//   });
// });
