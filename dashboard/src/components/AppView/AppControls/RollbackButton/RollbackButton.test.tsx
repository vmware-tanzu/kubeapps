// Copyright 2019-2023 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { CdsButton } from "@cds/react/button";
import { CdsModal } from "@cds/react/modal";
import { act, waitFor } from "@testing-library/react";
import actions from "actions";
import AlertGroup from "components/AlertGroup";
import {
  InstalledPackageReference,
  InstalledPackageStatus,
  InstalledPackageStatus_StatusReason,
} from "gen/kubeappsapis/core/packages/v1alpha1/packages_pb";
import { Plugin } from "gen/kubeappsapis/core/plugins/v1alpha1/plugins_pb";
import * as ReactRedux from "react-redux";
import { Tooltip } from "react-tooltip";
import { defaultStore, getStore, mountWrapper } from "shared/specs/mountWrapper";
import { IInstalledPackageState, RollbackError } from "shared/types";
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
const kubeActions = { ...actions.kube };
beforeEach(() => {
  actions.installedpackages = {
    ...actions.installedpackages,
    rollbackInstalledPackage: jest.fn(),
  };
  const mockDispatch = jest.fn();
  spyOnUseDispatch = jest.spyOn(ReactRedux, "useDispatch").mockReturnValue(mockDispatch);
});

afterEach(() => {
  actions.kube = { ...kubeActions };
  spyOnUseDispatch.mockRestore();
});

it("rolls back an application", async () => {
  const rollbackInstalledPackage = jest.fn();
  actions.installedpackages.rollbackInstalledPackage = rollbackInstalledPackage;
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
  expect(rollbackInstalledPackage).toHaveBeenCalledWith(defaultProps.installedPackageRef, 1);
});

it("renders an error", async () => {
  const store = getStore({
    apps: { error: new RollbackError("Boom!") },
  } as Partial<IInstalledPackageState>);
  const wrapper = mountWrapper(store, <RollbackButton {...defaultProps} />);
  // Open modal
  act(() => {
    (wrapper.find(CdsButton).prop("onClick") as any)();
  });
  wrapper.update();

  expect(wrapper.find(AlertGroup)).toIncludeText("Boom!");
});

it("should render a deactivated button if when passing an in-progress status", async () => {
  const disabledProps = {
    ...defaultProps,
    releaseStatus: {
      ready: false,
      reason: InstalledPackageStatus_StatusReason.PENDING,
      userReason: "Pending",
    } as InstalledPackageStatus,
  };
  const wrapper = mountWrapper(defaultStore, <RollbackButton {...disabledProps} />);

  expect(wrapper.find(CdsButton)).toBeDisabled();

  await waitFor(() => {
    expect(wrapper.find(Tooltip).prop("children")).toBe("The application is pending installation.");
  });
});
