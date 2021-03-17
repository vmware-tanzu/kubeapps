import { get } from "lodash";
import { useEffect, useState } from "react";
import { Link } from "react-router-dom";

import actions from "actions";
import Alert from "components/js/Alert";
import { useDispatch, useSelector } from "react-redux";
import * as url from "shared/url";
import { IStoreState } from "../../shared/types";
import LoadingWrapper from "../LoadingWrapper/LoadingWrapper";
import "./SelectRepoForm.css";

interface ISelectRepoFormProps {
  cluster: string;
  namespace: string;
  chartName: string;
}

function SelectRepoForm({ cluster, namespace, chartName }: ISelectRepoFormProps) {
  const dispatch = useDispatch();
  const {
    repos: {
      isFetching,
      repos,
      repo,
      errors: { fetch: fetchError },
    },
    charts: {
      selected: { error: chartError },
    },
    config: { kubeappsNamespace, kubeappsCluster },
  } = useSelector((state: IStoreState) => state);

  const [repoName, setRepoName] = useState(get(repo, "metadata.name", ""));

  useEffect(() => {
    if (namespace !== kubeappsNamespace) {
      // Normal namespace, show local and global repos
      dispatch(actions.repos.fetchRepos(namespace, true));
    } else {
      // Global namespace, show global repos only
      dispatch(actions.repos.fetchRepos(kubeappsNamespace));
    }
  }, [dispatch, namespace, kubeappsNamespace]);

  const handleChartRepoNameChange = async (e: React.ChangeEvent<HTMLSelectElement>) => {
    const [ns, name] = e.target.value.split("/");
    dispatch(actions.repos.checkChart(kubeappsCluster, ns, name, chartName));
    setRepoName(e.currentTarget.value);
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
          <h5>Chart repositories not found.</h5>
          Manage your Helm chart repositories in Kubeapps by visiting the{" "}
          <Link to={url.app.config.apprepositories(cluster, namespace)}>
            App repositories configuration
          </Link>{" "}
          page.
        </Alert>
      )}
      {repos.length > 0 && (
        <div className="select-repo-form">
          {chartError && <Alert theme="danger">An error occurred: {chartError.message}</Alert>}
          <h2>Select the source repository of {chartName}</h2>
          <label className="select-repo-form-label" htmlFor="chartRepoName">
            Chart Repository Name *
          </label>
          <div className="clr-select-wrapper">
            <select
              id="chartRepoName"
              onChange={handleChartRepoNameChange}
              value={repoName}
              required={true}
              className="clr-page-size-select"
            >
              {!repoName && <option key="" value="" />}
              {repos.map(r => {
                const value = `${r.metadata.namespace}/${r.metadata.name}`;
                return (
                  <option key={value} value={value}>
                    {value} ({getRepoURL(r.metadata.namespace, r.metadata.name)})
                  </option>
                );
              })}
            </select>
          </div>
          <p>
            {" "}
            * If the repository containing {chartName} is not in the list add it{" "}
            <Link to={url.app.config.apprepositories(cluster, namespace)}>here</Link>.{" "}
          </p>
        </div>
      )}
    </LoadingWrapper>
  );
}

export default SelectRepoForm;
