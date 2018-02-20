import * as React from "react";

import { IResource, IServiceSpec, IServiceStatus } from "../../shared/types";
import "./ServiceItem.css";

interface IServiceItemProps {
  service: IResource;
  extended: boolean;
}

class ServiceItem extends React.Component<IServiceItemProps> {
  public render() {
    const { extended } = this.props;

    return extended ? this.renderExtended() : this.renderSimple();
  }

  public renderSimple() {
    const { service } = this.props;
    return (
      <tr>
        <td>{service.metadata.name}</td>
        <td>
          {this.getPublicURLs().map(l => (
            <a key={l} href={l} target="_blank">
              <span className="ServiceItem__url padding-tiny padding-h-normal type-small margin-r-small">
                {l}
              </span>
            </a>
          ))}
        </td>
      </tr>
    );
  }

  public renderExtended() {
    const { service } = this.props;
    const spec: IServiceSpec = service.spec;
    return (
      <tr>
        <td>{service.metadata.name}</td>
        <td>{spec.type}</td>
        <td>{spec.clusterIP}</td>
        <td>{this.getExternalIP()}</td>
        <td>
          {spec.ports
            .map(p => `${p.port}${p.nodePort ? `:${p.nodePort}` : ""}/${p.protocol}`)
            .join(",")}
        </td>
      </tr>
    );
  }

  private getExternalIP(): string {
    const { service } = this.props;
    const spec: IServiceSpec = service.spec;
    const status: IServiceStatus = service.status;
    if (spec.type !== "LoadBalancer") {
      return "None";
    }
    if (status.loadBalancer.ingress) {
      return status.loadBalancer.ingress[0].ip;
    }
    return "Pending";
  }

  private getPublicURLs(): string[] {
    const { service } = this.props;
    const spec: IServiceSpec = service.spec;
    const externalIP = this.getExternalIP();
    if (externalIP === "None" || externalIP === "Pending") {
      return [externalIP];
    }
    return spec.ports.map(p => {
      switch (p.port) {
        case 80:
          return `http://${externalIP}`;
        case 443:
          return `https://${externalIP}`;
        default:
          return `http://${externalIP}:${p.port}`;
      }
    });
  }
}

export default ServiceItem;
