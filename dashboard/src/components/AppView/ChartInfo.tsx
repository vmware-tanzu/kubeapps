import * as React from "react";
import { ArrowUpCircle, CheckCircle } from "react-feather";

import { hapi } from "shared/hapi/release";
import Card, { CardContent, CardFooter, CardGrid, CardIcon } from "../Card";

import { Link } from "react-router-dom";
import placeholder from "../../placeholder.png";
import "./ChartInfo.css";

interface IChartInfoProps {
  app: hapi.release.Release;
  latest?: string;
}

class ChartInfo extends React.Component<IChartInfoProps> {
  public render() {
    const { app, latest } = this.props;
    const name = app.name;
    const metadata = app.chart && app.chart.metadata;
    const icon = metadata && metadata.icon;
    const iconSrc = icon ? icon : placeholder;
    let isUpdated = null;
    if (app.chart && app.chart.metadata && app.chart.metadata.version && latest) {
      if (app.chart.metadata.version === latest) {
        isUpdated = (
          <span>
            -{" "}
            <CheckCircle color="#20cb15" className="icon" size={15} style={{ bottom: "-0.2em" }} />{" "}
            Up to date
          </span>
        );
      } else {
        isUpdated = (
          // TODO: It should already include the repo found when clicking
          <Link to={`/apps/ns/${app.namespace}/upgrade/${app.name}`}>
            <span>
              -{" "}
              <ArrowUpCircle
                color="white"
                className="icon"
                fill="#20cb15"
                size={15}
                style={{ bottom: "-0.2em" }}
              />{" "}
              New version ({this.props.latest}) available
            </span>
          </Link>
        );
      }
    }
    let notes = <span />;
    if (metadata) {
      notes = (
        <div>
          {metadata.appVersion && <div>App Version: {metadata.appVersion}</div>}
          <div>
            Chart Version: {metadata.version} {isUpdated}
          </div>
        </div>
      );
    }
    return (
      <CardGrid className="ChartInfo">
        <Card>
          <CardIcon icon={iconSrc} />
          <CardContent>
            <h5>{name}</h5>
            <p className="margin-b-reset">{metadata && metadata.description}</p>
          </CardContent>
          <CardFooter>
            <small>{notes}</small>
          </CardFooter>
        </Card>
      </CardGrid>
    );
  }
}

export default ChartInfo;
