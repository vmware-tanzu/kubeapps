import * as React from "react";

import Spinner from "../js/Spinner";

export interface ILoadingWrapperProps {
  loaded?: boolean;
  size?: string;
  children?: any;
}

function LoadingWrapper(props: ILoadingWrapperProps) {
  return props.loaded ? (
    props.children
  ) : (
    <Spinner medium={props.size === "medium"} small={props.size === "small"} />
  );
}

export default LoadingWrapper;
