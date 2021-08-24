import { AvailablePackageSummary } from "gen/kubeappsapis/core/packages/v1alpha1/packages";
import { useMemo } from "react";
import { getIcon } from "shared/Operators";
import { IClusterServiceVersion, IRepo } from "shared/types";
import CatalogItem, { ICatalogItemProps } from "./CatalogItem";

interface ICatalogItemsProps {
  charts: AvailablePackageSummary[];
  csvs: IClusterServiceVersion[];
  cluster: string;
  namespace: string;
  page: number;
  isFetching: boolean;
  hasFinishedFetching: boolean;
}

export default function CatalogItems({
  charts,
  csvs,
  cluster,
  namespace,
  page,
  isFetching,
  hasFinishedFetching,
}: ICatalogItemsProps) {
  const chartItems: ICatalogItemProps[] = useMemo(
    () =>
      charts.map(c => {
        return {
          type: "chart",
          id: `chart/${c.availablePackageRef?.identifier}`,
          item: {
            id: `chart/${c.availablePackageRef?.identifier}/${c.latestVersion?.pkgVersion}`,
            name: c.displayName,
            icon: c.iconUrl,
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
    [charts, cluster, namespace],
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
    isFetching && page === 1
      ? []
      : chartItems
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
