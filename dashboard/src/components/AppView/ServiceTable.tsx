import * as React from "react";

import { IResource } from "../../shared/types";
import ServiceItem from "./ServiceItem";

interface IServiceTableProps {
  services: { [s: string]: IResource };
}

class ServiceTable extends React.Component<IServiceTableProps> {
  public render() {
    const { services } = this.props;
    const svcKeys = Object.keys(services);
    if (svcKeys.length > 0) {
      return (
        <table>
          <thead>
            <tr>
              <th>NAME</th>
              <th>TYPE</th>
              <th>CLUSTER-IP</th>
              <th>EXTERNAL-IP</th>
              <th>PORT(S)</th>
            </tr>
          </thead>
          <tbody>
            {svcKeys.map(k => (
              <ServiceItem key={k} service={services[k]} />
            ))}
          </tbody>
        </table>
      );
    } else {
      return <p>The current application does not contain any service.</p>;
    }
  }
}

export default ServiceTable;
