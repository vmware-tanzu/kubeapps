import * as url from "shared/url";
import { getPluginIcon, PluginNames, trimDescription } from "shared/utils";
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

  // Historically, this tag is used to show the repository a given package is from,
  // but each plugin as its own way to describe the repository right now.
  let repositoryName;
  switch (availablePackageSummary.availablePackageRef?.plugin?.name) {
    case PluginNames.PACKAGES_HELM:
      repositoryName = availablePackageSummary.availablePackageRef?.identifier.split("/")[0];
      break;
    // TODO: consider the fluxv2 plugin
    default:
      // Fallback to the plugin name
      repositoryName = availablePackageSummary.availablePackageRef?.plugin?.name;
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
      tag1Content={<span>{repositoryName}</span>}
      bgIcon={getPluginIcon(availablePackageSummary.availablePackageRef?.plugin)}
    />
  );
}
