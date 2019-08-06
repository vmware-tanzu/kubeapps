import { mount } from "enzyme";
import * as React from "react";
import * as ReactModal from "react-modal";
import { ConflictError, UnprocessableEntity } from "../../../shared/types";
import ErrorSelector from "../../ErrorAlert/ErrorSelector";
import { AppRepoAddButton } from "./AppRepoButton";
import { AppRepoForm } from "./AppRepoForm";

const defaultProps = {
  install: jest.fn(),
  kubeappsNamespace: "kubeapps",
};

it("should open a modal with the repository form", () => {
  const wrapper = mount(<AppRepoAddButton {...defaultProps} />);
  ReactModal.setAppElement(document.createElement("div"));
  wrapper.setState({ modalIsOpen: true });
  expect(wrapper).toMatchSnapshot();
});

it("should install a repository with a custom auth header", done => {
  const install = jest.fn(() => true);
  const wrapper = mount(<AppRepoAddButton {...defaultProps} install={install} />);
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

  const button = wrapper.find(AppRepoForm).find(".button");
  button.simulate("submit");

  expect(install).toBeCalledWith("my-repo", "http://foo.bar", "foo", "bar", "");
  // Wait for the Modal to be closed
  setTimeout(() => {
    expect(wrapper.state("modalIsOpen")).toBe(false);
    done();
  }, 1);
});

it("should install a repository with basic auth", done => {
  const install = jest.fn(() => true);
  const wrapper = mount(<AppRepoAddButton {...defaultProps} install={install} />);
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

  const button = wrapper.find(AppRepoForm).find(".button");
  button.simulate("submit");

  expect(install).toBeCalledWith("my-repo", "http://foo.bar", "Basic Zm9vOmJhcg==", "", "");
  // Wait for the Modal to be closed
  setTimeout(() => {
    expect(wrapper.state("modalIsOpen")).toBe(false);
    done();
  }, 1);
});

it("should install a repository with a bearer token", done => {
  const install = jest.fn(() => true);
  const wrapper = mount(<AppRepoAddButton {...defaultProps} install={install} />);
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

  const button = wrapper.find(AppRepoForm).find(".button");
  button.simulate("submit");

  expect(install).toBeCalledWith("my-repo", "http://foo.bar", "Bearer foobar", "", "");
  // Wait for the Modal to be closed
  setTimeout(() => {
    expect(wrapper.state("modalIsOpen")).toBe(false);
    done();
  }, 1);
});

it("should install a repository with a podSpecTemplate", done => {
  const install = jest.fn(() => true);
  const wrapper = mount(<AppRepoAddButton {...defaultProps} install={install} />);
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

  const button = wrapper.find(AppRepoForm).find(".button");
  button.simulate("submit");

  expect(install).toBeCalledWith("my-repo", "http://foo.bar", "Bearer ", "", "foo: bar");
  // Wait for the Modal to be closed
  setTimeout(() => {
    expect(wrapper.state("modalIsOpen")).toBe(false);
    done();
  }, 1);
});

describe("render error", () => {
  it("renders a conflict error", () => {
    const wrapper = mount(<AppRepoAddButton {...defaultProps} />);
    ReactModal.setAppElement(document.createElement("div"));
    wrapper.setState({ modalIsOpen: true });
    wrapper.update();
    wrapper.find(AppRepoForm).setState({ name: "my-repo" });

    const button = wrapper.find(AppRepoForm).find(".button");
    button.simulate("submit");
    wrapper.setProps({ error: new ConflictError("already exists!") });

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

  it("renders an 'unprocessable entity' error", () => {
    const wrapper = mount(<AppRepoAddButton {...defaultProps} />);
    ReactModal.setAppElement(document.createElement("div"));
    wrapper.setState({ modalIsOpen: true });
    wrapper.update();
    wrapper.find(AppRepoForm).setState({ name: "my-repo" });

    const button = wrapper.find(AppRepoForm).find(".button");
    button.simulate("submit");
    wrapper.setProps({ error: new UnprocessableEntity("cannot process this!") });

    expect(wrapper.find(ErrorSelector).text()).toContain(
      "Something went wrong processing App Repository my-repo",
    );
    expect(wrapper.find(ErrorSelector).text()).toContain("cannot process this!");
  });
});
