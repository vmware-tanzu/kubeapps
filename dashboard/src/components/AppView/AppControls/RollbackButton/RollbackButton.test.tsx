import { shallow, ShallowWrapper } from "enzyme";
import * as React from "react";
import * as ReactModal from "react-modal";
import RollbackButton from ".";
import RollbackDialog from "./RollbackDialog";

const defaultProps = {
  releaseName: "foo",
  namespace: "default",
  revision: 2,
  rollbackApp: jest.fn(),
  loading: false,
};

function openModal(wrapper: ShallowWrapper) {
  ReactModal.setAppElement(document.createElement("div"));
  wrapper.setState({ modalIsOpen: true });
  wrapper.update();
}

it("should perform the rollback", async () => {
  const rollbackApp = jest.fn();
  const wrapper = shallow(<RollbackButton {...defaultProps} rollbackApp={rollbackApp} />);
  openModal(wrapper);

  const dialog = wrapper.find(RollbackDialog);
  expect(dialog).toExist();
  const onConfirm = dialog.prop("onConfirm") as (revision: number) => Promise<any>;
  await onConfirm(1);
  expect(rollbackApp).toBeCalledWith(defaultProps.releaseName, defaultProps.namespace, 1);
});
