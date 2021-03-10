import { CdsModalContent } from "@cds/react/modal";
import { mount } from "enzyme";
import context from "jest-plugin-context";
import itBehavesLike from "../../shared/specs";
import ConfirmDialog from "./ConfirmDialog";

const defaultProps = {
  loading: false,
  modalIsOpen: true,
  onConfirm: jest.fn(),
  closeModal: jest.fn(),
};

context("when loading is true", () => {
  itBehavesLike("aLoadingComponent", {
    component: ConfirmDialog,
    props: { ...defaultProps, loading: true },
  });
});

it("should modify the default confirmation text", () => {
  const wrapper = mount(
    <ConfirmDialog {...defaultProps} confirmationText="Sure?" confirmationButtonText="Sure!" />,
  );
  expect(wrapper.find(CdsModalContent)).toIncludeText("Sure?");
  expect(wrapper.find(".btn").filterWhere(d => d.text() === "Sure!")).toExist();
});
