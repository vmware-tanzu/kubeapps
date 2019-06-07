import * as React from "react";

import ResourceRef from "shared/ResourceRef";
import WorkloadItem from "../../../containers/WorkloadItemContainer";

interface IWorkloadItemTableProps {
  title: string;
  resourceRefs: ResourceRef[];
  status: { [title: string]: string };
}

class WorkloadTable extends React.Component<IWorkloadItemTableProps> {
  public render() {
    const { resourceRefs, status } = this.props;
    let section = null;
    if (resourceRefs.length > 0) {
      section = (
        <React.Fragment>
          <h6>{this.props.title}</h6>
          <table>
            <thead>
              <tr>
                <th>NAME</th>
                {/* Print the desired columns */}
                {Object.keys(status).map(c => (
                  <th key={c}>{c}</th>
                ))}
              </tr>
            </thead>
            <tbody>
              {resourceRefs.map(r => (
                <WorkloadItem
                  key={r.getResourceURL()}
                  resourceRef={r}
                  statusFields={Object.values(status)}
                />
              ))}
            </tbody>
          </table>
        </React.Fragment>
      );
    }
    return section;
  }
}

export default WorkloadTable;
