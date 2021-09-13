import {
  AvailablePackageDetail,
  InstalledPackageDetail,
} from "gen/kubeappsapis/core/packages/v1alpha1/packages";
import { useMemo } from "react";
import { useSelector } from "react-redux";
import { CustomComponent } from "RemoteComponent";
import { IStoreState } from "shared/types";
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

  const url = remoteComponentsUrl
    ? remoteComponentsUrl
    : `${window.location.origin}/custom_components.js`;

  return useMemo(
    () => (
      <CustomComponent url={url} resourceRefs={resourceRefs} appDetails={appDetails} app={app} />
    ),
    [resourceRefs, app, appDetails, url],
  );
}

export default CustomAppView;
