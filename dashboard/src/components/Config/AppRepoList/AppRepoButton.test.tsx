import { mount } from "enzyme";
import * as React from "react";
import Modal from "react-modal";
import { ConflictError, UnprocessableEntity } from "../../../shared/types";
import { wait } from "../../../shared/utils";
import ErrorSelector from "../../ErrorAlert/ErrorSelector";
import { AppRepoAddButton } from "./AppRepoButton";
import { AppRepoForm } from "./AppRepoForm";

const defaultProps = {
  onSubmit: jest.fn(),
  validate: jest.fn().mockReturnValue(true),
  namespace: "kubeapps",
  kubeappsNamespace: "kubeapps",
  validating: false,
  errors: {},
  imagePullSecrets: [],
  fetchImagePullSecrets: jest.fn(),
  createDockerRegistrySecret: jest.fn(),
};

it("should open a modal with the repository form", () => {
  const wrapper = mount(<AppRepoAddButton {...defaultProps} />);
  Modal.setAppElement(document.createElement("div"));
  wrapper.setState({ modalIsOpen: true });
  expect(wrapper).toMatchSnapshot();
});

it("should install a repository with a custom auth header", async () => {
  const install = jest.fn().mockReturnValue(true);
  const wrapper = mount(<AppRepoAddButton {...defaultProps} onSubmit={install} />);
  Modal.setAppElement(document.createElement("div"));
  wrapper.setState({ modalIsOpen: true });
  wrapper.update();
  wrapper.find(AppRepoForm).setState({
    modalIsOpen: true,
    authMethod: "custom",
    name: "my-repo",
    url: "http://foo.bar",
    authHeader: "foo",
    customCA: "bar",
  });

  const button = wrapper.find(AppRepoForm).find(".button-primary");
  button.simulate("submit");

  // Wait for the Modal to be closed
  await wait(1);
  expect(install).toBeCalledWith("my-repo", "kubeapps", "http://foo.bar", "foo", "bar", "", []);
  expect(wrapper.state("modalIsOpen")).toBe(false);
});

it("should install a repository with basic auth", async () => {
  const install = jest.fn().mockReturnValue(true);
  const wrapper = mount(<AppRepoAddButton {...defaultProps} onSubmit={install} />);
  Modal.setAppElement(document.createElement("div"));
  wrapper.setState({ modalIsOpen: true });
  wrapper.update();
  wrapper.find(AppRepoForm).setState({
    modalIsOpen: true,
    authMethod: "basic",
    name: "my-repo",
    url: "http://foo.bar",
    user: "foo",
    password: "bar",
  });

  const button = wrapper.find(AppRepoForm).find(".button-primary");
  button.simulate("submit");

  // Wait for the Modal to be closed
  await wait(1);
  expect(install).toBeCalledWith(
    "my-repo",
    "kubeapps",
    "http://foo.bar",
    "Basic Zm9vOmJhcg==",
    "",
    "",
    [],
  );
  expect(wrapper.state("modalIsOpen")).toBe(false);
});

it("should install a repository with a bearer token", async () => {
  const install = jest.fn().mockReturnValue(true);
  const wrapper = mount(<AppRepoAddButton {...defaultProps} onSubmit={install} />);
  Modal.setAppElement(document.createElement("div"));
  wrapper.setState({ modalIsOpen: true });
  wrapper.update();
  wrapper.find(AppRepoForm).setState({
    modalIsOpen: true,
    authMethod: "bearer",
    name: "my-repo",
    url: "http://foo.bar",
    token: "foobar",
  });

  const button = wrapper.find(AppRepoForm).find(".button-primary");
  button.simulate("submit");

  // Wait for the Modal to be closed
  await wait(1);
  expect(install).toBeCalledWith(
    "my-repo",
    "kubeapps",
    "http://foo.bar",
    "Bearer foobar",
    "",
    "",
    [],
  );
  expect(wrapper.state("modalIsOpen")).toBe(false);
});

it("should install a repository with a podSpecTemplate", async () => {
  const install = jest.fn().mockReturnValue(true);
  const wrapper = mount(<AppRepoAddButton {...defaultProps} onSubmit={install} />);
  Modal.setAppElement(document.createElement("div"));
  wrapper.setState({ modalIsOpen: true });
  wrapper.update();
  wrapper.find(AppRepoForm).setState({
    modalIsOpen: true,
    authMethod: "bearer",
    name: "my-repo",
    url: "http://foo.bar",
    syncJobPodTemplate: "foo: bar",
  });

  const button = wrapper.find(AppRepoForm).find(".button-primary");
  button.simulate("submit");

  // Wait for the Modal to be closed
  await wait(1);
  expect(install).toBeCalledWith(
    "my-repo",
    "kubeapps",
    "http://foo.bar",
    "Bearer ",
    "",
    "foo: bar",
    [],
  );
  expect(wrapper.state("modalIsOpen")).toBe(false);
});

describe("render error", () => {
  it("renders a conflict error", async () => {
    const wrapper = mount(<AppRepoAddButton {...defaultProps} />);
    Modal.setAppElement(document.createElement("div"));
    wrapper.setState({ modalIsOpen: true });
    wrapper.update();
    wrapper.find(AppRepoForm).setState({ name: "my-repo" });

    const button = wrapper.find(AppRepoForm).find(".button-primary");
    button.simulate("submit");
    wrapper.setProps({ errors: { create: new ConflictError("already exists!") } });

    await wait(1);
    expect(wrapper.find(ErrorSelector).text()).toContain(
      "App Repository my-repo already exists, try a different name.",
    );
    // Now changing the name should not change the error message
    wrapper.setState({ name: "my-app-2" });
    wrapper.update();
    expect(wrapper.find(ErrorSelector).text()).toContain(
      "App Repository my-repo already exists, try a different name.",
    );
  });

  it("renders an 'unprocessable entity' error", async () => {
    const wrapper = mount(<AppRepoAddButton {...defaultProps} />);
    Modal.setAppElement(document.createElement("div"));
    wrapper.setState({ modalIsOpen: true });
    wrapper.update();
    wrapper.find(AppRepoForm).setState({ name: "my-repo" });

    const button = wrapper.find(AppRepoForm).find(".button-primary");
    button.simulate("submit");
    wrapper.setProps({ errors: { create: new UnprocessableEntity("cannot process this!") } });

    await wait(1);
    expect(wrapper.find(ErrorSelector).text()).toContain(
      "Something went wrong processing App Repository my-repo",
    );
    expect(wrapper.find(ErrorSelector).text()).toContain("cannot process this!");
  });
});

it("should modify the default button text", () => {
  const wrapper = mount(<AppRepoAddButton {...defaultProps} text="foo" />);
  expect(wrapper.text()).toContain("foo");
});
