import { CdsButton } from "@clr/react/button";
import { CdsIcon } from "@clr/react/icon";
import actions from "actions";
import FilterGroup from "components/FilterGroup/FilterGroup";
import Alert from "components/js/Alert";
import Column from "components/js/Column";
import Row from "components/js/Row";
import LoadingWrapper from "components/LoadingWrapper/LoadingWrapper";
import { push } from "connected-react-router";
import {
  debounce,
  flatten,
  get,
  intersection,
  isEqual,
  throttle,
  trimStart,
  uniq,
  without,
} from "lodash";
import { ParsedQs } from "qs";
import React, { useCallback, useEffect, useState } from "react";
import { useRef } from "react";
import { useDispatch, useSelector } from "react-redux";
import { Link } from "react-router-dom";
import { app } from "shared/url";
import { IChartState, IClusterServiceVersion, IStoreState } from "../../shared/types";
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
  fetchChartsWithPagination: (
    cluster: string,
    namespace: string,
    repo: string,
    query: string,
    nextPage: number,
    page: number,
    size: number,
  ) => void;
  fetchChartsSearch: (cluster: string, namespace: string, repo: string, query: string) => void;
  cluster: string;
  namespace: string;
  kubeappsNamespace: string;
  fetchChartCategories: (cluster: string, namespace: string, repo: string) => void;
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
      isFetching,
      nextPage,
      page,
      size,
      status,
      selected: { error },
      items: charts,
      search,
      categories,
    },
    fetchChartsWithPagination,
    fetchChartsSearch,
    fetchChartCategories,
    getCSVs,
    cluster,
    namespace,
    csvs,
    filter: propsFilter,
  } = props;
  const {
    repos: { repos },
    config: { kubeappsCluster, kubeappsNamespace },
  } = useSelector((state: IStoreState) => state);

  const dispatch = useDispatch();
  const [filters, setFilters] = useState(initialFilterState());
  const [currentSearchQuery, setCurrentSearchQuery] = useState("");
  const [currentRepo, setCurrentRepo] = useState("");

  const pushFilters = (newFilters: any, type: string) => {
    dispatch(push(app.catalog(cluster, namespace) + filtersToQuery(newFilters)));

    if (type === filterNames.REPO) {
      // if the repo changes, force reset
      dispatch(actions.charts.resetPaginaton());
      fetchChartsSearch(cluster, namespace, newFilters[filterNames.REPO], currentSearchQuery);
      fetchChartCategories(cluster, namespace, newFilters[filterNames.REPO]); // get corresponding categories
    }
  };
  const addFilter = (type: string, value: string) => {
    pushFilters(
      {
        ...filters,
        [type]: filters[type].concat(value),
      },
      type,
    );
  };
  const removeFilter = (type: string, value: string) => {
    pushFilters(
      {
        ...filters,
        [type]: without(filters[type], value),
      },
      type,
    );
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
    if (filters[filterNames.REPO]) {
      dispatch(actions.charts.resetPaginaton());
    }
    dispatch(actions.charts.resetChartsSearch());
    setCurrentSearchQuery("");
    pushFilters({}, "");
  };

  const allRepos = uniq(repos.map(c => c.metadata.name));
  const allProviders = uniq(csvs.map(c => c.spec.provider.name));
  const allCategories = uniq(
    categories
      .map(c => categoryToReadable(c.name))
      .concat(flatten(csvs.map(c => getOperatorCategories(c)))),
  ).sort();

  // We do not currently support app repositories on additional clusters.
  const supportedCluster = cluster === kubeappsCluster;
  // const fetchRepos: () => void = useCallback(() => {
  //   if (!namespace) {
  //     // All Namespaces
  //     dispatch(actions.repos.fetchRepos(""));
  //     return;
  //   }
  //   if (!supportedCluster || namespace === kubeappsNamespace) {
  //     // Global namespace or other cluster, show global repos only
  //     dispatch(actions.repos.fetchRepos(kubeappsNamespace));
  //     return;
  //   }
  //   // In other case, fetch global and namespace repos
  //   dispatch(actions.repos.fetchRepos(namespace, kubeappsNamespace));
  // }, [dispatch, supportedCluster, namespace, kubeappsNamespace]);

  useEffect(() => {
    const newFilters = {};
    Object.keys(propsFilter).forEach(filter => {
      newFilters[filter] = propsFilter[filter]?.toString().split(",");
    });
    // setCurrentRepo(propsFilter[filterNames.REPO].toString()); //FIXME(agamez): fix it once merged
    setFilters({
      ...initialFilterState(),
      ...newFilters,
    });
  }, [propsFilter]);

  const firstUpdate = useRef(true);

  const namespaceUpdate = useRef({ namespace });
  useEffect(() => {
    if (firstUpdate.current || namespaceUpdate.current.namespace !== namespace) {
      // initial actions or if only the namespace is changing, refresh
      firstUpdate.current = false;
      setCurrentSearchQuery("");
      setCurrentRepo("");
      setFilters({ ...initialFilterState() });
      dispatch(actions.charts.resetPaginaton());
      dispatch(actions.charts.resetChartsSearch());
      fetchChartsWithPagination(cluster, namespace, currentRepo, "", 1, size, nextPage); // get the first charts
      getCSVs(cluster, namespace);
      dispatch(actions.charts.resetPaginaton()); // start from page=1 again
    }
    namespaceUpdate.current.namespace = namespace;
  }, [
    dispatch,
    getCSVs,
    fetchChartsWithPagination,
    cluster,
    namespace,
    currentRepo,
    page,
    size,
    nextPage,
  ]);

  useEffect(() => {
    // when the current repo changes, re-fetch categories
    fetchChartCategories(cluster, namespace, currentRepo);
  }, [fetchChartCategories, cluster, namespace, currentRepo]);

  useEffect(() => {
    // when the namespace changes, re-fetch repos
    if (!namespace) {
      // All Namespaces
      dispatch(actions.repos.fetchRepos(""));
      return;
    }
    if (!supportedCluster || namespace === kubeappsNamespace) {
      // Global namespace or other cluster, show global repos only
      dispatch(actions.repos.fetchRepos(kubeappsNamespace));
      return;
    }
    // In other case, fetch global and namespace repos
    dispatch(actions.repos.fetchRepos(namespace, kubeappsNamespace));
  }, [dispatch, supportedCluster, namespace, kubeappsNamespace]);

  const debouncedfetchChartsSearch = useCallback(
    debounce((q: string) => {
      fetchChartsSearch(cluster, namespace, currentRepo, q);
    }, 500),
    [fetchChartsSearch, cluster, namespace, currentRepo],
  );

  const throttledfetchChartsSearch = useCallback(
    throttle((q: string) => {
      fetchChartsSearch(cluster, namespace, currentRepo, q);
    }, 300),
    [fetchChartsSearch, cluster, namespace, currentRepo],
  );

  const debouncedFetchChartsWithPagination = useCallback(
    debounce(() => {
      fetchChartsWithPagination(cluster, namespace, currentRepo, "", page, size, nextPage);
    }, 1000),
    [
      fetchChartsWithPagination,
      status,
      cluster,
      namespace,
      currentRepo,
      page,
      size,
      isFetching,
      nextPage,
    ],
  );

  // Only one search filter can be set
  const searchFilter = filters[filterNames.SEARCH][0] || "";

  const setSearchFilter = useCallback(
    (query: string) => {
      setCurrentSearchQuery(trimStart(query));
      if (currentSearchQuery.length) {
        if (currentSearchQuery.length < 5 || currentSearchQuery.endsWith(" ")) {
          throttledfetchChartsSearch(currentSearchQuery);
        } else {
          debouncedfetchChartsSearch(currentSearchQuery);
        }
      }
    },
    [currentSearchQuery, throttledfetchChartsSearch, debouncedfetchChartsSearch],
  );

  const filteredCharts = charts
    .filter(
      () => filters[filterNames.TYPE].length === 0 || filters[filterNames.TYPE].includes("Charts"),
    )
    .filter(() => filters[filterNames.OPERATOR_PROVIDER].length === 0)
    // .filter(c => new RegExp(escapeRegExp(searchFilter), "i").test(c.id))
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
    // .filter(c => new RegExp(escapeRegExp(search.query), "i").test(c.id))
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
    .filter(c => new RegExp(escapeRegExp(searchFilter), "i").test(c.metadata.name))
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

  const observeBorder = useCallback(
    // Check if the IntersectionAPI is enabled
    // TODO(agamez): add a "load more" manual button at the end if not
    node => {
      if (
        "IntersectionObserver" in window &&
        "IntersectionObserverEntry" in window &&
        "isIntersecting" in window.IntersectionObserverEntry.prototype
      ) {
        if (node !== null) {
          // https://developer.mozilla.org/en-US/docs/Web/API/IntersectionObserver
          new IntersectionObserver(
            entries => {
              // https://developer.mozilla.org/en-US/docs/Web/API/IntersectionObserverEntry
              entries.forEach(entry => {
                if (
                  entry.isIntersecting &&
                  !isFetching &&
                  !search.query.length &&
                  // status !== actions.charts.loadingStatus &&
                  status !== actions.charts.finishedStatus
                ) {
                  debouncedFetchChartsWithPagination();
                }
              });
            },
            {
              threshold: 0,
              rootMargin: "-50% 0px 0px 0px",
            },
          ).observe(node);
        }
      }
    },
    [debouncedFetchChartsWithPagination, search.query, status, isFetching],
  );

  const handleLoadMoreButton = useCallback(() => {
    fetchChartsWithPagination(cluster, namespace, currentRepo, "", page, size, nextPage);
  }, [fetchChartsWithPagination, cluster, namespace, currentRepo, page, size, nextPage]);

  const renderError = (err: Error) => {
    return (
      <Alert theme="danger">An error occurred while fetching the catalog: {err.message}</Alert>
    );
  };

  const renderEmptyCatalog = () => {
    return (
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
    );
  };

  const renderCatalogContent = () => {
    return (
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
        <Column span={8}>
          <>
            {/* <h1>{JSON.stringify(isFetching)}</h1>
            <h1>{JSON.stringify(status)}</h1> */}
            <div className="filter-summary">
              {Object.keys(filters).map(filterName => {
                if (filters[filterName].length) {
                  return filters[filterName].map((filterValue: string) => (
                    <span key={`${filterName}-${filterValue}`} className="label label-info">
                      {filterName}: {filterValue}{" "}
                      <CdsIcon shape="times" onClick={removeFilterFunc(filterName, filterValue)} />
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
                  <span>
                    {filteredChartsSearch.length} result
                    {filteredChartsSearch.length !== 1 ? "s" : ""}{" "}
                    {status === actions.charts.finishedStatus ? "" : "- loading more..."}
                  </span>
                </>
              ) : (
                <></>
              )}
            </div>
            <div className="catalogContainer">
              <Row>
                <>
                  <CatalogItems
                    charts={search.query.length > 0 ? filteredChartsSearch : filteredCharts}
                    csvs={filteredCSVs}
                    cluster={cluster}
                    namespace={namespace}
                    hasFinished={status === actions.charts.finishedStatus}
                  />
                  {status === actions.charts.errorStatus && (
                    <div className="endPageMessage">
                      <CdsButton status="primary" size="md" onClick={handleLoadMoreButton}>
                        Load more...{" "}
                      </CdsButton>
                    </div>
                  )}
                  {/* {status === actions.charts.loadingStatus && isFetching && ( */}
                  {status !== actions.charts.finishedStatus && (
                    // TODO(agamez): decide which one we prefer: placeholder or spinner

                    // <div className="catalogItemLoaderContainer">
                    //   <CatalogItemLoader />
                    //   <CatalogItemLoader />
                    //   <CatalogItemLoader />
                    //   <CatalogItemLoader />
                    // </div>
                    <div className="endPageMessage">
                      <LoadingWrapper loaded={false} />
                      <span>Scroll down to discover more applications</span>
                      {/* <Spinner text={"Loading more items..."} /> */}
                    </div>
                  )}
                  {!search.query.length && status === actions.charts.finishedStatus && (
                    <div className="endPageMessage">
                      <span>No remaining applications</span>
                    </div>
                  )}
                  {status !== actions.charts.finishedStatus && (
                    <div className="scrollHandler" ref={observeBorder} />
                  )}
                </>
              </Row>
            </div>
          </>
        </Column>
      </Row>
    );
  };
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
      <div>
        {error && renderError(error)}
        {isEqual(filters, initialFilterState()) &&
        !repos &&
        status !== actions.charts.unstartedStatus &&
        charts.length === 0 &&
        csvs.length === 0
          ? renderEmptyCatalog()
          : renderCatalogContent()}
      </div>
    </section>
  );
}

export default Catalog;
