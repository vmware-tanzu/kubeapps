// Copyright 2021-2023 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { CdsButton } from "@cds/react/button";
import {
  InstalledPackageStatus,
  InstalledPackageStatus_StatusReason,
} from "gen/kubeappsapis/core/packages/v1alpha1/packages_pb";
import { Tooltip } from "react-tooltip";

export interface IStatusAwareButtonProps {
  id: string;
  releaseStatus: InstalledPackageStatus | undefined | null;
  disabled?: boolean;
  statusesToDeactivate?: InstalledPackageStatus_StatusReason[];
  statusesToDeactivateTooltips?: { [key: string]: string };
}

export default function StatusAwareButton<T extends IStatusAwareButtonProps>(props: T) {
  const {
    id,
    releaseStatus,
    disabled,
    statusesToDeactivate,
    statusesToDeactivateTooltips,
    ...otherProps
  } = props;

  const defaultStatusesToDeactivate = [
    InstalledPackageStatus_StatusReason.UNINSTALLED,
    InstalledPackageStatus_StatusReason.PENDING,
  ];
  const defaultStatusesToDeactivateTooltips: { [index: string]: string } = {
    [InstalledPackageStatus_StatusReason.UNINSTALLED]: "The application is being deleted.",
    [InstalledPackageStatus_StatusReason.PENDING]: "The application is pending installation.",
  };

  // allow buttons to override the default statuses to deactivate
  const statuses = statusesToDeactivate?.length
    ? statusesToDeactivate
    : defaultStatusesToDeactivate;

  const tooltips = statusesToDeactivateTooltips
    ? statusesToDeactivateTooltips
    : defaultStatusesToDeactivateTooltips;

  // Deactivate the button if: the status code is undefined or null OR the status code is (uninstalled or pending)
  const isDisabled =
    disabled || releaseStatus?.reason == null ? true : statuses.includes(releaseStatus.reason);

  const tooltip = releaseStatus?.reason ? tooltips![releaseStatus.reason] : undefined;
  return (
    <>
      <CdsButton {...otherProps} disabled={isDisabled} data-tooltip-id={id} />
      <Tooltip id={id} place="bottom">
        {tooltip}
      </Tooltip>
    </>
  );
}
