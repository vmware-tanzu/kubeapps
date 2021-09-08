import { InstalledPackageDetail } from "gen/kubeappsapis/core/packages/v1alpha1/packages";
import { useMemo } from "react";
import { useSelector } from "react-redux";
import { CustomComponent } from "RemoteComponent";
import { IStoreState } from "shared/types";
import { IAppViewResourceRefs } from "../AppView";

export interface ICustomAppViewProps {
  resourceRefs: IAppViewResourceRefs;
  app: InstalledPackageDetail;
}

function CustomAppView({ resourceRefs, app }: ICustomAppViewProps) {
  const {
    config: { remoteComponentsUrl },
  } = useSelector((state: IStoreState) => state);

  const url = remoteComponentsUrl
    ? remoteComponentsUrl
    : `${window.location.origin}/custom_components.js`;

  return useMemo(
    () => <CustomComponent url={url} resourceRefs={resourceRefs} app={app} />,
    [resourceRefs, app, url],
  );
}

export default CustomAppView;
