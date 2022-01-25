// Copyright 2020-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { get } from "lodash";
import { IPort, IResource, IServiceSpec, IServiceStatus } from "shared/types";

function getServiceExternalIP(service: IResource): string {
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

function getServicePorts(service: IResource): string {
  if (service.spec.ports) {
    return service.spec.ports
      .map((p: IPort) => `${p.port}${p.nodePort ? `:${p.nodePort}` : ""}/${p.protocol || "TCP"}`)
      .join(", ");
  }
  return "";
}

export const ServiceColumns = [
  {
    accessor: "name",
    Header: "NAME",
    getter: (target: IResource) => get(target, "metadata.name"),
  },
  {
    accessor: "type",
    Header: "TYPE",
    getter: (target: IResource) => get(target, "spec.type"),
  },
  {
    accessor: "clusterIP",
    Header: "CLUSTER-IP",
    getter: (target: IResource) => get(target, "spec.clusterIP"),
  },
  {
    accessor: "externalIP",
    Header: "EXTERNAL-IP",
    getter: (target: IResource) => getServiceExternalIP(target),
  },
  {
    accessor: "ports",
    Header: "PORT(S)",
    getter: (target: IResource) => getServicePorts(target),
  },
];
