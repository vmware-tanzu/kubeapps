// Copyright 2020-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { CdsButton } from "@cds/react/button";
import actions from "actions";
import Alert from "components/js/Alert";
import Column from "components/js/Column";
import Row from "components/js/Row";
import OperatorSummary from "components/OperatorSummary/OperatorSummary";
import { push, RouterAction } from "connected-react-router";
import { useEffect, useState } from "react";
import { useDispatch, useSelector } from "react-redux";
import { Action } from "redux";
import { ThunkDispatch } from "redux-thunk";
import { Operators } from "shared/Operators";
import { IPackageManifest, IPackageManifestChannel, IStoreState } from "shared/types";
import { api, app } from "shared/url";
import { IOperatorsStateError } from "../../reducers/operators";
import LoadingWrapper from "../LoadingWrapper/LoadingWrapper";
import OperatorHeader from "../OperatorView/OperatorHeader";
import "./OperatorNew.css";

export interface IOperatorNewProps {
  operatorName: string;
  operator?: IPackageManifest;
  getOperator: (cluster: string, namespace: string, name: string) => Promise<void>;
  isFetching: boolean;
  cluster: string;
  namespace: string;
  errors: IOperatorsStateError;
  createOperator: (
    cluster: string,
    namespace: string,
    name: string,
    channel: string,
    installPlanApproval: string,
    csv: string,
  ) => Promise<boolean>;
  push: (location: string) => RouterAction;
}

export default function OperatorNew({ namespace, operatorName, cluster }: IOperatorNewProps) {
  const dispatch: ThunkDispatch<IStoreState, null, Action> = useDispatch();

  const [updateChannel, setUpdateChannel] = useState(
    undefined as IPackageManifestChannel | undefined,
  );
  const [updateChannelGlobal, setUpdateChannelGlobal] = useState(false);
  // Installation mode: true for global, false for namespaced
  const [installationModeGlobal, setInstallationModeGlobal] = useState(false);
  // Approval strategy: true for automatic, false for manual
  const [approvalStrategyAutomatic, setApprovalStrategyAutomatic] = useState(true);

  useEffect(() => {
    dispatch(actions.operators.getOperator(cluster, namespace, operatorName));
  }, [dispatch, cluster, namespace, operatorName]);

  const {
    operators: {
      operator,
      isFetching,
      errors: { operator: errors },
    },
  } = useSelector((state: IStoreState) => state);

  useEffect(() => {
    if (operator) {
      const defaultChannel = Operators.getDefaultChannel(operator);
      const global = Operators.global(defaultChannel);
      setUpdateChannel(defaultChannel);
      setUpdateChannelGlobal(global);
      setInstallationModeGlobal(global);
    }
  }, [operator]);

  if (errors.fetch) {
    return (
      <Alert theme="danger">
        An error occurred while fetching the operator {operatorName}: {errors.fetch.message}
      </Alert>
    );
  }
  if (errors.create) {
    return (
      <Alert theme="danger">
        An error occurred while creating the operator {operatorName}: {errors.create.message}
      </Alert>
    );
  }
  if (isFetching || !operator) {
    return <LoadingWrapper className="margin-t-xxl" loadingText={`Fetching ${operatorName}...`} />;
  }
  if (!updateChannel) {
    return (
      <Alert theme="danger">
        The Operator {operatorName} doesn't define a valid channel. This is needed to extract
        required info.
      </Alert>
    );
  }
  const { currentCSVDesc } = updateChannel;
  // It's not possible to install a namespaced operator in the "operators" ns
  const disableInstall = namespace === "operators" && !updateChannelGlobal;

  const selectChannel = (channel: string) => {
    const newChannel = operator?.status.channels.find(ch => ch.name === channel);
    const global = Operators.global(newChannel);
    return () => {
      setUpdateChannel(newChannel);
      setUpdateChannelGlobal(global);
      setInstallationModeGlobal(global);
    };
  };

  const setInstallationMode = (global: boolean) => {
    return () => setInstallationModeGlobal(global);
  };

  const setApprovalStrategy = (automatic: boolean) => {
    return () => setApprovalStrategyAutomatic(automatic);
  };

  const handleDeploy = async () => {
    const targetNS = installationModeGlobal ? "operators" : namespace;
    const approvalStrategy = approvalStrategyAutomatic ? "Automatic" : "Manual";
    const deployed = await dispatch(
      actions.operators.createOperator(
        cluster,
        targetNS,
        operator!.metadata.name,
        updateChannel!.name,
        approvalStrategy,
        updateChannel!.currentCSV,
      ),
    );
    if (deployed) {
      // Success, redirect to operator page
      dispatch(push(app.operators.list(cluster, namespace)));
    }
  };

  /* eslint-disable jsx-a11y/label-has-associated-control */
  return (
    <section>
      <OperatorHeader
        title={`${operator.metadata.name} by ${operator.status.provider.name}`}
        icon={api.operators.operatorIcon(cluster, namespace, operator.metadata.name)}
        version={currentCSVDesc.version}
      />
      <section>
        <Row>
          <Column span={3}>
            <OperatorSummary />
          </Column>
          <Column span={9}>
            <form onSubmit={handleDeploy} className="kubeapps-form">
              {disableInstall && (
                <Alert theme="danger">
                  It's not possible to install a namespaced operator in the "operators" namespace
                </Alert>
              )}
              <div className="clr-form-control">
                <label className="clr-control-label">Update Channel</label>
                <div className="clr-subtext-wrapper">
                  <span className="clr-subtext">The channel to track and receive updates from</span>
                </div>
                <div className="clr-control-container">
                  {operator.status.channels.map(channel => (
                    <div className="clr-radio-wrapper" key={`operator-channel-${channel.name}`}>
                      <input
                        type="radio"
                        id={`operator-channel-${channel.name}`}
                        name="channel"
                        checked={updateChannel.name === channel.name}
                        onChange={selectChannel(channel.name)}
                        className="clr-radio"
                      />
                      <label
                        htmlFor={`operator-channel-${channel.name}`}
                        className="clr-control-label"
                      >
                        {channel.name}
                      </label>
                    </div>
                  ))}
                </div>
              </div>

              <div className="clr-form-control">
                <label className="clr-control-label">Installation Mode</label>
                <div className="clr-control-container">
                  <div className="clr-radio-wrapper">
                    <input
                      type="radio"
                      id="operator-installation-mode-global"
                      name="installation-mode"
                      disabled={!updateChannelGlobal}
                      checked={installationModeGlobal}
                      onChange={setInstallationMode(true)}
                      className="clr-radio"
                    />
                    <label
                      htmlFor="operator-installation-mode-global"
                      className="clr-control-label"
                    >
                      <span className={!updateChannelGlobal ? "disabled" : ""}>
                        All namespaces on the cluster (default)
                      </span>
                    </label>
                    {!updateChannelGlobal && (
                      <div className="clr-subtext-wrapper">
                        <span className="clr-subtext disabled disabled-description">
                          This mode is not supported by this Operator and channel.
                        </span>
                      </div>
                    )}
                  </div>
                  <div className="clr-radio-wrapper">
                    <input
                      type="radio"
                      id="operator-installation-mode-namespaced"
                      name="installation-mode"
                      checked={!installationModeGlobal}
                      onChange={setInstallationMode(false)}
                      className="clr-radio"
                    />
                    <label
                      htmlFor="operator-installation-mode-namespaced"
                      className="clr-control-label"
                    >
                      The current namespace: {namespace}
                    </label>
                  </div>
                </div>
              </div>

              <div className="clr-form-control">
                <label className="clr-control-label">Approval Strategy</label>
                <div className="clr-subtext-wrapper">
                  <span className="clr-subtext">
                    The strategy to determine either manual or automatic updates
                  </span>
                </div>

                <div className="clr-control-container">
                  <div className="clr-radio-wrapper">
                    <input
                      type="radio"
                      id="operator-update-automatic"
                      name="operator-update"
                      checked={approvalStrategyAutomatic}
                      onChange={setApprovalStrategy(true)}
                      className="clr-radio"
                    />
                    <label htmlFor="operator-update-automatic" className="clr-control-label">
                      Automatic
                    </label>
                  </div>
                  <div className="clr-radio-wrapper">
                    <input
                      type="radio"
                      id="operator-update-manual"
                      name="operator-update"
                      checked={!approvalStrategyAutomatic}
                      onChange={setApprovalStrategy(false)}
                      className="clr-radio"
                    />
                    <label htmlFor="operator-update-manual" className="clr-control-label">
                      Manual
                    </label>
                  </div>
                </div>
              </div>
              <div className="clr-form-control">
                <CdsButton type="submit" disabled={disableInstall}>
                  Deploy
                </CdsButton>
              </div>
            </form>
          </Column>
        </Row>
      </section>
    </section>
  );
}
