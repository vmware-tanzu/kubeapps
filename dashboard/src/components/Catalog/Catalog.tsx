import { RouterAction } from "connected-react-router";
import * as React from "react";
import { Link } from "react-router-dom";

import { IChart, IChartState } from "../../shared/types";
import { escapeRegExp } from "../../shared/utils";
import { CardGrid } from "../Card";
import { MessageAlert } from "../ErrorAlert";
import LoadingWrapper from "../LoadingWrapper";
import PageHeader from "../PageHeader";
import SearchFilter from "../SearchFilter";
import CatalogItem from "./CatalogItem";

interface ICatalogProps {
  charts: IChartState;
  repo: string;
  filter: string;
  fetchCharts: (repo: string) => void;
  pushSearchFilter: (filter: string) => RouterAction;
}

interface ICatalogState {
  filter: string;
}

class Catalog extends React.Component<ICatalogProps, ICatalogState> {
  public state: ICatalogState = {
    filter: "",
  };

  public componentDidMount() {
    const { repo, fetchCharts, filter } = this.props;
    this.setState({ filter });
    fetchCharts(repo);
  }

  public componentWillReceiveProps(nextProps: ICatalogProps) {
    if (nextProps.filter !== this.state.filter) {
      this.setState({ filter: nextProps.filter });
    }
    if (this.props.repo !== nextProps.repo) {
      this.props.fetchCharts(nextProps.repo);
    }
  }

  public render() {
    const {
      charts: { isFetching, items: allItems },
      pushSearchFilter,
    } = this.props;
    const items = this.filteredCharts(allItems, this.state.filter);
    if (!isFetching && allItems.length === 0) {
      return (
        <MessageAlert
          level={"warning"}
          children={
            <div>
              <h5>Charts not found.</h5>
              Manage your Helm chart repositories in Kubeapps by visiting the{" "}
              <Link to={"/config/repos"}>App repositories configuration</Link> page.
            </div>
          }
        />
      );
    }
    const chartItems = items.map(c => <CatalogItem key={c.id} chart={c} />);
    return (
      <section className="Catalog">
        <PageHeader>
          <h1>Catalog</h1>
          <SearchFilter
            className="margin-l-big"
            placeholder="search charts..."
            onChange={this.handleFilterQueryChange}
            value={this.state.filter}
            onSubmit={pushSearchFilter}
          />
        </PageHeader>
        <LoadingWrapper loaded={!isFetching}>
          <CardGrid>{chartItems}</CardGrid>
        </LoadingWrapper>
      </section>
    );
  }

  private filteredCharts(charts: IChart[], filter: string) {
    return charts.filter(c => new RegExp(escapeRegExp(filter), "i").test(c.id));
  }

  private handleFilterQueryChange = (filter: string) => {
    this.setState({
      filter,
    });
  };
}

export default Catalog;
