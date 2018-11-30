import * as React from "react";

import { ISecret } from "../../../shared/types";
import SecretItem from "./SecretItem";

interface IServiceTableProps {
  secrets: ISecret[];
}

class SecretTable extends React.Component<IServiceTableProps> {
  public render() {
    const { secrets } = this.props;
    if (secrets.length > 0) {
      return (
        <table>
          <thead>
            <tr className="flex">
              <th className="col-2">NAME</th>
              <th className="col-2">TYPE</th>
              <th className="col-7">DATA</th>
            </tr>
          </thead>
          <tbody>
            {secrets.map(s => (
              <SecretItem key={s.metadata.name} secret={s} />
            ))}
          </tbody>
        </table>
      );
    } else {
      return <p>The current application does not contain any secret.</p>;
    }
  }
}

export default SecretTable;
