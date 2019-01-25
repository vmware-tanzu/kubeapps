import * as React from "react";

import ResourceRef from "shared/ResourceRef";
import DeploymentItem from "../../../containers/DeploymentItemContainer";

interface IDeploymentTableProps {
  deployRefs: ResourceRef[];
}

class DeploymentTable extends React.Component<IDeploymentTableProps> {
  public render() {
    return (
      <React.Fragment>
        <h6>Deployments</h6>
        {this.deploymentSection()}
      </React.Fragment>
    );
  }

  private deploymentSection() {
    const { deployRefs } = this.props;
    let deploymentSection = <p>The current application does not contain any Deployment objects.</p>;
    if (deployRefs.length > 0) {
      deploymentSection = (
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
            {deployRefs.map(d => (
              <DeploymentItem key={d.getResourceURL()} deployRef={d} />
            ))}
          </tbody>
        </table>
      );
    }
    return deploymentSection;
  }
}

export default DeploymentTable;
