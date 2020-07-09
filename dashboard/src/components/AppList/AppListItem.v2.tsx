import { bundleIcon, circleArrowIcon, ClarityIcons } from "@clr/core/icon-shapes";
import * as React from "react";
import helmIcon from "../../icons/helm.svg";
import placeholder from "../../placeholder.png";
import { IAppOverview } from "../../shared/types";
import * as url from "../../shared/url";
import { CdsIcon } from "../Clarity/clarity";
import InfoCard from "../InfoCard/InfoCard.v2";

import Column from "components/js/Column";
import Row from "components/js/Row";
import "./AppListItem.v2.css";

ClarityIcons.addIcons(bundleIcon, circleArrowIcon);

export interface IAppListItemProps {
  app: IAppOverview;
  cluster: string;
}

function AppListItem(props: IAppListItemProps) {
  const { app, cluster } = props;
  const icon = app.icon ? app.icon : placeholder;
  const appStatus = app.status.toLocaleLowerCase();
  const updateAvailable = app.updateInfo && !app.updateInfo.error && !app.updateInfo.upToDate;
  let tag2Content;
  if (app.updateInfo && updateAvailable) {
    if (app.updateInfo.appLatestVersion !== app.chartMetadata.appVersion) {
      tag2Content = `New App: ${app.updateInfo.appLatestVersion}`;
    } else {
      tag2Content = `New Chart: ${app.updateInfo.chartLatestVersion}`;
    }
  }
  return (
    <InfoCard
      key={app.releaseName}
      link={url.app.apps.get(cluster, app.namespace, app.releaseName)}
      title={app.releaseName}
      icon={icon}
      info={
        <Row aria-label="Chart version information">
          <Column span={2}>
            <div className={`info-icon ${updateAvailable ? "is-success" : ""}`}>
              <CdsIcon shape={updateAvailable ? "circle-arrow" : "bundle"} />
            </div>
          </Column>
          <Column span={10}>
            <span>App: {app.chartMetadata.appVersion}</span>
            <br />
            <span>Chart: {app.chartMetadata.version}</span>
          </Column>
        </Row>
      }
      description={app.chartMetadata.description}
      tag1Content={`Status: ${appStatus}`}
      tag1Class={appStatus === "deployed" ? "label-success" : "label-warning"}
      tag2Content={tag2Content}
      subIcon={helmIcon}
    />
  );
}

export default AppListItem;
