import actions from "actions";
import CardGrid from "components/Card/CardGrid.v2";
import { CdsIcon } from "components/Clarity/clarity";
import FilterGroup from "components/FilterGroup/FilterGroup";
import Alert from "components/js/Alert";
import Column from "components/js/Column";
import Row from "components/js/Row";
import { RouterAction } from "connected-react-router";
import { flatten, intersection, uniq } from "lodash";
import React, { useEffect, useState } from "react";
import { useDispatch, useSelector } from "react-redux";
import { IPackageManifestStatus, IStoreState } from "../../shared/types";
import { escapeRegExp } from "../../shared/utils";
import LoadingWrapper from "../LoadingWrapper/LoadingWrapper.v2";
import {
  AUTO_PILOT,
  BASIC_INSTALL,
  DEEP_INSIGHTS,
  FULL_LIFECYCLE,
  SEAMLESS_UPGRADES,
} from "../OperatorView/OperatorCapabilityLevel";
import PageHeader from "../PageHeader/PageHeader.v2";
import SearchFilter from "../SearchFilter/SearchFilter.v2";
import OLMNotFound from "./OLMNotFound.v2";
import OperatorItems from "./OperatorItems";
import OperatorNotSupported from "./OperatorsNotSupported.v2";

export interface IOperatorListProps {
  cluster: string;
  namespace: string;
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

export default function OperatorList({
  cluster,
  namespace,
  filter,
  pushSearchFilter,
}: IOperatorListProps) {
  const dispatch = useDispatch();
  const [searchFilter, setSearchFilter] = useState(filter);
  const [categoryFilter, setCategoryFilter] = useState([] as string[]);
  const [capabilitiesFilter, setCapabilitiesFilter] = useState([] as string[]);

  useEffect(() => {
    setSearchFilter(filter);
  }, [filter]);

  useEffect(() => {
    dispatch(actions.operators.checkOLMInstalled(cluster, namespace));
  }, [dispatch, cluster, namespace]);

  const {
    operators: {
      operators,
      isFetching,
      errors: {
        operator: { fetch: error },
      },
      isOLMInstalled,
    },
    config: { kubeappsCluster },
  } = useSelector((state: IStoreState) => state);

  useEffect(() => {
    if (isOLMInstalled) {
      dispatch(actions.operators.getOperators(cluster, namespace));
    }
  }, [dispatch, cluster, namespace, isOLMInstalled]);

  if (cluster !== kubeappsCluster) {
    return <OperatorNotSupported kubeappsCluster={kubeappsCluster} namespace={namespace} />;
  }

  const allCapabilities = [
    BASIC_INSTALL,
    SEAMLESS_UPGRADES,
    FULL_LIFECYCLE,
    DEEP_INSIGHTS,
    AUTO_PILOT,
  ];
  const allCategories = uniq(
    flatten(operators.map(operator => getCategories(operator.status))),
  ).sort();

  const filteredOperators = operators
    .filter(
      c =>
        capabilitiesFilter.length === 0 || capabilitiesFilter.includes(getCapabilities(c.status)),
    )
    .filter(c => new RegExp(escapeRegExp(searchFilter), "i").test(c.metadata.name))
    .filter(
      c =>
        categoryFilter.length === 0 || intersection(categoryFilter, getCategories(c.status)).length,
    );

  return (
    <section>
      <PageHeader>
        <Row>
          <h1>Operators</h1>
          <SearchFilter
            key="searchFilter"
            placeholder="search charts..."
            onChange={setSearchFilter}
            value={searchFilter}
            onSubmit={pushSearchFilter}
          />
        </Row>
      </PageHeader>
      <Alert theme="warning">
        <div>
          Operators integration is under heavy development and currently in beta state. If you find
          an issue please report it{" "}
          <a
            target="_blank"
            rel="noopener noreferrer"
            href="https://github.com/kubeapps/kubeapps/issues"
          >
            here.
          </a>
        </div>
      </Alert>
      <LoadingWrapper loaded={!isFetching}>
        {error && <Alert theme="danger">Found en error fetching operators: {error.message}</Alert>}
        {!isOLMInstalled ? (
          <OLMNotFound />
        ) : operators.length === 0 ? (
          <div className="section-not-found">
            <div>
              <CdsIcon shape="bundle" size="64" />
              <h4>The list of Operators is empty</h4>
              <p>
                This may mean that the OLM is still populating the catalog or that it found an
                error. Check the OLM logs for more information.
              </p>
            </div>
          </div>
        ) : (
          <Row>
            <Column span={2}>
              <div className="filters-menu">
                <h5>Filters</h5>
                {allCategories.length > 0 && (
                  <div className="filter-section">
                    <label className="filter-label">Category:</label>
                    <FilterGroup
                      name="category"
                      options={allCategories}
                      onChange={setCategoryFilter}
                    />
                  </div>
                )}
                {allCapabilities.length > 0 && (
                  <div className="filter-section">
                    <label>Application Repository:</label>
                    <FilterGroup
                      name="apprepo"
                      options={allCapabilities}
                      onChange={setCapabilitiesFilter}
                    />
                  </div>
                )}
              </div>
            </Column>
            <Column span={10}>
              <CardGrid>
                <OperatorItems
                  operators={filteredOperators}
                  cluster={cluster}
                  namespace={namespace}
                />
              </CardGrid>
            </Column>
          </Row>
        )}
      </LoadingWrapper>
    </section>
  );
}
