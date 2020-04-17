import { mount } from "enzyme";
import * as React from "react";
import * as ReactModal from "react-modal";
import { ConflictError, UnprocessableEntity } from "../../../shared/types";
import ErrorSelector from "../../ErrorAlert/ErrorSelector";
import { AppRepoAddButton } from "./AppRepoButton";
import { AppRepoForm } from "./AppRepoForm";

const defaultProps = {
  onSubmit: jest.fn(),
  validate: jest.fn(() => true),
  namespace: "kubeapps",
  validating: false,
  errors: {},
};

it("should open a modal with the repository form", () => {
  const wrapper = mount(<AppRepoAddButton {...defaultProps} />);
  ReactModal.setAppElement(document.createElement("div"));
  wrapper.setState({ modalIsOpen: true });
  expect(wrapper).toMatchSnapshot();
});

it("should install a repository with a custom auth header", done => {
  const install = jest.fn(() => true);
  const wrapper = mount(<AppRepoAddButton {...defaultProps} onSubmit={install} />);
  ReactModal.setAppElement(document.createElement("div"));
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
  setTimeout(() => {
    expect(install).toBeCalledWith("my-repo", "kubeapps", "http://foo.bar", "foo", "bar", "");
    expect(wrapper.state("modalIsOpen")).toBe(false);
    done();
  }, 1);
});

it("should install a repository with basic auth", done => {
  const install = jest.fn(() => true);
  const wrapper = mount(<AppRepoAddButton {...defaultProps} onSubmit={install} />);
  ReactModal.setAppElement(document.createElement("div"));
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
  setTimeout(() => {
    expect(install).toBeCalledWith(
      "my-repo",
      "kubeapps",
      "http://foo.bar",
      "Basic Zm9vOmJhcg==",
      "",
      "",
    );
    expect(wrapper.state("modalIsOpen")).toBe(false);
    done();
  }, 1);
});

it("should install a repository with a bearer token", done => {
  const install = jest.fn(() => true);
  const wrapper = mount(<AppRepoAddButton {...defaultProps} onSubmit={install} />);
  ReactModal.setAppElement(document.createElement("div"));
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
  setTimeout(() => {
    expect(install).toBeCalledWith(
      "my-repo",
      "kubeapps",
      "http://foo.bar",
      "Bearer foobar",
      "",
      "",
    );
    expect(wrapper.state("modalIsOpen")).toBe(false);
    done();
  }, 1);
});

it("should install a repository with a podSpecTemplate", done => {
  const install = jest.fn(() => true);
  const wrapper = mount(<AppRepoAddButton {...defaultProps} onSubmit={install} />);
  ReactModal.setAppElement(document.createElement("div"));
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
  setTimeout(() => {
    expect(install).toBeCalledWith(
      "my-repo",
      "kubeapps",
      "http://foo.bar",
      "Bearer ",
      "",
      "foo: bar",
    );
    expect(wrapper.state("modalIsOpen")).toBe(false);
    done();
  }, 1);
});

describe("render error", () => {
  it("renders a conflict error", done => {
    const wrapper = mount(<AppRepoAddButton {...defaultProps} />);
    ReactModal.setAppElement(document.createElement("div"));
    wrapper.setState({ modalIsOpen: true });
    wrapper.update();
    wrapper.find(AppRepoForm).setState({ name: "my-repo" });

    const button = wrapper.find(AppRepoForm).find(".button-primary");
    button.simulate("submit");
    wrapper.setProps({ errors: { create: new ConflictError("already exists!") } });

    setTimeout(() => {
      expect(wrapper.find(ErrorSelector).text()).toContain(
        "App Repository my-repo already exists, try a different name.",
      );
      // Now changing the name should not change the error message
      wrapper.setState({ name: "my-app-2" });
      wrapper.update();
      expect(wrapper.find(ErrorSelector).text()).toContain(
        "App Repository my-repo already exists, try a different name.",
      );
      done();
    }, 1);
  });

  it("renders an 'unprocessable entity' error", done => {
    const wrapper = mount(<AppRepoAddButton {...defaultProps} />);
    ReactModal.setAppElement(document.createElement("div"));
    wrapper.setState({ modalIsOpen: true });
    wrapper.update();
    wrapper.find(AppRepoForm).setState({ name: "my-repo" });

    const button = wrapper.find(AppRepoForm).find(".button-primary");
    button.simulate("submit");
    wrapper.setProps({ errors: { create: new UnprocessableEntity("cannot process this!") } });

    setTimeout(() => {
      expect(wrapper.find(ErrorSelector).text()).toContain(
        "Something went wrong processing App Repository my-repo",
      );
      expect(wrapper.find(ErrorSelector).text()).toContain("cannot process this!");
      done();
    }, 1);
  });
});

it("should modify the default button text", () => {
  const wrapper = mount(<AppRepoAddButton {...defaultProps} text="foo" />);
  expect(wrapper.text()).toContain("foo");
});
