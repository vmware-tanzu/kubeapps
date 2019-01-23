import * as React from "react";
import { AlertTriangle } from "react-feather";

import LoadingWrapper, { LoaderType } from "../../../components/LoadingWrapper";
import { IKubeItem, IResource, IServiceSpec, IServiceStatus } from "../../../shared/types";

interface IServiceItemProps {
  name: string;
  service?: IKubeItem<IResource>;
  getService: () => void;
}

class ServiceItem extends React.Component<IServiceItemProps> {
  public componentDidMount() {
    this.props.getService();
  }

  public render() {
    const { name, service } = this.props;
    return (
      <tr>
        <td>{name}</td>
        {this.renderServiceInfo(service)}
      </tr>
    );
  }

  private renderServiceInfo(service?: IKubeItem<IResource>) {
    if (service === undefined || service.isFetching) {
      return (
        <td colSpan={4}>
          <LoadingWrapper type={LoaderType.Placeholder} />
        </td>
      );
    }
    if (service.error) {
      return (
        <td colSpan={4}>
          <span className="flex">
            <AlertTriangle />
            <span className="flex margin-l-normal">Error: {service.error.message}</span>
          </span>
        </td>
      );
    }
    if (service.item) {
      const spec: IServiceSpec = service.item.spec;
      return (
        <React.Fragment>
          <td>{spec.type}</td>
          <td>{spec.clusterIP}</td>
          <td>{this.getExternalIP(service.item)}</td>
          <td>
            {spec.ports
              .map(p => `${p.port}${p.nodePort ? `:${p.nodePort}` : ""}/${p.protocol || "TCP"}`)
              .join(", ")}
          </td>
        </React.Fragment>
      );
    }
    return null;
  }

  private getExternalIP(service: IResource): string {
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
