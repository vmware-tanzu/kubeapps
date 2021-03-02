import { CdsButton } from "@cds/react/button";
import { CdsModal, CdsModalActions, CdsModalContent, CdsModalHeader } from "@cds/react/modal";
import { mount } from "enzyme";
import * as React from "react";
import Modal from "./Modal";

const modalSizes = ["sm", "default", "lg", "xl"];

const mockCloseFunc = jest.fn();

describe(Modal, () => {
  beforeEach(() => jest.resetAllMocks());

  it("shows modal when showModal is set to true", () => {
    const wrapper = mount(
      <Modal showModal={true} onModalClose={mockCloseFunc}>
        <span>child</span>
      </Modal>,
    );
    expect(wrapper).toExist();
  });
  it("hides modal when showModal is set to false", () => {
    const wrapper = mount(
      <Modal showModal={false} onModalClose={mockCloseFunc}>
        <span>child</span>
      </Modal>,
    );
    expect(wrapper.isEmptyRender()).toBeTruthy();
  });
  it("projects title, body and footer", () => {
    const wrapper = mount(
      <Modal
        showModal={true}
        onModalClose={mockCloseFunc}
        hideCloseButton={true}
        title={"Title"}
        footer={<section>Footer content</section>}
      >
        <p>Body content</p>
      </Modal>,
    );
    expect(wrapper.find(CdsModalHeader)).toHaveText("Title");
    expect(wrapper.find(CdsModalContent)).toHaveText("Body content");
    expect(wrapper.find(CdsModalActions)).toHaveText("Footer content");
  });
  it("matches all modal sizes", () => {
    Object.values(modalSizes).forEach(size => {
      const wrapper = mount(
        <Modal
          showModal={true}
          modalSize={size as "sm" | "default" | "lg" | "xl"}
          onModalClose={mockCloseFunc}
        >
          <span>child</span>
        </Modal>,
      );
      const modalDialog = wrapper.find(CdsModal);
      expect(modalDialog.prop("size")).toBe(size);
    });
  });
  it("hides close button when hideCloseButton set to true", () => {
    Object.keys(modalSizes).forEach(size => {
      const wrapper = mount(
        <Modal
          showModal={true}
          modalSize={size as "sm" | "default" | "lg" | "xl"}
          onModalClose={mockCloseFunc}
          hideCloseButton={true}
        >
          <span>child</span>
        </Modal>,
      );
      const closeButton = wrapper.find(CdsButton);
      expect(closeButton).not.toExist();
    });
  });
  it("shows close button when hideCloseButton set to false", () => {
    Object.keys(modalSizes).forEach(size => {
      const wrapper = mount(
        <Modal
          showModal={true}
          modalSize={size as "sm" | "default" | "lg" | "xl"}
          onModalClose={mockCloseFunc}
          hideCloseButton={false}
        >
          <span>child</span>
        </Modal>,
      );
      const closeButton = wrapper.find(CdsButton);
      expect(closeButton).toExist();
    });
  });
  it("close modal on close button click", () => {
    const wrapper = mount(
      <Modal showModal={true} onModalClose={mockCloseFunc}>
        <span>child</span>
      </Modal>,
    );
    const closeButton = wrapper.find(CdsButton);
    closeButton.simulate("click");
    wrapper.update();
    expect(mockCloseFunc).toHaveBeenCalled();
  });
});
