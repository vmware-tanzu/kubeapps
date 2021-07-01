import { CdsButton } from "@cds/react/button";
import actions from "actions";
import Alert from "components/js/Alert";
import { act } from "react-dom/test-utils";
import * as ReactRedux from "react-redux";
import { defaultStore, getStore, mountWrapper } from "shared/specs/mountWrapper";
import { ISecret } from "shared/types";
import AppRepoAddDockerCreds from "./AppRepoAddDockerCreds";
import { AppRepoForm } from "./AppRepoForm";

const defaultProps = {
  onSubmit: jest.fn(),
  namespace: "default",
  kubeappsNamespace: "kubeapps",
};

let spyOnUseDispatch: jest.SpyInstance;
const kubeaActions = { ...actions.kube };
beforeEach(() => {
  actions.repos = {
    ...actions.repos,
    validateRepo: jest.fn().mockReturnValue(true),
    fetchImagePullSecrets: jest.fn(),
    fetchRepoSecret: jest.fn(),
  };
  const mockDispatch = jest.fn(r => r);
  spyOnUseDispatch = jest.spyOn(ReactRedux, "useDispatch").mockReturnValue(mockDispatch);
});

afterEach(() => {
  actions.kube = { ...kubeaActions };
  spyOnUseDispatch.mockRestore();
});

it("fetches repos and imagePullSecrets", () => {
  mountWrapper(defaultStore, <AppRepoForm {...defaultProps} />);
  expect(actions.repos.fetchImagePullSecrets).toHaveBeenCalledWith(defaultProps.namespace);
});

it("disables the submit button while fetching", () => {
  const wrapper = mountWrapper(
    getStore({ repos: { validating: true } }),
    <AppRepoForm {...defaultProps} />,
  );
  expect(
    wrapper
      .find(CdsButton)
      .filterWhere(b => b.html().includes("Validating"))
      .prop("disabled"),
  ).toBe(true);
});

it("should show a validation error", () => {
  const wrapper = mountWrapper(
    getStore({ repos: { errors: { validate: new Error("Boom!") } } }),
    <AppRepoForm {...defaultProps} />,
  );
  expect(wrapper.find(Alert).text()).toContain("Boom!");
});

it("shows an error updating a repo", () => {
  const wrapper = mountWrapper(
    getStore({ repos: { errors: { update: new Error("boom!") } } }),
    <AppRepoForm {...defaultProps} />,
  );
  expect(wrapper.find(Alert)).toIncludeText("boom!");
});

it("should call the install method when the validation success", async () => {
  const validateRepo = jest.fn().mockReturnValue(true);
  const install = jest.fn().mockReturnValue(true);
  actions.repos = {
    ...actions.repos,
    validateRepo,
  };
  const wrapper = mountWrapper(defaultStore, <AppRepoForm {...defaultProps} onSubmit={install} />);
  const form = wrapper.find("form");
  await act(async () => {
    await (form.prop("onSubmit") as (e: any) => Promise<any>)({ preventDefault: jest.fn() });
  });
  wrapper.update();
  expect(install).toHaveBeenCalled();
});

it("should not call the install method when the validation fails unless forced", async () => {
  const validateRepo = jest.fn().mockReturnValue(false);
  const install = jest.fn().mockReturnValue(true);
  actions.repos = {
    ...actions.repos,
    validateRepo,
  };
  const wrapper = mountWrapper(defaultStore, <AppRepoForm {...defaultProps} onSubmit={install} />);
  const form = wrapper.find("form");
  await act(async () => {
    await (form.prop("onSubmit") as (e: any) => Promise<any>)({ preventDefault: jest.fn() });
  });
  wrapper.update();
  expect(install).not.toHaveBeenCalled();

  expect(
    wrapper
      .find(CdsButton)
      .filterWhere(b => b.text().includes("Install"))
      .text(),
  ).toContain("Install Repo (force)");

  // So disabling this test for the moment.
  await act(async () => {
    await (
      wrapper
        .find(CdsButton)
        .filterWhere(b => b.html().includes("Install Repo (force)"))
        .prop("onClick") as () => Promise<any>
    )();
  });
  expect(install).toHaveBeenCalled();
});

it("should call the install method with OCI information", async () => {
  const validateRepo = jest.fn().mockReturnValue(true);
  const install = jest.fn().mockReturnValue(true);
  actions.repos = {
    ...actions.repos,
    validateRepo,
  };
  const wrapper = mountWrapper(defaultStore, <AppRepoForm {...defaultProps} onSubmit={install} />);
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
  expect(install).toHaveBeenCalledWith(
    "",
    "https://oci.repo",
    "oci",
    "",
    "",
    "",
    "",
    "",
    [],
    ["apache", "jenkins"],
    false,
    false,
    undefined,
  );
});

it("should call the install skipping TLS verification", async () => {
  const validateRepo = jest.fn().mockReturnValue(true);
  const install = jest.fn().mockReturnValue(true);
  actions.repos = {
    ...actions.repos,
    validateRepo,
  };
  const wrapper = mountWrapper(defaultStore, <AppRepoForm {...defaultProps} onSubmit={install} />);
  wrapper.find("#kubeapps-repo-url").simulate("change", { target: { value: "helm.repo" } });
  wrapper.find("#kubeapps-repo-skip-tls").simulate("change");
  const form = wrapper.find("form");
  await act(async () => {
    await (form.prop("onSubmit") as (e: any) => Promise<any>)({ preventDefault: jest.fn() });
  });
  wrapper.update();
  expect(install).toHaveBeenCalledWith(
    "",
    "https://helm.repo",
    "helm",
    "",
    "",
    "",
    "",
    "",
    [],
    [],
    true,
    false,
    undefined,
  );
});

it("should call the install passing credentials", async () => {
  const validateRepo = jest.fn().mockReturnValue(true);
  const install = jest.fn().mockReturnValue(true);
  actions.repos = {
    ...actions.repos,
    validateRepo,
  };
  const wrapper = mountWrapper(defaultStore, <AppRepoForm {...defaultProps} onSubmit={install} />);
  wrapper.find("#kubeapps-repo-url").simulate("change", { target: { value: "helm.repo" } });
  wrapper.find("#kubeapps-repo-pass-credentials").simulate("change");
  const form = wrapper.find("form");
  await act(async () => {
    await (form.prop("onSubmit") as (e: any) => Promise<any>)({ preventDefault: jest.fn() });
  });
  wrapper.update();
  expect(install).toHaveBeenCalledWith(
    "",
    "https://helm.repo",
    "helm",
    "",
    "",
    "",
    "",
    "",
    [],
    [],
    false,
    true,
    undefined,
  );
});

describe("when using a filter", () => {
  it("should call the install method with a filter", async () => {
    const install = jest.fn().mockReturnValue(true);
    const wrapper = mountWrapper(
      defaultStore,
      <AppRepoForm {...defaultProps} onSubmit={install} />,
    );
    wrapper
      .find("#kubeapps-repo-url")
      .simulate("change", { target: { value: "https://helm.repo" } });
    wrapper
      .find("textarea")
      .at(0)
      .simulate("change", { target: { value: "nginx, wordpress" } });
    const form = wrapper.find("form");
    await act(async () => {
      await (form.prop("onSubmit") as (e: any) => Promise<any>)({ preventDefault: jest.fn() });
    });
    wrapper.update();
    expect(install).toHaveBeenCalledWith(
      "",
      "https://helm.repo",
      "helm",
      "",
      "",
      "",
      "",
      "",
      [],
      [],
      false,
      false,
      { jq: ".name == $var0 or .name == $var1", variables: { $var0: "nginx", $var1: "wordpress" } },
    );
  });

  it("should call the install method with a filter excluding a regex", async () => {
    const install = jest.fn().mockReturnValue(true);
    const wrapper = mountWrapper(
      defaultStore,
      <AppRepoForm {...defaultProps} onSubmit={install} />,
    );
    wrapper
      .find("#kubeapps-repo-url")
      .simulate("change", { target: { value: "https://helm.repo" } });
    wrapper
      .find("textarea")
      .at(0)
      .simulate("change", { target: { value: "nginx" } });
    wrapper.find('input[type="checkbox"]').at(0).simulate("change");
    wrapper.find('input[type="checkbox"]').at(1).simulate("change");
    const form = wrapper.find("form");
    await act(async () => {
      await (form.prop("onSubmit") as (e: any) => Promise<any>)({ preventDefault: jest.fn() });
    });
    wrapper.update();
    expect(install).toHaveBeenCalledWith(
      "",
      "https://helm.repo",
      "helm",
      "",
      "",
      "",
      "",
      "",
      [],
      [],
      false,
      false,
      { jq: ".name | test($var) | not", variables: { $var: "nginx" } },
    );
  });

  it("ignore the filter for the OCI case", async () => {
    const install = jest.fn().mockReturnValue(true);
    const wrapper = mountWrapper(
      defaultStore,
      <AppRepoForm {...defaultProps} onSubmit={install} />,
    );
    wrapper
      .find("#kubeapps-repo-url")
      .simulate("change", { target: { value: "https://oci.repo" } });
    wrapper
      .find("textarea")
      .at(0)
      .simulate("change", { target: { value: "nginx, wordpress" } });
    wrapper.find("#kubeapps-repo-type-oci").simulate("change");
    const form = wrapper.find("form");
    await act(async () => {
      await (form.prop("onSubmit") as (e: any) => Promise<any>)({ preventDefault: jest.fn() });
    });
    wrapper.update();
    expect(install).toHaveBeenCalledWith(
      "",
      "https://oci.repo",
      "oci",
      "",
      "",
      "",
      "",
      "",
      [],
      [],
      false,
      false,
      undefined,
    );
  });
});

describe("when using a description", () => {
  it("should call the install method with a description", async () => {
    const install = jest.fn().mockReturnValue(true);
    const wrapper = mountWrapper(
      defaultStore,
      <AppRepoForm {...defaultProps} onSubmit={install} />,
    );
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
    expect(install).toHaveBeenCalledWith(
      "",
      "https://helm.repo",
      "helm",
      "description test",
      "",
      "",
      "",
      "",
      [],
      [],
      false,
      false,
      undefined,
    );
  });
});

it("should disable the docker registry credentials section if the namespace is the global one", () => {
  const wrapper = mountWrapper(
    defaultStore,
    <AppRepoForm {...defaultProps} kubeappsNamespace={defaultProps.namespace} />,
  );
  expect(wrapper.find("select")).toBeDisabled();
  expect(wrapper.find(".docker-creds-subform-button button")).toBeDisabled();
});

it("should render the docker registry credentials section", () => {
  const wrapper = mountWrapper(defaultStore, <AppRepoForm {...defaultProps} />);
  expect(wrapper.find(AppRepoAddDockerCreds)).toExist();
});

it("should call the install method with the selected docker credentials", async () => {
  const validateRepo = jest.fn().mockReturnValue(true);
  const install = jest.fn().mockReturnValue(true);
  actions.repos = {
    ...actions.repos,
    validateRepo,
  };
  const secret = {
    metadata: {
      name: "repo-1",
    },
  } as ISecret;

  const wrapper = mountWrapper(
    getStore({
      repos: { imagePullSecrets: [secret] },
    }),
    <AppRepoForm {...defaultProps} onSubmit={install} />,
  );

  const label = wrapper.find("select");
  act(() => {
    label.simulate("change", { target: { value: "repo-1" } });
  });
  const radio = wrapper.find("#kubeapps-repo-auth-method-registry");
  act(() => {
    radio.simulate("change", { target: { value: "registry" } });
  });
  wrapper.find("#kubeapps-repo-url").simulate("change", { target: { value: "http://test" } });
  wrapper.update();

  await act(async () => {
    await (wrapper.find("form").prop("onSubmit") as (e: any) => Promise<any>)({
      preventDefault: jest.fn(),
    });
  });
  expect(install).toHaveBeenCalledWith(
    "",
    "http://test",
    "helm",
    "",
    "",
    "repo-1",
    "",
    "",
    ["repo-1"],
    [],
    false,
    false,
    undefined,
  );
});

it("should call the install reusing as auth the selected docker credentials", async () => {
  const validateRepo = jest.fn().mockReturnValue(true);
  const install = jest.fn().mockReturnValue(true);
  actions.repos = {
    ...actions.repos,
    validateRepo,
  };
  const secret = {
    metadata: {
      name: "repo-1",
    },
  } as ISecret;

  const wrapper = mountWrapper(
    getStore({
      repos: { imagePullSecrets: [secret] },
    }),
    <AppRepoForm {...defaultProps} onSubmit={install} />,
  );

  const label = wrapper.find("select");
  act(() => {
    label.simulate("change", { target: { value: "repo-1" } });
  });
  wrapper.find("#kubeapps-repo-url").simulate("change", { target: { value: "http://test" } });
  wrapper.update();

  await act(async () => {
    await (wrapper.find("form").prop("onSubmit") as (e: any) => Promise<any>)({
      preventDefault: jest.fn(),
    });
  });
  expect(install).toHaveBeenCalledWith(
    "",
    "http://test",
    "helm",
    "",
    "",
    "",
    "",
    "",
    ["repo-1"],
    [],
    false,
    false,
    undefined,
  );
});

it("should not show the list of OCI repositories if using a Helm repo (default)", () => {
  const wrapper = mountWrapper(defaultStore, <AppRepoForm {...defaultProps} />);
  expect(wrapper.find("#kubeapps-oci-repositories")).not.toExist();
});

describe("when the repository info is already populated", () => {
  it("should parse the existing name", () => {
    const repo = { metadata: { name: "foo" } } as any;
    const wrapper = mountWrapper(defaultStore, <AppRepoForm {...defaultProps} repo={repo} />);
    expect(wrapper.find("#kubeapps-repo-name").prop("value")).toBe("foo");
    // It should also disable the name input if it's already been set
    expect(wrapper.find("#kubeapps-repo-name").prop("disabled")).toBe(true);
  });

  it("should parse the existing url", () => {
    const repo = { metadata: { name: "foo" }, spec: { url: "http://repo" } } as any;
    const wrapper = mountWrapper(defaultStore, <AppRepoForm {...defaultProps} repo={repo} />);
    expect(wrapper.find("#kubeapps-repo-url").prop("value")).toBe("http://repo");
  });

  it("should parse the existing syncJobPodTemplate", () => {
    const repo = { metadata: { name: "foo" }, spec: { syncJobPodTemplate: { foo: "bar" } } } as any;
    const wrapper = mountWrapper(defaultStore, <AppRepoForm {...defaultProps} repo={repo} />);
    expect(wrapper.find("#kubeapps-repo-sync-job-tpl").prop("value")).toBe("foo: bar\n");
  });

  describe("when there is a secret associated to the repo", () => {
    it("should parse the existing CA cert", () => {
      const repo = {
        metadata: { name: "foo" },
        spec: { auth: { customCA: { secretKeyRef: { name: "bar" } } } },
      } as any;
      const secret = { data: { "ca.crt": "Zm9v" } } as any;
      const wrapper = mountWrapper(
        defaultStore,
        <AppRepoForm {...defaultProps} repo={repo} secret={secret} />,
      );
      expect(actions.repos.fetchRepoSecret).toHaveBeenCalledWith("default", "bar");
      expect(wrapper.find("#kubeapps-repo-custom-ca").prop("value")).toBe("foo");
    });

    it("should parse the existing auth header", () => {
      const repo = {
        metadata: { name: "foo" },
        spec: { auth: { header: { secretKeyRef: { name: "bar" } } } },
      } as any;
      const secret = { data: { authorizationHeader: "Zm9v" } } as any;
      const wrapper = mountWrapper(
        defaultStore,
        <AppRepoForm {...defaultProps} repo={repo} secret={secret} />,
      );
      expect(actions.repos.fetchRepoSecret).toHaveBeenCalledWith("default", "bar");
      expect(wrapper.find("#kubeapps-repo-custom-header").prop("value")).toBe("foo");
    });

    it("should parse the existing basic auth", () => {
      const repo = { metadata: { name: "foo" } } as any;
      const secret = { data: { authorizationHeader: "QmFzaWMgWm05dk9tSmhjZz09" } } as any;
      const wrapper = mountWrapper(
        defaultStore,
        <AppRepoForm {...defaultProps} repo={repo} secret={secret} />,
      );
      expect(wrapper.find("#kubeapps-repo-username").prop("value")).toBe("foo");
      expect(wrapper.find("#kubeapps-repo-password").prop("value")).toBe("bar");
    });

    it("should parse the existing type", () => {
      const repo = { metadata: { name: "foo" }, spec: { type: "oci" } } as any;
      const wrapper = mountWrapper(defaultStore, <AppRepoForm {...defaultProps} repo={repo} />);
      expect(wrapper.find("#kubeapps-repo-type-oci")).toBeChecked();
      expect(wrapper.find("#kubeapps-oci-repositories")).toExist();
    });

    it("should parse the existing skip tls config", () => {
      const repo = { metadata: { name: "foo" }, spec: { tlsInsecureSkipVerify: true } } as any;
      const wrapper = mountWrapper(defaultStore, <AppRepoForm {...defaultProps} repo={repo} />);
      expect(wrapper.find("#kubeapps-repo-skip-tls")).toBeChecked();
    });

    it("should parse the existing pass credentials config", () => {
      const repo = { metadata: { name: "foo" }, spec: { passCredentials: true } } as any;
      const wrapper = mountWrapper(defaultStore, <AppRepoForm {...defaultProps} repo={repo} />);
      expect(wrapper.find("#kubeapps-repo-pass-credentials")).toBeChecked();
    });

    it("should parse a bearer token", () => {
      const repo = { metadata: { name: "foo" } } as any;
      const secret = { data: { authorizationHeader: "QmVhcmVyIGZvbw==" } } as any;
      const wrapper = mountWrapper(
        defaultStore,
        <AppRepoForm {...defaultProps} repo={repo} secret={secret} />,
      );
      expect(wrapper.find("#kubeapps-repo-token").prop("value")).toBe("foo");
    });

    it("should select a docker secret as auth mechanism", () => {
      const repo = { metadata: { name: "foo" } } as any;
      const secret = { data: { ".dockerconfigjson": "QmVhcmVyIGZvbw==" } } as any;
      const wrapper = mountWrapper(
        defaultStore,
        <AppRepoForm {...defaultProps} repo={repo} secret={secret} />,
      );
      expect(wrapper.find("#kubeapps-repo-auth-method-registry")).toBeChecked();
    });

    it("should pre-select the existing docker registry secret", () => {
      const secret = {
        metadata: {
          name: "foo",
        },
      } as ISecret;
      const repo = { metadata: { name: "foo" }, spec: { dockerRegistrySecrets: ["foo"] } } as any;
      const wrapper = mountWrapper(
        getStore({
          repos: { imagePullSecrets: [secret] },
        }),
        <AppRepoForm {...defaultProps} repo={repo} />,
      );
      expect(wrapper.find("select").prop("value")).toBe("foo");
    });

    it("should parse the existing filter (simple)", () => {
      const repo = {
        metadata: { name: "foo" },
        spec: {
          type: "helm",
          filterRule: {
            jq: ".name == $var0 or .name == $var1",
            variables: { $var0: "nginx", $var1: "wordpress" },
          },
        },
      } as any;
      const wrapper = mountWrapper(defaultStore, <AppRepoForm {...defaultProps} repo={repo} />);
      expect(wrapper.find("textarea").at(0).prop("value")).toBe("nginx, wordpress");

      expect(wrapper.find('input[type="checkbox"]').at(0)).not.toBeChecked();
      expect(wrapper.find('input[type="checkbox"]').at(1)).not.toBeChecked();
    });
    it("should parse the existing filter (negated regex)", () => {
      const repo = {
        metadata: { name: "foo" },
        spec: {
          type: "helm",
          filterRule: { jq: ".name | test($var) | not", variables: { $var: "nginx" } },
        },
      } as any;
      const wrapper = mountWrapper(defaultStore, <AppRepoForm {...defaultProps} repo={repo} />);
      expect(wrapper.find("textarea").at(0).prop("value")).toBe("nginx");

      expect(wrapper.find('input[type="checkbox"]').at(0)).toBeChecked();
      expect(wrapper.find('input[type="checkbox"]').at(1)).toBeChecked();
    });
  });
});
