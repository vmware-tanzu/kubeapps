import * as React from "react";

import { IClusterServiceVersion, IPackageManifest } from "shared/types";
import { api, app } from "../../shared/url";
import { CardGrid } from "../Card";
import { ErrorSelector } from "../ErrorAlert";
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
  getCSVs: (namespace: string) => Promise<IClusterServiceVersion[]>;
  csvs: IClusterServiceVersion[];
}

class OperatorList extends React.Component<IOperatorListProps> {
  public componentDidMount() {
    this.props.checkOLMInstalled();
    this.props.getOperators(this.props.namespace);
    this.props.getCSVs(this.props.namespace);
  }

  public componentDidUpdate(prevProps: IOperatorListProps) {
    if (prevProps.namespace !== this.props.namespace) {
      this.props.getOperators(this.props.namespace);
      this.props.getCSVs(this.props.namespace);
    }
  }

  public render() {
    const { isFetching, isOLMInstalled } = this.props;
    return (
      <div>
        <PageHeader>
          <h1>Operators</h1>
        </PageHeader>
        <main>
          <LoadingWrapper loaded={!isFetching}>
            {isOLMInstalled ? this.renderOperators() : <OLMNotFound />}
          </LoadingWrapper>
        </main>
      </div>
    );
  }

  private renderOperators() {
    const { operators, error, csvs } = this.props;
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
    const csvNames = csvs.map(csv => csv.metadata.name);
    const installedOperators: IPackageManifest[] = [];
    const availableOperators: IPackageManifest[] = [];
    operators.forEach(operator => {
      const defaultChannel = operator.status.defaultChannel;
      const channel = operator.status.channels.find(ch => ch.name === defaultChannel);
      if (csvNames.some(csvName => csvName === channel?.currentCSV)) {
        installedOperators.push(operator);
      } else {
        availableOperators.push(operator);
      }
    });
    return (
      <>
        {installedOperators.length > 0 && (
          <>
            <h3>Installed</h3>
            <CardGrid>
              {installedOperators.map(operator => {
                return (
                  <InfoCard
                    key={operator.metadata.name}
                    link={app.operators.view(this.props.namespace, operator.metadata.name)}
                    title={operator.metadata.name}
                    icon={api.operators.operatorIcon(this.props.namespace, operator.metadata.name)}
                    info={`v${operator.status.channels[0].currentCSVDesc.version}`}
                    tag1Content={operator.status.channels[0].currentCSVDesc.annotations.categories}
                    tag2Content={operator.status.provider.name}
                  />
                );
              })}
            </CardGrid>
          </>
        )}
        <h3>Available Operators</h3>
        <CardGrid>
          {availableOperators.map(operator => {
            return (
              <InfoCard
                key={operator.metadata.name}
                link={app.operators.view(this.props.namespace, operator.metadata.name)}
                title={operator.metadata.name}
                icon={api.operators.operatorIcon(this.props.namespace, operator.metadata.name)}
                info={`v${operator.status.channels[0].currentCSVDesc.version}`}
                tag1Content={operator.status.channels[0].currentCSVDesc.annotations.categories}
                tag2Content={operator.status.provider.name}
              />
            );
          })}
        </CardGrid>
      </>
    );
  }
}

export default OperatorList;
