import * as React from "react";

import Spinner from "../js/Spinner";

export interface ILoadingWrapperProps {
  loaded?: boolean;
  small?: boolean;
  medium?: boolean;
  children?: any;
}

function LoadingWrapper(props: ILoadingWrapperProps) {
  return props.loaded ? props.children : <Spinner medium={props.medium} small={props.small} />;
}

export default LoadingWrapper;
