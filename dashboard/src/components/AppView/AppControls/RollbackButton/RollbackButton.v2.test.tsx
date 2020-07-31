import actions from "actions";
import { CdsButton } from "components/Clarity/clarity";
import Alert from "components/js/Alert";
import Modal from "components/js/Modal/Modal";
import * as React from "react";
import { act } from "react-dom/test-utils";
import * as ReactRedux from "react-redux";
import { defaultStore, getStore, mountWrapper } from "shared/specs/mountWrapper";
import RollbackButton from "./RollbackButton.v2";

const defaultProps = {
  cluster: "default",
  namespace: "kubeapps",
  releaseName: "foo",
  revision: 2,
};

let spyOnUseDispatch: jest.SpyInstance;
const kubeaActions = { ...actions.kube };
beforeEach(() => {
  actions.apps = {
    ...actions.apps,
    rollbackApp: jest.fn(),
  };
  const mockDispatch = jest.fn();
  spyOnUseDispatch = jest.spyOn(ReactRedux, "useDispatch").mockReturnValue(mockDispatch);
});

afterEach(() => {
  actions.kube = { ...kubeaActions };
  spyOnUseDispatch.mockRestore();
});

it("rolls back an application", async () => {
  const rollbackApp = jest.fn();
  actions.apps.rollbackApp = rollbackApp;
  const wrapper = mountWrapper(defaultStore, <RollbackButton {...defaultProps} />);
  act(() => {
    (wrapper.find(CdsButton).prop("onClick") as any)();
  });
  wrapper.update();
  expect(wrapper.find(Modal).prop("showModal")).toBe(true);
  await act(async () => {
    await (wrapper
      .find(CdsButton)
      .filterWhere(b => b.text() === "Rollback")
      .prop("onClick") as any)();
  });
  expect(rollbackApp).toHaveBeenCalledWith(
    defaultProps.cluster,
    defaultProps.namespace,
    defaultProps.releaseName,
    1,
  );
});

it("renders an error", async () => {
  const store = getStore({ apps: { error: new Error("Boom!") } });
  const wrapper = mountWrapper(store, <RollbackButton {...defaultProps} />);
  // Open modal
  act(() => {
    (wrapper.find(CdsButton).prop("onClick") as any)();
  });
  wrapper.update();

  expect(wrapper.find(Alert)).toIncludeText("Boom!");
});
