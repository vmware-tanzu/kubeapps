// Copyright 2021-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { CdsIcon } from "@cds/react/icon";
import {
  InstalledPackageReference,
  InstalledPackageStatus,
} from "gen/kubeappsapis/core/packages/v1alpha1/packages";
import { Link } from "react-router-dom";
import * as url from "../../../../shared/url";
import StatusAwareButton from "../StatusAwareButton/StatusAwareButton";

interface IUpgradeButtonProps {
  installedPackageRef: InstalledPackageReference;
  releaseStatus: InstalledPackageStatus | undefined | null;
  disabled?: boolean;
}

export default function UpgradeButton({
  installedPackageRef,
  releaseStatus,
  disabled,
}: IUpgradeButtonProps) {
  const button = (
    <StatusAwareButton
      id="upgrade-button"
      status="primary"
      disabled={disabled}
      releaseStatus={releaseStatus}
    >
      <CdsIcon shape="upload-cloud" /> Upgrade
    </StatusAwareButton>
  );
  return disabled ? (
    <>{button}</>
  ) : (
    <Link to={url.app.apps.upgrade(installedPackageRef)}>{button}</Link>
  );
}
