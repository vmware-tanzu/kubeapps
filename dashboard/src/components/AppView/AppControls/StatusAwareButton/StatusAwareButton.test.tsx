// Copyright 2021-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { CdsButton } from "@cds/react/button";
import {
  InstalledPackageStatus,
  InstalledPackageStatus_StatusReason,
} from "gen/kubeappsapis/core/packages/v1alpha1/packages_pb";
import { Tooltip as ReactTooltip } from "react-tooltip";
import { defaultStore, mountWrapper } from "shared/specs/mountWrapper";
import StatusAwareButton, { IStatusAwareButtonProps } from "./StatusAwareButton";

type TProps = IStatusAwareButtonProps & {
  code: InstalledPackageStatus_StatusReason | null | undefined;
  tooltip?: string;
};

it("tests the disabled flag and tooltip for each release with default and custom status condition", async () => {
  // this should cover all conditions
  const testsProps: TProps[] = [
    {
      code: InstalledPackageStatus_StatusReason.STATUS_REASON_PENDING,
      disabled: true,
      tooltip: "The application is pending installation.",
      id: "",
      releaseStatus: {} as InstalledPackageStatus,
    },
    {
      code: InstalledPackageStatus_StatusReason.STATUS_REASON_UNINSTALLED,
      disabled: true,
      tooltip: "The application is being deleted.",
      id: "",
      releaseStatus: {} as InstalledPackageStatus,
    },
    {
      code: undefined,
      disabled: true,
      tooltip: undefined,
      id: "",
      releaseStatus: {} as InstalledPackageStatus,
    },
    {
      code: null,
      disabled: true,
      tooltip: undefined,
      id: "",
      releaseStatus: {} as InstalledPackageStatus,
    },
    {
      code: InstalledPackageStatus_StatusReason.STATUS_REASON_UNINSTALLED,
      disabled: true,
      tooltip: "test tooltip for uninstalled",
      id: "",
      releaseStatus: {} as InstalledPackageStatus,
      statusesToDeactivate: [InstalledPackageStatus_StatusReason.STATUS_REASON_UNINSTALLED],
      statusesToDeactivateTooltips: {
        [InstalledPackageStatus_StatusReason.STATUS_REASON_UNINSTALLED]:
          "test tooltip for uninstalled",
      },
    },
    {
      code: InstalledPackageStatus_StatusReason.STATUS_REASON_PENDING,
      disabled: false,
      id: "",
      releaseStatus: {} as InstalledPackageStatus,
      statusesToDeactivate: [InstalledPackageStatus_StatusReason.STATUS_REASON_UNINSTALLED],
      statusesToDeactivateTooltips: {
        [InstalledPackageStatus_StatusReason.STATUS_REASON_UNINSTALLED]:
          "test tooltip for uninstalled",
      },
    },
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
    const wrapper = mountWrapper(
      defaultStore,
      <StatusAwareButton
        id={testProps.id}
        releaseStatus={releaseStatus}
        disabled={testProps.disabled}
        statusesToDeactivate={testProps.statusesToDeactivate}
        statusesToDeactivateTooltips={testProps.statusesToDeactivateTooltips}
      />,
    );

    // test disabled flag
    expect(wrapper.find(CdsButton).prop("disabled")).toBe(testProps.disabled);

    // test tooltip
    const tooltipUI = wrapper.find(ReactTooltip);
    if (testProps.tooltip) {
      expect(tooltipUI).toExist();
      expect(tooltipUI).toIncludeText(testProps.tooltip);
    } else {
      expect(tooltipUI.exists()).toBeFalsy();
    }
  }
});
