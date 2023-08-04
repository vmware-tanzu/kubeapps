// Copyright 2021-2023 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { CustomComponent } from "RemoteComponent";
import actions from "actions";
import { usePush } from "hooks/push";
import {
  AvailablePackageDetail,
  InstalledPackageDetail,
} from "gen/kubeappsapis/core/packages/v1alpha1/packages_pb";
import { useCallback, useMemo } from "react";
import { useDispatch, useSelector } from "react-redux";
import { IStoreState } from "shared/types";
import * as urls from "shared/url";
import { IAppViewResourceRefs } from "../AppView";

export interface ICustomAppViewProps {
  resourceRefs: IAppViewResourceRefs;
  app: InstalledPackageDetail;
  appDetails: AvailablePackageDetail;
}

function CustomAppView({ resourceRefs, app, appDetails }: ICustomAppViewProps) {
  const {
    config: { remoteComponentsUrl },
  } = useSelector((state: IStoreState) => state);

  const dispatch = useDispatch();

  const handleDelete = useCallback(
    () => dispatch(actions.installedpackages.deleteInstalledPackage(app.installedPackageRef!)),
    [dispatch, app.installedPackageRef],
  );

  const handleRollback = useCallback(
    () => dispatch(actions.installedpackages.rollbackInstalledPackage(app.installedPackageRef!, 1)),
    [dispatch, app.installedPackageRef],
  );

  const push = usePush();
  const handleRedirect = useCallback(url => push(url), [push]);

  const url = remoteComponentsUrl
    ? remoteComponentsUrl
    : `${window.location.origin}/custom_components.js`;

  return useMemo(
    () => (
      <CustomComponent
        url={url}
        resourceRefs={resourceRefs}
        appDetails={appDetails}
        app={app}
        componentType="AppView"
        appId={app.availablePackageRef?.identifier}
        handleDelete={handleDelete}
        handleRollback={handleRollback}
        handleRedirect={handleRedirect}
        urls={urls}
      />
    ),
    [resourceRefs, app, appDetails, url, handleDelete, handleRollback, handleRedirect],
  );
}

export default CustomAppView;
