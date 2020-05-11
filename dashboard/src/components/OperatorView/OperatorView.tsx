import * as React from "react";

import { RouterAction } from "connected-react-router";
import { IPackageManifest } from "../../shared/types";
import { api } from "../../shared/url";
import { ErrorSelector } from "../ErrorAlert";
import UnexpectedErrorPage from "../ErrorAlert/UnexpectedErrorAlert";
import LoadingWrapper from "../LoadingWrapper";
import CapabiliyLevel from "./OperatorCapabilityLevel";
import OperatorDescription from "./OperatorDescription";
import OperatorHeader from "./OperatorHeader";

interface IOperatorViewProps {
  operatorName: string;
  operator?: IPackageManifest;
  getOperator: (namespace: string, name: string) => Promise<void>;
  isFetching: boolean;
  namespace: string;
  error?: Error;
  push: (location: string) => RouterAction;
}

class OperatorView extends React.Component<IOperatorViewProps> {
  public componentDidMount() {
    const { operatorName, namespace, getOperator } = this.props;
    getOperator(namespace, operatorName);
  }

  public render() {
    const { isFetching, namespace, operatorName, operator, error, push } = this.props;
    if (error) {
      return <ErrorSelector error={error} resource={`Operator ${operatorName}`} />;
    }
    if (isFetching || !operator) {
      return <LoadingWrapper />;
    }
    const channel = operator.status.channels.find(ch => ch.name === operator.status.defaultChannel);
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
          namespace={namespace}
          provider={operator.status.provider.name}
          namespaced={!namespaced?.supported}
          push={push}
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
                        <a href={currentCSVDesc.annotations.repository} target="_blank">
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
