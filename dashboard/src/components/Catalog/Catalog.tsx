import { CdsButton } from "@clr/react/button";
import { CdsIcon } from "@clr/react/icon";
import actions from "actions";
import FilterGroup from "components/FilterGroup/FilterGroup";
import Alert from "components/js/Alert";
import Column from "components/js/Column";
import Row from "components/js/Row";
import { push } from "connected-react-router";
import { debounce, flatten, get, intersection, trimStart, uniq, without } from "lodash";
import { ParsedQs } from "qs";
import React, { useCallback, useEffect, useState } from "react";
import { useDispatch } from "react-redux";
import { Link } from "react-router-dom";
import { app } from "shared/url";
import { IChartState, IClusterServiceVersion } from "../../shared/types";
import { escapeRegExp } from "../../shared/utils";
import PageHeader from "../PageHeader/PageHeader";
import SearchFilter from "../SearchFilter/SearchFilter";
import "./Catalog.css";
import CatalogItems from "./CatalogItems";

function categoryToReadable(category: string) {
  return category === "" ? "Unknown" : category.replace(/([a-z])([A-Z][a-z])/g, "$1 $2").trimLeft();
}

function getOperatorCategories(c: IClusterServiceVersion): string[] {
  return get(c, "metadata.annotations.categories", "")
    .split(",")
    .map((category: string) => categoryToReadable(category));
}

interface ICatalogProps {
  charts: IChartState;
  repo: string;
  filter: ParsedQs;
  fetchCharts: (cluster: string, namespace: string, repos: string, query: string) => void;
  cluster: string;
  namespace: string;
  kubeappsNamespace: string;
  fetchChartCategories: (cluster: string, namespace: string) => void;
  getCSVs: (cluster: string, namespace: string) => void;
  csvs: IClusterServiceVersion[];
}

export const filterNames = {
  SEARCH: "Search",
  TYPE: "Type",
  REPO: "Repository",
  CATEGORY: "Category",
  OPERATOR_PROVIDER: "Provider",
};

export function initialFilterState() {
  const result = {};
  Object.values(filterNames).forEach(f => (result[f] = []));
  return result;
}

export function filtersToQuery(filters: any) {
  let query = "";
  const activeFilters = Object.keys(filters).filter(f => filters[f].length);
  if (activeFilters.length) {
    const filterQueries = activeFilters.map(
      filter => `${filter}=${filters[filter].map((f: string) => encodeURIComponent(f)).join(",")}`,
    );
    query = "?" + filterQueries.join("&");
  }
  return query;
}

function Catalog(props: ICatalogProps) {
  const {
    charts: {
      status,
      selected: { error },
      items: charts,
      categories,
      search,
    },
    fetchCharts,
    cluster,
    namespace,
    fetchChartCategories,
    getCSVs,
    csvs,
    repo,
    filter: propsFilter,
  } = props;

  const dispatch = useDispatch();
  const [filters, setFilters] = useState(initialFilterState());
  const [currentSearchQuery, setCurrentSearchQuery] = useState("");

  useEffect(() => {
    const newFilters = {};
    Object.keys(propsFilter).forEach(filter => {
      newFilters[filter] = propsFilter[filter]?.toString().split(",");
    });
    setFilters({
      ...initialFilterState(),
      ...newFilters,
    });
  }, [propsFilter]);

  const pushFilters = (newFilters: any, type?: string) => {
    dispatch(push(app.catalog(cluster, namespace) + filtersToQuery(newFilters)));
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
  const removeSearchQuery = () => {
    return () => {
      dispatch(actions.charts.resetChartsSearch());
      setCurrentSearchQuery("");
    };
  };
  const clearAllFilters = () => {
    dispatch(actions.charts.resetChartsSearch());
    setCurrentSearchQuery("");
    pushFilters({}, "");
  };

  const allRepos = uniq(charts.map(c => c.attributes.repo.name));
  const allProviders = uniq(csvs.map(c => c.spec.provider.name));
  const allCategories = uniq(
    categories
      .map(c => categoryToReadable(c.name))
      .concat(flatten(csvs.map(c => getOperatorCategories(c)))),
  ).sort();

  useEffect(() => {
    fetchChartCategories(cluster, namespace);
    getCSVs(cluster, namespace);
  }, [dispatch, getCSVs, fetchChartCategories, cluster, namespace]);

  useEffect(() => {
    fetchCharts(cluster, namespace, repo, currentSearchQuery);
  }, [dispatch, fetchCharts, cluster, namespace, repo, currentSearchQuery]);

  const debouncedfetchChartsSearch = useCallback(
    debounce((q: string) => {
      fetchCharts(cluster, namespace, repo, q);
    }, 500),
    [fetchCharts, cluster, namespace, repo],
  );

  // Only one search filter can be set
  const searchFilter = filters[filterNames.SEARCH][0] || "";
  const setSearchFilter = useCallback(
    (searchTerm: string) => {
      setCurrentSearchQuery(trimStart(searchTerm));
      if (currentSearchQuery.length) {
        debouncedfetchChartsSearch(currentSearchQuery);
      }
    },
    [currentSearchQuery, debouncedfetchChartsSearch],
  );

  const filteredCharts = charts
    .filter(
      () => filters[filterNames.TYPE].length === 0 || filters[filterNames.TYPE].includes("Charts"),
    )
    .filter(() => filters[filterNames.OPERATOR_PROVIDER].length === 0)
    .filter(c => new RegExp(escapeRegExp(searchFilter), "i").test(c.id))
    .filter(
      c =>
        filters[filterNames.REPO].length === 0 ||
        filters[filterNames.REPO].includes(c.attributes.repo.name),
    )
    .filter(
      c =>
        filters[filterNames.CATEGORY].length === 0 ||
        filters[filterNames.CATEGORY].includes(categoryToReadable(c.attributes.category)),
    );
  const filteredChartsSearch = search.items
    .filter(
      () => filters[filterNames.TYPE].length === 0 || filters[filterNames.TYPE].includes("Charts"),
    )
    .filter(() => filters[filterNames.OPERATOR_PROVIDER].length === 0)
    .filter(
      c =>
        filters[filterNames.REPO].length === 0 ||
        filters[filterNames.REPO].includes(c.attributes.repo.name),
    )
    .filter(
      c =>
        filters[filterNames.CATEGORY].length === 0 ||
        filters[filterNames.CATEGORY].includes(categoryToReadable(c.attributes.category)),
    );
  const filteredCSVs = csvs
    .filter(
      () =>
        filters[filterNames.TYPE].length === 0 || filters[filterNames.TYPE].includes("Operators"),
    )
    .filter(() => filters[filterNames.REPO].length === 0)
    .filter(c => {
      const regex = new RegExp(escapeRegExp(currentSearchQuery), "i");
      return (
        regex.test(c.metadata.name) ||
        c?.spec?.customresourcedefinitions?.owned.find(crd => regex.test(crd.displayName))
      );
    })
    .filter(
      c =>
        filters[filterNames.OPERATOR_PROVIDER].length === 0 ||
        filters[filterNames.OPERATOR_PROVIDER].includes(c.spec.provider.name),
    )
    .filter(
      c =>
        filters[filterNames.CATEGORY].length === 0 ||
        intersection(filters[filterNames.CATEGORY], getOperatorCategories(c)).length,
    );

  return (
    <section>
      <PageHeader
        title="Catalog"
        filter={
          <SearchFilter
            key="searchFilter"
            placeholder="search charts..."
            onChange={setSearchFilter}
            value={currentSearchQuery}
            submitFilters={setSearchFilter}
          />
        }
      />
      {error && (
        <Alert theme="danger">An error occurred while fetching the catalog: {error.message}</Alert>
      )}
      {charts.length === 0 && filteredChartsSearch.length === 0 && csvs.length === 0 ? (
        <div className="empty-catalog">
          <CdsIcon shape="bundle" />
          <p>The current catalog is empty.</p>
          <p>
            Manage your Helm chart repositories in Kubeapps by visiting the App repositories
            configuration page.
          </p>
          <Link to={app.config.apprepositories(cluster, namespace)}>
            <CdsButton>Manage App Repositories</CdsButton>
          </Link>
        </div>
      ) : (
        <Row>
          <Column span={2}>
            <div className="filters-menu">
              <h5>
                Filters{" "}
                {flatten(Object.values(filters)).length || search.query.length ? (
                  <CdsButton size="sm" action="flat" onClick={clearAllFilters}>
                    Clear All
                  </CdsButton>
                ) : (
                  <></>
                )}{" "}
              </h5>
              {csvs.length > 0 && (
                <div className="filter-section">
                  <label>Application Type</label>
                  <FilterGroup
                    name={filterNames.TYPE}
                    options={["Operators", "Charts"]}
                    currentFilters={filters[filterNames.TYPE]}
                    onAddFilter={addFilter}
                    onRemoveFilter={removeFilter}
                  />
                </div>
              )}
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
              {allRepos.length > 0 && (
                <div className="filter-section">
                  <label>Application Repository</label>
                  <FilterGroup
                    name={filterNames.REPO}
                    options={allRepos}
                    currentFilters={filters[filterNames.REPO]}
                    onAddFilter={addFilter}
                    onRemoveFilter={removeFilter}
                  />
                </div>
              )}
              {allProviders.length > 0 && (
                <div className="filter-section">
                  <label className="filter-label">Operator Provider</label>
                  <FilterGroup
                    name={filterNames.OPERATOR_PROVIDER}
                    options={allProviders}
                    currentFilters={filters[filterNames.OPERATOR_PROVIDER]}
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
                {search.query.length ? (
                  <>
                    <span key={`query-${search.query}`} className="label label-info">
                      Query: {search.query}
                      <CdsIcon shape="times" onClick={removeSearchQuery()} />
                    </span>
                  </>
                ) : (
                  <></>
                )}
              </div>
              <Row>
                <CatalogItems
                  charts={search.query.length > 0 ? filteredChartsSearch : filteredCharts}
                  csvs={filteredCSVs}
                  cluster={cluster}
                  namespace={namespace}
                  hasFinishedFetching={status === actions.charts.finishedStatus}
                />
              </Row>
            </>
          </Column>
        </Row>
      )}
    </section>
  );
}

export default Catalog;
