// Copyright 2018-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { CdsButton } from "@cds/react/button";
import { CdsIcon } from "@cds/react/icon";
import { CdsToggle, CdsToggleGroup } from "@cds/react/toggle";
import actions from "actions";
import { filterNames, filtersToQuery } from "components/Catalog/Catalog";
import Alert from "components/js/Alert";
import Table from "components/js/Table";
import Tooltip from "components/js/Tooltip";
import PageHeader from "components/PageHeader/PageHeader";
import { push } from "connected-react-router";
import qs from "qs";
import { useCallback, useEffect, useState } from "react";
import { useDispatch, useSelector } from "react-redux";
import { Link, useLocation } from "react-router-dom";
import { Kube } from "shared/Kube";
import { IAppRepository, IStoreState } from "shared/types";
import { app } from "shared/url";
import LoadingWrapper from "../../LoadingWrapper/LoadingWrapper";
import { AppRepoAddButton } from "./AppRepoButton";
import { AppRepoControl } from "./AppRepoControl";
import { AppRepoDisabledControl } from "./AppRepoDisabledControl";
import "./AppRepoList.css";

function AppRepoList() {
  const dispatch = useDispatch();
  const location = useLocation();
  const {
    repos: { errors, isFetchingElem, repos },
    clusters: { clusters, currentCluster },
    config: { kubeappsCluster, kubeappsNamespace, globalReposNamespace },
  } = useSelector((state: IStoreState) => state);
  const cluster = currentCluster;
  const { currentNamespace } = clusters[cluster];
  const allNSQuery =
    qs.parse(location.search, { ignoreQueryPrefix: true }).allns === "yes" ? true : false;
  const [allNS, setAllNS] = useState(allNSQuery);
  const [canSetAllNS, setCanSetAllNS] = useState(false);
  const [canEditGlobalRepos, setCanEditGlobalRepos] = useState(false);
  const [namespace, setNamespace] = useState(allNSQuery ? "" : currentNamespace);

  // We do not currently support package repositories on additional clusters.
  const supportedCluster = cluster === kubeappsCluster;
  // useCallback stores the reference to the function, not the function execution
  // so calling several times to refetchRepos would run the code inside, even
  // if the dependencies do not change.
  const refetchRepos: () => void = useCallback(() => {
    if (!namespace) {
      // All Namespaces
      dispatch(actions.repos.fetchRepos(""));
      return () => {};
    }
    if (!supportedCluster || namespace === globalReposNamespace) {
      // Global namespace or other cluster, show global repos only
      dispatch(actions.repos.fetchRepos(globalReposNamespace));
      return () => {};
    }
    // In other case, fetch global and namespace repos
    dispatch(actions.repos.fetchRepos(namespace, true));
    return () => {};
  }, [dispatch, supportedCluster, namespace, globalReposNamespace]);

  useEffect(() => {
    refetchRepos();
  }, [refetchRepos]);

  const submitFilters = (allns: boolean) => {
    if (allns) {
      dispatch(push("?allns=yes"));
    } else {
      dispatch(push("?allns=no"));
    }
  };
  const toggleListAllNS = () => {
    submitFilters(!allNS);
    setAllNS(!allNS);
  };
  useEffect(() => {
    if (allNS) {
      setNamespace("");
    } else {
      setNamespace(currentNamespace);
    }
  }, [allNS, currentNamespace]);

  useEffect(() => {
    Kube.canI(cluster, "kubeapps.com", "apprepositories", "list", "")
      .then(allowed => setCanSetAllNS(allowed))
      ?.catch(() => setCanSetAllNS(false));
    Kube.canI(kubeappsCluster, "kubeapps.com", "apprepositories", "update", globalReposNamespace)
      .then(allowed => setCanEditGlobalRepos(allowed))
      ?.catch(() => setCanEditGlobalRepos(false));
  }, [cluster, kubeappsCluster, kubeappsNamespace, globalReposNamespace]);

  const globalRepos: IAppRepository[] = [];
  const namespacedRepos: IAppRepository[] = [];
  repos.forEach(repo => {
    repo.metadata.namespace === globalReposNamespace
      ? globalRepos.push(repo)
      : namespacedRepos.push(repo);
  });

  const tableColumns = [
    { accessor: "name", Header: "Name" },
    { accessor: "url", Header: "URL" },
    { accessor: "plugin", Header: "Package Type" },
    { accessor: "accessLevel", Header: "Access Level" },
    { accessor: "namespace", Header: "Namespace" },
    { accessor: "status", Header: "Status" },
    { accessor: "actions", Header: "Actions" },
  ];
  const getTableData = (targetRepos: IAppRepository[], disableControls: boolean) => {
    return targetRepos.map(repo => {
      return {
        name: getRepoNameLinkAndTooltip(cluster, repo),
        url: repo.spec?.url,
        // TODO(agamez): the PackageRepositorySummary API doesn't expose this field. It will be added in upcoming PRs; in the meantime, set to "unknown"
        // accessLevel: repo.type?.auth?.header ? "Private" : "Public",
        accessLevel: "unknown",
        namespace: repo.metadata.namespace,
        plugin: "Helm",
        // eslint-disable-next-line no-constant-condition
        status: true ? (
          <>Ready</>
        ) : (
          <>
            <CdsButton action="flat-inline" onClick={refetchRepos}>
              <CdsIcon shape="refresh" />
              Refresh
            </CdsButton>
            <p>Not ready</p>
            {false && (
              <Tooltip
                label="notready-tooltip"
                id={`${repo.metadata.name}-notready-tooltip`}
                icon="info-circle"
                position="top-right"
                small={true}
                iconProps={{ solid: true, size: "sm" }}
              >
                {"unknown"}
              </Tooltip>
            )}
          </>
        ),
        actions: disableControls ? (
          <AppRepoDisabledControl />
        ) : (
          <AppRepoControl
            repo={repo}
            refetchRepos={refetchRepos}
            kubeappsNamespace={globalReposNamespace}
          />
        ),
      };
    });
  };

  /* eslint-disable jsx-a11y/label-has-associated-control */
  return (
    <>
      <PageHeader
        title="Package Repositories"
        buttons={[
          <AppRepoAddButton
            title="Add a Package Repository"
            key="add-repo-button"
            namespace={currentNamespace}
            kubeappsNamespace={globalReposNamespace}
          />,
        ]}
        filter={
          canSetAllNS ? (
            <CdsToggleGroup className="flex-v-center">
              <CdsToggle>
                <label>Show repositories in all namespaces</label>
                <input type="checkbox" onChange={toggleListAllNS} checked={allNS} />
              </CdsToggle>
            </CdsToggleGroup>
          ) : (
            <></>
          )
        }
      />
      <div className="catalog-container">
        {!supportedCluster ? (
          <Alert theme="warning">
            <h5>Package Repositories can't be managed from this cluster.</h5>
            <p>
              Currently, the Package Repositories must be managed from the default cluster (the one
              on which Kubeapps has been installed).
            </p>
            <p>
              Any <i>global</i> Package Repository defined in the default cluster can be later used
              across any target cluster.
              <br />
              However, <i>namespaced</i> Package Repositories can only be used on the default
              cluster.
            </p>
          </Alert>
        ) : (
          <div className="page-content">
            {errors.fetch && (
              <Alert theme="danger">
                An error occurred while fetching repositories: {errors.fetch.message}
              </Alert>
            )}
            {errors.delete && (
              <Alert theme="danger">
                An error occurred while deleting the repository: {errors.delete.message}
              </Alert>
            )}
            {!errors.fetch && (
              <>
                <LoadingWrapper
                  className="margin-t-xxl"
                  loadingText="Fetching Package Repositories..."
                  loaded={!isFetchingElem.repositories}
                >
                  <h3>Global Repositories:</h3>
                  <p>Global Package Repositories are available for all Kubeapps users.</p>
                  {globalRepos.length ? (
                    <Table
                      valign="center"
                      columns={tableColumns}
                      data={getTableData(globalRepos, !canEditGlobalRepos)}
                    />
                  ) : (
                    <p>
                      There are no <i>global</i> Package Repositories yet. Click on the "Add Package
                      Repository" button to create one.
                    </p>
                  )}
                  {namespace !== globalReposNamespace && (
                    <>
                      <h3>Namespaced Repositories: {namespace}</h3>
                      <p>
                        Namespaced Package Repositories are available in their namespace only. To
                        switch to a different one, use the "Current Context" selector in the top
                        navigation.
                      </p>
                      {namespacedRepos.length ? (
                        <Table
                          valign="center"
                          columns={tableColumns}
                          data={getTableData(namespacedRepos, false)}
                        />
                      ) : (
                        <p>
                          There are no <i>namespaced</i> Package Repositories in the '{namespace}'
                          namespace yet. Click on the "Add Package Repository" button to create one.
                        </p>
                      )}
                    </>
                  )}
                </LoadingWrapper>
              </>
            )}
          </div>
        )}
      </div>
    </>
  );
}

function getRepoNameLinkAndTooltip(cluster: string, repo: IAppRepository) {
  const linkObj = (
    <Link
      to={
        app.catalog(cluster, repo.metadata.namespace) +
        filtersToQuery({ [filterNames.REPO]: [repo.metadata.name] })
      }
    >
      {repo.metadata.name}
    </Link>
  );
  return repo.spec?.description ? (
    <div className="color-icon-info">
      <span className="tooltip-wrapper">
        {linkObj}
        <Tooltip
          label="pending-tooltip"
          id={`${repo.metadata.name}-pending-tooltip`}
          icon="info-circle"
          position="bottom-left"
          small={true}
          iconProps={{ solid: true, size: "sm" }}
        >
          {repo.spec?.description}
        </Tooltip>
      </span>
    </div>
  ) : (
    linkObj
  );
}

export default AppRepoList;
