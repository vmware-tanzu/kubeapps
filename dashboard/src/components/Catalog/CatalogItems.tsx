import React from "react";
import { getIcon } from "shared/Operators";
import { IChart, IClusterServiceVersion } from "shared/types";
import CatalogItem, { ICatalogItemProps } from "./CatalogItem.v2";

interface ICatalogItemsProps {
  charts: IChart[];
  csvs: IClusterServiceVersion[];
  namespace: string;
}

export default function CatalogItems({ charts, csvs, namespace }: ICatalogItemsProps) {
  const chartItems: ICatalogItemProps[] = charts.map(c => {
    return {
      type: "chart",
      key: `chart/${c.attributes.repo.name}/${c.id}`,
      item: {
        id: c.id,
        name: c.attributes.name,
        icon: c.attributes.icon ? `api/assetsvc/${c.attributes.icon}` : undefined,
        version: c.relationships.latestChartVersion.data.app_version,
        description: c.attributes.description,
        repo: c.attributes.repo,
        namespace,
      },
    };
  });
  const crdItems: ICatalogItemProps[] = csvs
    .map(csv => {
      if (csv.spec.customresourcedefinitions?.owned) {
        return csv.spec.customresourcedefinitions.owned.map(crd => {
          return {
            type: "operator",
            key: `operator/${csv.metadata.name}/${crd.name}`,
            item: {
              id: crd.name,
              name: crd.displayName || crd.name,
              icon: getIcon(csv),
              version: crd.version,
              description: crd.description,
              csv: csv.metadata.name,
              namespace,
            },
          };
        });
      } else {
        return [];
      }
    })
    .flat();

  const sortedItems = chartItems
    .concat(crdItems)
    .sort((a, b) => (a.item.name.toLowerCase() > b.item.name.toLowerCase() ? 1 : -1));

  return (
    <>
      {sortedItems.map(i => (
        <CatalogItem type={i.type} key={i.key} item={i.item} />
      ))}
    </>
  );
}
