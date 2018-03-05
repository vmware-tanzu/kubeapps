import * as React from "react";

import { IResource } from "../../shared/types";
import DeploymentTable from "./DeploymentTable";
import ServiceTable from "./ServiceTable";

interface IAppDetailsProps {
  deployments: Map<string, IResource>;
  services: Map<string, IResource>;
  otherResources: Map<string, IResource>;
}

class AppDetails extends React.Component<IAppDetailsProps> {
  public render() {
    return (
      <div className="AppDetails">
        <h2>Details</h2>
        <hr />
        <div className="AppDetails__content margin-h-big">
          {Object.keys(this.props.deployments).length > 0 && (
            <div>
              <h6>Deployments</h6>
              <DeploymentTable deployments={this.props.deployments} />
            </div>
          )}
          {Object.keys(this.props.services).length > 0 && (
            <div>
              <h6>Services</h6>
              <ServiceTable services={this.props.services} extended={true} />
            </div>
          )}
          <h6>Other Resources</h6>
          <table>
            <tbody>
              {this.props.otherResources &&
                Object.keys(this.props.otherResources).map((k: string) => {
                  const r = this.props.otherResources[k];
                  return (
                    <tr key={k}>
                      <td>{r.kind}</td>
                      <td>{r.metadata.namespace}</td>
                      <td>{r.metadata.name}</td>
                    </tr>
                  );
                })}
            </tbody>
          </table>
        </div>
      </div>
    );
  }
}

export default AppDetails;
