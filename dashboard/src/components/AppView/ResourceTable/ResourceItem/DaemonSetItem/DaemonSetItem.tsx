import * as React from "react";

import { IDaemonsetStatus, IResource } from "shared/types";

interface IDaemonSetItemRow {
  resource: IResource;
}

export const DaemonSetColumns: React.SFC = () => {
  return (
    <React.Fragment>
      <th className="col-6">NAME</th>
      <th className="col-3">DESIRED</th>
      <th className="col-3">AVAILABLE</th>
    </React.Fragment>
  );
};

const DaemonSetItemRow: React.SFC<IDaemonSetItemRow> = props => {
  const status = props.resource.status as IDaemonsetStatus;
  return (
    <React.Fragment>
      <td className="col-6">{props.resource.metadata.name}</td>
      <td className="col-3">{status.currentNumberScheduled || 0}</td>
      <td className="col-3">{status.numberReady || 0}</td>
    </React.Fragment>
  );
};

export default DaemonSetItemRow;
