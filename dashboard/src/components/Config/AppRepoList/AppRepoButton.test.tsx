import { mount } from "enzyme";
import * as React from "react";
import * as ReactModal from "react-modal";
import { ConflictError, UnprocessableEntity } from "../../../shared/types";
import ErrorSelector from "../../ErrorAlert/ErrorSelector";
import { AppRepoAddButton, AppRepoForm } from "./AppRepoButton";

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

it("should install a repository", done => {
  const install = jest.fn(() => true);
  const wrapper = mount(<AppRepoAddButton {...defaultProps} install={install} />);
  ReactModal.setAppElement(document.createElement("div"));
  wrapper.setState({
    modalIsOpen: true,
    name: "my-repo",
    url: "http://foo.bar",
    authHeader: "foo",
    customCA: "bar",
  });
  wrapper.update();

  const button = wrapper.find(AppRepoForm).find(".button");
  button.simulate("submit");

  expect(install).toBeCalledWith("my-repo", "http://foo.bar", "foo", "bar");
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
    wrapper.setState({ modalIsOpen: true, name: "my-repo" });
    wrapper.update();

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
    wrapper.setState({ modalIsOpen: true, name: "my-repo" });
    wrapper.update();

    const button = wrapper.find(AppRepoForm).find(".button");
    button.simulate("submit");
    wrapper.setProps({ error: new UnprocessableEntity("cannot process this!") });

    expect(wrapper.find(ErrorSelector).text()).toContain(
      "Something went wrong processing App Repository my-repo",
    );
    expect(wrapper.find(ErrorSelector).text()).toContain("cannot process this!");
  });
});
