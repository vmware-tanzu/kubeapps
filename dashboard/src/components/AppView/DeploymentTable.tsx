import * as React from "react";

import { IResource } from "../../shared/types";
import DeploymentItem from "./DeploymentItem";

interface IDeploymentTableProps {
  deployments: IResource[];
}

class DeploymentTable extends React.Component<IDeploymentTableProps> {
  public render() {
    const { deployments } = this.props;
    if (deployments.length > 0) {
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
            {deployments.map(d => (
              <DeploymentItem key={d.metadata.name} deployment={d} />
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
