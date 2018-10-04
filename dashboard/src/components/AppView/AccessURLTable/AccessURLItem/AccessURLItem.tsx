import * as React from "react";

import { IResource, IServiceSpec, IServiceStatus } from "../../../../shared/types";
import "./AccessURLItem.css";

interface IAccessURLItem {
  loadBalancerService: IResource;
}

class AccessURLItem extends React.Component<IAccessURLItem> {
  public render() {
    const { loadBalancerService } = this.props;
    const isLink = this.isLink();
    return (
      <tr>
        <td>{loadBalancerService.metadata.name}</td>
        <td>{loadBalancerService.spec.type}</td>
        <td>
          {this.URLs().map(l => (
            <span
              key={l}
              className={`ServiceItem ${
                isLink ? "ServiceItem--with-link" : ""
              } type-small margin-r-small padding-tiny padding-h-normal`}
            >
              {isLink ? (
                <a className="padding-tiny padding-h-normal" href={l} target="_blank">
                  {l}
                </a>
              ) : (
                l
              )}
            </span>
          ))}
        </td>
      </tr>
    );
  }

  // isLink returns true if there are any link in the Item
  private isLink(): boolean {
    if (
      this.props.loadBalancerService.status.loadBalancer.ingress &&
      this.props.loadBalancerService.status.loadBalancer.ingress.length
    ) {
      return true;
    }
    return false;
  }

  // URLs returns the list of URLs obtained from the service status
  private URLs(): string[] {
    const URLs: string[] = [];
    const { loadBalancerService } = this.props;
    const status: IServiceStatus = loadBalancerService.status;
    if (status.loadBalancer.ingress && status.loadBalancer.ingress.length) {
      status.loadBalancer.ingress.forEach(i => {
        (loadBalancerService.spec as IServiceSpec).ports.forEach(port => {
          if (i.hostname) {
            URLs.push(this.getURL(i.hostname, port.port));
          }
          if (i.ip) {
            URLs.push(this.getURL(i.ip, port.port));
          }
        });
      });
    } else {
      URLs.push("Pending");
    }
    return URLs;
  }

  // getURL returns a full URL adding the protocol and the port if needed
  private getURL(base: string, port: number) {
    const protocol = port === 443 ? "https" : "http";
    // Only show the port in the URL if it's not a standard HTTP/HTTPS port
    const portSuffix = port === 443 || port === 80 ? "" : `:${port}`;
    return `${protocol}://${base}${portSuffix}`;
  }
}

export default AccessURLItem;
