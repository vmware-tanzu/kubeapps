// Copyright 2018-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { IResource, IServiceSpec, IServiceStatus } from "shared/types";
import { IURLItem } from "./IURLItem";

// isLink returns true if there are any link in the Item
function isLink(loadBalancerService?: IResource): boolean {
  if (
    loadBalancerService &&
    loadBalancerService.status &&
    loadBalancerService.status.loadBalancer.ingress &&
    loadBalancerService.status.loadBalancer.ingress.length
  ) {
    return true;
  }
  return false;
}

// URLs returns the list of URLs obtained from the service status
function URLs(loadBalancerService?: IResource): string[] {
  if (!loadBalancerService) {
    return ["Pending"];
  }
  const res: string[] = [];
  const status: IServiceStatus = loadBalancerService.status;
  if (status && status.loadBalancer.ingress && status.loadBalancer.ingress.length) {
    status.loadBalancer.ingress.forEach(i => {
      (loadBalancerService.spec as IServiceSpec).ports.forEach(port => {
        if (i.hostname) {
          res.push(getURL(i.hostname, port.port));
        }
        if (i.ip) {
          res.push(getURL(i.ip, port.port));
        }
      });
    });
  } else {
    res.push("Pending");
  }
  return res;
}

// getURL returns a full URL adding the protocol and the port if needed
function getURL(base: string, port: number) {
  const protocol = port === 443 ? "https" : "http";
  // Only show the port in the URL if it's not a standard HTTP/HTTPS port
  const portSuffix = port === 443 || port === 80 ? "" : `:${port}`;
  return `${protocol}://${base}${portSuffix}`;
}

export function GetURLItemFromService(loadBalancerService?: IResource) {
  return {
    name: loadBalancerService?.metadata.name,
    type: "Service LoadBalancer",
    isLink: isLink(loadBalancerService),
    URLs: URLs(loadBalancerService),
  } as IURLItem;
}
