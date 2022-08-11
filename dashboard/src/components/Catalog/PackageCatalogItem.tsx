// Copyright 2021-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import * as url from "shared/url";
import { getPluginIcon, getPluginName, trimDescription } from "shared/utils";
import placeholder from "icons/placeholder.svg";
import InfoCard from "../InfoCard/InfoCard";
import { IPackageCatalogItem } from "./CatalogItem";

export default function PackageCatalogItem(props: IPackageCatalogItem) {
  const { cluster, namespace, availablePackageSummary } = props;

  // Use the current cluster/namespace in the URL (passed as props here),
  // but, if it is global a "global" segement will be included in the generated URL.
  const packageViewLink = url.app.packages.get(
    cluster,
    namespace,
    availablePackageSummary.availablePackageRef!,
  );

  let pkgRepository;
  const pkgPluginName = getPluginName(availablePackageSummary.availablePackageRef?.plugin);

  // Get the pkg repository for the plugins that have one.
  // Assuming an identifier will always be like: "repo/pkgName"
  const splitIdentifier = availablePackageSummary.availablePackageRef?.identifier.split("/");
  if (splitIdentifier && splitIdentifier?.length > 1) {
    pkgRepository = splitIdentifier[0];
  }

  return (
    <InfoCard
      key={availablePackageSummary.availablePackageRef?.identifier}
      title={decodeURIComponent(availablePackageSummary.displayName)}
      link={packageViewLink}
      info={availablePackageSummary?.latestVersion?.pkgVersion || ""}
      icon={availablePackageSummary.iconUrl || placeholder}
      description={trimDescription(availablePackageSummary.shortDescription)}
      tag1Content={pkgRepository}
      tag2Content={pkgPluginName}
      tag2Class={"label-info-secondary"}
      bgIcon={getPluginIcon(availablePackageSummary.availablePackageRef?.plugin)}
    />
  );
}
