import * as React from "react";

import { IResource, IServiceSpec } from "../../../shared/types";
import AccessURLItem from "./AccessURLItem";
import { GetURLItemFromIngress } from "./AccessURLItem/AccessURLIngressHelper";
import { GetURLItemFromService } from "./AccessURLItem/AccessURLServiceHelper";

interface IServiceTableProps {
  services: { [s: string]: IResource };
  ingresses: { [i: string]: IResource };
}

class AccessURLTable extends React.Component<IServiceTableProps> {
  public render() {
    const { services, ingresses } = this.props;
    const publicServices = this.publicServices();
    if (publicServices.length > 0 || Object.keys(ingresses).length > 0) {
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
              {Object.keys(ingresses).map((k: string) => (
                <AccessURLItem key={k} URLItem={GetURLItemFromIngress(ingresses[k])} />
              ))}
              {publicServices.map((k: string) => (
                <AccessURLItem key={k} URLItem={GetURLItemFromService(services[k])} />
              ))}
            </tbody>
          </table>
        </div>
      );
    } else {
      return <p>The current application does not expose a public URL.</p>;
    }
  }

  private publicServices(): string[] {
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
}

export default AccessURLTable;
