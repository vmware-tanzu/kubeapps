// Copyright 2021-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { CdsButton } from "@cds/react/button";
import {
  InstalledPackageStatus,
  InstalledPackageStatus_StatusReason,
} from "gen/kubeappsapis/core/packages/v1alpha1/packages";
import ReactTooltip from "react-tooltip";

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
    InstalledPackageStatus_StatusReason.STATUS_REASON_UNINSTALLED,
    InstalledPackageStatus_StatusReason.STATUS_REASON_PENDING,
  ];
  const defaultStatusesToDeactivateTooltips = {
    [InstalledPackageStatus_StatusReason.STATUS_REASON_UNINSTALLED]:
      "The application is being deleted.",
    [InstalledPackageStatus_StatusReason.STATUS_REASON_PENDING]:
      "The application is pending installation.",
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
      <CdsButton {...otherProps} disabled={isDisabled} data-for={id} data-tip={true} />
      {tooltip && (
        <ReactTooltip id={id} effect="solid" place="bottom">
          {tooltip}
        </ReactTooltip>
      )}
    </>
  );
}
