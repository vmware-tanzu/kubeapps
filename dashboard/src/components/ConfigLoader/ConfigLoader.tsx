import * as React from "react";

import UnexpectedErrorPage from "../../components/ErrorAlert/UnexpectedErrorAlert";
import LoadingWrapper, {
  ILoadingWrapperProps,
} from "../../components/LoadingWrapper/LoadingWrapper";

interface IConfigLoaderProps extends ILoadingWrapperProps {
  getConfig: () => void;
  error?: Error;
  children?: any;
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
