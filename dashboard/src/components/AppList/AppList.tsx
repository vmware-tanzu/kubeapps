// Copyright 2018-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { CdsButton } from "@cds/react/button";
import { CdsIcon } from "@cds/react/icon";
import { CdsToggle, CdsToggleGroup } from "@cds/react/toggle";
import actions from "actions";
import ErrorAlert from "components/ErrorAlert";
import LoadingWrapper from "components/LoadingWrapper/LoadingWrapper";
import { push } from "connected-react-router";
import qs from "qs";
import { useEffect, useState } from "react";
import { useDispatch, useSelector } from "react-redux";
import * as ReactRouter from "react-router-dom";
import { Link } from "react-router-dom";
import { Kube } from "shared/Kube";
import { IStoreState } from "shared/types";
import * as url from "shared/url";
import PageHeader from "../PageHeader/PageHeader";
import SearchFilter from "../SearchFilter/SearchFilter";
import "./AppList.css";
import AppListGrid from "./AppListGrid";

function AppList() {
  const location = ReactRouter.useLocation();
  const searchQuery = qs.parse(location.search, { ignoreQueryPrefix: true }).q?.toString() || "";
  const allNSQuery = qs.parse(location.search, { ignoreQueryPrefix: true }).allns || "";
  const dispatch = useDispatch();

  const {
    apps: { error, isFetching, listOverview },
    clusters: { clusters, currentCluster },
    operators: { isFetching: isFetchingResources, resources: customResources, csvs },
    config: { appVersion, featureFlags },
  } = useSelector((state: IStoreState) => state);
  const cluster = currentCluster;
  const { currentNamespace } = clusters[cluster];

  const [searchFilter, setSearchFilter] = useState("");
  const [allNS, setAllNS] = useState(false);
  const [canSetAllNS, setCanSetAllNS] = useState(false);
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
  }, [currentNamespace]);

  useEffect(() => {
    if (allNS) {
      setNamespace("");
    } else {
      setNamespace(currentNamespace);
    }
  }, [allNS, currentNamespace]);

  useEffect(() => {
    dispatch(actions.installedpackages.fetchInstalledPackages(cluster, namespace));
    if (featureFlags?.operators) {
      dispatch(actions.operators.getResources(cluster, namespace));
    }
  }, [dispatch, cluster, namespace, featureFlags]);

  useEffect(() => {
    // In order to be able to list applications in all namespaces, it's necessary to be able
    // to list/get secrets in all of them.
    Kube.canI(cluster, "", "secrets", "list", "")
      .then(allowed => setCanSetAllNS(allowed))
      ?.catch(() => setCanSetAllNS(false));
  }, [cluster]);

  useEffect(() => {
    setSearchFilter(searchQuery);
  }, [searchQuery]);

  useEffect(() => {
    setAllNS(allNSQuery === "yes" ? true : false);
  }, [allNSQuery]);

  /* eslint-disable jsx-a11y/label-has-associated-control */
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
            {canSetAllNS && (
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
            )}
          </>
        }
        buttons={[
          <Link to={url.app.catalog(cluster, namespace)} key="deploy-button">
            <CdsButton status="primary">
              <CdsIcon shape="deploy" /> Deploy
            </CdsButton>
          </Link>,
        ]}
      />
      <LoadingWrapper
        loaded={!isFetching && !isFetchingResources}
        loadingText="Getting the list of applications..."
        className="margin-t-xl"
      >
        {error ? (
          <ErrorAlert error={error} />
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
