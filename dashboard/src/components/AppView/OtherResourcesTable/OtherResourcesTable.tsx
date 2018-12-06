import * as React from "react";

import { IResource } from "../../../shared/types";

interface IAppDetailsProps {
  otherResources: IResource[];
}

class OtherResourcesTable extends React.Component<IAppDetailsProps> {
  public render() {
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
                <tr key={r.metadata.name}>
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
    return (
      <React.Fragment>
        <h6>Other Resources</h6>
        {otherResourcesSection}
      </React.Fragment>
    );
  }
}

export default OtherResourcesTable;
