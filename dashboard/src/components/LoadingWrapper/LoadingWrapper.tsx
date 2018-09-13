import * as React from "react";
import LoaderSpinner from "../LoaderSpinner";

interface ILoadingWrapperProps {
  loaded?: boolean;
}

// TODO(miguel) Move these kind of components to stateless compontents once we upgrade ts definitions
// Currently, I am having issues transforming it via const LoadingWrapper: React.SFC<ILoadingWrapperProps>
class LoadingWrapper extends React.Component<ILoadingWrapperProps> {
  public render() {
    return this.props.loaded ? this.props.children : <LoaderSpinner />;
  }
}

export default LoadingWrapper;
