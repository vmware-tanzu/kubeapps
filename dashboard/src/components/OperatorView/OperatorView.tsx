import * as React from "react";

import OperatorNotSupported from "components/OperatorList/OperatorsNotSupported";
import { RouterAction } from "connected-react-router";
import { Operators } from "../../shared/Operators";
import { IClusterServiceVersion, IPackageManifest } from "../../shared/types";
import { api } from "../../shared/url";
import { ErrorSelector } from "../ErrorAlert";
import UnexpectedErrorPage from "../ErrorAlert/UnexpectedErrorAlert";
import LoadingWrapper from "../LoadingWrapper";
import CapabiliyLevel from "./OperatorCapabilityLevel";
import OperatorDescription from "./OperatorDescription";
import OperatorHeader from "./OperatorHeader";

export interface IOperatorViewProps {
  operatorName: string;
  operator?: IPackageManifest;
  getOperator: (cluster: string, namespace: string, name: string) => Promise<void>;
  isFetching: boolean;
  cluster: string;
  namespace: string;
  kubeappsCluster: string;
  error?: Error;
  push: (location: string) => RouterAction;
  getCSV: (
    cluster: string,
    namespace: string,
    name: string,
  ) => Promise<IClusterServiceVersion | undefined>;
  csv?: IClusterServiceVersion;
}

class OperatorView extends React.Component<IOperatorViewProps> {
  public componentDidMount() {
    const { cluster, operatorName, namespace, getOperator } = this.props;
    getOperator(cluster, namespace, operatorName);
  }

  public componentDidUpdate(prevProps: IOperatorViewProps) {
    const { cluster, namespace, getOperator, getCSV } = this.props;
    if (prevProps.operator !== this.props.operator && this.props.operator) {
      const defaultChannel = Operators.getDefaultChannel(this.props.operator);
      if (defaultChannel) {
        getCSV(cluster, namespace, defaultChannel.currentCSV);
      }
    }
    if (prevProps.namespace !== this.props.namespace) {
      getOperator(cluster, this.props.namespace, this.props.operatorName);
    }
  }

  public render() {
    const {
      isFetching,
      cluster,
      namespace,
      kubeappsCluster,
      operatorName,
      operator,
      error,
      push,
      csv,
    } = this.props;
    if (cluster !== kubeappsCluster) {
      return <OperatorNotSupported kubeappsCluster={kubeappsCluster} namespace={namespace} />;
    }
    if (error) {
      return <ErrorSelector error={error} resource={`Operator ${operatorName}`} />;
    }
    if (isFetching || !operator) {
      return <LoadingWrapper />;
    }
    const channel = Operators.getDefaultChannel(operator);
    if (!channel) {
      return (
        <UnexpectedErrorPage
          text={`Operator ${operatorName} doesn't define a valid channel. This is needed to extract required info.`}
        />
      );
    }
    const { currentCSVDesc } = channel;
    const namespaced = currentCSVDesc.installModes.find(m => m.type === "AllNamespaces");
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
          namespaced={!namespaced?.supported}
          push={push}
          disableButton={!!csv}
        />
        <main>
          <div className="container container-fluid">
            <div className="row">
              <div className="col-9 ChartView__readme-container">
                <OperatorDescription description={currentCSVDesc.description} />
              </div>
              <div className="col-3 ChartView__sidebar-container">
                <aside className="ChartViewSidebar bg-light margin-v-big padding-h-normal padding-b-normal">
                  <div className="ChartViewSidebar__section">
                    <h2>Capability Level</h2>
                    <div>
                      <ul className="remove-style padding-l-reset margin-b-reset">
                        <li>
                          <CapabiliyLevel level={currentCSVDesc.annotations.capabilities} />
                        </li>
                      </ul>
                    </div>
                  </div>
                  <div className="ChartViewSidebar__section">
                    <h2>Repository</h2>
                    <div className="margin-l-big">
                      <span>
                        <a
                          href={currentCSVDesc.annotations.repository}
                          target="_blank"
                          rel="noopener noreferrer"
                        >
                          {currentCSVDesc.annotations.repository}
                        </a>
                      </span>
                    </div>
                  </div>
                  <div className="ChartViewSidebar__section">
                    <h2>Provider</h2>
                    <div className="margin-l-big">
                      <span>{operator.status.provider.name}</span>
                    </div>
                  </div>
                  <div className="ChartViewSidebar__section">
                    <h2>Container Image</h2>
                    <div className="margin-l-big">
                      <span>{currentCSVDesc.annotations.containerImage}</span>
                    </div>
                  </div>
                  <div className="ChartViewSidebar__section">
                    <h2>Created At</h2>
                    <div className="margin-l-big">
                      <span>{currentCSVDesc.annotations.createdAt}</span>
                    </div>
                  </div>
                </aside>
              </div>
            </div>
          </div>
        </main>
      </section>
    );
  }
}

export default OperatorView;
