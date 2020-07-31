import actions from "actions";
import { CdsButton } from "components/Clarity/clarity";
import ConfirmDialog from "components/ConfirmDialog/ConfirmDialog.v2";
import Alert from "components/js/Alert";
import * as React from "react";
import { act } from "react-dom/test-utils";
import * as ReactRedux from "react-redux";
import { defaultStore, getStore, mountWrapper } from "shared/specs/mountWrapper";
import DeleteButton from "./DeleteButton.v2";

const defaultProps = {
  cluster: "default",
  namespace: "kubeapps",
  releaseName: "foo",
};

let spyOnUseDispatch: jest.SpyInstance;
const kubeaActions = { ...actions.kube };
beforeEach(() => {
  actions.apps = {
    ...actions.apps,
    deleteApp: jest.fn(),
  };
  const mockDispatch = jest.fn();
  spyOnUseDispatch = jest.spyOn(ReactRedux, "useDispatch").mockReturnValue(mockDispatch);
});

afterEach(() => {
  actions.kube = { ...kubeaActions };
  spyOnUseDispatch.mockRestore();
});

it("deletes an application", async () => {
  const deleteApp = jest.fn();
  actions.apps.deleteApp = deleteApp;
  const wrapper = mountWrapper(defaultStore, <DeleteButton {...defaultProps} />);
  act(() => {
    (wrapper.find(CdsButton).prop("onClick") as any)();
  });
  wrapper.update();
  expect(wrapper.find(ConfirmDialog).prop("modalIsOpen")).toBe(true);
  await act(async () => {
    await (wrapper
      .find(CdsButton)
      .filterWhere(b => b.text() === "Delete")
      .prop("onClick") as any)();
  });
  expect(deleteApp).toHaveBeenCalledWith(
    defaultProps.cluster,
    defaultProps.namespace,
    defaultProps.releaseName,
    true,
  );
});

it("renders an error", async () => {
  const store = getStore({ apps: { deleteError: new Error("Boom!") } });
  const wrapper = mountWrapper(store, <DeleteButton {...defaultProps} />);
  // Open modal
  act(() => {
    (wrapper.find(CdsButton).prop("onClick") as any)();
  });
  wrapper.update();

  expect(wrapper.find(Alert)).toIncludeText("Boom!");
});
