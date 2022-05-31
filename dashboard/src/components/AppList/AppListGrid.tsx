// Copyright 2020-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import Row from "components/js/Row";
import { InstalledPackageSummary } from "gen/kubeappsapis/core/packages/v1alpha1/packages";
import { Link } from "react-router-dom";
import { IClusterServiceVersion, IResource } from "../../shared/types";
import * as url from "../../shared/url";
import { escapeRegExp } from "../../shared/utils";
import Alert from "../js/Alert";
import "./AppList.css";
import AppListItem from "./AppListItem";
import CustomResourceListItem from "./CustomResourceListItem";

export interface IAppListProps {
  appList: InstalledPackageSummary[] | undefined;
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
    new RegExp(escapeRegExp(filter), "i").test(a.name),
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
            href={`https://github.com/vmware-tanzu/kubeapps/blob/${appVersion}/site/content/docs/latest/tutorials/getting-started.md`}
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
    <Row>
      <>
        {filteredReleases.map(r => {
          return (
            <AppListItem
              key={`${r.installedPackageRef?.context?.namespace}-${r.installedPackageRef?.identifier}`}
              app={r}
              cluster={cluster}
            />
          );
        })}
        {filteredCRs.map(r => {
          const csv = props.csvs.find(c =>
            c.spec.customresourcedefinitions.owned?.some(crd => crd.kind === r.kind),
          );
          return (
            <CustomResourceListItem
              cluster={cluster}
              key={r.metadata.name + "_" + r.metadata.namespace}
              resource={r}
              csv={csv!}
            />
          );
        })}
      </>
    </Row>
  );
}

export default AppListGrid;
