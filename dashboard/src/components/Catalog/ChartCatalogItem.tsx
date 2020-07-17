import * as React from "react";
import { useSelector } from "react-redux";
import { trimDescription } from "shared/utils";
import helmIcon from "../../icons/helm.svg";
import placeholder from "../../placeholder.png";
import { IRepo, IStoreState } from "../../shared/types";
import * as url from "../../shared/url";
import InfoCard from "../InfoCard/InfoCard.v2";
import { IChartCatalogItem } from "./CatalogItem.v2";

export default function ChartCatalogItem(props: IChartCatalogItem) {
  const { icon, name, repo, version, description, namespace, id } = props;
  const iconSrc = icon || placeholder;
  const cluster = useSelector((state: IStoreState) => state.clusters.currentCluster);
  const link = url.app.charts.get(cluster, namespace, name, repo || ({} as IRepo));
  const subIcon = helmIcon;

  return (
    <InfoCard
      key={id}
      title={name}
      link={link}
      info={version || ""}
      icon={iconSrc}
      description={trimDescription(description)}
      tag1Content={<span>{repo.name}</span>}
      subIcon={subIcon}
    />
  );
}
