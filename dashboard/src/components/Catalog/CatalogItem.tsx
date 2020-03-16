import * as React from "react";
import { Link } from "react-router-dom";

import placeholder from "../../placeholder.png";
import InfoCard from "../InfoCard";

import "./CatalogItem.css";

interface ICatalogItemProps {
  item: ICatalogItem;
}

export interface ICatalogItem {
  id: string;
  name: string;
  version: string;
  description: string;
  type: "chart" | "operator";
  namespace: string;
  icon?: string;
  repoName?: string;
  csv?: string;
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
  const { item } = props;
  const { icon, name, repoName, version, description, type, namespace, id, csv } = item;
  const iconSrc = icon || placeholder;
  let link;
  let tag1;
  if (type === "chart") {
    tag1 = (
      <Link className="ListItem__content__info_tag_link" to={`/catalog/${repoName}`}>
        {repoName}
      </Link>
    );
    link = `/charts/${repoName}/${name}`;
  } else {
    // Cosmetic change, remove the version from the csv name
    const csvName = csv?.split(".v")[0];
    tag1 = <span>{csvName}</span>;
    link = `/operators-instances/ns/${namespace}/new/${csv}/${id}`;
  }
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
      tag1Class={repoName}
      tag2Content={type}
    />
  );
};

export default CatalogItem;
