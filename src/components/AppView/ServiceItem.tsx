import * as React from "react";

import { IResource, IServiceSpec } from "../../shared/types";

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
        <td>
          {spec.ports
            .map(p => `${p.port}${p.nodePort ? `:${p.nodePort}` : ""}/${p.protocol}`)
            .join(",")}
        </td>
      </tr>
    );
  }
}

export default ServiceItem;
