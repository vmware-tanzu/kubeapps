import * as React from "react";

interface ILoadingWrapperProps {
  loaded: boolean;
}

class LoadingWrapper extends React.Component<ILoadingWrapperProps> {
  public render() {
    return this.props.loaded ? this.props.children : <div>Loading</div>;
  }
}

export default LoadingWrapper;
