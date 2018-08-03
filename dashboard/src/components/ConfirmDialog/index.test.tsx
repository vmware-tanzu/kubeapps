import { mount } from "enzyme";
import * as React from "react";
import * as Modal from "react-modal";

import ConfirmDialog from "./index";

it("renders loading message", () => {
  const wrapper = mount(
    <ConfirmDialog
      modalIsOpen={true}
      loading={true}
      onConfirm={jest.fn()}
      closeModal={jest.fn()}
    />,
  );
  expect(wrapper.find(Modal).text()).toContain("Loading ...");
});

it("perform action on confirm", () => {
  const spyOnConfirm = jest.fn();
  const wrapper = mount(
    <ConfirmDialog
      modalIsOpen={true}
      loading={false}
      onConfirm={spyOnConfirm}
      closeModal={jest.fn()}
    />,
  );
  expect(wrapper.find(Modal).text()).toContain("Are you sure");
  const confirmButton = wrapper.find("#delete");
  expect(confirmButton.exists()).toBe(true);
  confirmButton.simulate("click");
  expect(spyOnConfirm).toHaveBeenCalled();
});

it("closes modal", () => {
  const spyCloseModal = jest.fn();
  const wrapper = mount(
    <ConfirmDialog
      modalIsOpen={true}
      loading={false}
      onConfirm={jest.fn()}
      closeModal={spyCloseModal}
    />,
  );
  expect(wrapper.find(Modal).text()).toContain("Are you sure");
  const confirmButton = wrapper.find("#cancel");
  expect(confirmButton.exists()).toBe(true);
  confirmButton.simulate("click");
  expect(spyCloseModal).toHaveBeenCalled();
});
