import React from "react";
import Modal, { modalSizes } from "./Modal";

import { mount } from "enzyme";

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
        title={"Title"}
        footer={<section>Footer content</section>}
      >
        <p>Body content</p>
      </Modal>,
    );

    expect(wrapper.find("section")).toHaveText("Footer content");
    expect(wrapper.find("h3")).toHaveText("Title");
    expect(wrapper.find("p")).toHaveText("Body content");
  });

  it("matches all modal sizes", () => {
    Object.keys(modalSizes).forEach(size => {
      const wrapper = mount(
        <Modal showModal={true} modalSize={size} onModalClose={mockCloseFunc}>
          <span>child</span>
        </Modal>,
      );

      const modalDialog = wrapper.find(".modal-dialog");
      expect(modalDialog).toHaveClassName(modalSizes[size]);
    });
  });

  it("hides close button when showCLoseButton set to false", () => {
    Object.keys(modalSizes).forEach(size => {
      const wrapper = mount(
        <Modal showModal={true} modalSize={size} onModalClose={mockCloseFunc} hideCloseButton>
          <span>child</span>
        </Modal>,
      );

      const closeButton = wrapper.find(".close");
      expect(closeButton).not.toExist();
    });
  });

  it("shows close button when showCLoseButton set to true", () => {
    Object.keys(modalSizes).forEach(size => {
      const wrapper = mount(
        <Modal
          showModal={true}
          showCloseButton={true}
          modalSize={size}
          onModalClose={mockCloseFunc}
        >
          <span>child</span>
        </Modal>,
      );

      const closeButton = wrapper.find(".close");
      expect(closeButton).toExist();
    });
  });

  it("close modal on close button click", () => {
    const wrapper = mount(
      <Modal showModal={true} onModalClose={mockCloseFunc}>
        <span>child</span>
      </Modal>,
    );

    const closeButton = wrapper.find(".close");

    closeButton.simulate("click");

    wrapper.update();

    expect(mockCloseFunc).toHaveBeenCalled();
  });

  it("close modal on backdrop click", () => {
    const wrapper = mount(
      <Modal showModal={true} staticBackdrop={false} onModalClose={mockCloseFunc}>
        <span>child</span>
      </Modal>,
    );

    const closeButton = wrapper.find(".modal-backdrop");

    closeButton.simulate("click");

    wrapper.update();

    expect(mockCloseFunc).toHaveBeenCalled();
  });

  it("does not close modal on backdrop if staticBackdrop is set to true", () => {
    const wrapper = mount(
      <Modal showModal={true} onModalClose={mockCloseFunc}>
        <span>child</span>
      </Modal>,
    );

    const closeButton = wrapper.find(".modal-backdrop");

    closeButton.simulate("click");

    wrapper.update();

    expect(mockCloseFunc).not.toHaveBeenCalled();
  });
});
