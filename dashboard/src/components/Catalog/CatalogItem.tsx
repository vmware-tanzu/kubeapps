import * as React from "react";
import { Link } from "react-router-dom";
import helmIcon from "../../icons/helm.svg";
import operatorIcon from "../../icons/operator-framework.svg";
import placeholder from "../../placeholder.png";
import { IRepo } from "../../shared/types";
import * as url from "../../shared/url";
import InfoCard from "../InfoCard";
import "./CatalogItem.css";

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

// 3 lines description max
const MAX_DESC_LENGTH = 90;

function trimDescription(desc: string): string {
  if (desc.length > MAX_DESC_LENGTH) {
    // Trim to the last word under the max length
    return desc.substr(0, desc.lastIndexOf(" ", MAX_DESC_LENGTH)).concat("...");
  }
  return desc;
}

const CatalogItem: React.SFC<ICatalogItemProps> = props => {
  if (props.type === "operator") {
    const item = props.item as IOperatorCatalogItem;
    return OperatorCatalogItem(item); 
  } else {
    const item = props.item as IChartCatalogItem;
    return ChartCatalogItem(item);
  }
};

const OperatorCatalogItem: React.SFC<IOperatorCatalogItem> = props => {
  const { icon, name, csv, version, description, namespace, id } = props;
  const iconSrc = icon || placeholder;
  // Cosmetic change, remove the version from the csv name
  const csvName = props.csv.split(".v")[0];
  const tag1 = <span>{csvName}</span>;
  const link = `/ns/${namespace}/operators-instances/new/${csv}/${id}`;
  const subIcon = operatorIcon;
  const descriptionC = (
    <div className="ListItem__content__description">{trimDescription(description)}</div>
  );
  return (
    <InfoCard
      key={id}
      title={name}
      link={link}
      info={version || "-"}
      icon={iconSrc}
      description={descriptionC}
      tag1Content={tag1}
      tag1Class={""}
      subIcon={subIcon}
    />
  );
};

const ChartCatalogItem: React.SFC<IChartCatalogItem> = props => {
  const { icon, name, repo, version, description, namespace, id } = props;
  const iconSrc = icon || placeholder;
  const tag1 = (
    <Link
      className="ListItem__content__info_tag_link"
      to={url.app.repo(repo.name, namespace)}
    >
      {repo.name}
    </Link>
  );
  const link = url.app.charts.get(name, repo || {} as IRepo, namespace);
  const subIcon = helmIcon;

  const descriptionC = (
    <div className="ListItem__content__description">{trimDescription(description)}</div>
  );
  return (
    <InfoCard
      key={id}
      title={name}
      link={link}
      info={version || "-"}
      icon={iconSrc}
      description={descriptionC}
      tag1Content={tag1}
      tag1Class={repo ? repo.name : ""}
      subIcon={subIcon}
    />
  );
};

export default CatalogItem;
