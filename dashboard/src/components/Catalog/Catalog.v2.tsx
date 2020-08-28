import { RouterAction } from "connected-react-router";
import { flatten, get, intersection, uniq } from "lodash";
import React, { useEffect, useState } from "react";
import { Link } from "react-router-dom";
import { CdsIcon } from "../Clarity/clarity";

import FilterGroup from "components/FilterGroup/FilterGroup";
import Alert from "components/js/Alert";
import Column from "components/js/Column";
import Row from "components/js/Row";
import { IFeatureFlags } from "shared/Config";
import { IChartState, IClusterServiceVersion } from "../../shared/types";
import { escapeRegExp } from "../../shared/utils";
import CardGrid from "../Card/CardGrid.v2";
import LoadingWrapper from "../LoadingWrapper/LoadingWrapper.v2";
import PageHeader from "../PageHeader/PageHeader.v2";
import SearchFilter from "../SearchFilter/SearchFilter.v2";

import { CdsButton } from "components/Clarity/clarity";
import { app } from "shared/url";
import "./Catalog.v2.css";
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
  filter: string;
  fetchCharts: (namespace: string, repo: string) => void;
  pushSearchFilter: (filter: string) => RouterAction;
  cluster: string;
  namespace: string;
  kubeappsNamespace: string;
  getCSVs: (cluster: string, namespace: string) => void;
  csvs: IClusterServiceVersion[];
  featureFlags: IFeatureFlags;
}

function Catalog(props: ICatalogProps) {
  const {
    charts: {
      isFetching,
      selected: { error },
      items: charts,
    },
    fetchCharts,
    cluster,
    namespace,
    pushSearchFilter,
    getCSVs,
    csvs,
    repo,
    filter: propsFilter,
  } = props;
  const [searchFilter, setSearchFilter] = useState(propsFilter);
  const [typeFilter, setTypeFilter] = useState([] as string[]);
  const [repoFilter, setRepoFilter] = useState([] as string[]);
  const [categoryFilter, setCategoryFilter] = useState([] as string[]);
  const [operatorProviderFilter, setOperatorProviderFilter] = useState([] as string[]);

  const allRepos = uniq(charts.map(c => c.attributes.repo.name));
  const allProviders = uniq(csvs.map(c => c.spec.provider.name));
  const allCategories = uniq(
    charts
      .map(c => categoryToReadable(c.attributes.category))
      .concat(flatten(csvs.map(c => getOperatorCategories(c)))),
  ).sort();

  useEffect(() => {
    fetchCharts(namespace, repo);
    getCSVs(cluster, namespace);
  }, [cluster, namespace, repo, fetchCharts, getCSVs]);

  useEffect(() => {
    setSearchFilter(propsFilter);
  }, [propsFilter]);

  const filteredCharts = charts
    .filter(() => typeFilter.length === 0 || typeFilter.includes("Charts"))
    .filter(() => operatorProviderFilter.length === 0)
    .filter(c => new RegExp(escapeRegExp(searchFilter), "i").test(c.id))
    .filter(c => repoFilter.length === 0 || repoFilter.includes(c.attributes.repo.name))
    .filter(
      c =>
        categoryFilter.length === 0 ||
        categoryFilter.includes(categoryToReadable(c.attributes.category)),
    );
  const filteredCSVs = csvs
    .filter(() => typeFilter.length === 0 || typeFilter.includes("Operators"))
    .filter(() => repoFilter.length === 0)
    .filter(c => new RegExp(escapeRegExp(searchFilter), "i").test(c.metadata.name))
    .filter(
      c =>
        operatorProviderFilter.length === 0 ||
        operatorProviderFilter.includes(c.spec.provider.name),
    )
    .filter(
      c =>
        categoryFilter.length === 0 ||
        intersection(categoryFilter, getOperatorCategories(c)).length,
    );

  return (
    <section>
      <PageHeader>
        <Row>
          <h1>Catalog</h1>
          <SearchFilter
            key="searchFilter"
            placeholder="search charts..."
            onChange={setSearchFilter}
            value={searchFilter}
            onSubmit={pushSearchFilter}
          />
        </Row>
      </PageHeader>
      <LoadingWrapper loaded={!isFetching}>
        {error && (
          <Alert theme="danger">Found en error fetching the catalog: {error.message}</Alert>
        )}
        {charts.length === 0 && csvs.length === 0 ? (
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
                <h5>Filters</h5>
                {csvs.length > 0 && (
                  <div className="filter-section">
                    <label>Application Type:</label>
                    <FilterGroup
                      name="apptype"
                      options={["Operators", "Charts"]}
                      onChange={setTypeFilter}
                    />
                  </div>
                )}
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
                {allRepos.length > 0 && (
                  <div className="filter-section">
                    <label>Application Repository:</label>
                    <FilterGroup name="apprepo" options={allRepos} onChange={setRepoFilter} />
                  </div>
                )}
                {allProviders.length > 0 && (
                  <div className="filter-section">
                    <label className="filter-label">Operator Provider:</label>
                    <FilterGroup
                      name="operator-provider"
                      options={allProviders}
                      onChange={setOperatorProviderFilter}
                    />
                  </div>
                )}
              </div>
            </Column>
            <Column span={10}>
              <CardGrid>
                <CatalogItems
                  charts={filteredCharts}
                  csvs={filteredCSVs}
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

export default Catalog;
