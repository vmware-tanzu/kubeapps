import React, { useEffect } from "react";

import actions from "actions";
import { filterNames, filtersToQuery } from "components/Catalog/Catalog";
import Alert from "components/js/Alert";
import Table from "components/js/Table";
import PageHeader from "components/PageHeader/PageHeader";
import { useDispatch, useSelector } from "react-redux";
import { Link } from "react-router-dom";
import { app } from "shared/url";
import { IAppRepository, IStoreState } from "../../../shared/types";
import LoadingWrapper from "../../LoadingWrapper/LoadingWrapper";
import { AppRepoAddButton } from "./AppRepoButton";
import { AppRepoControl } from "./AppRepoControl";
import { AppRepoDisabledControl } from "./AppRepoDisabledControl";
import "./AppRepoList.css";
import { AppRepoRefreshAllButton } from "./AppRepoRefreshAllButton";

export interface IAppRepoListProps {
  cluster: string;
  namespace: string;
  kubeappsCluster: string;
  kubeappsNamespace: string;
}

function AppRepoList({
  cluster,
  namespace,
  kubeappsCluster,
  kubeappsNamespace,
}: IAppRepoListProps) {
  const dispatch = useDispatch();
  // We do not currently support app repositories on additional clusters.
  const supportedCluster = cluster === kubeappsCluster;

  useEffect(() => {
    if (!supportedCluster || namespace === kubeappsNamespace) {
      // If we are not in the supported cluster, only fetch global namespaces
      // TODO(andresmgot): It will likely fail fetching secrets
      dispatch(actions.repos.fetchRepos(kubeappsNamespace));
    } else {
      dispatch(actions.repos.fetchRepos(namespace, kubeappsNamespace));
    }
  }, [dispatch, namespace, kubeappsNamespace, supportedCluster]);

  const { errors, isFetching, repos, repoSecrets } = useSelector(
    (state: IStoreState) => state.repos,
  );

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
            namespace={namespace}
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
            namespace={namespace}
            kubeappsNamespace={kubeappsNamespace}
          />,
          <AppRepoRefreshAllButton key="refresh-all-button" />,
        ]}
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
                <h3>Global Repositories</h3>
                <p>
                  Global repositories are available for all Kubeapps users.{" "}
                  {kubeappsCluster &&
                    (kubeappsCluster !== cluster || namespace !== kubeappsNamespace) && (
                      <>
                        Administrators can go to the{" "}
                        <Link to={app.config.apprepositories(kubeappsCluster, kubeappsNamespace)}>
                          {kubeappsNamespace}
                        </Link>{" "}
                        namespace to manage them.
                      </>
                    )}
                </p>
                {globalRepos.length ? (
                  <Table
                    valign="center"
                    columns={tableColumns}
                    data={getTableData(globalRepos, namespace !== kubeappsNamespace)}
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
