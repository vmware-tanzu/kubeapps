import * as React from "react";

import { IResource, IServiceSpec, IServiceStatus } from "../../../../../shared/types";

interface IServiceItemRow {
  resource: IResource;
}

export const ServiceColumns: React.SFC = () => {
  return (
    <React.Fragment>
      <th className="col-3">NAME</th>
      <th className="col-3">TYPE</th>
      <th className="col-2">CLUSTER-IP</th>
      <th className="col-2">EXTERNAL-IP</th>
      <th className="col-2">PORT(S)</th>
    </React.Fragment>
  );
};

function getExternalIP(service: IResource): string {
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

const ServiceItem: React.SFC<IServiceItemRow> = props => {
  const { resource } = props;
  const spec: IServiceSpec = resource.spec;
  return (
    <React.Fragment>
      <td className="col-3">{resource.metadata.name}</td>
      <td className="col-3">{spec.type}</td>
      <td className="col-2">{spec.clusterIP}</td>
      <td className="col-2">{getExternalIP(resource)}</td>
      <td className="col-2">
        {spec.ports
          .map(p => `${p.port}${p.nodePort ? `:${p.nodePort}` : ""}/${p.protocol || "TCP"}`)
          .join(", ")}
      </td>
    </React.Fragment>
  );
};

export default ServiceItem;
