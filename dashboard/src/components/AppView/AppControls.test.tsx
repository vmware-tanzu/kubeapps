import { shallow } from "enzyme";
import * as React from "react";
import { Redirect } from "react-router";
import { hapi } from "../../shared/hapi/release";
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

  expect(wrapper.state().deleting).toBe(true);
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
