import { CdsButton } from "@cds/react/button";
import { CdsModal } from "@cds/react/modal";
import actions from "actions";
import Alert from "components/js/Alert";
import {
  InstalledPackageReference,
  InstalledPackageStatus,
  InstalledPackageStatus_StatusReason,
} from "gen/kubeappsapis/core/packages/v1alpha1/packages";
import { Plugin } from "gen/kubeappsapis/core/plugins/v1alpha1/plugins";
import { act } from "react-dom/test-utils";
import * as ReactRedux from "react-redux";
import ReactTooltip from "react-tooltip";
import { defaultStore, getStore, mountWrapper } from "shared/specs/mountWrapper";
import { RollbackError } from "shared/types";
import RollbackButton from "./RollbackButton";

const defaultProps = {
  installedPackageRef: {
    context: { cluster: "default", namespace: "kubeapps" },
    identifier: " foo",
    plugin: { name: "my.plugin", version: "0.0.1" },
  } as InstalledPackageReference,
  revision: 3,
  releaseStatus: null,
  plugin: { name: "my.plugin", version: "0.0.1" } as Plugin,
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
  expect(wrapper.find(CdsModal)).toExist();
  wrapper
    .find("select")
    .at(0)
    .simulate("change", { target: { value: "1" } });
  await act(async () => {
    await (
      wrapper
        .find(CdsButton)
        .filterWhere(b => b.text() === "Rollback")
        .prop("onClick") as any
    )();
  });
  expect(rollbackApp).toHaveBeenCalledWith(defaultProps.installedPackageRef, 1);
});

it("renders an error", async () => {
  const store = getStore({ apps: { error: new RollbackError("Boom!") } });
  const wrapper = mountWrapper(store, <RollbackButton {...defaultProps} />);
  // Open modal
  act(() => {
    (wrapper.find(CdsButton).prop("onClick") as any)();
  });
  wrapper.update();

  expect(wrapper.find(Alert)).toIncludeText("Boom!");
});

it("should render a disabled button if when passing an in-progress status", async () => {
  const disabledProps = {
    ...defaultProps,
    releaseStatus: {
      ready: false,
      reason: InstalledPackageStatus_StatusReason.STATUS_REASON_PENDING,
      userReason: "Pending",
    } as InstalledPackageStatus,
  };
  const wrapper = mountWrapper(defaultStore, <RollbackButton {...disabledProps} />);

  expect(wrapper.find(CdsButton)).toBeDisabled();
  expect(wrapper.find(ReactTooltip)).toExist();
});
