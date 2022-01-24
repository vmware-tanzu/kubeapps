// Copyright 2021-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { CdsControlMessage } from "@cds/react/forms";
import { CdsSelect } from "@cds/react/select";
import { PackageAppVersion } from "gen/kubeappsapis/core/packages/v1alpha1/packages";
import React from "react";
interface IPackageVersionSelectorProps {
  versions: PackageAppVersion[];
  onSelect: (event: React.ChangeEvent<HTMLSelectElement>) => void;
  currentVersion?: string;
  selectedVersion?: string;
  label?: JSX.Element | string;
  message?: string;
}

export default function PackageVersionSelector({
  versions,
  onSelect,
  currentVersion,
  selectedVersion,
  label,
  message,
}: IPackageVersionSelectorProps) {
  /* eslint-disable jsx-a11y/label-has-associated-control */
  return (
    <CdsSelect>
      <label>{label}</label>
      <select
        name="package-versions"
        value={selectedVersion || currentVersion || (versions.length ? versions[0].pkgVersion : "")}
        onChange={onSelect}
      >
        {versions.map(v => (
          <option key={`package-version-selector-${v.pkgVersion}`} value={v.pkgVersion}>
            {v.pkgVersion} / App Version {v.appVersion}
            {currentVersion === v.pkgVersion ? " (current)" : ""}
          </option>
        ))}
      </select>
      {message ? <CdsControlMessage>{message}</CdsControlMessage> : <></>}
    </CdsSelect>
  );
}
