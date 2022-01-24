// Copyright 2019-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { CdsModalContent } from "@cds/react/modal";
import LoadingWrapper from "components/LoadingWrapper";
import { mount } from "enzyme";
import ConfirmDialog from "./ConfirmDialog";

const defaultProps = {
  loading: false,
  modalIsOpen: true,
  confirmationText: "",
  onConfirm: jest.fn(),
  closeModal: jest.fn(),
};

it("should render a loading wrapper", () => {
  const wrapper = mount(<ConfirmDialog {...defaultProps} loading={true} />);
  expect(wrapper.find(LoadingWrapper).prop("loaded")).toBe(false);
});

it("should modify the default confirmation text", () => {
  const wrapper = mount(
    <ConfirmDialog {...defaultProps} confirmationText="Sure?" confirmationButtonText="Sure!" />,
  );
  expect(wrapper.find(CdsModalContent)).toIncludeText("Sure?");
  expect(wrapper.find(".btn").filterWhere(d => d.text() === "Sure!")).toExist();
});
