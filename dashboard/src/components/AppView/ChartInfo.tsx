import * as React from "react";
import { AlertTriangle, CheckCircle } from "react-feather";
import { Link } from "react-router-dom";

import { IRelease } from "shared/types";
import Card, { CardContent, CardFooter, CardGrid, CardIcon } from "../Card";

import placeholder from "../../placeholder.png";
import "./ChartInfo.css";

interface IChartInfoProps {
  app: IRelease;
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
          <div>
            <span>Chart Version: {metadata.version}</span>
            {this.updateStatusInfo()}
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

  private updateStatusInfo() {
    const { app } = this.props;
    // If update is not set yet we cannot know if there is
    // an update available or not
    if (app.updateInfo) {
      let updateContent = null;
      if (app.updateInfo.error) {
        updateContent = (
          <React.Fragment>
            <AlertTriangle
              color="white"
              fill="#FDBA12"
              className="icon"
              size={15}
              style={{ bottom: "-0.2em" }}
            />{" "}
            <span>Update check failed. {app.updateInfo.error.message}</span>
          </React.Fragment>
        );
      } else {
        if (app.updateInfo.upToDate) {
          updateContent = (
            <React.Fragment>
              <CheckCircle
                color="#82C341"
                className="icon"
                size={15}
                style={{ bottom: "-0.2em" }}
              />{" "}
              Up to date
            </React.Fragment>
          );
        } else {
          const update =
            app.chart &&
            app.chart.metadata &&
            app.chart.metadata.appVersion !== app.updateInfo.appLatestVersion ? (
              // A new version for the app is available
              <span>
                A new version for {app.chart.metadata.name} is available:{" "}
                {app.updateInfo.appLatestVersion}.
              </span>
            ) : (
              // Just a new chart version
              <span>A new chart version is available: {app.updateInfo.chartLatestVersion}.</span>
            );
          updateContent = (
            <React.Fragment>
              <h5 className="ChartInfoUpdate">Update Available</h5>
              {update}
              <br />
              <span>
                Click <Link to={`/apps/ns/${app.namespace}/upgrade/${app.name}`}>here</Link> to
                upgrade.
              </span>
            </React.Fragment>
          );
        }
      }
      return (
        <div>
          <hr className="separator-small" />
          {updateContent}
        </div>
      );
    }
    return;
  }
}

export default ChartInfo;
