import * as React from "react";

import { IResource } from "../../../shared/types";

interface IAppDetailsProps {
  otherResources: IResource[];
}

class OtherResourcesTable extends React.Component<IAppDetailsProps> {
  public render() {
    return (
      <React.Fragment>
        <h6>Other Resources</h6>
        {this.otherResourcesSection()}
      </React.Fragment>
    );
  }

  private otherResourcesSection() {
    const { otherResources } = this.props;
    let otherResourcesSection = (
      <p>The current application does not contain any additional resource.</p>
    );
    if (otherResources.length > 0) {
      otherResourcesSection = (
        <table>
          <tbody>
            {otherResources.map((r: IResource) => {
              return (
                <tr key={`otherResources/${r.kind}/${r.metadata.name}`}>
                  <td>{r.kind}</td>
                  <td>{r.metadata.namespace}</td>
                  <td>{r.metadata.name}</td>
                </tr>
              );
            })}
          </tbody>
        </table>
      );
    }
    return otherResourcesSection;
  }
}

export default OtherResourcesTable;
