import * as React from "react";
import { Link } from "react-router-dom";

import placeholder from "../../placeholder.png";
import { IChart } from "../../shared/types";
import InfoCard from "../InfoCard";

import "./CatalogItem.css";

interface ICatalogItemProps {
  chart: IChart;
}

// 3 lines description max
const MAX_DESC_LENGTH = 90;

interface ICatalogItemState {
  fullDescription: boolean;
}

class CatalogItem extends React.Component<ICatalogItemProps, ICatalogItemState> {
  public state: ICatalogItemState = {
    fullDescription: false,
  };

  public render() {
    const { chart } = this.props;
    const { icon, name, repo } = chart.attributes;
    const iconSrc = icon ? `/api/chartsvc/${icon}` : placeholder;
    const latestAppVersion = chart.relationships.latestChartVersion.data.app_version;
    const repoTag = (
      <Link className="ListItem__content__info_tag_link" to={`/catalog/${repo.name}`}>
        {repo.name}
      </Link>
    );
    const description = (
      <div
        className={`ListItem__content__description
          ${this.state.fullDescription ? "ListItem__content__description-full" : ""}`}
      >
        {this.state.fullDescription
          ? chart.attributes.description
          : this.trimDescription(chart.attributes.description)}
        {chart.attributes.description.length > MAX_DESC_LENGTH && (
          <a onClick={this.toggleFullDescription}>
            {" "}
            {this.state.fullDescription ? "Hide" : "Show more"}
          </a>
        )}
      </div>
    );
    return (
      <InfoCard
        key={`${repo}/${name}`}
        title={name}
        link={`/charts/${chart.id}`}
        info={latestAppVersion || "-"}
        icon={iconSrc}
        description={description}
        tag1Content={repoTag}
        tag1Class={repo.name}
      />
    );
  }

  private toggleFullDescription = () => {
    this.setState({ fullDescription: !this.state.fullDescription });
  };

  private trimDescription(desc: string): string {
    if (desc.length > MAX_DESC_LENGTH) {
      // Trim to the last word under the max length
      return desc.substr(0, desc.lastIndexOf(" ", MAX_DESC_LENGTH)).concat("... ");
    }
    return desc;
  }
}

export default CatalogItem;
