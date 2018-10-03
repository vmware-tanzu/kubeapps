import * as React from "react";

import { IResource, IServiceSpec } from "../../../shared/types";
import AccessURLItem from "./AccessURLItem";

interface IServiceTableProps {
  services: { [s: string]: IResource };
}

class AccessURLTable extends React.Component<IServiceTableProps> {
  get publicServices() {
    const { services } = this.props;
    const publicServices: string[] = [];
    Object.keys(services).forEach(key => {
      const spec = services[key].spec as IServiceSpec;
      if (spec.type === "LoadBalancer") {
        publicServices.push(key);
      }
    });
    return publicServices;
  }

  public render() {
    const { services } = this.props;
    return (
      this.publicServices.length > 0 && (
        <table>
          <thead>
            <tr>
              <th>NAME</th>
              <th>TYPE</th>
              <th>URL</th>
            </tr>
          </thead>
          <tbody>
            {this.publicServices.map((k: string) => (
              <AccessURLItem key={k} service={services[k]} />
            ))}
          </tbody>
        </table>
      )
    );
  }
}

export default AccessURLTable;
