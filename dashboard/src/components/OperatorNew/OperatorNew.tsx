// Copyright 2020-2023 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { CdsButton } from "@cds/react/button";
import actions from "actions";
import AlertGroup from "components/AlertGroup";
import Column from "components/Column";
import LoadingWrapper from "components/LoadingWrapper";
import OperatorSummary from "components/OperatorSummary/OperatorSummary";
import Row from "components/Row";
import { usePush } from "hooks/push";
import { useEffect, useState } from "react";
import { useDispatch, useSelector } from "react-redux";
import { useParams } from "react-router-dom";
import { Action } from "redux";
import { ThunkDispatch } from "redux-thunk";
import { Operators } from "shared/Operators";
import { IPackageManifestChannel, IStoreState } from "shared/types";
import { api, app } from "shared/url";
import OperatorHeader from "../OperatorView/OperatorHeader";
import "./OperatorNew.css";

export default function OperatorNew() {
  const dispatch: ThunkDispatch<IStoreState, null, Action> = useDispatch();

  const [updateChannel, setUpdateChannel] = useState(
    undefined as IPackageManifestChannel | undefined,
  );
  const [updateChannelGlobal, setUpdateChannelGlobal] = useState(false);
  // Installation mode: true for global, false for namespaced
  const [installationModeGlobal, setInstallationModeGlobal] = useState(false);
  // Approval strategy: true for automatic, false for manual
  const [approvalStrategyAutomatic, setApprovalStrategyAutomatic] = useState(true);

  type OperatorNewParams = {
    operator: string;
  };
  const params = useParams<OperatorNewParams>();
  const operatorName = params.operator || "";
  const push = usePush();

  const {
    operators: {
      operator,
      isFetching,
      errors: { operator: errors },
    },
    clusters: { currentCluster, clusters },
  } = useSelector((state: IStoreState) => state);
  const namespace = clusters[currentCluster].currentNamespace;
  const cluster = currentCluster;

  useEffect(() => {
    dispatch(actions.operators.getOperator(cluster, namespace, operatorName));
  }, [dispatch, cluster, namespace, operatorName]);

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
      <AlertGroup status="danger">
        An error occurred while fetching the operator {operatorName}: {errors.fetch.message}.
      </AlertGroup>
    );
  }
  if (errors.create) {
    return (
      <AlertGroup status="danger">
        An error occurred while creating the operator {operatorName}: {errors.create.message}.
      </AlertGroup>
    );
  }
  if (isFetching || !operator) {
    return <LoadingWrapper className="margin-t-xxl" loadingText={`Fetching ${operatorName}...`} />;
  }
  if (!updateChannel) {
    return (
      <AlertGroup status="danger">
        The Operator {operatorName} doesn't define a valid channel. This is needed to extract
        required info.
      </AlertGroup>
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
      push(app.operators.list(cluster, namespace));
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
                <AlertGroup status="danger">
                  It's not possible to install a namespaced operator in the "operators" namespace.
                </AlertGroup>
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
