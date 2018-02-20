import * as React from "react";

import { IResource } from "../../shared/types";
import DeploymentItem from "./DeploymentItem";

interface IDeploymentTableProps {
  deployments: Map<string, IResource>;
}

class DeploymentTable extends React.Component<IDeploymentTableProps> {
  public render() {
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
          {this.props.deployments &&
            Object.keys(this.props.deployments).map((k: string) => (
              <DeploymentItem key={k} deployment={this.props.deployments[k]} />
            ))}
        </tbody>
      </table>
    );
  }
}

export default DeploymentTable;
