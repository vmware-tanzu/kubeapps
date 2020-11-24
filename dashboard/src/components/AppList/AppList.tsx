import { CdsButton } from "@clr/react/button";
import { CdsIcon } from "@clr/react/icon";
import { CdsToggle, CdsToggleGroup } from "@clr/react/toggle";
import actions from "actions";
import Alert from "components/js/Alert";
import LoadingWrapper from "components/LoadingWrapper/LoadingWrapper";
import { push } from "connected-react-router";
import * as qs from "qs";
import * as React from "react";
import { useEffect, useState } from "react";
import { useDispatch, useSelector } from "react-redux";
import { useLocation } from "react-router";
import { Link } from "react-router-dom";
import { IStoreState } from "../../shared/types";
import * as url from "../../shared/url";
import PageHeader from "../PageHeader/PageHeader";
import SearchFilter from "../SearchFilter/SearchFilter";
import "./AppList.css";
import AppListGrid from "./AppListGrid";

function AppList() {
  const location = useLocation();
  const searchQuery = qs.parse(location.search, { ignoreQueryPrefix: true }).q || "";
  const allNSQuery = qs.parse(location.search, { ignoreQueryPrefix: true }).allns || "";
  const dispatch = useDispatch();

  const {
    apps: { error, isFetching, listOverview },
    clusters: { clusters, currentCluster },
    operators: { isFetching: isFetchingResources, resources: customResources, csvs },
    config: { appVersion },
  } = useSelector((state: IStoreState) => state);
  const cluster = currentCluster;
  const { currentNamespace } = clusters[cluster];

  const [searchFilter, setSearchFilter] = useState("");
  const [allNS, setAllNS] = useState(false);
  const [namespace, setNamespace] = useState(currentNamespace);
  const toggleListAllNS = () => {
    submitFilters(!allNS);
    setAllNS(!allNS);
  };

  const submitFilters = (allns: boolean) => {
    const filters = [];
    if (allns) {
      filters.push("allns=yes");
    } else {
      filters.push("allns=no");
    }
    if (searchFilter) {
      filters.push(`q=${searchFilter}`);
    }
    dispatch(push(`?${filters.join("&")}`));
  };
  const submitSearchFilter = () => submitFilters(allNS);

  useEffect(() => {
    setNamespace(currentNamespace);
    setAllNS(false);
  }, [currentNamespace]);

  useEffect(() => {
    if (allNS) {
      setNamespace("");
    } else {
      setNamespace(currentNamespace);
    }
  }, [allNS, currentNamespace]);

  useEffect(() => {
    dispatch(actions.apps.fetchAppsWithUpdateInfo(cluster, namespace));
    dispatch(actions.operators.getResources(cluster, namespace));
  }, [dispatch, cluster, namespace]);

  useEffect(() => {
    setSearchFilter(searchQuery);
  }, [searchQuery]);

  useEffect(() => {
    setAllNS(allNSQuery === "yes" ? true : false);
  }, [allNSQuery]);

  return (
    <section>
      <PageHeader
        title="Applications"
        filter={
          <>
            <SearchFilter
              key="searchFilter"
              placeholder="search apps..."
              onChange={setSearchFilter}
              value={searchFilter}
              submitFilters={submitSearchFilter}
            />
            <CdsToggleGroup className="flex-v-center">
              <CdsToggle>
                <label>Show apps in all namespaces</label>
                <input
                  type="checkbox"
                  onChange={toggleListAllNS}
                  checked={allNSQuery === "yes" || allNS}
                />
              </CdsToggle>
            </CdsToggleGroup>
          </>
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
            filter={searchFilter}
            csvs={csvs}
          />
        )}
      </LoadingWrapper>
    </section>
  );
}

export default AppList;
