import { useSelector } from "react-redux";
import { IRepo, IStoreState } from "shared/types";
import * as url from "shared/url";
import { getPluginIcon, trimDescription } from "shared/utils";
import placeholder from "../../placeholder.png";
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
    props.plugin,
  );
  const bgIcon = getPluginIcon(props.plugin);

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
