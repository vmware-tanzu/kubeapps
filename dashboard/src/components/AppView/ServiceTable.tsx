import * as React from "react";

import { IResource, IServiceSpec } from "../../shared/types";
import ServiceItem from "./ServiceItem";

interface IServiceTableProps {
  services: Map<string, IResource>;
  extended: boolean;
}

class ServiceTable extends React.Component<IServiceTableProps> {
  private simpleHeaders = (
    <tr>
      <th>SERVICE</th>
      <th>URL</th>
    </tr>
  );

  private extendedHeaders = (
    <tr>
      <th>NAME</th>
      <th>TYPE</th>
      <th>CLUSTER-IP</th>
      <th>EXTERNAL-IP</th>
      <th>PORT(S)</th>
    </tr>
  );

  public render() {
    const { extended, services } = this.props;
    return (
      <table>
        <thead>{extended ? this.extendedHeaders : this.simpleHeaders}</thead>
        <tbody>
          {services &&
            Object.keys(services)
              .filter(k => extended || (services[k].spec as IServiceSpec).type === "LoadBalancer")
              .map((k: string) => (
                <ServiceItem key={k} service={services[k]} extended={extended} />
              ))}
        </tbody>
      </table>
    );
  }
}

export default ServiceTable;
