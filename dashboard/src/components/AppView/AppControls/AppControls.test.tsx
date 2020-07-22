import { mount, shallow } from "enzyme";
import context from "jest-plugin-context";
import * as React from "react";
import Modal from "react-modal";

import { getStore, mountWrapper } from "shared/specs/mountWrapper";
import { IRelease } from "shared/types";
import RollbackButtonContainer from "../../../containers/RollbackButtonContainer";
import { hapi } from "../../../shared/hapi/release";
import itBehavesLike from "../../../shared/specs";
import * as url from "../../../shared/url";
import ConfirmDialog from "../../ConfirmDialog";
import AppControls, { IAppControlsProps } from "./AppControls";
import UpgradeButton from "./UpgradeButton";

const namespace = "bar";
const defaultProps = {
  cluster: "default",
  app: new hapi.release.Release({ name: "foo", namespace }),
  deleteApp: jest.fn(),
  push: jest.fn(),
} as IAppControlsProps;

it("calls delete function without purge when clicking the button", done => {
  const store = getStore({});
  const push = jest.fn();
  const deleteApp = jest.fn().mockReturnValue(true);
  const props = {
    ...defaultProps,
    deleteApp,
    push,
  };
  const wrapper = mountWrapper(store, <AppControls {...props} />);
  const appControls = wrapper.find(AppControls);
  const button = appControls.children().find(".button-danger");
  expect(button.exists()).toBe(true);
  expect(button.text()).toBe("Delete");
  button.simulate("click");

  const confirm = appControls.children().find(ConfirmDialog);
  expect(confirm.exists()).toBe(true);
  confirm.props().onConfirm(); // Simulate confirmation

  expect(appControls.state("deleting")).toBe(true);

  // Wait for the async action to finish
  setTimeout(() => {
    wrapper.update();
    expect(push.mock.calls.length).toBe(1);
    expect(push.mock.calls[0]).toEqual([url.app.apps.list(defaultProps.cluster, namespace)]);
    done();
  }, 1);
  expect(deleteApp).toHaveBeenCalledWith(false);
});

it("calls delete function with additional purge", () => {
  // Return "false" to avoid redirect when mounting
  const deleteApp = jest.fn().mockReturnValue(false);
  const props = { ...defaultProps, deleteApp };
  const store = getStore({});
  const wrapper = mountWrapper(store, <AppControls {...props} />);
  Modal.setAppElement(document.createElement("div"));
  const appControls = wrapper.find(AppControls);
  const button = appControls.children().find(".button-danger");
  expect(button.exists()).toBe(true);
  expect(button.text()).toBe("Delete");
  button.simulate("click");

  // Check that the checkbox changes the AppControls state
  const confirm = wrapper.find(ConfirmDialog);
  expect(confirm.exists()).toBe(true);
  const checkbox = wrapper.find('input[type="checkbox"]');
  expect(checkbox.exists()).toBe(true);
  expect(appControls.state("purge")).toBe(false);
  checkbox.simulate("change");
  expect(appControls.state("purge")).toBe(true);

  // Check that the "purge" state is forwarded to deleteApp
  confirm.props().onConfirm(); // Simulate confirmation
  expect(deleteApp).toHaveBeenCalledWith(true);
});

context("when name or namespace do not exist", () => {
  const props = {
    ...defaultProps,
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
    ...defaultProps,
    app: new hapi.release.Release({ name: "name", namespace: "my-ns", info: { deleted: {} } }),
    deleteApp: jest.fn().mockReturnValue(false), // Return "false" to avoid redirect when mounting
  };

  it("should show Purge instead of Delete in the button title", () => {
    const wrapper = shallow(<AppControls {...props} />);
    const button = wrapper.find(".button-danger");
    expect(button.text()).toBe("Purge");
  });

  it("should not show the purge checkbox", () => {
    // mount() is necessary to render the Modal
    const wrapper = mount(<AppControls {...props} push={jest.fn()} />);
    Modal.setAppElement(document.createElement("div"));
    wrapper.setState({ modalIsOpen: true });
    wrapper.update();

    const confirm = wrapper.find(ConfirmDialog);
    expect(confirm.exists()).toBe(true);
    const checkbox = wrapper.find('input[type="checkbox"]');
    expect(checkbox).not.toExist();
  });

  it("should purge when clicking on delete", () => {
    // mount() is necessary to render the Modal
    const deleteApp = jest.fn().mockReturnValue(false);
    const wrapper = mount(<AppControls {...props} deleteApp={deleteApp} />);
    Modal.setAppElement(document.createElement("div"));
    wrapper.setState({ modalIsOpen: true, purge: false });
    wrapper.update();

    const confirm = wrapper.find(ConfirmDialog);
    expect(confirm.exists()).toBe(true);

    // Check that the "purge" is forwarded to deleteApp
    confirm.props().onConfirm(); // Simulate confirmation
    expect(deleteApp).toHaveBeenCalledWith(true);
  });

  it("should not show the Upgrade button", () => {
    const deleteApp = jest.fn().mockReturnValue(false);
    const wrapper = shallow(<AppControls {...props} deleteApp={deleteApp} push={jest.fn()} />);
    const buttons = wrapper.find("button");
    expect(buttons.length).toBe(1);
    expect(buttons.text()).toBe("Purge");
  });
});

context("when there is a new version available", () => {
  it("should forward the latest version", () => {
    const name = "foo";
    const app = {
      name,
      namespace,
      updateInfo: {
        upToDate: false,
        chartLatestVersion: "1.0.0",
        appLatestVersion: "1.0.0",
      },
    } as IRelease;
    const props = { ...defaultProps, app };
    const wrapper = shallow(<AppControls {...props} />);

    expect(wrapper.find(UpgradeButton).prop("newVersion")).toBe(true);
  });
});

context("when the application is up to date", () => {
  it("should not forward the latest version", () => {
    const name = "foo";
    const app = {
      name,
      namespace,
      updateInfo: {
        upToDate: true,
        chartLatestVersion: "1.1.0",
        appLatestVersion: "1.1.0",
      },
    } as IRelease;
    const props = { ...defaultProps, app };
    const wrapper = shallow(<AppControls {...props} />);

    expect(wrapper.find(UpgradeButton).prop("updateVersion")).toBe(undefined);
  });
});

context("Rollback button", () => {
  it("should show the RollbackButton when there is more than one revision", () => {
    const props = {
      ...defaultProps,
      app: new hapi.release.Release({
        name: "name",
        namespace: "my-ns",
        version: 2,
        info: {},
      }),
      deleteApp: jest.fn().mockReturnValue(false), // Return "false" to avoid redirect when mounting
    };
    const wrapper = shallow(<AppControls {...props} />);
    const button = wrapper.find(RollbackButtonContainer);
    expect(button).toExist();
  });
  it("should not show the RollbackButton when there is only one revision", () => {
    const props = {
      ...defaultProps,
      app: new hapi.release.Release({
        name: "name",
        namespace: "my-ns",
        version: 1,
        info: {},
      }),
      deleteApp: jest.fn().mockReturnValue(false), // Return "false" to avoid redirect when mounting
    };
    const wrapper = shallow(<AppControls {...props} push={jest.fn()} />);
    const button = wrapper.find(RollbackButtonContainer);
    expect(button).not.toExist();
  });
});
