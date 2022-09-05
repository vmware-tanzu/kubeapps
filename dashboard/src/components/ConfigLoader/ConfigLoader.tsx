// Copyright 2019-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import Alert from "components/js/Alert";
import LoadingWrapper, { ILoadingWrapperProps } from "components/LoadingWrapper/LoadingWrapper";
import React from "react";
import { useIntl } from "react-intl";

interface IConfigLoaderProps extends ILoadingWrapperProps {
  children?: JSX.Element;
  getConfig: () => void;
  error?: Error;
}

function ConfigLoader({ getConfig, error, ...otherProps }: IConfigLoaderProps) {
  const intl = useIntl();
  React.useEffect(() => getConfig(), [getConfig]);
  const kubeappsTitle = intl.formatMessage({ id: "Kubeapps", defaultMessage: "Kubeapps" });

  return (
    <>
      {error ? (
        <Alert theme="danger">
          Unable to load {kubeappsTitle} configuration: {error?.message}
        </Alert>
      ) : (
        <LoadingWrapper
          className="margin-t-xxl"
          loadingText={<h1>{kubeappsTitle}</h1>}
          {...otherProps}
        />
      )}
    </>
  );
}

export default ConfigLoader;
