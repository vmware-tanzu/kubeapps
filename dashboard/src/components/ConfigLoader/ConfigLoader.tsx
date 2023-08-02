// Copyright 2019-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import actions from "actions";
import Alert from "components/js/Alert";
import LoadingWrapper, { ILoadingWrapperProps } from "components/LoadingWrapper/LoadingWrapper";
import React from "react";
import { useIntl } from "react-intl";
import { useDispatch, useSelector } from "react-redux";
import { ThunkDispatch } from "redux-thunk";
import { IStoreState } from "shared/types";
import { Action } from "typesafe-actions";

interface IConfigLoaderProps extends ILoadingWrapperProps {
  children?: JSX.Element;
}

function ConfigLoader({ ...otherProps }: IConfigLoaderProps) {
  const dispatch: ThunkDispatch<IStoreState, null, Action> = useDispatch();
  const intl = useIntl();
  React.useEffect(() => {
    dispatch(actions.config.getConfig());
  });
  const kubeappsTitle = intl.formatMessage({ id: "Kubeapps", defaultMessage: "Kubeapps" });

  const {
    config: { error, loaded },
  } = useSelector((state: IStoreState) => state);

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
          loaded={loaded}
          {...otherProps}
        />
      )}
    </>
  );
}

export default ConfigLoader;
