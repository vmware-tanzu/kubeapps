import * as React from "react";

import LoadingWrapper from "../../../components/LoadingWrapper";
import { IKubeItem, IResource } from "../../../shared/types";
import isSomeResourceLoading from "../helpers";
import DeploymentItem from "./DeploymentItem";

interface IDeploymentTableProps {
  deployments: Array<IKubeItem<IResource>>;
}

class DeploymentTable extends React.Component<IDeploymentTableProps> {
  public render() {
    return (
      <React.Fragment>
        <h6>Deployments</h6>
        <LoadingWrapper loaded={!isSomeResourceLoading(this.props.deployments)} size="small">
          {this.deploymentSection()}
        </LoadingWrapper>
      </React.Fragment>
    );
  }

  private deploymentSection() {
    const { deployments } = this.props;
    let deploymentSection = <p>The current application does not contain any deployment.</p>;
    if (deployments.length > 0) {
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
            {deployments.map(
              d =>
                d.item && (
                  <DeploymentItem key={`deployments/${d.item.metadata.name}`} deployment={d.item} />
                ),
            )}
          </tbody>
        </table>
      );
    }
    return deploymentSection;
  }
}

export default DeploymentTable;
