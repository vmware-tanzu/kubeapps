import * as React from "react";

import { ISecret } from "../../../shared/types";
import SecretItem from "./SecretItem";

interface IServiceTableProps {
  secrets: { [s: string]: ISecret };
}

class SecretTable extends React.Component<IServiceTableProps> {
  public render() {
    const { secrets } = this.props;
    const secretKeys = Object.keys(secrets);
    if (secretKeys.length > 0) {
      return (
        <div>
          <table>
            <thead>
              <tr className="flex">
                <th className="col-2">NAME</th>
                <th className="col-2">TYPE</th>
                <th className="col-7">DATA</th>
              </tr>
            </thead>
            <tbody>
              {secretKeys.map(k => (
                <SecretItem key={k} secret={secrets[k]} />
              ))}
            </tbody>
          </table>
        </div>
      );
    } else {
      return <p>The current application does not contain any secret.</p>;
    }
  }
}

export default SecretTable;
