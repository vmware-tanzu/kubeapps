import { RouterAction } from "connected-react-router";
import * as React from "react";

import { ForbiddenError, IClusterServiceVersion, IPackageManifest } from "../../shared/types";
import { api, app } from "../../shared/url";
import { escapeRegExp } from "../../shared/utils";
import { CardGrid } from "../Card";
import { ErrorSelector, MessageAlert } from "../ErrorAlert";
import InfoCard from "../InfoCard";
import LoadingWrapper from "../LoadingWrapper";
import PageHeader from "../PageHeader";
import SearchFilter from "../SearchFilter";
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
  filter: string;
  pushSearchFilter: (filter: string) => RouterAction;
}

export interface IOperatorListState {
  filter: string;
}

class OperatorList extends React.Component<IOperatorListProps, IOperatorListState> {
  public state: IOperatorListState = {
    filter: "",
  };

  public componentDidMount() {
    this.props.checkOLMInstalled();
    this.props.getOperators(this.props.namespace);
    this.props.getCSVs(this.props.namespace);
    this.setState({ filter: this.props.filter });
  }

  public componentDidUpdate(prevProps: IOperatorListProps) {
    if (prevProps.namespace !== this.props.namespace) {
      this.props.getOperators(this.props.namespace);
      this.props.getCSVs(this.props.namespace);
    }
    if (this.props.filter !== prevProps.filter) {
      this.props.getOperators(this.props.namespace);
      this.props.getCSVs(this.props.namespace);
      this.setState({ filter: this.props.filter });
    }
  }

  public render() {
    const { isFetching, pushSearchFilter } = this.props;
    return (
      <div>
        <PageHeader>
          <h1>Operators</h1>
          <SearchFilter
            className="margin-l-big"
            placeholder="search operators..."
            onChange={this.handleFilterQueryChange}
            value={this.state.filter}
            onSubmit={pushSearchFilter}
          />
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
          <LoadingWrapper loaded={!isFetching}>{this.renderOperators()}</LoadingWrapper>
        </main>
      </div>
    );
  }

  private renderOperators() {
    const { operators, error, csvs, isOLMInstalled } = this.props;
    const { filter } = this.state;
    if (error && error.constructor === ForbiddenError) {
      return (
        <ErrorSelector
          error={error}
          action="list"
          resource="Operators"
          namespace={this.props.namespace}
        />
      );
    }
    if (!isOLMInstalled) {
      return <OLMNotFound />;
    }
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
    const filteredOperators = operators.filter(c =>
      new RegExp(escapeRegExp(filter), "i").test(c.metadata.name),
    );
    if (filteredOperators.length === 0) {
      return <p>No Operator found</p>;
    }
    filteredOperators.forEach(operator => {
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

  private handleFilterQueryChange = (filter: string) => {
    this.setState({
      filter,
    });
  };
}

export default OperatorList;
