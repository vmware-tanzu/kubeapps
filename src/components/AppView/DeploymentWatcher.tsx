import * as React from "react";

import { IResource } from "../../shared/types";
import DeploymentItem from "./DeploymentItem";

interface IDeploymentWatcherProps {
  deployments: Map<string, IResource>;
}

class DeploymentWatcher extends React.Component<IDeploymentWatcherProps> {
  public render() {
    return (
      <div>
        <h6>Deployments</h6>
        <table>
          <thead>
            <tr>
              <th>NAME</th>
              <th>DESIRED</th>
              <th>UP-TO-DATE</th>
              <th>AVAILABLE</th>
            </tr>
          </thead>
          <tbody>
            {this.props.deployments &&
              Object.keys(this.props.deployments).map((k: string) => (
                <DeploymentItem key={k} deployment={this.props.deployments[k]} />
              ))}
          </tbody>
        </table>
      </div>
    );
  }
}

export default DeploymentWatcher;
