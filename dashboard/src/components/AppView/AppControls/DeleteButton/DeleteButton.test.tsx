// Copyright 2020-2023 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { CdsButton } from "@cds/react/button";
import { act, waitFor } from "@testing-library/react";
import actions from "actions";
import AlertGroup from "components/AlertGroup";
import ConfirmDialog from "components/ConfirmDialog";
import {
  InstalledPackageReference,
  InstalledPackageStatus,
  InstalledPackageStatus_StatusReason,
} from "gen/kubeappsapis/core/packages/v1alpha1/packages_pb";
import * as ReactRedux from "react-redux";
import { Tooltip } from "react-tooltip";
import { defaultStore, getStore, mountWrapper } from "shared/specs/mountWrapper";
import { DeleteError, IInstalledPackageState, IStoreState } from "shared/types";
import DeleteButton from "./DeleteButton";

const defaultProps = {
  installedPackageRef: {
    context: { cluster: "default", namespace: "kubeapps" },
    identifier: " foo",
    plugin: { name: "my.plugin", version: "0.0.1" },
  } as InstalledPackageReference,
  releaseStatus: null,
};

let spyOnUseDispatch: jest.SpyInstance;
const kubeActions = { ...actions.kube };
beforeEach(() => {
  actions.installedpackages = {
    ...actions.installedpackages,
    deleteInstalledPackage: jest.fn(),
  };
  const mockDispatch = jest.fn();
  spyOnUseDispatch = jest.spyOn(ReactRedux, "useDispatch").mockReturnValue(mockDispatch);
});

afterEach(() => {
  actions.kube = { ...kubeActions };
  spyOnUseDispatch.mockRestore();
});

it("deletes an application", async () => {
  const deleteInstalledPackage = jest.fn();
  actions.installedpackages.deleteInstalledPackage = deleteInstalledPackage;
  const wrapper = mountWrapper(defaultStore, <DeleteButton {...defaultProps} />);
  act(() => {
    (wrapper.find(CdsButton).prop("onClick") as any)();
  });
  wrapper.update();
  expect(wrapper.find(ConfirmDialog).prop("modalIsOpen")).toBe(true);
  await act(async () => {
    await (
      wrapper
        .find(CdsButton)
        .filterWhere(b => b.text() === "Delete")
        .prop("onClick") as any
    )();
  });
  expect(deleteInstalledPackage).toHaveBeenCalledWith(defaultProps.installedPackageRef);
});

it("renders an error", async () => {
  const store = getStore({
    apps: { error: new DeleteError("Boom!") } as Partial<IInstalledPackageState>,
  } as Partial<IStoreState>);
  const wrapper = mountWrapper(store, <DeleteButton {...defaultProps} />);
  // Open modal
  act(() => {
    (wrapper.find(CdsButton).prop("onClick") as any)();
  });
  wrapper.update();

  expect(wrapper.find(AlertGroup)).toIncludeText("Boom!");
});

it("should render an enabled button and tooltip if when passing a pending status", async () => {
  const disabledProps = {
    ...defaultProps,
    releaseStatus: {
      ready: false,
      reason: InstalledPackageStatus_StatusReason.PENDING,
      userReason: "Pending",
    } as InstalledPackageStatus,
  };
  const wrapper = mountWrapper(defaultStore, <DeleteButton {...disabledProps} />);

  expect(wrapper.find(CdsButton)).not.toBeDisabled();

  await waitFor(() => {
    expect(wrapper.find(Tooltip).prop("children")).toBe("The application is pending installation.");
  });
});

it("should render a deactivated button if when passing a uninstalled status", async () => {
  const disabledProps = {
    ...defaultProps,
    releaseStatus: {
      ready: false,
      reason: InstalledPackageStatus_StatusReason.UNINSTALLED,
      userReason: "Uninstalling",
    } as InstalledPackageStatus,
  };
  const wrapper = mountWrapper(defaultStore, <DeleteButton {...disabledProps} />);

  expect(wrapper.find(CdsButton)).toBeDisabled();

  await waitFor(() => {
    expect(wrapper.find(Tooltip).prop("children")).toBe("The application is being deleted.");
  });
});
