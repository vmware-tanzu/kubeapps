import * as React from "react";

import { IResource, IServiceSpec, IServiceStatus } from "../../../shared/types";

interface IServiceItemProps {
  service: IResource;
}

class ServiceItem extends React.Component<IServiceItemProps> {
  public render() {
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
            .map(p => `${p.port}${p.nodePort ? `:${p.nodePort}` : ""}/${p.protocol || "TCP"}`)
            .join(", ")}
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
    if (status.loadBalancer.ingress && status.loadBalancer.ingress.length > 0) {
      return (
        status.loadBalancer.ingress[0].hostname || status.loadBalancer.ingress[0].ip || "Pending"
      );
    }
    return "Pending";
  }
}

export default ServiceItem;
