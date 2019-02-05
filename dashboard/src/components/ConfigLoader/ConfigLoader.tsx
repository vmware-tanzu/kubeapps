import * as React from "react";

import UnexpectedErrorPage from "../../components/ErrorAlert/UnexpectedErrorAlert";
import LoadingWrapper, { ILoadingWrapperProps } from "../../components/LoadingWrapper";

export interface IConfigLoaderProps extends ILoadingWrapperProps {
  error?: Error;
}

const ConfigLoader: React.SFC<IConfigLoaderProps> = props => {
  const { error, ...otherProps } = props;
  if (props.error) {
    return (
      <UnexpectedErrorPage
        raw={true}
        showGenericMessage={true}
        text={`Unable to load Kubeapps configuration: ${props.error.message}`}
      />
    );
  }
  return <LoadingWrapper {...otherProps} />;
};

export default ConfigLoader;
