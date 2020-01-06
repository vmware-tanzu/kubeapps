import { shallow } from "enzyme";
import context from "jest-plugin-context";
import * as React from "react";
import itBehavesLike from "../../shared/specs";
import ConfirmDialog from "./ConfirmDialog";

context("when loading is true", () => {
  const props = {
    loading: true,
  };

  itBehavesLike("aLoadingComponent", { component: ConfirmDialog, props });
});

const defaultProps = {
  loading: false,
  modalIsOpen: true,
  onConfirm: jest.fn(),
  closeModal: jest.fn(),
};

it("should modify the default confirmation text", () => {
  const wrapper = shallow(
    <ConfirmDialog {...defaultProps} confirmationText="Sure?" confirmationButtonText="Sure!" />,
  );
  expect(wrapper.find(".margin-b-normal").filterWhere(d => d.text() === "Sure?")).toExist();
  expect(wrapper.find("button").filterWhere(d => d.text() === "Sure!")).toExist();
});
