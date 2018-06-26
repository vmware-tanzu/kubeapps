import * as React from "react";

import { hapi } from "shared/hapi/release";
import Card, { CardContent, CardFooter, CardGrid, CardIcon } from "../Card";

import placeholder from "../../placeholder.png";
import "./ChartInfo.css";

interface IChartInfoProps {
  app: hapi.release.Release;
}

class ChartInfo extends React.Component<IChartInfoProps> {
  public render() {
    const { app } = this.props;
    const name = app.name;
    const metadata = app.chart && app.chart.metadata;
    const icon = metadata && metadata.icon;
    const iconSrc = icon ? icon : placeholder;
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
