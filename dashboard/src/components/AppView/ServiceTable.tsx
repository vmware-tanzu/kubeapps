import * as React from "react";

import { IResource } from "../../shared/types";
import ServiceItem from "./ServiceItem";

interface IServiceTableProps {
  services: IResource[];
}

class ServiceTable extends React.Component<IServiceTableProps> {
  public render() {
    const { services } = this.props;
    if (services.length > 0) {
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
            {services.map(s => (
              <ServiceItem key={s.metadata.name} service={s} />
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
