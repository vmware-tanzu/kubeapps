import * as React from "react";
import { Link } from "react-router-dom";

import { IAppOverview, IClusterServiceVersion, IResource } from "../../shared/types";
import * as url from "../../shared/url";
import { escapeRegExp } from "../../shared/utils";
import CardGrid from "../Card/CardGrid.v2";
import AppListItem from "./AppListItem.v2";

import Alert from "../js/Alert";
import "./AppList.v2.css";
import CustomResourceListItem from "./CustomResourceListItem.v2";

export interface IAppListProps {
  appList: IAppOverview[] | undefined;
  cluster: string;
  namespace: string;
  filter: string;
  customResources: IResource[];
  csvs: IClusterServiceVersion[];
  appVersion: string;
}

function AppListGrid(props: IAppListProps) {
  const { appList, customResources, cluster, namespace, appVersion, filter } = props;
  const filteredReleases = (appList || []).filter(a =>
    new RegExp(escapeRegExp(filter), "i").test(a.releaseName),
  );
  const filteredCRs = customResources.filter(cr =>
    new RegExp(escapeRegExp(filter), "i").test(cr.metadata.name),
  );

  if (filteredReleases.length === 0 && filteredCRs.length === 0) {
    return (
      <div className="applist-empty">
        <Alert>Deploy applications on your Kubernetes cluster with a single click.</Alert>
        <h2>Welcome To Kubeapps</h2>
        <p>
          Start browsing your <Link to={url.app.catalog(cluster, namespace)}>favourite apps</Link>{" "}
          or check the{" "}
          <a
            href={`https://github.com/kubeapps/kubeapps/blob/${appVersion}/docs/user/getting-started.md`}
            target="_blank"
            rel="noopener noreferrer"
          >
            documentation
          </a>
          .
        </p>
      </div>
    );
  }
  return (
    <CardGrid>
      <>
        {filteredReleases.map(r => {
          return <AppListItem key={r.releaseName} app={r} cluster={cluster} />;
        })}
        {filteredCRs.map(r => {
          const csv = props.csvs.find(c =>
            c.spec.customresourcedefinitions.owned.some(crd => crd.kind === r.kind),
          );
          return <CustomResourceListItem key={r.metadata.name} resource={r} csv={csv!} />;
        })}
      </>
    </CardGrid>
  );
}

export default AppListGrid;
