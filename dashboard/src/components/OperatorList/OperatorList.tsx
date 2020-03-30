import * as React from "react";

import { IPackageManifest } from "shared/types";
import { api } from "../../shared/url";
import { CardGrid } from "../Card";
import { ErrorSelector, MessageAlert } from "../ErrorAlert";
import InfoCard from "../InfoCard";
import LoadingWrapper from "../LoadingWrapper";
import PageHeader from "../PageHeader";
import OLMNotFound from "./OLMNotFound";

export interface IOperatorListProps {
  isFetching: boolean;
  checkOLMInstalled: () => Promise<boolean>;
  isOLMInstalled: boolean;
  namespace: string;
  getOperators: (namespace: string) => Promise<void>;
  operators: IPackageManifest[];
  error?: Error;
}

class OperatorList extends React.Component<IOperatorListProps> {
  public componentDidMount() {
    this.props.checkOLMInstalled();
    this.props.getOperators(this.props.namespace);
  }

  public render() {
    const { isFetching, isOLMInstalled } = this.props;
    return (
      <div>
        <PageHeader>
          <h1>Operators</h1>
        </PageHeader>
        <main>
          <MessageAlert level="warning">
            <div>
              Operators integration is under heavy development and currently in alpha state. If you
              find an issue please report it{" "}
              <a target="_blank" href="https://github.com/kubeapps/kubeapps/issues">
                here.
              </a>
            </div>
          </MessageAlert>
          <LoadingWrapper loaded={!isFetching}>
            {isOLMInstalled ? this.renderOperators() : <OLMNotFound />}
          </LoadingWrapper>
        </main>
      </div>
    );
  }

  private renderOperators() {
    const { operators, error } = this.props;
    if (error) {
      return (
        <ErrorSelector
          error={error}
          action="list"
          resource="Operators"
          namespace={this.props.namespace}
        />
      );
    }
    return (
      <CardGrid>
        {operators.map(operator => {
          return (
            <InfoCard
              key={operator.metadata.name}
              link={`/operators/ns/${this.props.namespace}/${operator.metadata.name}`}
              title={operator.metadata.name}
              icon={api.operators.operatorIcon(this.props.namespace, operator.metadata.name)}
              info={`v${operator.status.channels[0].currentCSVDesc.version}`}
              tag1Content={operator.status.channels[0].currentCSVDesc.annotations.categories}
              tag2Content={operator.status.provider.name}
            />
          );
        })}
      </CardGrid>
    );
  }
}

export default OperatorList;
