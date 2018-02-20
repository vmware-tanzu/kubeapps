import * as React from "react";
import { Link } from "react-router-dom";

import { IApp } from "shared/types";
import Card, { CardContent, CardFooter, CardGrid, CardIcon } from "../Card";

import placeholder from "../../placeholder.png";
import "./ChartInfo.css";

interface IChartInfoProps {
  app: IApp;
}

class ChartInfo extends React.Component<IChartInfoProps> {
  public render() {
    const { app } = this.props;
    const name = app.data.name;
    let iconSrc = placeholder;
    if (app.chart && app.chart.attributes.icon) {
      iconSrc = `/api/chartsvc/${app.chart && app.chart.attributes.icon}`;
    }
    let chartID = "";
    const metadata = app.data.chart && app.data.chart.metadata;
    if (app.repo && app.repo.name && app.chart) {
      chartID = `${app.repo.name}/${app.chart.attributes.name}`;
      if (metadata) {
        chartID = `${chartID}/versions/${metadata.version}`;
      }
    }
    let notes = <span />;
    if (metadata) {
      notes = (
        <div>
          {metadata.appVersion && <div>App Version: {metadata.appVersion}</div>}
          <div>Chart Version: {metadata.version}</div>
        </div>
      );
    }
    return (
      <CardGrid className="ChartInfo">
        <Link to={`/charts/${chartID}`}>
          <Card>
            <CardIcon icon={iconSrc} />
            <CardContent>
              <h5>{name}</h5>
              <p className="margin-b-reset">{app.chart && app.chart.attributes.description}</p>
            </CardContent>
            <CardFooter>
              <small>{notes}</small>
            </CardFooter>
          </Card>
        </Link>
      </CardGrid>
    );
  }
}

export default ChartInfo;
