import { get } from "lodash";
import React, { useEffect, useState } from "react";
import { Link } from "react-router-dom";

import Alert from "components/js/Alert";
import * as url from "shared/url";
import { IAppRepository } from "../../shared/types";
import LoadingWrapper from "../LoadingWrapper/LoadingWrapper.v2";
import "./SelectRepoForm.css";

interface ISelectRepoFormProps {
  isFetching: boolean;
  cluster: string;
  namespace: string;
  repoError?: Error;
  error?: Error;
  repo: IAppRepository;
  repos: IAppRepository[];
  chartName: string;
  checkChart: (namespace: string, repo: string, chartName: string) => any;
  fetchRepositories: (namespace: string) => void;
}

function SelectRepoForm({
  isFetching,
  cluster,
  namespace,
  repoError,
  error,
  repo,
  repos,
  chartName,
  checkChart,
  fetchRepositories,
}: ISelectRepoFormProps) {
  const [repoName, setRepoName] = useState(get(repo, "metadata.name", ""));

  useEffect(() => {
    fetchRepositories(namespace);
  }, [fetchRepositories, namespace]);

  const handleChartRepoNameChange = async (e: React.ChangeEvent<HTMLSelectElement>) => {
    checkChart(namespace, e.target.value, chartName);
    setRepoName(e.currentTarget.value);
  };

  const getRepoURL = (name: string) => {
    let res = "";
    repos.forEach(r => {
      if (r.metadata.name === name && r.spec) {
        res = r.spec.url;
      }
    });
    return res;
  };

  return (
    <LoadingWrapper loaded={!isFetching}>
      {repoError && <Alert theme="danger">Found error: {repoError.message}</Alert>}
      {!repoError && repos.length === 0 && (
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
          {error && <Alert theme="danger">Found error: {error.message}</Alert>}
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
              {repos.map(r => (
                <option key={r.metadata.name} value={r.metadata.name}>
                  {r.metadata.name} ({getRepoURL(r.metadata.name)})
                </option>
              ))}
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
