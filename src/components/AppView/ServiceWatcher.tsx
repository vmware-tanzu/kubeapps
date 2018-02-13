import * as React from "react";

import { IResource } from "../../shared/types";
import ServiceItem from "./ServiceItem";

interface IServiceWatcherProps {
  services: Map<string, IResource>;
}

class ServiceWatcher extends React.Component<IServiceWatcherProps> {
  public render() {
    return (
      <div>
        <h6>Services</h6>
        <table>
          <thead>
            <tr>
              <th>NAME</th>
              <th>TYPE</th>
              <th>CLUSTER-IP</th>
              <th>PORT(S)</th>
            </tr>
          </thead>
          <tbody>
            {this.props.services &&
              Object.keys(this.props.services).map((k: string) => (
                <ServiceItem service={this.props.services[k]} />
              ))}
          </tbody>
        </table>
      </div>
    );
  }
}

export default ServiceWatcher;
