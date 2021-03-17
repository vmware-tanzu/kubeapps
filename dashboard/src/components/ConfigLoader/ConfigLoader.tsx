import Alert from "components/js/Alert";
import React from "react";

import LoadingWrapper, {
  ILoadingWrapperProps,
} from "../../components/LoadingWrapper/LoadingWrapper";

interface IConfigLoaderProps extends ILoadingWrapperProps {
  children?: JSX.Element;
  getConfig: () => void;
  error?: Error;
}

function ConfigLoader({ getConfig, error, ...otherProps }: IConfigLoaderProps) {
  React.useEffect(() => {
    getConfig();
  }, [getConfig]);
  if (error) {
    return <Alert theme="danger">Unable to load Kubeapps configuration: {error.message}</Alert>;
  }
  return (
    <LoadingWrapper className="margin-t-xxl" loadingText={<h1>Kubeapps</h1>} {...otherProps} />
  );
}

export default ConfigLoader;
