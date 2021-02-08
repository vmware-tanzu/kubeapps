import { CdsButton } from "@clr/react/button";
import actions from "actions";
import Alert from "components/js/Alert";
import * as React from "react";
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
    validateRepo: jest.fn(),
  };
  const mockDispatch = jest.fn(r => r);
  spyOnUseDispatch = jest.spyOn(ReactRedux, "useDispatch").mockReturnValue(mockDispatch);
});

afterEach(() => {
  actions.kube = { ...kubeaActions };
  spyOnUseDispatch.mockRestore();
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
    await (wrapper
      .find(CdsButton)
      .filterWhere(b => b.html().includes("Install Repo (force)"))
      .prop("onClick") as () => Promise<any>)();
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
    [],
    ["apache", "jenkins"],
  );
});

it("should not show the docker registry credentials section if the namespace is the global one", () => {
  const wrapper = mountWrapper(
    defaultStore,
    <AppRepoForm {...defaultProps} kubeappsNamespace={defaultProps.namespace} />,
  );
  expect(wrapper.html()).not.toContain("Associate Docker Registry Credentials");
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

  const label = wrapper.find("#app-repo-secret-repo-1");
  act(() => {
    label.simulate("change");
  });
  wrapper.find("#kubeapps-repo-url").simulate("change", { target: { value: "http://test" } });
  wrapper.update();

  await act(async () => {
    await (wrapper.find("form").prop("onSubmit") as (e: any) => Promise<any>)({
      preventDefault: jest.fn(),
    });
  });
  expect(install).toHaveBeenCalledWith("", "http://test", "helm", "", "", "", ["repo-1"], []);
});

it("should not show the custom CA field if using an OCI registry", () => {
  const wrapper = mountWrapper(defaultStore, <AppRepoForm {...defaultProps} />);
  wrapper.find("#kubeapps-repo-type-oci").simulate("change");
  const section = wrapper
    .find(".clr-form-control")
    .filterWhere(c => c.text().includes("Custom CA Certificate"));
  expect(section.prop("hidden")).toBeTruthy();
});

it("should not show the list of OCI repositories if using a Helm repo (default)", () => {
  const wrapper = mountWrapper(defaultStore, <AppRepoForm {...defaultProps} />);
  const section = wrapper
    .find(".clr-control-container")
    .filterWhere(c => c.text().includes("List of Repositories"));
  expect(section.prop("hidden")).toBeTruthy();
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
      const repo = { metadata: { name: "foo" } } as any;
      const secret = { data: { "ca.crt": "Zm9v" } } as any;
      const wrapper = mountWrapper(
        defaultStore,
        <AppRepoForm {...defaultProps} repo={repo} secret={secret} />,
      );
      expect(wrapper.find("#kubeapps-repo-custom-ca").prop("value")).toBe("foo");
    });

    it("should parse the existing auth header", () => {
      const repo = { metadata: { name: "foo" } } as any;
      const secret = { data: { authorizationHeader: "Zm9v" } } as any;
      const wrapper = mountWrapper(
        defaultStore,
        <AppRepoForm {...defaultProps} repo={repo} secret={secret} />,
      );
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
      expect(wrapper.find("#app-repo-secret-foo").prop("checked")).toBe(true);
    });
  });
});
