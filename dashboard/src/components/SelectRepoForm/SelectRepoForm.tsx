import { CdsButton } from "@cds/react/button";
import actions from "actions";
import Alert from "components/js/Alert";
import { InstalledPackageDetail } from "gen/kubeappsapis/core/packages/v1alpha1/packages";
import { useEffect, useState } from "react";
import { useDispatch, useSelector } from "react-redux";
import { Link } from "react-router-dom";
import { IStoreState } from "shared/types";
import * as url from "shared/url";
import LoadingWrapper from "../LoadingWrapper/LoadingWrapper";
import "./SelectRepoForm.css";

interface ISelectRepoFormProps {
  cluster: string;
  namespace: string;
  app?: InstalledPackageDetail;
}

function SelectRepoForm({ cluster, namespace, app }: ISelectRepoFormProps) {
  const dispatch = useDispatch();
  const {
    repos: {
      isFetching,
      repos,
      repo,
      errors: { fetch: fetchError },
    },
    packages: {
      selected: { error: packageError },
    },
    config: { kubeappsNamespace, kubeappsCluster },
  } = useSelector((state: IStoreState) => state);

  const [userRepoName, setUserRepoName] = useState(repo?.metadata?.name ?? "");
  const [userRepoNamespace, setUserRepoNamepace] = useState(repo?.metadata?.namespace ?? "");

  useEffect(() => {
    if (namespace !== kubeappsNamespace) {
      // Normal namespace, show local and global repos
      dispatch(actions.repos.fetchRepos(namespace, true));
    } else {
      // Global namespace, show global repos only
      dispatch(actions.repos.fetchRepos(kubeappsNamespace));
    }
  }, [dispatch, namespace, kubeappsNamespace]);

  const handleRepoNameChange = (e: React.ChangeEvent<HTMLSelectElement>) => {
    if (e.target.value) {
      const [ns, name] = e.target.value.split("/");
      setUserRepoName(name);
      setUserRepoNamepace(ns);
    }
  };

  const handleRepoNameSelection = async () => {
    if (userRepoNamespace && userRepoName) {
      dispatch(
        actions.repos.findPackageInRepo(kubeappsCluster, userRepoNamespace, userRepoName, app),
      );
    }
  };

  const findRepo = (ns: string, name: string) => {
    return repos.find(r => r.metadata.name === name && r.metadata.namespace === ns);
  };

  const getRepoURL = (ns: string, name: string) => {
    const r = findRepo(ns, name);
    return r && r.spec ? r.spec.url : "";
  };

  return (
    <LoadingWrapper
      className="margin-t-xxl"
      loadingText="Fetching Application Repositories..."
      loaded={!isFetching}
    >
      {fetchError && <Alert theme="danger">An error occurred: {fetchError.message}</Alert>}
      {!fetchError && repos.length === 0 && (
        <Alert theme="warning">
          <h5>Repositories not found. </h5>
          Manage your repositories in Kubeapps by visiting the{" "}
          <Link to={url.app.config.apprepositories(cluster, namespace)}>
            App repositories configuration
          </Link>{" "}
          page.
        </Alert>
      )}
      {repos.length > 0 && (
        <div className="select-repo-form">
          {packageError && <Alert theme="danger">An error occurred: {packageError.message}</Alert>}
          <h2>
            Select the source repository of '{app?.availablePackageRef?.identifier ?? app?.name}'
          </h2>
          <label className="select-repo-form-label" htmlFor="repoNameSelector">
            Repository Name *
          </label>
          <div className="clr-select-wrapper">
            <select
              id="repoNameSelector"
              onChange={handleRepoNameChange}
              value={`${userRepoNamespace}/${userRepoName}`}
              required={true}
              className="clr-page-size-select"
            >
              {!userRepoName && <option key="" value="" />}
              {repos.map(r => {
                const value = `${r.metadata.namespace}/${r.metadata.name}`;
                return (
                  <option key={value} value={value}>
                    {value} ({getRepoURL(r.metadata.namespace, r.metadata.name)})
                  </option>
                );
              })}
            </select>
            <CdsButton size="sm" action="flat" onClick={handleRepoNameSelection} type="button">
              Select
            </CdsButton>
          </div>
          <p>
            {" "}
            * If the repository containing '{app?.availablePackageRef?.identifier ?? app?.name}' is
            not in the list, add it{" "}
            <Link to={url.app.config.apprepositories(cluster, namespace)}>here</Link>.{" "}
          </p>
        </div>
      )}
    </LoadingWrapper>
  );
}

export default SelectRepoForm;
