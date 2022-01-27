// Copyright 2021-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { CdsButton } from "@cds/react/button";
import {
  InstalledPackageStatus,
  InstalledPackageStatus_StatusReason,
} from "gen/kubeappsapis/core/packages/v1alpha1/packages";
import ReactTooltip from "react-tooltip";
import { defaultStore, mountWrapper } from "shared/specs/mountWrapper";
import StatusAwareButton from "./StatusAwareButton";

it("tests the disabled flag and tooltip for each release status condition", async () => {
  type TProps = {
    code: InstalledPackageStatus_StatusReason | null | undefined;
    disabled: boolean;
    tooltip?: string;
  };

  // this should cover all conditions
  const testsProps: TProps[] = [
    {
      code: InstalledPackageStatus_StatusReason.STATUS_REASON_PENDING,
      disabled: true,
      tooltip: "The application is pending installation.",
    },
    {
      code: InstalledPackageStatus_StatusReason.STATUS_REASON_UNINSTALLED,
      disabled: true,
      tooltip: "The application is being deleted.",
    },
    { code: undefined, disabled: true, tooltip: undefined },
    { code: null, disabled: true, tooltip: undefined },
  ];

  for (const testProps of testsProps) {
    let releaseStatus;
    switch (testProps.code) {
      case null:
        releaseStatus = null;
        break;
      case undefined:
        releaseStatus = undefined;
        break;
      default:
        releaseStatus = {
          reason: testProps.code,
        } as InstalledPackageStatus;
    }
    const disabled = testProps.disabled;
    const tooltip = testProps.tooltip;
    const wrapper = mountWrapper(
      defaultStore,
      <StatusAwareButton id="test" releaseStatus={releaseStatus} />,
    );

    // test disabled flag
    expect(wrapper.find(CdsButton).prop("disabled")).toBe(disabled);

    // test tooltip
    const tooltipUI = wrapper.find(ReactTooltip);
    if (tooltip) {
      expect(tooltipUI).toExist();
      expect(tooltipUI).toIncludeText(tooltip);
    } else {
      expect(tooltipUI.exists()).toBeFalsy();
    }
  }
});
