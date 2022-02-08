// Copyright 2018-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import actions from "actions";
import Alert from "components/js/Alert";
import { InstalledPackageReference } from "gen/kubeappsapis/core/packages/v1alpha1/packages";
import { Plugin } from "gen/kubeappsapis/core/plugins/v1alpha1/plugins";
import { useEffect, useState } from "react";
import { useDispatch, useSelector } from "react-redux";
import * as ReactRouter from "react-router-dom";
import { Action } from "redux";
import { ThunkDispatch } from "redux-thunk";
import { FetchError, IStoreState } from "shared/types";
import LoadingWrapper from "../LoadingWrapper/LoadingWrapper";
import SelectRepoForm from "../SelectRepoForm/SelectRepoForm";
import UpgradeForm from "../UpgradeForm/UpgradeForm";

interface IRouteParams {
  cluster: string;
  namespace: string;
  releaseName: string;
  pluginName: string;
  pluginVersion: string;
  version?: string;
}

function AppUpgrade() {
  const dispatch: ThunkDispatch<IStoreState, null, Action> = useDispatch();
  const { cluster, namespace, releaseName, pluginName, pluginVersion, version } =
    ReactRouter.useParams() as IRouteParams;

  const {
    apps: {
      selected: installedAppInstalledPackageDetail,
      isFetching: appsIsFetching,
      error,
      selectedDetails: installedAppAvailablePackageDetail,
    },
    packages: { isFetching: chartsIsFetching, selected: selectedPackage },
  } = useSelector((state: IStoreState) => state);

  const isFetching = appsIsFetching || chartsIsFetching;

  const [pluginObj] = useState({ name: pluginName, version: pluginVersion } as Plugin);

  // Initial fetch using the params in the URL
  useEffect(() => {
    dispatch(actions.availablepackages.resetSelectedAvailablePackageDetail());
    dispatch(
      actions.installedpackages.getInstalledPackage({
        context: { cluster: cluster, namespace: namespace },
        identifier: releaseName,
        plugin: pluginObj,
      } as InstalledPackageReference),
    );
  }, [dispatch, cluster, namespace, pluginObj, releaseName]);

  if (error && error.constructor === FetchError) {
    return <Alert theme="danger">Unable to retrieve the current app: {error.message}</Alert>;
  }

  if (isFetching || !installedAppInstalledPackageDetail) {
    const loadingPkgName =
      selectedPackage.availablePackageDetail?.availablePackageRef?.identifier ??
      installedAppInstalledPackageDetail?.installedPackageRef?.identifier ??
      "package";
    return (
      <LoadingWrapper
        loadingText={`Fetching ${decodeURIComponent(loadingPkgName)} version...`}
        className="margin-t-xxl"
        loaded={false}
      />
    );
  }
  if (installedAppAvailablePackageDetail && installedAppInstalledPackageDetail && selectedPackage) {
    return (
      <div>
        <UpgradeForm version={version} />
      </div>
    );
  }
  return (
    <SelectRepoForm
      cluster={cluster}
      namespace={namespace}
      app={installedAppInstalledPackageDetail}
    />
  );
}

export default AppUpgrade;
