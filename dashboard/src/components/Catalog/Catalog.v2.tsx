import { RouterAction } from "connected-react-router";
import { uniq } from "lodash";
import React, { useEffect, useState } from "react";
import { Link } from "react-router-dom";

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

import "./Catalog.v2.css";
import CatalogItems from "./CatalogItems";

interface ICatalogProps {
  charts: IChartState;
  repo: string;
  filter: string;
  fetchCharts: (namespace: string, repo: string) => void;
  pushSearchFilter: (filter: string) => RouterAction;
  namespace: string;
  kubeappsNamespace: string;
  getCSVs: (namespace: string) => void;
  csvs: IClusterServiceVersion[];
  featureFlags: IFeatureFlags;
}

function Catalog(props: ICatalogProps) {
  const [searchFilter, setSearchFilter] = useState("");
  const [typeFilter, setTypeFilter] = useState([] as string[]);
  const [repoFilter, setRepoFilter] = useState([] as string[]);

  const {
    charts: {
      isFetching,
      selected: { error },
      items: charts,
    },
    fetchCharts,
    namespace,
    pushSearchFilter,
    getCSVs,
    csvs,
    repo,
    filter: propsFilter,
  } = props;
  const allRepos = uniq(charts.map(c => c.attributes.repo.name));

  useEffect(() => {
    fetchCharts(namespace, repo);
    getCSVs(namespace);
  }, [namespace, repo, fetchCharts, getCSVs]);

  useEffect(() => {
    setSearchFilter(propsFilter);
  }, [propsFilter]);

  const filteredCharts = charts
    .filter(() => typeFilter.length === 0 || typeFilter.includes("Charts"))
    .filter(c => new RegExp(escapeRegExp(searchFilter), "i").test(c.id))
    .filter(c => repoFilter.length === 0 || repoFilter.includes(c.attributes.repo.name));
  const filteredCSVs = csvs
    .filter(() => typeFilter.length === 0 || typeFilter.includes("Operators"))
    .filter(() => repoFilter.length === 0)
    .filter(c => new RegExp(escapeRegExp(searchFilter), "i").test(c.metadata.name));

  return (
    <section>
      <PageHeader>
        <div className="kubeapps-header">
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
        </div>
      </PageHeader>
      <LoadingWrapper loaded={!isFetching}>
        {error && <Alert theme="danger">Unable to fetch catalog: {error.message}</Alert>}
        {charts.length === 0 && csvs.length === 0 && (
          <Alert theme="warning">
            Charts not found. Manage your Helm chart repositories in Kubeapps by visiting the{" "}
            <Link to={`/config/ns/${namespace}/repos`}>App repositories configuration</Link> page.
          </Alert>
        )}
        <Row>
          <Column span={2}>
            <div className="filters-menu">
              <h5>Filters</h5>
              {csvs.length > 0 && (
                <>
                  <div className="filter-label">Application Type:</div>
                  <FilterGroup
                    name="apptype"
                    options={["Operators", "Charts"]}
                    onChange={setTypeFilter}
                  />
                </>
              )}
              <div className="filter-label">Application Repository:</div>
              <FilterGroup name="apprepo" options={allRepos} onChange={setRepoFilter} />
            </div>
          </Column>
          <Column span={10}>
            <CardGrid>
              <CatalogItems charts={filteredCharts} csvs={filteredCSVs} namespace={namespace} />
            </CardGrid>
          </Column>
        </Row>
      </LoadingWrapper>
    </section>
  );
}

export default Catalog;
