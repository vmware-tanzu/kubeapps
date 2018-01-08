export interface Chart {
  id: string;
  attributes: ChartAttributes;
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
  items: Array<Chart>;
}

export interface StoreState {
  charts: ChartState;
}
