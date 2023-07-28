// Copyright 2021-2023 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import PageHeader from "components/PageHeader/PageHeader";
import {
  AvailablePackageDetail,
  PackageAppVersion,
} from "gen/kubeappsapis/core/packages/v1alpha1/packages_pb";
import placeholder from "icons/placeholder.svg";
import React from "react";
import { Tooltip } from "react-tooltip";
import "./PackageHeader.css";
import PackageVersionSelector from "./PackageVersionSelector";
import { CdsIcon } from "@cds/react/icon";

export interface IPackageHeaderProps {
  availablePackageDetail: AvailablePackageDetail;
  versions: PackageAppVersion[];
  onSelect: (event: React.ChangeEvent<HTMLSelectElement>) => void;
  releaseName?: string;
  currentVersion?: string;
  selectedVersion?: string;
  deployButton?: JSX.Element;
  hideVersionsSelector?: boolean;
}

export default function PackageHeader({
  availablePackageDetail,
  versions,
  onSelect,
  releaseName,
  currentVersion,
  deployButton,
  selectedVersion,
  hideVersionsSelector,
}: IPackageHeaderProps) {
  return availablePackageDetail?.availablePackageRef?.identifier ? (
    <PageHeader
      title={
        releaseName
          ? `${releaseName} (${decodeURIComponent(
              availablePackageDetail.availablePackageRef.identifier,
            )})`
          : `${decodeURIComponent(availablePackageDetail.displayName)}`
      }
      titleSize="md"
      icon={availablePackageDetail?.iconUrl ? availablePackageDetail.iconUrl : placeholder}
      plugin={availablePackageDetail.availablePackageRef.plugin}
      version={
        hideVersionsSelector ? (
          <></>
        ) : (
          <>
            <PackageVersionSelector
              versions={versions}
              onSelect={onSelect}
              selectedVersion={selectedVersion}
              currentVersion={currentVersion}
              label={
                <>
                  <span data-tooltip-id="package-versions-tooltip">
                    Package Version <CdsIcon shape="info-circle" size="sm" solid={true} />
                  </span>
                </>
              }
            />
            <Tooltip id="package-versions-tooltip" place="bottom-end" clickable={true}>
              Package and application versions can be increased independently.{" "}
              <a
                href="https://helm.sh/docs/topics/charts/#charts-and-versioning"
                target="_blank"
                rel="noopener noreferrer"
              >
                More info here
              </a>
              .{" "}
            </Tooltip>
          </>
        )
      }
      buttons={deployButton ? [deployButton] : undefined}
    />
  ) : (
    <></>
  );
}
