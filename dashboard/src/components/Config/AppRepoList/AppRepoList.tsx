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
import { useLocation, Link } from "react-router-dom";
import { Kube } from "shared/Kube";
import { IAppRepository, IStoreState } from "shared/types";
import { app } from "shared/url";
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

  // We do not currently support app repositories on additional clusters.
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
    Kube.canI(cluster, "kubeapps.com", "apprepositories", "list", "").then(allowed =>
      setCanSetAllNS(allowed),
    );
    Kube.canI(
      kubeappsCluster,
      "kubeapps.com",
      "apprepositories",
      "update",
      globalReposNamespace,
    ).then(allowed => setCanEditGlobalRepos(allowed));
  }, [cluster, kubeappsCluster, kubeappsNamespace, globalReposNamespace]);

  const globalRepos: IAppRepository[] = [];
  const namespaceRepos: IAppRepository[] = [];
  repos.forEach(repo => {
    repo.metadata.namespace === globalReposNamespace
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
        name: getRepoNameLinkAndTooltip(cluster, repo),
        url: repo.spec?.url,
        accessLevel: repo.spec?.auth?.header ? "Private" : "Public",
        namespace: repo.metadata.namespace,
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
        title="Application Repositories"
        buttons={[
          <AppRepoAddButton
            title="Add an App Repository"
            key="add-repo-button"
            namespace={currentNamespace}
            kubeappsNamespace={globalReposNamespace}
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
            The catalog of packages from AppRepositories on the default cluster which are available
            for all namespaces will be available on additional clusters also, but you can not
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
              <LoadingWrapper
                className="margin-t-xxl"
                loadingText="Fetching Application Repositories..."
                loaded={!isFetchingElem.repositories}
              >
                <h3>Global Repositories:</h3>
                <p>Global repositories are available for all Kubeapps users.</p>
                {globalRepos.length ? (
                  <Table
                    valign="center"
                    columns={tableColumns}
                    data={getTableData(globalRepos, !canEditGlobalRepos)}
                  />
                ) : (
                  <p>No global repositories found.</p>
                )}
                {namespace !== globalReposNamespace && (
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
