import * as React from "react";

import { IResource, IServiceSpec, IServiceStatus } from "../../shared/types";
import "./AccessURLItem.css";

interface IAccessURLItem {
  service: IResource;
}

interface IAccessURLState {
  isLink: boolean;
  URLs: string[];
}

class AccessURLItem extends React.Component<IAccessURLItem, IAccessURLState> {
  public state: IAccessURLState = {
    isLink: false,
    URLs: [],
  };

  public componentDidMount() {
    const { service } = this.props;
    const status: IServiceStatus = service.status;
    if (status.loadBalancer.ingress && status.loadBalancer.ingress.length) {
      const URLs: string[] = [];
      status.loadBalancer.ingress.forEach(i => {
        (service.spec as IServiceSpec).ports.forEach(port => {
          if (i.hostname) {
            URLs.push(this.getURL(i.hostname, port.port));
          }
          if (i.ip) {
            URLs.push(this.getURL(i.ip, port.port));
          }
        });
      });
      this.setState({
        isLink: true,
        URLs,
      });
    } else {
      this.setState({
        isLink: false,
        URLs: ["Pending"],
      });
    }
  }

  public render() {
    const { service } = this.props;
    const { isLink, URLs } = this.state;
    return (
      <tr>
        <td>{service.metadata.name}</td>
        <td>{service.spec.type}</td>
        <td>
          {URLs.map(l => (
            <span
              key={l}
              className={`${
                isLink ? "ServiceItem__url" : "ServiceItem__not-url"
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

  private getURL(base: string, port: number) {
    const protocol = port === 443 ? "https" : "http";
    const portText = port === 443 || port === 80 ? "" : `:${port}`;
    return `${protocol}://${base}${portText}`;
  }
}

export default AccessURLItem;
