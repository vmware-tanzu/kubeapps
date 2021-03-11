import { app } from "shared/url";
import { trimDescription } from "shared/utils";
import operatorIcon from "../../icons/operator-framework.svg";
import placeholder from "../../placeholder.png";
import InfoCard from "../InfoCard/InfoCard";
import { IOperatorCatalogItem } from "./CatalogItem";

export default function OperatorCatalogItem(props: IOperatorCatalogItem) {
  const { icon, name, csv, version, description, cluster, namespace, id } = props;
  const iconSrc = icon || placeholder;
  // Cosmetic change, remove the version from the csv name
  const csvName = props.csv.split(".")[0];
  const tag1 = <span>{csvName}</span>;
  const link = app.operatorInstances.new(cluster, namespace, csv, id);
  const bgIcon = operatorIcon;
  return (
    <InfoCard
      key={id}
      title={name}
      link={link}
      info={version || "-"}
      icon={iconSrc}
      description={trimDescription(description)}
      tag1Content={tag1}
      bgIcon={bgIcon}
    />
  );
}
