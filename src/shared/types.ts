export interface IChartVersion {
  id: string;
  attributes: IChartVersionAttributes;
  relationships: {
    chart: {
      data: IChartAttributes;
    };
  };
}

export interface IChartVersionAttributes {
  version: string;
  app_version: string;
}

export interface IChart {
  id: string;
  attributes: IChartAttributes;
  relationships: {
    latestChartVersion: {
      data: IChartVersionAttributes;
    };
  };
}

export interface IChartAttributes {
  name: string;
  description: string;
  home: string;
  icon: string;
  keywords: string[];
  maintainers: Array<{}>;
  repo: {
    url: string;
  };
  sources: string[];
}

export interface IChartState {
  isFetching: boolean;
  selectedChart: IChart | null;
  selectedVersion: IChartVersion | null;
  items: IChart[];
}

export interface IStoreState {
  charts: IChartState;
}
