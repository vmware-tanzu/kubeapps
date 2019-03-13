import * as React from "react";
import { Link } from "react-router-dom";

import { IChartVersion } from "shared/types";
import ChartIcon from "../ChartIcon";
import ChartDeployButton from "./ChartDeployButton";
import "./ChartHeader.css";

interface IChartHeaderProps {
  id: string;
  icon?: string;
  repo: string;
  description: string;
  version: IChartVersion;
  namespace: string;
}

class ChartHeader extends React.Component<IChartHeaderProps> {
  public render() {
    const { id, icon, repo, description, version, namespace } = this.props;
    const appVersion = version.attributes.app_version;
    return (
      <header>
        <div className="ChartView__heading margin-normal row">
          <div className="col-1 ChartHeader__icon">
            <ChartIcon icon={icon} />
          </div>
          <div className="col-9">
            <div className="title margin-l-small">
              <h1 className="margin-t-reset">{id}</h1>
              <h5 className="subtitle margin-b-normal">
                {appVersion && <span>{appVersion} - </span>}
                <Link to={`/catalog/${repo}`}>{repo}</Link>
              </h5>
              <h5 className="subtitle margin-b-reset">{description}</h5>
            </div>
          </div>
          <div className="col-2 ChartHeader__button">
            <ChartDeployButton version={version} namespace={namespace} />
          </div>
        </div>
        <hr />
      </header>
    );
  }
}

export default ChartHeader;
