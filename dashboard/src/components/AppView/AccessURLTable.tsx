import * as React from "react";

import { IResource, IServiceSpec } from "../../shared/types";
import AccessURLItem from "./AccessURLItem";

interface IServiceTableProps {
  services: Map<string, IResource>;
}

class AccessURLTable extends React.Component<IServiceTableProps> {
  public render() {
    const { services } = this.props;
    const publicServices: string[] = [];
    services.forEach((svc, key) => {
      const spec = svc.spec as IServiceSpec;
      if (spec.type === "LoadBalancer") {
        publicServices.push(key);
      }
    });
    if (publicServices.length > 0) {
      return (
        <table>
          <thead>
            <tr>
              <th>NAME</th>
              <th>TYPE</th>
              <th>URL</th>
            </tr>
          </thead>
          <tbody>
            {publicServices.map((k: string) => (
              <AccessURLItem key={k} service={services.get(k) as IResource} />
            ))}
          </tbody>
        </table>
      );
    } else {
      return null;
    }
  }
}

export default AccessURLTable;
