import { useMemo } from "react";
import { IAppViewResourceRefs } from "../AppView";
import { IRelease, IStoreState } from "shared/types";
import { useSelector } from "react-redux";
import { CustomComponent } from "RemoteComponent";

export interface ICustomAppViewProps {
  resourceRefs: IAppViewResourceRefs;
  app: IRelease;
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
