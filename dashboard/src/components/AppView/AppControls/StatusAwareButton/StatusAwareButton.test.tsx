// Copyright 2021-2023 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { CdsButton } from "@cds/react/button";
import { act, waitFor } from "@testing-library/react";
import {
  InstalledPackageStatus,
  InstalledPackageStatus_StatusReason,
} from "gen/kubeappsapis/core/packages/v1alpha1/packages_pb";
import { Tooltip } from "react-tooltip";
import { defaultStore, mountWrapper } from "shared/specs/mountWrapper";
import StatusAwareButton, { IStatusAwareButtonProps } from "./StatusAwareButton";

type TProps = IStatusAwareButtonProps & {
  code: InstalledPackageStatus_StatusReason | null | undefined;
  tooltip?: string;
};

describe("tests the disabled flag and tooltip for each release with default and custom status condition", () => {
  // this should cover all conditions
  const testsProps: TProps[] = [
    {
      code: InstalledPackageStatus_StatusReason.PENDING,
      disabled: true,
      tooltip: "The application is pending installation.",
      id: "",
      releaseStatus: {} as InstalledPackageStatus,
    },
    {
      code: InstalledPackageStatus_StatusReason.UNINSTALLED,
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
      code: InstalledPackageStatus_StatusReason.UNINSTALLED,
      disabled: true,
      tooltip: "test tooltip for uninstalled",
      id: "",
      releaseStatus: {} as InstalledPackageStatus,
      statusesToDeactivate: [InstalledPackageStatus_StatusReason.UNINSTALLED],
      statusesToDeactivateTooltips: {
        [InstalledPackageStatus_StatusReason.UNINSTALLED]: "test tooltip for uninstalled",
      },
    },
    {
      code: InstalledPackageStatus_StatusReason.PENDING,
      disabled: false,
      id: "",
      releaseStatus: {} as InstalledPackageStatus,
      statusesToDeactivate: [InstalledPackageStatus_StatusReason.UNINSTALLED],
      statusesToDeactivateTooltips: {
        [InstalledPackageStatus_StatusReason.UNINSTALLED]: "test tooltip for uninstalled",
      },
    },
  ];

  for (const testProps of testsProps) {
    it(`shows tooltip when code is ${testProps.code} and tooltip "${testProps.tooltip}"`, async () => {
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

      act(() => {
        wrapper.find(CdsButton).simulate("mouseover");
      });
      wrapper.update();

      // test tooltip
      await waitFor(() => {
        expect(wrapper.find(Tooltip).prop("children")).toBe(testProps.tooltip);
      });
    });
  }
});
