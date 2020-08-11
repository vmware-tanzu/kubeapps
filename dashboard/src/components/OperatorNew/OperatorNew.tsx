import * as React from "react";

import OperatorNotSupported from "components/OperatorList/OperatorsNotSupported";
import { RouterAction } from "connected-react-router";
import { IOperatorsStateError } from "../../reducers/operators";
import { Operators } from "../../shared/Operators";
import { IPackageManifest, IPackageManifestChannel } from "../../shared/types";
import { api, app } from "../../shared/url";
import { ErrorSelector } from "../ErrorAlert";
import UnexpectedErrorPage from "../ErrorAlert/UnexpectedErrorAlert";
import LoadingWrapper from "../LoadingWrapper";
import OperatorHeader from "../OperatorView/OperatorHeader";

import "./OperatorNew.css";

interface IOperatorNewProps {
  operatorName: string;
  operator?: IPackageManifest;
  getOperator: (namespace: string, name: string) => Promise<void>;
  isFetching: boolean;
  cluster: string;
  namespace: string;
  errors: IOperatorsStateError;
  createOperator: (
    namespace: string,
    name: string,
    channel: string,
    installPlanApproval: string,
    csv: string,
  ) => Promise<boolean>;
  push: (location: string) => RouterAction;
}

interface IOperatorNewState {
  updateChannel?: IPackageManifestChannel;
  updateChannelGlobal: boolean;
  // Instalation mode: true for global, false for namespaced
  installationModeGlobal: boolean;
  // Approval strategy: true for automatic, false for manual
  approvalStrategyAutomatic: boolean;
}

class OperatorNew extends React.Component<IOperatorNewProps, IOperatorNewState> {
  public state: IOperatorNewState = {
    updateChannelGlobal: false,
    installationModeGlobal: false,
    approvalStrategyAutomatic: true,
  };

  public componentDidMount() {
    const { operatorName, namespace, getOperator } = this.props;
    getOperator(namespace, operatorName);
  }

  public componentDidUpdate(prevProps: IOperatorNewProps) {
    if (prevProps.operator !== this.props.operator && this.props.operator) {
      const defaultChannel = Operators.getDefaultChannel(this.props.operator);
      const global = Operators.global(defaultChannel);
      this.setState({
        updateChannel: defaultChannel,
        updateChannelGlobal: global,
        installationModeGlobal: global,
      });
    }
    if (prevProps.namespace !== this.props.namespace) {
      this.props.getOperator(this.props.namespace, this.props.operatorName);
    }
  }

  public render() {
    const { cluster, isFetching, namespace, operatorName, operator, errors, push } = this.props;
    if (cluster !== "default") {
      return <OperatorNotSupported namespace={namespace} />;
    }
    const {
      updateChannel,
      updateChannelGlobal,
      installationModeGlobal,
      approvalStrategyAutomatic,
    } = this.state;
    const error = errors.fetch || errors.create;
    if (error) {
      return <ErrorSelector error={error} resource={`Operator ${operatorName}`} />;
    }
    if (isFetching || !operator) {
      return <LoadingWrapper />;
    }
    if (!updateChannel) {
      return (
        <UnexpectedErrorPage
          text={`Operator ${operatorName} doesn't define a valid channel. This is needed to extract required info.`}
        />
      );
    }
    const { currentCSVDesc } = updateChannel;
    // It's not possible to install a namespaced operator in the "operators" ns
    const disableInstall = namespace === "operators" && !updateChannelGlobal;
    return (
      <section className="ChartView padding-b-big">
        <OperatorHeader
          id={operator.metadata.name}
          description={currentCSVDesc.displayName}
          icon={api.operators.operatorIcon(this.props.namespace, operator.metadata.name)}
          version={currentCSVDesc.version}
          cluster={cluster}
          namespace={namespace}
          provider={operator.status.provider.name}
          namespaced={!updateChannelGlobal}
          push={push}
          hideButton={true}
        />
        <main>
          <div className="container container-fluid">
            <div className="row">
              <div className="col-9">
                <form className="container" onSubmit={this.handleDeploy}>
                  <div className="margin-b-normal">
                    <div>
                      <strong>Update Channel</strong>
                    </div>
                    <div className="margin-b-small">
                      The channel to track and receive updates from.
                    </div>
                    {operator.status.channels.map(channel => {
                      return (
                        <label
                          className="margin-l-big"
                          htmlFor={`operator-channel-${channel.name}`}
                          key={channel.name}
                        >
                          <input
                            type="radio"
                            id={`operator-channel-${channel.name}`}
                            name="channel"
                            checked={updateChannel.name === channel.name}
                            onClick={this.selectChannel(channel.name)}
                          />
                          {channel.name}
                          <br />
                        </label>
                      );
                    })}
                  </div>
                  <div className="margin-b-normal">
                    <div className="margin-b-small">
                      <strong>Installation Mode</strong>
                    </div>
                    <label className="margin-l-big" htmlFor="operator-installation-mode-global">
                      <input
                        type="radio"
                        id="operator-installation-mode-global"
                        name="installation-mode"
                        disabled={!updateChannelGlobal}
                        checked={installationModeGlobal}
                        onClick={this.setInstallationMode(true)}
                      />
                      <span className={!updateChannelGlobal ? "disabled" : ""}>
                        All namespaces on the cluster (default)
                      </span>
                      <br />
                      {!updateChannelGlobal && (
                        <div className="disabled margin-l-enormous">
                          <i>This mode is not supported by this Operator and channel.</i>
                        </div>
                      )}
                    </label>
                    <label className="margin-l-big" htmlFor="operator-installation-mode-namespaced">
                      <input
                        type="radio"
                        id="operator-installation-mode-namespaced"
                        name="installation-mode"
                        checked={!installationModeGlobal}
                        onClick={this.setInstallationMode(false)}
                      />
                      The current namespace: {namespace}
                      <br />
                    </label>
                  </div>
                  <div className="margin-b-normal">
                    <div>
                      <strong>Approval Strategy</strong>
                    </div>
                    <div className="margin-b-small">
                      The strategy to determine either manual or automatic updates.
                    </div>
                    <label className="margin-l-big" htmlFor="operator-update-automatic">
                      <input
                        type="radio"
                        id="operator-update-automatic"
                        name="operator-update"
                        checked={approvalStrategyAutomatic}
                        onClick={this.setApprovalStrategy(true)}
                      />
                      Automatic
                      <br />
                    </label>
                    <label className="margin-l-big" htmlFor="operator-update-manual">
                      <input
                        type="radio"
                        id="operator-update-manual"
                        name="operator-update"
                        checked={!approvalStrategyAutomatic}
                        onClick={this.setApprovalStrategy(false)}
                      />
                      Manual
                      <br />
                    </label>
                  </div>
                  {disableInstall && (
                    <UnexpectedErrorPage
                      title={
                        'It\'s not possible to install a namespaced operator in the "operators" namespace'
                      }
                    />
                  )}
                  <button className="button button-primary" type="submit" disabled={disableInstall}>
                    Submit
                  </button>
                </form>
              </div>
            </div>
          </div>
        </main>
      </section>
    );
  }

  private selectChannel = (channel: string) => {
    const newChannel = this.props.operator?.status.channels.find(ch => ch.name === channel);
    const global = Operators.global(newChannel);
    return () => {
      this.setState({
        updateChannel: newChannel,
        updateChannelGlobal: global,
        installationModeGlobal: global,
      });
    };
  };

  private setInstallationMode = (global: boolean) => {
    return () => {
      this.setState({ installationModeGlobal: global });
    };
  };

  private setApprovalStrategy = (automatic: boolean) => {
    return () => {
      this.setState({ approvalStrategyAutomatic: !this.state.approvalStrategyAutomatic });
    };
  };

  private handleDeploy = async () => {
    const { cluster, namespace, operator, createOperator, push } = this.props;
    const { updateChannel, installationModeGlobal, approvalStrategyAutomatic } = this.state;
    const targetNS = installationModeGlobal ? "operators" : namespace;
    const approvalStrategy = approvalStrategyAutomatic ? "Automatic" : "Manual";
    const deployed = await createOperator(
      targetNS,
      operator!.metadata.name,
      updateChannel!.name,
      approvalStrategy,
      updateChannel!.currentCSV,
    );
    if (deployed) {
      // Success, redirect to operator page
      push(app.operators.list(cluster, namespace));
    }
  };
}

export default OperatorNew;
