// Copyright 2018-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { CdsButton } from "@cds/react/button";
import { CdsIcon } from "@cds/react/icon";
import actions from "actions";
import FilterGroup from "components/FilterGroup/FilterGroup";
import Alert from "components/js/Alert";
import Column from "components/js/Column";
import Row from "components/js/Row";
import { push } from "connected-react-router";
import { flatten, get, intersection, isEqual, trimStart, uniq, without } from "lodash";
import qs from "qs";
import React, { useEffect } from "react";
import { useDispatch, useSelector } from "react-redux";
import * as ReactRouter from "react-router-dom";
import { Link } from "react-router-dom";
import { IClusterServiceVersion, IStoreState } from "shared/types";
import { app } from "shared/url";
import { escapeRegExp, getPluginPackageName } from "shared/utils";
import LoadingWrapper from "../LoadingWrapper/LoadingWrapper";
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

export const filterNames = {
  SEARCH: "Search",
  TYPE: "Type",
  REPO: "Repository",
  CATEGORY: "Category",
  OPERATOR_PROVIDER: "Provider",
  PKG_TYPE: "Plugin",
};

export function initialFilterState() {
  const result = {};
  Object.values(filterNames).forEach(f => (result[f] = []));
  return result;
}

const tmpStrRegex = /__/g;
const tmpStr = "__";
const commaRegex = /,/g;

export function filtersToQuery(filters: any) {
  let query = "";
  const activeFilters = Object.keys(filters).filter(f => filters[f].length);
  if (activeFilters.length) {
    // https://github.com/vmware-tanzu/kubeapps/pull/2279
    // get parameters from the parsed and decoded query params
    // since some search filters could eventually have a ','
    // we need to temporary replace it by other arbitrary string '__'.
    const filterQueries = activeFilters.map(filter => {
      return `${filter}=${filters[filter]
        .map((f: string) => encodeURIComponent(f?.replace(commaRegex, tmpStr)))
        .join(",")}`;
    });
    query = "?" + filterQueries.join("&");
  }
  return query;
}

interface IRouteParams {
  cluster: string;
  namespace: string;
}

export default function Catalog() {
  const {
    packages: {
      hasFinishedFetching,
      nextPageToken,
      selected: { error },
      items: availablePackageSummaries,
      categories,
      size,
      isFetching,
    },
    operators,
    repos: { reposSummaries: repos },
    config: {
      appVersion,
      kubeappsCluster,
      helmGlobalNamespace,
      carvelGlobalNamespace,
      featureFlags,
      configuredPlugins,
    },
  } = useSelector((state: IStoreState) => state);
  const { cluster, namespace } = ReactRouter.useParams() as IRouteParams;
  const location = ReactRouter.useLocation();
  const dispatch = useDispatch();

  const [filters, setFilters] = React.useState(initialFilterState());
  // localNextPageToken is only used to avoid flicker in the CatalogItems.
  // It is no longer used for calculating pagination, which is now handled
  // by the server's opaque nextPageToken etc.
  const [localNextPageToken, setLocalNextPageToken] = React.useState("");
  const [hasRequestedFirstPage, setHasRequestedFirstPage] = React.useState(false);
  const [hasLoadedFirstPage, setHasLoadedFirstPage] = React.useState(false);
  const [isFirstPage, setIsFirstPage] = React.useState(false);
  const localIsFetchingRef = React.useRef(isFetching);

  const csvs = operators.csvs;

  // Only one search filter can be set
  const searchFilter = filters[filterNames.SEARCH]?.toString().replace(tmpStrRegex, ",") || "";
  const reposFilter = filters[filterNames.REPO]?.join(",") || "";

  const timeout = (ms: number) => new Promise(r => setTimeout(r, ms));

  // Detect changes in cluster/ns/repos/search and reset the current package
  // list Note: useEffect is called on every render - the initial render and any
  // re-renders due to updates. Additionally, any cleanup function returned by
  // an effect is run prior to re-running the effect as well as on unmount.
  // https://reactjs.org/docs/hooks-effect.html#example-using-hooks-1
  // Therefore the reset effect below only has a cleanup function so that we a)
  // don't reset when the component is initially rendered, and b) do reset
  // whenever the effect is re-triggered or the component unmounted.
  useEffect(() => {
    // Ensure that when this component is unmounted, we remove any catalog state
    // so that it is reloaded cleanly for the given inputs next time it is
    // loaded.
    return function cleanup() {
      setLocalNextPageToken("");
      setIsFirstPage(false);
      dispatch(actions.availablepackages.resetAvailablePackageSummaries());
      dispatch(actions.availablepackages.resetSelectedAvailablePackageDetail());
    };
  }, [dispatch, cluster, namespace, reposFilter, searchFilter]);

  useEffect(() => {
    const propsFilter = qs.parse(location.search, { ignoreQueryPrefix: true });
    const newFilters = {};
    Object.keys(propsFilter).forEach(filter => {
      const filterValue = propsFilter[filter]?.toString() || "";
      newFilters[filter] = filterValue.split(",").map(a => a.replace(tmpStrRegex, ","));
    });
    setFilters({
      ...initialFilterState(),
      ...newFilters,
    });
  }, [location.search]);

  // using the local reference to the isFetching state to avoid re-rendering on each isFetching actual change
  useEffect(() => {
    localIsFetchingRef.current = isFetching;
  }, [isFetching]);

  useEffect(() => {
    const isFetchingAwareFetching = async () => {
      if (hasFinishedFetching) {
        return;
      }
      // if the local isFetching is true, we need to wait for the rest of the actions to take effect.
      // this state seldom happens (eg. clicking buttons quickly) but it is possible, so we need to wait,
      // otherwise the UI will be in an inconsistent state due to race conditions
      // Note that the "receiveAvailablePackageSummaries" is ignoring the received data if it wasn't previously fetching.
      if (localIsFetchingRef.current) {
        await timeout(1500);
        localIsFetchingRef.current = false;
      }
      dispatch(
        actions.availablepackages.fetchAvailablePackageSummaries(
          cluster,
          namespace,
          reposFilter,
          localNextPageToken,
          size,
          searchFilter,
        ),
      );
    };
    isFetchingAwareFetching();
  }, [
    dispatch,
    size,
    cluster,
    localIsFetchingRef,
    namespace,
    reposFilter,
    searchFilter,
    hasFinishedFetching,
    localNextPageToken,
  ]);

  // hasLoadedFirstPage is used to not bump the current page until the first page is fully
  // requested first
  useEffect(() => {
    if (isFetching) {
      setHasRequestedFirstPage(true);
      setIsFirstPage(true);
    }
    if (hasRequestedFirstPage && !isFetching) {
      setHasLoadedFirstPage(true);
      setIsFirstPage(false);
    }
  }, [hasRequestedFirstPage, isFetching]);

  const pushFilters = (newFilters: any) => {
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
  const clearAllFilters = () => {
    pushFilters({});
  };
  const submitFilters = () => {
    pushFilters(filters);
  };

  const allRepos = uniq(
    repos
      .filter(r => !r.namespaceScoped || r.packageRepoRef?.context?.namespace === namespace)
      .map(r => r.name),
  ).sort();
  const allPlugins = uniq(
    configuredPlugins
      .filter(p => p.name.includes(".packages"))
      .map(p => getPluginPackageName(p, true) || ""),
  ).sort();
  const allProviders = uniq(csvs.map(c => c.spec.provider.name));
  const allCategories = uniq(
    categories
      .map(c => categoryToReadable(c))
      .concat(flatten(csvs.map(c => getOperatorCategories(c)))),
  ).sort();

  // We do not currently support package repositories on additional clusters.
  const supportedCluster = cluster === kubeappsCluster;
  useEffect(() => {
    if (
      !namespace ||
      !supportedCluster ||
      [helmGlobalNamespace, carvelGlobalNamespace].includes(namespace)
    ) {
      // All Namespaces. Global namespace or other cluster, show global repos only
      dispatch(actions.repos.fetchRepoSummaries(""));
      return () => {};
    }
    // In other case, fetch global and namespace repos
    dispatch(actions.repos.fetchRepoSummaries(namespace, true));
    return () => {};
  }, [dispatch, supportedCluster, namespace, helmGlobalNamespace, carvelGlobalNamespace]);

  useEffect(() => {
    // Ignore operators if specified
    if (featureFlags?.operators) {
      dispatch(actions.operators.getCSVs(cluster, namespace));
    }
  }, [dispatch, cluster, namespace, featureFlags]);

  const setSearchFilter = (searchTerm: string) => {
    const newFilters = {
      ...filters,
      [filterNames.SEARCH]: trimStart(searchTerm) ? [trimStart(searchTerm)] : [],
    };
    setFilters(newFilters);
    pushFilters(newFilters);
  };

  const filteredAvailablePackageSummaries = availablePackageSummaries
    .filter(
      () =>
        filters[filterNames.TYPE].length === 0 || filters[filterNames.TYPE].includes("Packages"),
    )
    .filter(
      c =>
        filters[filterNames.PKG_TYPE].length === 0 ||
        filters[filterNames.PKG_TYPE].includes(
          getPluginPackageName(c.availablePackageRef?.plugin, true),
        ),
    )
    .filter(() => filters[filterNames.OPERATOR_PROVIDER].length === 0)
    .filter(
      c =>
        filters[filterNames.REPO].length === 0 ||
        // TODO(agamez): get the repo name once available
        // https://github.com/vmware-tanzu/kubeapps/issues/3165#issuecomment-884574732
        filters[filterNames.REPO].includes(c.availablePackageRef?.identifier.split("/")[0]),
    )
    .filter(
      c =>
        filters[filterNames.CATEGORY].length === 0 ||
        c.categories?.some(category =>
          filters[filterNames.CATEGORY].includes(categoryToReadable(category)),
        ),
    );
  const filteredCSVs = csvs
    .filter(
      () =>
        filters[filterNames.TYPE].length === 0 || filters[filterNames.TYPE].includes("Operators"),
    )
    .filter(() => filters[filterNames.REPO].length === 0)
    .filter(c => {
      const regex = new RegExp(escapeRegExp(searchFilter), "i");
      return (
        regex.test(c.metadata.name) ||
        c?.spec?.customresourcedefinitions?.owned?.find(crd => regex.test(crd.displayName))
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

  // Required to have the latest value of page
  const setPageWithContext = () => {
    if (!error) {
      increaseRequestedPage();
    }
  };

  const forceRetry = () => {
    dispatch(actions.availablepackages.clearErrorPackage());
    dispatch(
      actions.availablepackages.fetchAvailablePackageSummaries(
        cluster,
        namespace,
        reposFilter,
        localNextPageToken,
        size,
        searchFilter,
      ),
    );
  };

  const increaseRequestedPage = () => {
    setLocalNextPageToken(nextPageToken);
  };

  const observeBorder = (node: any) => {
    // Check if the IntersectionAPI is enabled
    if ("IntersectionObserver" in window && "IntersectionObserverEntry" in window) {
      if (node !== null) {
        // https://developer.mozilla.org/en-US/docs/Web/API/IntersectionObserver
        new IntersectionObserver(
          entries => {
            // https://developer.mozilla.org/en-US/docs/Web/API/IntersectionObserverEntry
            entries.forEach(entry => {
              if (
                entry.isIntersecting &&
                // Deactivate scrolling when only operators are selected
                (!filters[filterNames.TYPE].length ||
                  filters[filterNames.TYPE].find((type: string) => type === "Packages")) &&
                // Deactivate scrolling if all the packages have been fetched
                !isFetching &&
                !hasFinishedFetching &&
                hasLoadedFirstPage
              ) {
                setPageWithContext();
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
  };

  /* eslint-disable jsx-a11y/label-has-associated-control */
  return (
    <section>
      <PageHeader
        title="Catalog"
        filter={
          <SearchFilter
            key="searchFilter"
            placeholder="search packages..."
            onChange={setSearchFilter}
            value={searchFilter}
            submitFilters={submitFilters}
          />
        }
      />
      {error && (
        <Alert theme="danger">
          An error occurred while fetching the catalog: {error.message}.{" "}
          {!hasFinishedFetching && (
            <CdsButton size="sm" action="flat" onClick={forceRetry} type="button">
              {" "}
              Try again{" "}
            </CdsButton>
          )}
        </Alert>
      )}
      {isEqual(filters, initialFilterState()) &&
      hasFinishedFetching &&
      searchFilter.length === 0 &&
      availablePackageSummaries.length === 0 &&
      csvs.length === 0 ? (
        <div className="empty-catalog">
          <CdsIcon shape="bundle" />
          <p>The current catalog is empty.</p>
          <p>
            Manage your Package Repositories in Kubeapps by visiting the Package repositories
            configuration page.
          </p>
          <Link to={app.config.pkgrepositories(cluster, namespace)}>
            <CdsButton>Manage Package Repositories</CdsButton>
          </Link>
          <p>
            For help managing other packaging formats, such as Flux or Carvel, please refer to the{" "}
            <a
              target="_blank"
              rel="noopener noreferrer"
              href={`https://github.com/vmware-tanzu/kubeapps/tree/${appVersion}/site/content/docs/latest`}
            >
              Kubeapps documentation
            </a>
            .
          </p>
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
              {csvs.length > 0 && (
                <div className="filter-section">
                  <label>Application Type</label>
                  <FilterGroup
                    name={filterNames.TYPE}
                    options={["Operators", "Packages"]}
                    currentFilters={filters[filterNames.TYPE]}
                    onAddFilter={addFilter}
                    onRemoveFilter={removeFilter}
                    disabled={isFetching}
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
                    disabled={isFetching}
                  />
                </div>
              )}
              {allRepos.length > 0 && (
                <div className="filter-section">
                  <label>Package Repository</label>
                  <FilterGroup
                    name={filterNames.REPO}
                    options={allRepos}
                    currentFilters={filters[filterNames.REPO]}
                    onAddFilter={addFilter}
                    onRemoveFilter={removeFilter}
                    disabled={isFetching}
                  />
                </div>
              )}
              {allPlugins.length > 0 && (
                <div className="filter-section">
                  <label className="filter-label">Package Type</label>
                  <FilterGroup
                    name={filterNames.PKG_TYPE}
                    options={allPlugins}
                    currentFilters={filters[filterNames.PKG_TYPE]}
                    onAddFilter={addFilter}
                    onRemoveFilter={removeFilter}
                    disabled={isFetching}
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
                    disabled={isFetching}
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
                    return filters[filterName].map((filterValue: string) =>
                      filterValue.length ? (
                        <span key={`${filterName}-${filterValue}`} className="label label-info">
                          {filterName}: {filterValue}{" "}
                          <CdsIcon
                            shape="times"
                            onClick={removeFilterFunc(filterName, filterValue)}
                          />
                        </span>
                      ) : (
                        ""
                      ),
                    );
                  }
                  return null;
                })}
              </div>
              <div className="catalog-container">
                <Row>
                  <>
                    <CatalogItems
                      availablePackageSummaries={filteredAvailablePackageSummaries}
                      csvs={filteredCSVs}
                      cluster={cluster}
                      namespace={namespace}
                      isFirstPage={isFirstPage}
                      hasLoadedFirstPage={hasLoadedFirstPage}
                      hasFinishedFetching={hasFinishedFetching && !localIsFetchingRef.current}
                    />
                    {!hasFinishedFetching &&
                      (!filters[filterNames.TYPE].length ||
                        filters[filterNames.TYPE].find((type: string) => type === "Packages")) && (
                        <div className="end-page-message">
                          <LoadingWrapper loaded={false} />
                          {error && !hasFinishedFetching && (
                            <CdsButton size="sm" action="flat" onClick={forceRetry} type="button">
                              {" "}
                              Try again{" "}
                            </CdsButton>
                          )}
                        </div>
                      )}
                    {!hasFinishedFetching && !isFetching && (
                      <div className="scroll-handler" ref={observeBorder} />
                    )}
                  </>
                </Row>
              </div>
            </>
          </Column>
        </Row>
      )}
    </section>
  );
}
