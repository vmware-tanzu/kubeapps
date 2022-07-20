// Copyright 2020-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { CdsButton } from "@cds/react/button";
import actions from "actions";
import ConfirmDialog from "components/ConfirmDialog/ConfirmDialog";
import Alert from "components/js/Alert";
import {
  InstalledPackageReference,
  InstalledPackageStatus,
  InstalledPackageStatus_StatusReason,
} from "gen/kubeappsapis/core/packages/v1alpha1/packages";
import { act } from "react-dom/test-utils";
import * as ReactRedux from "react-redux";
import ReactTooltip from "react-tooltip";
import { defaultStore, getStore, mountWrapper } from "shared/specs/mountWrapper";
import { StartError } from "shared/types";
import StopButton from "./StopButton";

const defaultProps = {
  installedPackageRef: {
    context: { cluster: "default", namespace: "kubeapps" },
    identifier: " foo",
    plugin: { name: "my.plugin", version: "0.0.1" },
  } as InstalledPackageReference,
  releaseStatus: null,
};

let spyOnUseDispatch: jest.SpyInstance;
const kubeaActions = { ...actions.kube };
beforeEach(() => {
  actions.installedpackages = {
    ...actions.installedpackages,
    stopInstalledPackage: jest.fn(),
  };
  const mockDispatch = jest.fn();
  spyOnUseDispatch = jest.spyOn(ReactRedux, "useDispatch").mockReturnValue(mockDispatch);
});

afterEach(() => {
  actions.kube = { ...kubeaActions };
  spyOnUseDispatch.mockRestore();
});

it("stops an application", async () => {
  const stopInstalledPackage = jest.fn();
  actions.installedpackages.stopInstalledPackage = stopInstalledPackage;
  const wrapper = mountWrapper(defaultStore, <StopButton {...defaultProps} />);
  act(() => {
    (wrapper.find(CdsButton).prop("onClick") as any)();
  });
  wrapper.update();
  expect(wrapper.find(ConfirmDialog).prop("modalIsOpen")).toBe(true);
  await act(async () => {
    await (
      wrapper
        .find(".btn")
        .filterWhere(b => b.text() === "Stop")
        .prop("onClick") as any
    )();
  });
  expect(stopInstalledPackage).toHaveBeenCalledWith(defaultProps.installedPackageRef);
});

it("renders an error", async () => {
  const store = getStore({ apps: { error: new StartError("Boom!") } });
  const wrapper = mountWrapper(store, <StopButton {...defaultProps} />);
  // Open modal
  act(() => {
    (wrapper.find(CdsButton).prop("onClick") as any)();
  });
  wrapper.update();

  expect(wrapper.find(Alert)).toIncludeText("Boom!");
});

it("should render a deactivated button if when passing a stopped status", async () => {
  const disabledProps = {
    ...defaultProps,
    releaseStatus: {
      ready: false,
      reason: InstalledPackageStatus_StatusReason.STATUS_REASON_STOPPED,
      userReason: "Stopped",
    } as InstalledPackageStatus,
  };
  const wrapper = mountWrapper(defaultStore, <StopButton {...disabledProps} />);

  expect(wrapper.find(CdsButton)).toBeDisabled();
  expect(wrapper.find(ReactTooltip)).toExist();
});
