import * as React from "react";

import { IResource, IServiceSpec } from "../../../shared/types";
import AccessURLItem from "./AccessURLItem";
import { GetURLItemFromIngress } from "./AccessURLItem/AccessURLIngressHelper";
import { GetURLItemFromService } from "./AccessURLItem/AccessURLServiceHelper";

interface IServiceTableProps {
  services: IResource[];
  ingresses: IResource[];
}

class AccessURLTable extends React.Component<IServiceTableProps> {
  public render() {
    const { ingresses } = this.props;
    const publicServices = this.publicServices();
    if (publicServices.length > 0 || ingresses.length > 0) {
      return (
        <div>
          <table>
            <thead>
              <tr>
                <th>RESOURCE</th>
                <th>TYPE</th>
                <th>URL</th>
              </tr>
            </thead>
            <tbody>
              {ingresses.map(i => (
                <AccessURLItem key={i.metadata.name} URLItem={GetURLItemFromIngress(i)} />
              ))}
              {publicServices.map(s => (
                <AccessURLItem key={s.metadata.name} URLItem={GetURLItemFromService(s)} />
              ))}
            </tbody>
          </table>
        </div>
      );
    } else {
      return <p>The current application does not expose a public URL.</p>;
    }
  }

  private publicServices(): IResource[] {
    const { services } = this.props;
    const publicServices: IResource[] = [];
    services.forEach(s => {
      const spec = s.spec as IServiceSpec;
      if (spec.type === "LoadBalancer") {
        publicServices.push(s);
      }
    });
    return publicServices;
  }
}

export default AccessURLTable;
