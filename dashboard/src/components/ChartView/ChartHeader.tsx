import * as React from "react";
import { Link } from "react-router-dom";

import ChartIcon from "../ChartIcon";

interface IChartHeaderProps {
  appVersion?: string;
  id: string;
  icon?: string;
  repo: string;
  description: string;
}

class ChartHeader extends React.Component<IChartHeaderProps> {
  public render() {
    const { appVersion, id, icon, repo, description } = this.props;
    return (
      <header>
        <div className="ChartView__heading margin-normal">
          <ChartIcon icon={icon} />
          <div className="title margin-l-small">
            <h1 className="margin-t-reset">{id}</h1>
            <h5 className="subtitle margin-b-normal">
              {appVersion && <span>{appVersion} - </span>}
              <Link to={`/catalog/${repo}`}>{repo}</Link>
            </h5>
            <h5 className="subtitle margin-b-reset">{description}</h5>
          </div>
        </div>
        <hr />
      </header>
    );
  }
}

export default ChartHeader;
