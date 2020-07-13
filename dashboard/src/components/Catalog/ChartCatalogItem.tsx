import * as React from "react";
import { useSelector } from "react-redux";
import { trimDescription } from "shared/utils";
import helmIcon from "../../icons/helm.svg";
import placeholder from "../../placeholder.png";
import { IRepo, IStoreState } from "../../shared/types";
import * as url from "../../shared/url";
import InfoCard from "../InfoCard/InfoCard.v2";

export interface ICatalogItem {
  id: string;
  name: string;
  version: string;
  description: string;
  namespace: string;
  icon?: string;
}

export interface IChartCatalogItem extends ICatalogItem {
  repo: IRepo;
}

export interface IOperatorCatalogItem extends ICatalogItem {
  csv: string;
}

export interface ICatalogItemProps {
  type: string;
  item: IChartCatalogItem | IOperatorCatalogItem;
}

export default function ChartCatalogItem(props: IChartCatalogItem) {
  const { icon, name, repo, version, description, namespace, id } = props;
  const iconSrc = icon || placeholder;
  const cluster = useSelector((state: IStoreState) => state.clusters.currentCluster);
  // const tag1 = <Link to={url.app.repo(cluster, namespace, repo.name)}>{repo.name}</Link>;
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
