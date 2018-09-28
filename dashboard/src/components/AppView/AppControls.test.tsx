import { mount, shallow } from "enzyme";
import context from "jest-plugin-context";
import * as React from "react";
import * as ReactModal from "react-modal";
import { Redirect } from "react-router";
import { hapi } from "../../shared/hapi/release";
import itBehavesLike from "../../shared/specs";
import ConfirmDialog from "../ConfirmDialog";

import AppControls from "./AppControls";

it("renders a redirect when clicking upgrade", () => {
  const name = "foo";
  const namespace = "bar";
  const app = new hapi.release.Release({ name, namespace });
  const wrapper = shallow(<AppControls app={app} deleteApp={jest.fn()} />);
  const button = wrapper
    .find(".AppControls")
    .children()
    .find(".button")
    .filterWhere(i => i.text() === "Upgrade");
  expect(button.exists()).toBe(true);
  expect(wrapper.find(Redirect).exists()).toBe(false);

  button.simulate("click");
  const redirect = wrapper.find(Redirect);
  expect(redirect.exists()).toBe(true);
  expect(redirect.props()).toMatchObject({
    push: true,
    to: `/apps/ns/${namespace}/upgrade/${name}`,
  });
});

it("calls delete function when clicking the button", done => {
  const name = "foo";
  const namespace = "bar";
  const app = new hapi.release.Release({ name, namespace });
  const wrapper = shallow(<AppControls app={app} deleteApp={jest.fn(() => true)} />);
  const button = wrapper
    .find(".AppControls")
    .children()
    .find(".button-danger");
  expect(button.exists()).toBe(true);
  expect(button.text()).toBe("Delete");
  button.simulate("click");

  const confirm = wrapper
    .find(".AppControls")
    .children()
    .find(ConfirmDialog);
  expect(confirm.exists()).toBe(true);
  confirm.props().onConfirm(); // Simulate confirmation

  expect(wrapper.state("deleting")).toBe(true);
  // Wait for the async action to finish
  setTimeout(() => {
    wrapper.update();
    const redirect = wrapper
      .find(".AppControls")
      .children()
      .find(Redirect);
    expect(redirect.exists()).toBe(true);
    expect(redirect.props()).toMatchObject({
      push: false,
      to: `/apps/ns/${namespace}`,
    });
    done();
  }, 1);
});

it("calls delete function with additional purge", () => {
  const name = "foo";
  const namespace = "bar";
  const app = new hapi.release.Release({ name, namespace });
  const deleteApp = jest.fn(() => false); // Return "false" to avoid redirect when mounting
  // mount() is necessary to render the Modal
  const wrapper = mount(<AppControls app={app} deleteApp={deleteApp} />);
  ReactModal.setAppElement(document.createElement("div"));
  wrapper.setState({ modalIsOpen: true });
  wrapper.update();

  // Check that the checkbox changes the AppControls state
  const confirm = wrapper.find(ConfirmDialog);
  expect(confirm.exists()).toBe(true);
  const checkbox = wrapper.find('input[type="checkbox"]');
  expect(checkbox.exists()).toBe(true);
  expect(wrapper.state("purge")).toBe(false);
  checkbox.simulate("change");
  expect(wrapper.state("purge")).toBe(true);

  // Check that the "purge" state is forwarded to deleteApp
  confirm.props().onConfirm(); // Simulate confirmation
  expect(deleteApp).toHaveBeenCalledWith(true);
});

context("when name or namespace do not exist", () => {
  const props = {
    app: new hapi.release.Release({ name: "name", namespace: "my-ns" }),
  };

  itBehavesLike("aLoadingComponent", {
    component: AppControls,
    props: { ...props, app: { ...props.app, name: null } },
  });
  itBehavesLike("aLoadingComponent", {
    component: AppControls,
    props: { ...props, app: { ...props.app, namespace: null } },
  });
});

context("when the application has been already deleted", () => {
  const props = {
    app: new hapi.release.Release({ name: "name", namespace: "my-ns", info: { deleted: {} } }),
    deleteApp: jest.fn(() => false), // Return "false" to avoid redirect when mounting
  };

  it("should show Purge instead of Delete in the button title", () => {
    const wrapper = shallow(<AppControls {...props} />);
    const button = wrapper.find(".button-danger");
    expect(button.text()).toBe("Purge");
  });

  it("should not show the purge checkbox", () => {
    // mount() is necessary to render the Modal
    const wrapper = mount(<AppControls {...props} />);
    ReactModal.setAppElement(document.createElement("div"));
    wrapper.setState({ modalIsOpen: true });
    wrapper.update();

    const confirm = wrapper.find(ConfirmDialog);
    expect(confirm.exists()).toBe(true);
    const checkbox = wrapper.find('input[type="checkbox"]');
    expect(checkbox).not.toExist();
  });

  it("should purge when clicking on delete", () => {
    // mount() is necessary to render the Modal
    const deleteApp = jest.fn(() => false);
    const wrapper = mount(<AppControls {...props} deleteApp={deleteApp} />);
    ReactModal.setAppElement(document.createElement("div"));
    wrapper.setState({ modalIsOpen: true, purge: false });
    wrapper.update();

    const confirm = wrapper.find(ConfirmDialog);
    expect(confirm.exists()).toBe(true);

    // Check that the "purge" is forwarded to deleteApp
    confirm.props().onConfirm(); // Simulate confirmation
    expect(deleteApp).toHaveBeenCalledWith(true);
  });
});
