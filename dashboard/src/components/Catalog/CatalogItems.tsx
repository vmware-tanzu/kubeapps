// Copyright 2020-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { AvailablePackageSummary } from "gen/kubeappsapis/core/packages/v1alpha1/packages";
import { useMemo } from "react";
import { getIcon } from "shared/Operators";
import { IClusterServiceVersion } from "shared/types";
import CatalogItem, {
  ICatalogItemProps,
  IOperatorCatalogItem,
  IPackageCatalogItem,
} from "./CatalogItem";
export interface ICatalogItemsProps {
  availablePackageSummaries: AvailablePackageSummary[];
  csvs: IClusterServiceVersion[];
  cluster: string;
  namespace: string;
  isFirstPage: boolean;
  hasLoadedFirstPage: boolean;
  hasFinishedFetching: boolean;
}

export default function CatalogItems({
  availablePackageSummaries,
  csvs,
  cluster,
  namespace,
  isFirstPage,
  hasLoadedFirstPage,
  hasFinishedFetching,
}: ICatalogItemsProps) {
  const packageItems: ICatalogItemProps[] = useMemo(
    () =>
      availablePackageSummaries.map(c => {
        return {
          // TODO: this should be simplified once the operators are also implemented as a plugin
          type: `${c.availablePackageRef?.plugin?.name}/${c.availablePackageRef?.plugin?.version}`,
          id: `package/${c.availablePackageRef?.identifier}`,
          item: {
            name: c.displayName,
            cluster,
            namespace,
            availablePackageSummary: c,
          } as IPackageCatalogItem,
        } as ICatalogItemProps;
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
                } as IOperatorCatalogItem,
              } as ICatalogItemProps;
            });
          } else {
            return [];
          }
        })
        .flat(),
    [csvs, cluster, namespace],
  );

  const sortedItems =
    !hasLoadedFirstPage && isFirstPage
      ? []
      : packageItems
          .concat(crdItems)
          .sort((a, b) =>
            a.item.name.toLowerCase() > b.item.name.toLowerCase()
              ? 1
              : b.item.name.toLowerCase() > a.item.name.toLowerCase()
              ? -1
              : 0,
          );

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
