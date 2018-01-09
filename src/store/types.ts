import { Dispatch } from 'redux';

export interface ChartVersion {
  id: string;
  attributes: ChartAttributes;
  relationships: {
    chart: {
      data: ChartAttributes;
    }
  };
}

export interface ChartVersionAttributes {
  version: string;
  app_version: string;
}

export interface Chart {
  id: string;
  attributes: ChartAttributes;
  relationships: {
    latestChartVersion: {
      data: ChartVersionAttributes;
    }
  };
}

export interface ChartAttributes {
  name: string;
  description: string;
  home: string;
  icon: string;
  keywords: string[];
  maintainers: {}[];
  repo: {};
  sources: string[];
}

export interface ChartState {
  isFetching: boolean;
  selectedChart: Chart | null;
  selectedVersion: ChartVersion | null;
  items: Chart[];
}

export interface StoreState {
  charts: ChartState;
}

export type AsyncAction = (dispatch: Dispatch<StoreState>) => Promise<{}>;
