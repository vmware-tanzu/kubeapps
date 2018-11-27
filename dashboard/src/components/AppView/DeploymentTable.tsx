import * as React from "react";

import { IResource } from "../../shared/types";
import DeploymentItem from "./DeploymentItem";

interface IDeploymentTableProps {
  deployments: { [d: string]: IResource };
}

class DeploymentTable extends React.Component<IDeploymentTableProps> {
  public render() {
    const depKeys = Object.keys(this.props.deployments);
    if (depKeys.length > 0) {
      return (
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
            {depKeys.map((k: string) => (
              <DeploymentItem key={k} deployment={this.props.deployments[k]} />
            ))}
          </tbody>
        </table>
      );
    } else {
      return <p>The current application does not contain any deployment.</p>;
    }
  }
}

export default DeploymentTable;
