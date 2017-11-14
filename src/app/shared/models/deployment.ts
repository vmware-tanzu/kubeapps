export class Deployment {
  id: string;
  type: string;
  attributes: DeploymentAttributes
}

class DeploymentAttributes {
  chartName: string;
  chartVersion: string;
  chartIcon: string;
  name: string;
  namespace: string;
  status: string;
  updated: Date;
  notes: string;
  resources: string;
  urls: string[];
}
