import * as React from "react";

import { IResource, IStatefulsetStatus } from "shared/types";

interface IStatefulSetItemRow {
  resource: IResource;
}

export const StatefulSetColumns: React.SFC = () => {
  return (
    <React.Fragment>
      <th className="col-6">NAME</th>
      <th className="col-2">DESIRED</th>
      <th className="col-2">UP-TO-DATE</th>
      <th className="col-2">READY</th>
    </React.Fragment>
  );
};

const StatefulSetItemRow: React.SFC<IStatefulSetItemRow> = props => {
  const status = props.resource.status as IStatefulsetStatus;
  return (
    <React.Fragment>
      <td className="col-6">{props.resource.metadata.name}</td>
      <td className="col-2">{status.replicas || 0}</td>
      <td className="col-2">{status.updatedReplicas || 0}</td>
      <td className="col-2">{status.readyReplicas || 0}</td>
    </React.Fragment>
  );
};

export default StatefulSetItemRow;
