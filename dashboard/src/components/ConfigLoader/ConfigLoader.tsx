// Copyright 2019-2023 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import actions from "actions";
import AlertGroup from "components/AlertGroup";
import Column from "components/Column";
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
  }, [dispatch]);
  const kubeappsTitle = intl.formatMessage({ id: "Kubeapps", defaultMessage: "Kubeapps" });

  const {
    config: { error, loaded },
  } = useSelector((state: IStoreState) => state);

  return (
    <>
      {error ? (
        <Column>
          <AlertGroup status="danger" closable={false}>
            Unable to load the {kubeappsTitle} configuration: {error?.message}.
          </AlertGroup>
        </Column>
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
