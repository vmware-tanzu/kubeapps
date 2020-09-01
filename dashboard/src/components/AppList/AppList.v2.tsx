import * as React from "react";
import { useEffect, useState } from "react";
import { Link } from "react-router-dom";
import { CdsButton, CdsIcon } from "../Clarity/clarity";

import { IFeatureFlags } from "../../shared/Config";
import { IAppState, IClusterServiceVersion, IResource } from "../../shared/types";
import * as url from "../../shared/url";
import Column from "../js/Column";
import PageHeader from "../PageHeader/PageHeader.v2";
import SearchFilter from "../SearchFilter/SearchFilter.v2";

import Alert from "components/js/Alert";
import Row from "components/js/Row";
import LoadingWrapper from "components/LoadingWrapper/LoadingWrapper.v2";
import "./AppList.v2.css";
import AppListGrid from "./AppListGrid";

export interface IAppListProps {
  apps: IAppState;
  fetchAppsWithUpdateInfo: (cluster: string, ns: string, all: boolean) => void;
  cluster: string;
  namespace: string;
  pushSearchFilter: (filter: string) => any;
  filter: string;
  getCustomResources: (cluster: string, ns: string) => void;
  customResources: IResource[];
  isFetchingResources: boolean;
  csvs: IClusterServiceVersion[];
  featureFlags: IFeatureFlags;
  appVersion: string;
}

function AppList(props: IAppListProps) {
  const [filter, setFilter] = useState("");
  const {
    fetchAppsWithUpdateInfo,
    filter: filterProps,
    namespace,
    getCustomResources,
    apps: { error, isFetching, listOverview },
    cluster,
    isFetchingResources,
    pushSearchFilter,
    customResources,
    appVersion,
    csvs,
  } = props;

  useEffect(() => {
    fetchAppsWithUpdateInfo(cluster, namespace, true);
    getCustomResources(cluster, namespace);
  }, [cluster, namespace, fetchAppsWithUpdateInfo, getCustomResources]);

  useEffect(() => {
    setFilter(filterProps);
  }, [filterProps]);

  return (
    <section>
      <PageHeader>
        <Column span={10}>
          <Row>
            <h1>Applications</h1>
            <SearchFilter
              key="searchFilter"
              placeholder="search apps..."
              onChange={setFilter}
              value={filter}
              onSubmit={pushSearchFilter}
            />
          </Row>
        </Column>
        <Column span={2}>
          <div className="header-button">
            <Link to={url.app.catalog(cluster, namespace)}>
              <CdsButton status="primary">
                <CdsIcon shape="deploy" inverse={true} /> Deploy
              </CdsButton>
            </Link>
          </div>
        </Column>
      </PageHeader>
      <LoadingWrapper loaded={!isFetching && !isFetchingResources}>
        {error ? (
          <Alert theme="danger">Unable to list apps: {error.message}</Alert>
        ) : (
          <AppListGrid
            appList={listOverview}
            customResources={customResources}
            cluster={cluster}
            namespace={namespace}
            appVersion={appVersion}
            filter={filter}
            csvs={csvs}
          />
        )}
      </LoadingWrapper>
    </section>
  );
}

export default AppList;
