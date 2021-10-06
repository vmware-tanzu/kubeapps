import { AvailablePackageSummary } from "gen/kubeappsapis/core/packages/v1alpha1/packages";
import { Plugin } from "gen/kubeappsapis/core/plugins/v1alpha1/plugins";
import { useMemo } from "react";
import { getIcon } from "shared/Operators";
import { IClusterServiceVersion, IRepo } from "shared/types";
import placeholder from "../../placeholder.png";
import CatalogItem, { ICatalogItemProps } from "./CatalogItem";
export interface ICatalogItemsProps {
  availablePackageSummaries: AvailablePackageSummary[];
  csvs: IClusterServiceVersion[];
  cluster: string;
  namespace: string;
  page: number;
  hasLoadedFirstPage: boolean;
  hasFinishedFetching: boolean;
}

export default function CatalogItems({
  availablePackageSummaries,
  csvs,
  cluster,
  namespace,
  page,
  hasLoadedFirstPage,
  hasFinishedFetching,
}: ICatalogItemsProps) {
  const packageItems: ICatalogItemProps[] = useMemo(
    () =>
      availablePackageSummaries.map(c => {
        return {
          type: `${c.availablePackageRef?.plugin?.name}/${c.availablePackageRef?.plugin?.version}`,
          id: `package/${c.availablePackageRef?.identifier}`,
          item: {
            plugin: c.availablePackageRef?.plugin ?? ({ name: "", version: "" } as Plugin),
            id: `package/${c.availablePackageRef?.identifier}/${c.latestVersion?.pkgVersion}`,
            name: c.displayName,
            icon: c.iconUrl ?? placeholder,
            version: c.latestVersion?.pkgVersion ?? "",
            description: c.shortDescription,
            // TODO(agamez): get the repo name once available
            // https://github.com/kubeapps/kubeapps/issues/3165#issuecomment-884574732
            repo: {
              name: c.availablePackageRef?.identifier.split("/")[0],
              namespace: c.availablePackageRef?.context?.namespace,
            } as IRepo,
            cluster,
            namespace,
          },
        };
      }),
    [availablePackageSummaries, cluster, namespace],
  );
  const crdItems: ICatalogItemProps[] = useMemo(
    () =>
      csvs
        .map(csv => {
          if (csv.spec.customresourcedefinitions?.owned) {
            return csv.spec.customresourcedefinitions.owned.map(crd => {
              return {
                type: "operator",
                id: `operator/${csv.metadata.name}/${crd.name}`,
                item: {
                  id: crd.name,
                  name: crd.displayName || crd.name,
                  icon: getIcon(csv),
                  version: crd.version,
                  description: crd.description,
                  csv: csv.metadata.name,
                  cluster,
                  namespace,
                },
              };
            });
          } else {
            return [];
          }
        })
        .flat(),
    [csvs, cluster, namespace],
  );

  const sortedItems =
    !hasLoadedFirstPage && page === 1
      ? []
      : packageItems
          .concat(crdItems)
          .sort((a, b) => (a.item.name.toLowerCase() > b.item.name.toLowerCase() ? 1 : -1));

  if (hasFinishedFetching && sortedItems.length === 0) {
    return <p>No application matches the current filter.</p>;
  }
  return (
    <>
      {sortedItems.map(i => (
        <CatalogItem key={i.id} {...i} />
      ))}
    </>
  );
}
