// Copyright 2019-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

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
      reposSummaries: repos,
      repoDetail: repo,
      errors: { fetch: fetchError },
    },
    packages: {
      selected: { error: packageError },
    },
    config: { kubeappsNamespace, kubeappsCluster, helmGlobalNamespace, carvelGlobalNamespace },
  } = useSelector((state: IStoreState) => state);

  const [userRepoName, setUserRepoName] = useState(repo?.name ?? "");
  const [userRepoNamespace, setUserRepoNamepace] = useState(
    repo.packageRepoRef?.context?.namespace ?? "",
  );
  // We do not currently support package repositories on additional clusters.
  const supportedCluster = cluster === kubeappsCluster;
  useEffect(() => {
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
  }, [
    dispatch,
    namespace,
    kubeappsNamespace,
    helmGlobalNamespace,
    carvelGlobalNamespace,
    supportedCluster,
  ]);

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
    return repos.find(r => r.name === name && r.packageRepoRef?.context?.namespace === ns);
  };

  const getRepoURL = (ns: string, name: string) => {
    const r = findRepo(ns, name);
    return r?.url || "";
  };

  return (
    <LoadingWrapper
      className="margin-t-xxl"
      loadingText="Fetching Package Repositories..."
      loaded={!isFetching}
    >
      {fetchError && <Alert theme="danger">An error occurred: {fetchError.message}</Alert>}
      {!fetchError && repos.length === 0 && (
        <Alert theme="warning">
          <h5>Repositories not found. </h5>
          Manage your repositories in Kubeapps by visiting the{" "}
          <Link to={url.app.config.pkgrepositories(cluster, namespace)}>
            Package Repositories configuration
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
                const value = `${r.packageRepoRef?.context?.namespace}/${r.name}`;
                return (
                  <option key={value} value={value}>
                    {value} ({getRepoURL(r.packageRepoRef?.context?.namespace || "", r.name)})
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
            <Link to={url.app.config.pkgrepositories(cluster, namespace)}>here</Link>.{" "}
          </p>
        </div>
      )}
    </LoadingWrapper>
  );
}

export default SelectRepoForm;
