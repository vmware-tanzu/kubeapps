import * as React from "react";

import LoadingSpinner from "../LoadingSpinner";
import LoadingPlaceholder from "./LoadingPlaceholder";

export enum LoaderType {
  Spinner,
  Placeholder,
}

export interface ILoadingWrapperProps {
  type?: LoaderType;
  loaded?: boolean;
  size?: string;
}

// TODO(miguel) Move these kind of components to stateless compontents once we upgrade ts definitions
// Currently, I am having issues transforming it via const LoadingWrapper: React.SFC<ILoadingWrapperProps>
class LoadingWrapper extends React.Component<ILoadingWrapperProps> {
  public static defaultProps: ILoadingWrapperProps = {
    type: LoaderType.Spinner,
    loaded: false,
  };

  public render() {
    return this.props.loaded ? this.props.children : this.renderLoader();
  }

  private renderLoader() {
    switch (this.props.type) {
      case LoaderType.Spinner:
        return <LoadingSpinner size={this.props.size} />;
      case LoaderType.Placeholder:
        return <LoadingPlaceholder />;
    }
    return;
  }
}

export default LoadingWrapper;
