import { ChartAttributes } from "./chart"
export class ChartVersion {
  id: string;
  type: string;
  attributes: ChartVersionAttributes;
  relationships: ChartVersionRelationships;
}

export class ChartVersionAttributes {
  created: Date;
  digest: string;
  icons: ChartVersionIcon[];
  readme: string;
  version: string;
  app_version: string;
  urls: string[];
}

class ChartVersionIcon {
  name: string;
  path: string;
}

class ChartVersionRelationships {
  chart: ChartVersionChart;
}

class ChartVersionChart {
  data: ChartAttributes
  links: {
    self: string
  }
}
