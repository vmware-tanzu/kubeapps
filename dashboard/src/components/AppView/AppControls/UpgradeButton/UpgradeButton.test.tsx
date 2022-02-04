// Copyright 2021-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { CdsButton } from "@cds/react/button";
import actions from "actions";
import {
  InstalledPackageReference,
  InstalledPackageStatus,
  InstalledPackageStatus_StatusReason,
} from "gen/kubeappsapis/core/packages/v1alpha1/packages";
import * as ReactRedux from "react-redux";
import ReactTooltip from "react-tooltip";
import { defaultStore, mountWrapper } from "shared/specs/mountWrapper";
import UpgradeButton from "./UpgradeButton";

const defaultProps = {
  installedPackageRef: {
    context: { cluster: "default", namespace: "kubeapps" },
    identifier: "foo",
    plugin: { name: "my.plugin", version: "0.0.1" },
  } as InstalledPackageReference,
  releaseStatus: null,
};

let spyOnUseDispatch: jest.SpyInstance;
const kubeaActions = { ...actions.kube };
beforeEach(() => {
  actions.installedpackages = {
    ...actions.installedpackages,
    updateInstalledPackage: jest.fn(),
  };
  const mockDispatch = jest.fn();
  spyOnUseDispatch = jest.spyOn(ReactRedux, "useDispatch").mockReturnValue(mockDispatch);
});

afterEach(() => {
  actions.kube = { ...kubeaActions };
  spyOnUseDispatch.mockRestore();
});

it("should render a deactivated button if when passing an in-progress status", async () => {
  const disabledProps = {
    ...defaultProps,
    releaseStatus: {
      ready: false,
      reason: InstalledPackageStatus_StatusReason.STATUS_REASON_PENDING,
      userReason: "Pending",
    } as InstalledPackageStatus,
  };
  const wrapper = mountWrapper(defaultStore, <UpgradeButton {...disabledProps} />);

  expect(wrapper.find(CdsButton)).toBeDisabled();
  expect(wrapper.find(ReactTooltip)).toExist();
});
