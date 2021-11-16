import * as url from "shared/url";
import { getPluginIcon, getPluginName, PluginNames, trimDescription } from "shared/utils";
import placeholder from "../../placeholder.png";
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
  switch (availablePackageSummary.availablePackageRef?.plugin?.name) {
    case PluginNames.PACKAGES_HELM:
      pkgRepository = availablePackageSummary.availablePackageRef?.identifier.split("/")[0];
      break;
    case PluginNames.PACKAGES_FLUX:
      // TODO: get repo from flux
      break;
    case PluginNames.PACKAGES_KAPP:
      // TODO: get repo from kapp-controller
      break;
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
