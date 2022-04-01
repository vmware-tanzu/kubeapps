// Copyright 2020-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { CdsButton } from "@cds/react/button";
import { CdsIcon } from "@cds/react/icon";
import actions from "actions";
import { filtersToQuery } from "components/Catalog/Catalog";
import FilterGroup from "components/FilterGroup/FilterGroup";
import Alert from "components/js/Alert";
import Column from "components/js/Column";
import Row from "components/js/Row";
import { push } from "connected-react-router";
import { flatten, get, intersection, uniq, without } from "lodash";
import { ParsedQs } from "qs";
import { useEffect, useState } from "react";
import { useDispatch, useSelector } from "react-redux";
import { IPackageManifest, IPackageManifestStatus, IStoreState } from "shared/types";
import { app } from "shared/url";
import { escapeRegExp } from "shared/utils";
import LoadingWrapper from "../LoadingWrapper/LoadingWrapper";
import {
  AUTO_PILOT,
  BASIC_INSTALL,
  DEEP_INSIGHTS,
  FULL_LIFECYCLE,
  SEAMLESS_UPGRADES,
} from "../OperatorView/OperatorCapabilityLevel";
import PageHeader from "../PageHeader/PageHeader";
import SearchFilter from "../SearchFilter/SearchFilter";
import OLMNotFound from "./OLMNotFound";
import OperatorItems from "./OperatorItems";
import "./OperatorList.css";

export interface IOperatorListProps {
  cluster: string;
  namespace: string;
  filter: ParsedQs;
}

function getDefaultChannel(packageStatus: IPackageManifestStatus) {
  const defaultChannel = packageStatus.defaultChannel;
  const channel = packageStatus.channels.find(ch => ch.name === defaultChannel);
  return channel!;
}

function getCategories(packageStatus: IPackageManifestStatus) {
  const channel = getDefaultChannel(packageStatus);
  return get(channel, "currentCSVDesc.annotations.categories", "Unknown")
    .split(",")
    .map((c: string) => c.trim());
}

function getCapabilities(packageStatus: IPackageManifestStatus) {
  const channel = getDefaultChannel(packageStatus);
  return get(channel, "currentCSVDesc.annotations.capabilities", BASIC_INSTALL);
}

function getProvider(packageStatus: IPackageManifestStatus) {
  const channel = getDefaultChannel(packageStatus);
  return get(channel, "currentCSVDesc.provider.name", "");
}

export const filterNames = {
  SEARCH: "Search",
  CAPABILITY: "Capability",
  CATEGORY: "Category",
  PROVIDER: "Provider",
};

function initialFilterState() {
  const result = {};
  Object.values(filterNames).forEach(f => (result[f] = []));
  return result;
}

export default function OperatorList({
  cluster,
  namespace,
  filter: propsFilter,
}: IOperatorListProps) {
  const dispatch = useDispatch();
  const [filters, setFilters] = useState(initialFilterState());

  useEffect(() => {
    const tmpStrRegex = /__/g;
    const newFilters = {};
    Object.keys(propsFilter).forEach(filter => {
      const filterValue = propsFilter[filter]?.toString() || "";
      newFilters[filter] = filterValue.split(",").map(a => a.replace(tmpStrRegex, ","));
    });
    setFilters({
      ...initialFilterState(),
      ...newFilters,
    });
  }, [propsFilter]);

  const pushFilters = (newFilters: any) => {
    dispatch(push(app.operators.list(cluster, namespace) + filtersToQuery(newFilters)));
  };
  const addFilter = (type: string, value: string) => {
    pushFilters({
      ...filters,
      [type]: filters[type].concat(value),
    });
  };
  const removeFilter = (type: string, value: string) => {
    pushFilters({
      ...filters,
      [type]: without(filters[type], value),
    });
  };
  const removeFilterFunc = (type: string, value: string) => {
    return () => removeFilter(type, value);
  };
  const clearAllFilters = () => {
    pushFilters({});
  };
  const submitFilters = () => {
    pushFilters(filters);
  };

  // Only one search filter can be set
  const searchFilter = filters[filterNames.SEARCH][0] || "";
  const setSearchFilter = (searchTerm: string) => {
    setFilters({
      ...filters,
      [filterNames.SEARCH]: [searchTerm],
    });
  };

  useEffect(() => {
    dispatch(actions.operators.checkOLMInstalled(cluster, namespace));
  }, [dispatch, cluster, namespace]);

  const {
    operators: {
      operators,
      isFetching,
      errors: {
        operator: { fetch: opError },
        subscriptions: { fetch: subsError },
      },
      subscriptions,
      isOLMInstalled,
    },
  } = useSelector((state: IStoreState) => state);
  const error = opError || subsError;

  useEffect(() => {
    if (isOLMInstalled) {
      dispatch(actions.operators.getOperators(cluster, namespace));
      dispatch(actions.operators.listSubscriptions(cluster, namespace));
    }
  }, [dispatch, cluster, namespace, isOLMInstalled]);

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
  const allProviders = uniq(operators.map(operator => getProvider(operator.status))).sort();

  const subscriptionNames = subscriptions.map(subscription => subscription.spec.name);
  const installedOperators: IPackageManifest[] = [];
  const availableOperators: IPackageManifest[] = [];
  const filteredOperators = operators
    .filter(
      c =>
        filters[filterNames.CAPABILITY].length === 0 ||
        filters[filterNames.CAPABILITY].includes(getCapabilities(c.status)),
    )
    .filter(c => new RegExp(escapeRegExp(searchFilter), "i").test(c.metadata.name))
    .filter(
      c =>
        filters[filterNames.CATEGORY].length === 0 ||
        intersection(filters[filterNames.CATEGORY], getCategories(c.status)).length,
    )
    .filter(
      c =>
        filters[filterNames.PROVIDER].length === 0 ||
        filters[filterNames.PROVIDER].includes(getProvider(c.status)),
    );
  filteredOperators.forEach(operator => {
    if (subscriptionNames.includes(operator.metadata.name)) {
      installedOperators.push(operator);
    } else {
      availableOperators.push(operator);
    }
  });

  /* eslint-disable jsx-a11y/label-has-associated-control */
  return (
    <section>
      <PageHeader
        title="Operators"
        filter={
          <SearchFilter
            key="searchFilter"
            placeholder="search operators..."
            onChange={setSearchFilter}
            value={searchFilter}
            submitFilters={submitFilters}
          />
        }
      />
      <Alert theme="warning">
        <div>
          Operators integration is under heavy development and currently in beta state. If you find
          an issue please report it{" "}
          <a
            target="_blank"
            rel="noopener noreferrer"
            href="https://github.com/vmware-tanzu/kubeapps/issues"
          >
            here.
          </a>
        </div>
      </Alert>
      <LoadingWrapper
        className="margin-t-xxl"
        loadingText="Fetching Operators..."
        loaded={!isFetching}
      >
        {!isOLMInstalled ? (
          <OLMNotFound />
        ) : (
          <>
            {error && (
              <Alert theme="danger">
                An error occurred while fetching Operators: {error.message}
              </Alert>
            )}
            {operators.length === 0 ? (
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
                    <h5>
                      Filters{" "}
                      {flatten(Object.values(filters)).length ? (
                        <CdsButton size="sm" action="flat" onClick={clearAllFilters}>
                          Clear All
                        </CdsButton>
                      ) : (
                        <></>
                      )}{" "}
                    </h5>
                    {allCategories.length > 0 && (
                      <div className="filter-section">
                        <label className="filter-label">Category</label>
                        <FilterGroup
                          name={filterNames.CATEGORY}
                          options={allCategories}
                          currentFilters={filters[filterNames.CATEGORY]}
                          onAddFilter={addFilter}
                          onRemoveFilter={removeFilter}
                        />
                      </div>
                    )}
                    {allCapabilities.length > 0 && (
                      <div className="filter-section">
                        <label>Capability</label>
                        <FilterGroup
                          name={filterNames.CAPABILITY}
                          options={allCapabilities}
                          currentFilters={filters[filterNames.CAPABILITY]}
                          onAddFilter={addFilter}
                          onRemoveFilter={removeFilter}
                        />
                      </div>
                    )}
                    {allProviders.length > 0 && (
                      <div className="filter-section">
                        <label>Provider</label>
                        <FilterGroup
                          name={filterNames.PROVIDER}
                          options={allProviders}
                          currentFilters={filters[filterNames.PROVIDER]}
                          onAddFilter={addFilter}
                          onRemoveFilter={removeFilter}
                        />
                      </div>
                    )}
                  </div>
                </Column>
                <Column span={10}>
                  <>
                    <div className="filter-summary">
                      {Object.keys(filters).map(filterName => {
                        if (filters[filterName].length) {
                          return filters[filterName].map((filterValue: string) => (
                            <span key={`${filterName}-${filterValue}`} className="label label-info">
                              {filterName}: {filterValue}{" "}
                              <CdsIcon
                                shape="times"
                                onClick={removeFilterFunc(filterName, filterValue)}
                              />
                            </span>
                          ));
                        }
                        return null;
                      })}
                    </div>
                    {installedOperators.length > 0 && (
                      <>
                        <div className="operator-list-container">
                          <h3>Installed</h3>
                          <Row>
                            <OperatorItems operators={installedOperators} cluster={cluster} />
                          </Row>
                        </div>
                      </>
                    )}
                    <div className="operator-list-container">
                      <h3>Available Operators</h3>
                      <Row>
                        <OperatorItems operators={availableOperators} cluster={cluster} />
                      </Row>
                    </div>
                  </>
                </Column>
              </Row>
            )}
          </>
        )}
      </LoadingWrapper>
    </section>
  );
}
