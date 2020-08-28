import { RouterAction } from "connected-react-router";
import { flatten, intersection, uniq } from "lodash";
import * as React from "react";

import {
  ForbiddenError,
  IClusterServiceVersion,
  IPackageManifest,
  IPackageManifestStatus,
} from "../../shared/types";
import { api, app } from "../../shared/url";
import { CardGrid } from "../Card";
import { ErrorSelector, MessageAlert } from "../ErrorAlert";
import InfoCard from "../InfoCard";
import LoadingWrapper from "../LoadingWrapper";
import {
  AUTO_PILOT,
  BASIC_INSTALL,
  DEEP_INSIGHTS,
  FULL_LIFECYCLE,
  SEAMLESS_UPGRADES,
} from "../OperatorView/OperatorCapabilityLevel";
import PageHeader from "../PageHeader";
import SearchFilter from "../SearchFilter";
import OLMNotFound from "./OLMNotFound";
import OperatorNotSupported from "./OperatorsNotSupported";

import "./OperatorList.css";

export interface IOperatorListProps {
  isFetching: boolean;
  checkOLMInstalled: (cluster: string, namespace: string) => Promise<boolean>;
  isOLMInstalled: boolean;
  cluster: string;
  namespace: string;
  kubeappsCluster: string;
  getOperators: (cluster: string, namespace: string) => Promise<void>;
  operators: IPackageManifest[];
  error?: Error;
  getCSVs: (cluster: string, namespace: string) => Promise<IClusterServiceVersion[]>;
  csvs: IClusterServiceVersion[];
  filter: string;
  pushSearchFilter: (filter: string) => RouterAction;
}

export interface IOperatorListState {
  filter: string;
  categories: string[];
  filterCategories: { [key: string]: boolean };
  filterCapabilities: { [key: string]: boolean };
}

function getDefaultChannel(packageStatus: IPackageManifestStatus) {
  const defaultChannel = packageStatus.defaultChannel;
  const channel = packageStatus.channels.find(ch => ch.name === defaultChannel);
  return channel!;
}

function getCategories(packageStatus: IPackageManifestStatus) {
  const channel = getDefaultChannel(packageStatus);
  return channel.currentCSVDesc.annotations.categories.split(",").map(c => c.trim());
}

function getCapabilities(packageStatus: IPackageManifestStatus) {
  const channel = getDefaultChannel(packageStatus);
  return channel.currentCSVDesc.annotations.capabilities;
}

class OperatorList extends React.Component<IOperatorListProps, IOperatorListState> {
  public state: IOperatorListState = {
    filter: "",
    categories: [],
    filterCategories: {},
    filterCapabilities: {
      [BASIC_INSTALL]: false,
      [SEAMLESS_UPGRADES]: false,
      [FULL_LIFECYCLE]: false,
      [DEEP_INSIGHTS]: false,
      [AUTO_PILOT]: false,
    },
  };

  public componentDidMount() {
    this.props.checkOLMInstalled(this.props.cluster, this.props.namespace);
    this.props.getOperators(this.props.cluster, this.props.namespace);
    this.props.getCSVs(this.props.cluster, this.props.namespace);
    this.setState({ filter: this.props.filter });
  }

  public componentDidUpdate(prevProps: IOperatorListProps) {
    if (prevProps.namespace !== this.props.namespace) {
      this.props.getOperators(this.props.cluster, this.props.namespace);
      this.props.getCSVs(this.props.cluster, this.props.namespace);
    }
    if (this.props.filter !== prevProps.filter) {
      this.props.getOperators(this.props.cluster, this.props.namespace);
      this.props.getCSVs(this.props.cluster, this.props.namespace);
      this.setState({ filter: this.props.filter });
    }

    if (this.props.operators !== prevProps.operators) {
      const categories = uniq(
        flatten(this.props.operators.map(operator => getCategories(operator.status))),
      );
      const filterCategories = {};
      categories.forEach(category => {
        filterCategories[category] = false;
      });
      this.setState({ categories, filterCategories });
    }
  }

  public render() {
    const { cluster, kubeappsCluster, namespace, isFetching, pushSearchFilter } = this.props;
    if (cluster !== kubeappsCluster) {
      return <OperatorNotSupported kubeappsCluster={kubeappsCluster} namespace={namespace} />;
    }
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
              <a
                target="_blank"
                rel="noopener noreferrer"
                href="https://github.com/kubeapps/kubeapps/issues"
              >
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
    const { filter, filterCategories, filterCapabilities } = this.state;
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
    const filteredOperators = operators.filter(operator => {
      if (filter && !operator.metadata.name.match(filter)) {
        return false;
      }
      const hasFilteredCategories = Object.values(filterCategories).some(
        filteredCategory => filteredCategory,
      );
      if (hasFilteredCategories) {
        const allowedCategories = Object.keys(filterCategories).filter(
          cat => filterCategories[cat],
        );
        const categories = getCategories(operator.status);
        if (intersection(allowedCategories, categories).length === 0) {
          return false;
        }
      }
      const hasFilteredCapabilities = Object.values(filterCapabilities).some(
        filterCapability => filterCapability,
      );
      if (hasFilteredCapabilities) {
        const allowedCapabilities = Object.keys(filterCapabilities).filter(
          capability => filterCapabilities[capability],
        );
        if (
          !allowedCapabilities.some(capability => capability === getCapabilities(operator.status))
        ) {
          return false;
        }
      }
      return true;
    });
    filteredOperators.forEach(operator => {
      if (csvNames.some(csvName => csvName === getDefaultChannel(operator.status).currentCSV)) {
        installedOperators.push(operator);
      } else {
        availableOperators.push(operator);
      }
    });
    return (
      <div className="row margin-t-big">
        <div className="col-2 margin-t-big horizontal-column">
          <div className="margin-b-normal ">
            <b>Categories</b>
          </div>
          {this.state.categories.map(category => {
            return (
              <div key={category}>
                <label
                  className="checkbox"
                  key={category}
                  onChange={this.toggleFilterCategory(category)}
                >
                  <input type="checkbox" />
                  <span>{category}</span>
                </label>
              </div>
            );
          })}
          <div className="margin-v-normal ">
            <b>Capability Level</b>
          </div>
          {[BASIC_INSTALL, SEAMLESS_UPGRADES, FULL_LIFECYCLE, DEEP_INSIGHTS, AUTO_PILOT].map(
            capability => {
              return (
                <div key={capability}>
                  <label
                    className="checkbox"
                    key={capability}
                    onChange={this.toggleFilterCapability(capability)}
                  >
                    <input type="checkbox" />
                    <span>{capability}</span>
                  </label>
                </div>
              );
            },
          )}
        </div>
        <div className="col-10">
          <div className="padding-l-normal">
            {filteredOperators.length === 0 ? (
              <p>No Operator found</p>
            ) : (
              this.renderCardGrid(installedOperators, availableOperators)
            )}
          </div>
        </div>
      </div>
    );
  }

  private renderCardGrid(
    installedOperators: IPackageManifest[],
    availableOperators: IPackageManifest[],
  ) {
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
                    link={app.operators.view(
                      this.props.cluster,
                      this.props.namespace,
                      operator.metadata.name,
                    )}
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
                link={app.operators.view(
                  this.props.cluster,
                  this.props.namespace,
                  operator.metadata.name,
                )}
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

  private toggleFilterCategory = (category: string) => {
    return () => {
      const { filterCategories } = this.state;
      this.setState({
        filterCategories: {
          ...filterCategories,
          [category]: !filterCategories[category],
        },
      });
    };
  };

  private toggleFilterCapability = (capability: string) => {
    return () => {
      const { filterCapabilities } = this.state;
      this.setState({
        filterCapabilities: {
          ...filterCapabilities,
          [capability]: !filterCapabilities[capability],
        },
      });
    };
  };
}

export default OperatorList;
