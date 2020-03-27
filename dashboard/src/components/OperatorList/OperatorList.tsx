import * as React from "react";
import { Tab, TabList, TabPanel, Tabs } from "react-tabs";

import { IClusterServiceVersion, IPackageManifest } from "shared/types";
import { api } from "../../shared/url";
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
    return (
      <Tabs className="margin-t-big">
        <TabList>
          <Tab>Browse</Tab>
          <Tab>Installed</Tab>
        </TabList>
        <TabPanel>
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
        </TabPanel>
        <TabPanel>
          <CardGrid>
            {csvs.map(csv => {
              const op = operators.find(operator => {
                const defaultChannel = operator.status.defaultChannel;
                const channel = operator.status.channels.find(ch => ch.name === defaultChannel);
                return channel?.currentCSV === csv.metadata.name;
              });
              return (
                <InfoCard
                  key={csv.metadata.name}
                  link={`/operators/ns/${this.props.namespace}/${op?.metadata.name}`}
                  title={op?.metadata.name || csv.metadata.name}
                  icon={`data:${csv.spec.icon[0].mediatype};base64,${csv.spec.icon[0].base64data}`}
                  info={`v${csv.spec.version}`}
                  tag1Content={csv.spec.provider.name}
                  tag2Content={csv.metadata.namespace}
                />
              );
            })}
          </CardGrid>
        </TabPanel>
      </Tabs>
    );
  }
}

export default OperatorList;
