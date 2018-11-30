import * as React from "react";
import LoadingSpinner from "../LoadingSpinner";

export interface ILoadingWrapperProps {
  loaded?: boolean;
  size?: string;
}

// TODO(miguel) Move these kind of components to stateless compontents once we upgrade ts definitions
// Currently, I am having issues transforming it via const LoadingWrapper: React.SFC<ILoadingWrapperProps>
class LoadingWrapper extends React.Component<ILoadingWrapperProps> {
  public static defaultProps: ILoadingWrapperProps = {
    loaded: false,
  };

  public render() {
    return this.props.loaded ? this.props.children : <LoadingSpinner size={this.props.size} />;
  }
}

export default LoadingWrapper;
