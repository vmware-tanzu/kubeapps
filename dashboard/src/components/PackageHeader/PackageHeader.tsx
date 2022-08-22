// Copyright 2021-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import Tooltip from "components/js/Tooltip";
import PageHeader from "components/PageHeader/PageHeader";
import {
  AvailablePackageDetail,
  PackageAppVersion,
} from "gen/kubeappsapis/core/packages/v1alpha1/packages";
import React from "react";
import placeholder from "icons/placeholder.svg";
import "./PackageHeader.css";
import PackageVersionSelector from "./PackageVersionSelector";

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
                  Package Version{" "}
                  <Tooltip
                    label="package-versions-tooltip"
                    id="package-versions-tooltip"
                    position="bottom-left"
                    iconProps={{ solid: true, size: "sm" }}
                  >
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
              }
            />
          </>
        )
      }
      buttons={deployButton ? [deployButton] : undefined}
    />
  ) : (
    <></>
  );
}
