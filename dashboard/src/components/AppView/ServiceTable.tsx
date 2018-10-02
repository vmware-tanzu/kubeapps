import * as React from "react";

import { IResource } from "../../shared/types";
import ServiceItem from "./ServiceItem";

interface IServiceTableProps {
  services: Map<string, IResource>;
}

class ServiceTable extends React.Component<IServiceTableProps> {
  public render() {
    const { services } = this.props;
    const svcList: JSX.Element[] = [];
    services.forEach((svc: IResource, k: string) =>
      svcList.push(<ServiceItem key={k} service={svc} />),
    );
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
        <tbody>{svcList}</tbody>
      </table>
    );
  }
}

export default ServiceTable;
