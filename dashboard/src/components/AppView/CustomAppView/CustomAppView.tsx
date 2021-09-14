import actions from "actions";
import { push } from "connected-react-router";
import {
  AvailablePackageDetail,
  InstalledPackageDetail,
} from "gen/kubeappsapis/core/packages/v1alpha1/packages";
import { useCallback, useEffect, useMemo } from "react";
import { useDispatch, useSelector } from "react-redux";
import { CustomComponent } from "RemoteComponent";
import { IStoreState } from "shared/types";
import { IAppViewResourceRefs, IRouteParams } from "../AppView";
import * as ReactRouter from "react-router";
import * as urls from "../../../shared/url";

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
  const { cluster, namespace, releaseName } = ReactRouter.useParams() as IRouteParams;

  useEffect(() => {
    dispatch(actions.apps.getApp(cluster, namespace, releaseName));
  }, [cluster, dispatch, namespace, releaseName]);

  const handleDelete = useCallback(
    () => dispatch(actions.apps.deleteApp(cluster, namespace, releaseName, true)),
    [dispatch, cluster, namespace, releaseName],
  );

  const handleRollback = useCallback(
    () => dispatch(actions.apps.rollbackApp(cluster, namespace, releaseName, 1)),
    [dispatch, cluster, namespace, releaseName],
  );

  const handleRedirect = useCallback(url => dispatch(push(url)), [dispatch]);

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
