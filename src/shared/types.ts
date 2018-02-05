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
  created: string;
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
  home?: string;
  icon?: string;
  keywords: string[];
  maintainers: Array<{
    name: string;
    email?: string;
  }>;
  repo: {
    name: string;
    url: string;
  };
  sources: string[];
}

export interface IChartState {
  isFetching: boolean;
  selected: {
    version?: IChartVersion;
    versions: IChartVersion[];
    readme?: string;
    values?: string;
  };
  items: IChart[];
}

export interface IStoreState {
  charts: IChartState;
}
