import * as React from "react";

import { IResource } from "../../shared/types";

interface IAppDetailsProps {
  otherResources: { [r: string]: IResource };
}

class OtherResourcesTable extends React.Component<IAppDetailsProps> {
  public render() {
    const otherResourcesKeys = Object.keys(this.props.otherResources);
    if (otherResourcesKeys.length > 0) {
      return (
        <table>
          <tbody>
            {this.props.otherResources &&
              otherResourcesKeys.map((k: string) => {
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
      );
    } else {
      return <p>The current application does not contain any additional resource.</p>;
    }
  }
}

export default OtherResourcesTable;
