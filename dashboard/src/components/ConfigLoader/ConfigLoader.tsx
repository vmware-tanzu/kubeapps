import * as React from "react";

import UnexpectedErrorPage from "../../components/ErrorAlert/UnexpectedErrorAlert";
import LoadingWrapper from "../../components/LoadingWrapper";
import { ILoadingWrapperProps } from "../LoadingWrapper/LoadingWrapper";

interface IConfigLoaderProps extends ILoadingWrapperProps {
  getConfig: () => void;
  error?: Error;
}

class ConfigLoader extends React.Component<IConfigLoaderProps> {
  public componentDidMount() {
    this.props.getConfig();
  }

  public render() {
    const { error, ...otherProps } = this.props;
    if (error) {
      return (
        <UnexpectedErrorPage
          raw={true}
          showGenericMessage={true}
          text={`Unable to load Kubeapps configuration: ${error.message}`}
        />
      );
    }
    return <LoadingWrapper {...otherProps} />;
  }
}

export default ConfigLoader;
