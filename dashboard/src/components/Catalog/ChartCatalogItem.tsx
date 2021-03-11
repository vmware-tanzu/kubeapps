import { useSelector } from "react-redux";
import { trimDescription } from "shared/utils";
import helmIcon from "../../icons/helm.svg";
import placeholder from "../../placeholder.png";
import { IRepo, IStoreState } from "../../shared/types";
import * as url from "../../shared/url";
import InfoCard from "../InfoCard/InfoCard";
import { IChartCatalogItem } from "./CatalogItem";

export default function ChartCatalogItem(props: IChartCatalogItem) {
  const { icon, name, repo, version, description, namespace, id } = props;
  const {
    config: { kubeappsNamespace },
  } = useSelector((state: IStoreState) => state);
  const iconSrc = icon || placeholder;
  const cluster = useSelector((state: IStoreState) => state.clusters.currentCluster);
  const link = url.app.charts.get(
    cluster,
    namespace,
    name,
    repo || ({} as IRepo),
    kubeappsNamespace,
  );
  const bgIcon = helmIcon;

  return (
    <InfoCard
      key={id}
      title={decodeURIComponent(name)}
      link={link}
      info={version || ""}
      icon={iconSrc}
      description={trimDescription(description)}
      tag1Content={<span>{repo.name}</span>}
      bgIcon={bgIcon}
    />
  );
}
