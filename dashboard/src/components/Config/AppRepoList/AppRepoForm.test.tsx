import { shallow } from "enzyme";
import * as React from "react";
import UnexpectedErrorPage from "../../ErrorAlert/UnexpectedErrorAlert";
import { AppRepoForm } from "./AppRepoForm";

const defaultProps = {
  onSubmit: jest.fn(),
  validate: jest.fn(),
  validating: false,
  imagePullSecrets: [],
  namespace: "default",
  kubeappsNamespace: "kubeapps",
  fetchImagePullSecrets: jest.fn(),
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
  const validate = jest.fn(() => true);
  const install = jest.fn(() => true);
  const wrapper = shallow(<AppRepoForm {...defaultProps} validate={validate} onSubmit={install} />);
  const button = wrapper.find("form");
  button.simulate("submit", { preventDefault: jest.fn() });
  // wait for async functions
  await new Promise(s => s());
  expect(install).toHaveBeenCalled();
});

it("should not call the install method when the validation fails unless forced", async () => {
  const validate = jest.fn(() => false);
  const install = jest.fn(() => true);
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
  });
});
