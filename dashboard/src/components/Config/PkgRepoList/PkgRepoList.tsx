// Copyright 2018-2023 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { CdsButton } from "@cds/react/button";
import { CdsIcon } from "@cds/react/icon";
import { CdsToggle, CdsToggleGroup } from "@cds/react/toggle";
import actions from "actions";
import AlertGroup from "components/AlertGroup";
import { filterNames, filtersToQuery } from "components/Catalog/Catalog";
import LoadingWrapper from "components/LoadingWrapper";
import PageHeader from "components/PageHeader/PageHeader";
import Table from "components/Table";
import {
  PackageRepositoriesPermissions,
  PackageRepositoryReference,
  PackageRepositorySummary,
} from "gen/kubeappsapis/core/packages/v1alpha1/repositories_pb";
import { Plugin } from "gen/kubeappsapis/core/plugins/v1alpha1/plugins_pb";
import { usePush } from "hooks/push";
import qs from "qs";
import { useCallback, useEffect, useState } from "react";
import { useDispatch, useSelector } from "react-redux";
import { Link, useLocation } from "react-router-dom";
import { Tooltip } from "react-tooltip";
import { IStoreState } from "shared/types";
import { app } from "shared/url";
import { getPluginName } from "shared/utils";
import { PkgRepoAddButton } from "./PkgRepoButton";
import { PkgRepoControl } from "./PkgRepoControl";
import { PkgRepoDisabledControl } from "./PkgRepoDisabledControl";
import "./PkgRepoList.css";
function PkgRepoList() {
  const dispatch = useDispatch();
  const location = useLocation();
  const {
    repos: { errors, isFetching, reposSummaries: repos, reposPermissions },
    clusters: { clusters, currentCluster },
    config: { kubeappsCluster, helmGlobalNamespace, carvelGlobalNamespace },
  } = useSelector((state: IStoreState) => state);
  const cluster = currentCluster;
  const { currentNamespace } = clusters[cluster];
  const allNSQuery =
    qs.parse(location.search, { ignoreQueryPrefix: true }).allns === "yes" ? true : false;
  const [allNS, setAllNS] = useState(allNSQuery);
  const [canSetAllNS, setCanSetAllNS] = useState(false);
  const [reposRBAC, setReposRBAC] = useState(new Map<string, PackageRepositoriesPermissions>());
  const [namespace, setNamespace] = useState(allNSQuery ? "" : currentNamespace);

  // We do not currently support package repositories on additional clusters.
  const supportedCluster = cluster === kubeappsCluster;
  // useCallback stores the reference to the function, not the function execution
  // so calling several times to refetchRepos would run the code inside, even
  // if the dependencies do not change.
  const refetchRepos: () => void = useCallback(() => {
    if (
      !namespace ||
      !supportedCluster ||
      [helmGlobalNamespace, carvelGlobalNamespace].includes(namespace)
    ) {
      // All Namespaces. Global namespace or other cluster, show global repos only
      dispatch(actions.repos.fetchRepoSummaries(""));
      return () => {};
    }
    // In other case, fetch global and namespace repos
    dispatch(actions.repos.fetchRepoSummaries(namespace, true));
    return () => {};
  }, [dispatch, supportedCluster, namespace, helmGlobalNamespace, carvelGlobalNamespace]);

  useEffect(() => {
    refetchRepos();
  }, [refetchRepos]);

  const fecthPermissions: () => void = useCallback(() => {
    if (supportedCluster) {
      dispatch(actions.repos.fetchReposPermissions(kubeappsCluster, namespace));
    }
    return () => {};
  }, [dispatch, supportedCluster, kubeappsCluster, namespace]);

  useEffect(() => {
    fecthPermissions();
  }, [fecthPermissions]);

  const push = usePush();
  const submitFilters = (allns: boolean) => {
    if (allns) {
      push("?allns=yes");
    } else {
      push("?allns=no");
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
    const rbac = new Map<string, PackageRepositoriesPermissions>();
    reposPermissions.forEach(p => rbac.set(JSON.stringify(p.plugin), p));
    setReposRBAC(rbac);

    // Cluster-wide check
    setCanSetAllNS([...rbac.values()].some(r => r.global["list"]));
  }, [cluster, reposPermissions]);

  const canEditGlobalRepos = (plugin?: Plugin): boolean => {
    if (!plugin) {
      return false;
    }
    return reposRBAC.get(JSON.stringify(plugin))?.global["update"] || false;
  };

  const canEditNamespacedRepos = (plugin?: Plugin): boolean => {
    if (!plugin) {
      return false;
    }
    return reposRBAC.get(JSON.stringify(plugin))?.namespace["update"] || false;
  };

  const canAddRepos = () => {
    return [...reposRBAC.values()].some(r => r.global["create"] || r.namespace["create"]);
  };

  const globalRepos: PackageRepositorySummary[] = [];
  const namespacedRepos: PackageRepositorySummary[] = [];
  repos.forEach(repo => {
    if (!repo.namespaceScoped) {
      globalRepos.push(repo);
      // ensure listed namespaced repos are those in the current namespace
    } else if (allNS || repo.packageRepoRef?.context?.namespace === namespace) {
      namespacedRepos.push(repo);
    }
  });

  const tableColumns = [
    { accessor: "name", Header: "Name" },
    { accessor: "url", Header: "URL" },
    { accessor: "packageFormat", Header: "Packaging Format" },
    { accessor: "accessLevel", Header: "Access Level" },
    { accessor: "namespace", Header: "Namespace" },
    { accessor: "status", Header: "Status" },
    { accessor: "actions", Header: "Actions" },
  ];
  const getTableData = (
    targetRepos: PackageRepositorySummary[],
    disableControls: (repoRef?: PackageRepositoryReference) => boolean,
  ) => {
    return targetRepos.map(repo => {
      return {
        name: getRepoNameLinkAndTooltip(cluster, repo),
        url: repo.url,
        accessLevel: repo.requiresAuth ? "Private" : "Public",
        namespace: repo.packageRepoRef?.context?.namespace,
        packageFormat: `${getPluginName(repo.packageRepoRef?.plugin)} (${repo.type})`,
        status: repo.status?.ready ? (
          <>Ready</>
        ) : (
          <>
            <CdsButton action="flat-inline" onClick={refetchRepos}>
              <CdsIcon shape="refresh" />
              Refresh
            </CdsButton>
            <p>Not ready</p>
            {repo?.status?.userReason && (
              <>
                <span data-tooltip-id={`${repo.name}-notready-tooltip`}>
                  <CdsIcon shape="info-circle" size="sm" solid={true} />
                </span>
                <Tooltip id={`${repo.name}-notready-tooltip`} place="top-end">
                  {repo?.status?.userReason}
                </Tooltip>
              </>
            )}
          </>
        ),
        actions: disableControls(repo.packageRepoRef) ? (
          <PkgRepoDisabledControl />
        ) : (
          <PkgRepoControl
            repo={repo}
            refetchRepos={refetchRepos}
            helmGlobalNamespace={helmGlobalNamespace}
            carvelGlobalNamespace={carvelGlobalNamespace}
          />
        ),
      };
    });
  };

  const getGlobalReposTable = (
    globalRepos: PackageRepositorySummary[],
    disableControls: boolean,
  ) => {
    return (
      <>
        <h3>Global Repositories:</h3>
        <p>Global Package Repositories are available for all Kubeapps users.</p>
        {globalRepos.length ? (
          <Table
            valign="center"
            columns={tableColumns}
            data={getTableData(
              globalRepos,
              repoRef => disableControls || !canEditGlobalRepos(repoRef?.plugin),
            )}
          />
        ) : (
          <p>
            There are no <i>global</i> Package Repositories yet. Click on the "Add Package
            Repository" button to create one.
          </p>
        )}
      </>
    );
  };

  /* eslint-disable jsx-a11y/label-has-associated-control */
  return (
    <>
      <PageHeader
        title="Package Repositories"
        buttons={[
          <PkgRepoAddButton
            title="Add a Package Repository"
            key="add-repo-button"
            namespace={currentNamespace}
            helmGlobalNamespace={helmGlobalNamespace}
            carvelGlobalNamespace={carvelGlobalNamespace}
            disabled={!supportedCluster || !canAddRepos()}
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
          <div className="page-content">
            <AlertGroup status="warning" closable={false}>
              Package Repositories can't be managed from this cluster.
              <br />
              <br />
              Currently, the <b>
                Package Repositories must be managed from the default cluster
              </b>{" "}
              (the one on which Kubeapps has been installed).
              <br />
              <br />
              Any <i>global</i> Package Repository defined in the default cluster can be later used
              across any target cluster.
              <br />
              However, <i>namespaced</i> Package Repositories can only be used on the default
              cluster.
            </AlertGroup>
            {getGlobalReposTable(globalRepos, true)}
          </div>
        ) : (
          <div className="page-content">
            {errors.fetch && (
              <AlertGroup status="danger">
                An error occurred while fetching repositories: {errors.fetch.message}.
              </AlertGroup>
            )}
            {errors.delete && (
              <AlertGroup status="danger">
                An error occurred while deleting the repository: {errors.delete.message}.
              </AlertGroup>
            )}
            {!errors.fetch && (
              <>
                <LoadingWrapper
                  className="margin-t-xxl"
                  loadingText="Fetching Package Repositories..."
                  loaded={!isFetching}
                >
                  {getGlobalReposTable(globalRepos, !canEditGlobalRepos)}
                  {![helmGlobalNamespace, carvelGlobalNamespace].includes(namespace) && (
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
                          data={getTableData(
                            namespacedRepos,
                            repoRef => !canEditNamespacedRepos(repoRef?.plugin),
                          )}
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

function getRepoNameLinkAndTooltip(cluster: string, repo: PackageRepositorySummary) {
  const linkObj = (
    <Link
      to={
        app.catalog(cluster, repo.packageRepoRef?.context?.namespace || "") +
        filtersToQuery({ [filterNames.REPO]: [repo.name] })
      }
    >
      {repo.name}
    </Link>
  );
  return repo.description ? (
    <div className="color-icon-info">
      <span data-tooltip-id={`${repo.name}-pending-tooltip`} className="tooltip-wrapper">
        {linkObj} <CdsIcon shape="info-circle" size="sm" solid={true} />
        <Tooltip id={`${repo.name}-pending-tooltip`} place="top-end">
          {repo.description}
        </Tooltip>
      </span>
    </div>
  ) : (
    linkObj
  );
}

export default PkgRepoList;
