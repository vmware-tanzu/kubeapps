import { shallow } from "enzyme";
import * as React from "react";
import { ISecret } from "shared/types";
import { wait } from "../../../shared/utils";
import UnexpectedErrorPage from "../../ErrorAlert/UnexpectedErrorAlert";
import AppRepoAddDockerCreds from "./AppRepoAddDockerCreds";
import { AppRepoForm } from "./AppRepoForm";

const defaultProps = {
  onSubmit: jest.fn(),
  validate: jest.fn(),
  validating: false,
  imagePullSecrets: [],
  namespace: "default",
  kubeappsNamespace: "kubeapps",
  fetchImagePullSecrets: jest.fn(),
  createDockerRegistrySecret: jest.fn(),
};

it("should render the repo form", () => {
  const wrapper = shallow(<AppRepoForm {...defaultProps} />);
  expect(wrapper).toMatchSnapshot();
});

it("disables the submit button while fetching", () => {
  const wrapper = shallow(<AppRepoForm {...defaultProps} validating={true} />);
  expect(wrapper.find("button").prop("disabled")).toBe(true);
});

it("should show a validation error", () => {
  const wrapper = shallow(<AppRepoForm {...defaultProps} validationError={new Error("Boom!")} />);
  expect(
    wrapper
      .find(UnexpectedErrorPage)
      .dive()
      .text(),
  ).toContain("Boom!");
});

it("should call the install method when the validation success", async () => {
  const validate = jest.fn().mockReturnValue(true);
  const install = jest.fn().mockReturnValue(true);
  const wrapper = shallow(<AppRepoForm {...defaultProps} validate={validate} onSubmit={install} />);
  const button = wrapper.find("form");
  button.simulate("submit", { preventDefault: jest.fn() });
  // wait for async functions
  await new Promise(s => s());
  expect(install).toHaveBeenCalled();
});

it("should not call the install method when the validation fails unless forced", async () => {
  const validate = jest.fn().mockReturnValue(false);
  const install = jest.fn().mockReturnValue(true);
  const wrapper = shallow(<AppRepoForm {...defaultProps} validate={validate} onSubmit={install} />);
  let button = wrapper.find("form");

  button.simulate("submit", { preventDefault: jest.fn() });
  // wait for async functions
  await new Promise(s => s());
  expect(install).not.toHaveBeenCalled();
  wrapper.update();
  button = wrapper.find("button");
  expect(button.text()).toContain("Install Repo (force)");

  wrapper.find("form").simulate("submit", { preventDefault: jest.fn() });
  // wait for async functions
  await new Promise(s => s());
  expect(install).toHaveBeenCalled();
});

it("should not show the docker registry credentials section if the namespace is the global one", () => {
  const wrapper = shallow(
    <AppRepoForm {...defaultProps} kubeappsNamespace={defaultProps.namespace} />,
  );
  expect(wrapper.text()).not.toContain("Associate Docker Registry Credentials");
});

it("should render the docker registry credentials section", () => {
  const wrapper = shallow(<AppRepoForm {...defaultProps} />);
  expect(wrapper.find(AppRepoAddDockerCreds)).toExist();
});

it("should call the install method with the selected docker credentials", async () => {
  const validate = jest.fn().mockReturnValue(true);
  const install = jest.fn().mockReturnValue(true);
  const wrapper = shallow(<AppRepoForm {...defaultProps} validate={validate} onSubmit={install} />);
  wrapper.setState({ selectedImagePullSecrets: { "repo-1": true } });

  const button = wrapper.find("form");
  button.simulate("submit", { preventDefault: jest.fn() });
  // wait for async functions
  await wait(1);
  expect(install).toHaveBeenCalledWith("", "", "", "", "", ["repo-1"]);
});

it("should parse and preserve the selection of docker credentials", () => {
  const secret1 = {
    metadata: {
      name: "foo",
    },
  } as ISecret;
  const secret2 = {
    metadata: {
      name: "bar",
    },
  } as ISecret;
  const wrapper = shallow(<AppRepoForm {...defaultProps} imagePullSecrets={[secret1]} />);
  expect(wrapper.state()).toMatchObject({ selectedImagePullSecrets: { foo: false } });

  // Select secret
  wrapper.setState({ selectedImagePullSecrets: { foo: true } });

  // Add new secret, force re-render
  wrapper.setProps({ imagePullSecrets: [secret1, secret2] });

  expect(wrapper.state()).toMatchObject({ selectedImagePullSecrets: { foo: true, bar: false } });
});

describe("when the repository info is already populated", () => {
  it("should parse the existing name", () => {
    const repo = { metadata: { name: "foo" } } as any;
    const wrapper = shallow(<AppRepoForm {...defaultProps} repo={repo} />);
    expect(wrapper.state()).toMatchObject({ name: "foo" });
    // It should also disable the name input if it's already been set
    expect(
      wrapper
        .find("input")
        .findWhere(i => i.prop("id") === "kubeapps-repo-name")
        .prop("disabled"),
    ).toBe(true);
  });

  it("should parse the existing url", () => {
    const repo = { metadata: { name: "foo" }, spec: { url: "http://repo" } } as any;
    const wrapper = shallow(<AppRepoForm {...defaultProps} repo={repo} />);
    expect(wrapper.state()).toMatchObject({ url: "http://repo" });
  });

  it("should parse the existing syncJobPodTemplate", () => {
    const repo = { metadata: { name: "foo" }, spec: { syncJobPodTemplate: { foo: "bar" } } } as any;
    const wrapper = shallow(<AppRepoForm {...defaultProps} repo={repo} />);
    expect(wrapper.state()).toMatchObject({ syncJobPodTemplate: "foo: bar\n" });
  });

  describe("when there is a secret associated to the repo", () => {
    it("should parse the existing CA cert", () => {
      const repo = { metadata: { name: "foo" } } as any;
      const secret = { data: { "ca.crt": "Zm9v" } } as any;
      const wrapper = shallow(<AppRepoForm {...defaultProps} repo={repo} secret={secret} />);
      expect(wrapper.state()).toMatchObject({ customCA: "foo" });
    });

    it("should parse the existing auth header", () => {
      const repo = { metadata: { name: "foo" } } as any;
      const secret = { data: { authorizationHeader: "Zm9v" } } as any;
      const wrapper = shallow(<AppRepoForm {...defaultProps} repo={repo} secret={secret} />);
      expect(wrapper.state()).toMatchObject({ authHeader: "foo", authMethod: "custom" });
    });

    it("should parse the existing basic auth", () => {
      const repo = { metadata: { name: "foo" } } as any;
      const secret = { data: { authorizationHeader: "QmFzaWMgWm05dk9tSmhjZz09" } } as any;
      const wrapper = shallow(<AppRepoForm {...defaultProps} repo={repo} secret={secret} />);
      expect(wrapper.state()).toMatchObject({
        authHeader: "Basic Zm9vOmJhcg==",
        user: "foo",
        password: "bar",
        authMethod: "basic",
      });
    });

    it("should parse a bearer token", () => {
      const repo = { metadata: { name: "foo" } } as any;
      const secret = { data: { authorizationHeader: "QmVhcmVyIGZvbw==" } } as any;
      const wrapper = shallow(<AppRepoForm {...defaultProps} repo={repo} secret={secret} />);
      expect(wrapper.state()).toMatchObject({
        authHeader: "Bearer foo",
        token: "foo",
        authMethod: "bearer",
      });
    });

    it("should pre-select the existing docker registry secret", () => {
      const secret = {
        metadata: {
          name: "foo",
        },
      } as ISecret;
      const repo = { metadata: { name: "foo" }, spec: { dockerRegistrySecrets: ["foo"] } } as any;
      const wrapper = shallow(
        <AppRepoForm {...defaultProps} imagePullSecrets={[secret]} repo={repo} />,
      );
      expect(wrapper.state()).toMatchObject({ selectedImagePullSecrets: { foo: true } });

      // Add new secret, force re-render
      const secret2 = {
        metadata: {
          name: "bar",
        },
      } as ISecret;
      wrapper.setProps({ imagePullSecrets: [secret, secret2] });

      expect(wrapper.state()).toMatchObject({
        selectedImagePullSecrets: { foo: true, bar: false },
      });
    });
  });
});
