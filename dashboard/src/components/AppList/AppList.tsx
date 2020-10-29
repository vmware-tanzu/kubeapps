import { CdsButton } from "@clr/react/button";
import { CdsIcon } from "@clr/react/icon";
import * as React from "react";
import { useEffect, useState } from "react";
import { Link } from "react-router-dom";

import { IAppState, IClusterServiceVersion, IResource } from "../../shared/types";
import * as url from "../../shared/url";
import PageHeader from "../PageHeader/PageHeader";
import SearchFilter from "../SearchFilter/SearchFilter";

import Alert from "components/js/Alert";
import LoadingWrapper from "components/LoadingWrapper/LoadingWrapper";
import "./AppList.css";
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
  appVersion: string;
}

function AppList(props: IAppListProps) {
  const [filter, setFilter] = useState("");
  const {
    fetchAppsWithUpdateInfo,
    filter: filterProps,
    namespace,
    getCustomResources,
    apps: { getError: error, isFetching, listOverview },
    cluster,
    isFetchingResources,
    pushSearchFilter,
    customResources,
    appVersion,
    csvs,
  } = props;
  const submitFilters = () => pushSearchFilter(filter);

  useEffect(() => {
    fetchAppsWithUpdateInfo(cluster, namespace, true);
    getCustomResources(cluster, namespace);
  }, [cluster, namespace, fetchAppsWithUpdateInfo, getCustomResources]);

  useEffect(() => {
    setFilter(filterProps);
  }, [filterProps]);

  return (
    <section>
      <PageHeader
        title="Applications"
        filter={
          <SearchFilter
            key="searchFilter"
            placeholder="search apps..."
            onChange={setFilter}
            value={filter}
            submitFilters={submitFilters}
          />
        }
        buttons={[
          <Link to={url.app.catalog(cluster, namespace)} key="deploy-button">
            <CdsButton status="primary">
              <CdsIcon shape="deploy" inverse={true} /> Deploy
            </CdsButton>
          </Link>,
        ]}
      />
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
