import * as React from "react";
import { trimDescription } from "shared/utils";
import operatorIcon from "../../icons/operator-framework.svg";
import placeholder from "../../placeholder.png";
import { IRepo } from "../../shared/types";
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

export default function OperatorCatalogItem(props: IOperatorCatalogItem) {
  const { icon, name, csv, version, description, namespace, id } = props;
  const iconSrc = icon || placeholder;
  // Cosmetic change, remove the version from the csv name
  const csvName = props.csv.split(".")[0];
  const tag1 = <span>{csvName}</span>;
  const link = `/ns/${namespace}/operators-instances/new/${csv}/${id}`;
  const subIcon = operatorIcon;
  return (
    <InfoCard
      key={id}
      title={name}
      link={link}
      info={version || "-"}
      icon={iconSrc}
      description={trimDescription(description)}
      tag1Content={tag1}
      subIcon={subIcon}
    />
  );
}
