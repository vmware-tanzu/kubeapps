import React, { useCallback, useEffect, useState } from "react";

import { CdsToggle, CdsToggleGroup } from "@clr/react/toggle";
import actions from "actions";
import { filterNames, filtersToQuery } from "components/Catalog/Catalog";
import Alert from "components/js/Alert";
import Table from "components/js/Table";
import PageHeader from "components/PageHeader/PageHeader";
import { push } from "connected-react-router";
import * as qs from "qs";
import { useDispatch, useSelector } from "react-redux";
import { useLocation } from "react-router";
import { Link } from "react-router-dom";
import { Kube } from "shared/Kube";
import { app } from "shared/url";
import { IAppRepository, IStoreState } from "../../../shared/types";
import LoadingWrapper from "../../LoadingWrapper/LoadingWrapper";
import { AppRepoAddButton } from "./AppRepoButton";
import { AppRepoControl } from "./AppRepoControl";
import { AppRepoDisabledControl } from "./AppRepoDisabledControl";
import "./AppRepoList.css";
import { AppRepoRefreshAllButton } from "./AppRepoRefreshAllButton";

function AppRepoList() {
  const dispatch = useDispatch();
  const location = useLocation();
  const {
    repos: { errors, isFetching, repos, repoSecrets },
    clusters: { clusters, currentCluster },
    config: { kubeappsCluster, kubeappsNamespace },
  } = useSelector((state: IStoreState) => state);
  const cluster = currentCluster;
  const { currentNamespace } = clusters[cluster];
  const allNSQuery =
    qs.parse(location.search, { ignoreQueryPrefix: true }).allns === "yes" ? true : false;
  const [allNS, setAllNS] = useState(allNSQuery);
  const [canSetAllNS, setCanSetAllNS] = useState(false);
  const [canEditKubeappsRepos, setCanEditKubeappsRepos] = useState(false);
  const [namespace, setNamespace] = useState(allNSQuery ? "" : currentNamespace);

  // We do not currently support app repositories on additional clusters.
  const supportedCluster = cluster === kubeappsCluster;
  const refetchRepos: () => void = useCallback(() => {
    if (!namespace) {
      // All Namespaces
      dispatch(actions.repos.fetchRepos(""));
      return;
    }
    if (!supportedCluster || namespace === kubeappsNamespace) {
      // Global namespace or other cluster, show global repos only
      dispatch(actions.repos.fetchRepos(kubeappsNamespace));
      return;
    }
    // In other case, fetch global and namespace repos
    dispatch(actions.repos.fetchRepos(namespace, kubeappsNamespace));
  }, [dispatch, supportedCluster, namespace, kubeappsNamespace]);

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
    Kube.canI(cluster, "kubeapps.com", "apprepositories", "list", "").then(allowed =>
      setCanSetAllNS(allowed),
    );
    Kube.canI(
      cluster,
      "kubeapps.com",
      "apprepositories",
      "update",
      kubeappsNamespace,
    ).then(allowed => setCanEditKubeappsRepos(allowed));
  }, [cluster, kubeappsNamespace]);

  useEffect(() => {
    if (repos) {
      dispatch(actions.repos.fetchImagePullSecrets(namespace));
    }
  }, [dispatch, repos, namespace]);

  const globalRepos: IAppRepository[] = [];
  const namespaceRepos: IAppRepository[] = [];
  repos.forEach(repo => {
    repo.metadata.namespace === kubeappsNamespace
      ? globalRepos.push(repo)
      : namespaceRepos.push(repo);
  });

  const tableColumns = [
    { accessor: "name", Header: "Name" },
    { accessor: "url", Header: "URL" },
    { accessor: "accessLevel", Header: "Access Level" },
    { accessor: "namespace", Header: "Namespace" },
    { accessor: "actions", Header: "Actions" },
  ];
  const getTableData = (targetRepos: IAppRepository[], disableControls: boolean) => {
    return targetRepos.map(repo => {
      return {
        name: (
          <Link
            to={
              app.catalog(cluster, repo.metadata.namespace) +
              filtersToQuery({ [filterNames.REPO]: [repo.metadata.name] })
            }
          >
            {repo.metadata.name}
          </Link>
        ),
        url: repo.spec?.url,
        accessLevel: repo.spec?.auth?.header ? "Private" : "Public",
        namespace: repo.metadata.namespace,
        actions: disableControls ? (
          <AppRepoDisabledControl />
        ) : (
          <AppRepoControl
            repo={repo}
            secret={repoSecrets.find(secret =>
              secret.metadata.ownerReferences?.some(
                ownerRef => ownerRef.name === repo.metadata.name,
              ),
            )}
            refetchRepos={refetchRepos}
            kubeappsNamespace={kubeappsNamespace}
          />
        ),
      };
    });
  };
  return (
    <>
      <PageHeader
        title="Application Repositories"
        buttons={[
          <AppRepoAddButton
            key="add-repo-button"
            namespace={currentNamespace}
            kubeappsNamespace={kubeappsNamespace}
          />,
          <AppRepoRefreshAllButton key="refresh-all-button" />,
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
      {!supportedCluster ? (
        <Alert theme="warning">
          <h5>App Repositories are available on the default cluster only</h5>
          <p>
            Currently the multi-cluster support in Kubeapps supports AppRepositories on the default
            cluster only.
          </p>
          <p>
            The catalog of charts from AppRepositories on the default cluster which are available
            for all namespaces will be avaialble on additional clusters also, but you can not
            currently create a private AppRepository for a particular namespace of an additional
            cluster. We may in the future support AppRepositories on additional clusters but for now
            you will need to switch back to your default cluster.
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
              <LoadingWrapper loaded={!isFetching}>
                <h3>Global Repositories:</h3>
                <p>Global repositories are available for all Kubeapps users.</p>
                {globalRepos.length ? (
                  <Table
                    valign="center"
                    columns={tableColumns}
                    data={getTableData(globalRepos, !canEditKubeappsRepos)}
                  />
                ) : (
                  <p>No global repositories found.</p>
                )}
                {namespace !== kubeappsNamespace && (
                  <>
                    <h3>Namespace Repositories: {namespace}</h3>
                    <p>
                      Namespaced Repositories are available in their namespace only. To switch to a
                      different one, use the "Current Context" selector in the top navigation.
                    </p>
                    {namespaceRepos.length ? (
                      <Table
                        valign="center"
                        columns={tableColumns}
                        data={getTableData(namespaceRepos, false)}
                      />
                    ) : (
                      <p>
                        The current namespace doesn't have any repositories. Click on the button
                        "Add app repository" above to create the first one.
                      </p>
                    )}
                  </>
                )}
              </LoadingWrapper>
            </>
          )}
        </div>
      )}
    </>
  );
}

export default AppRepoList;
