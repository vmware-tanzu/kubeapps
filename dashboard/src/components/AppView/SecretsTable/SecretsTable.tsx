import * as _ from "lodash";
import * as React from "react";

import { ErrorSelector } from "../../../components/ErrorAlert";
import { IKubeItem, ISecret } from "../../../shared/types";
import SecretItem from "./SecretItem";

interface IServiceTableProps {
  namespace: string;
  secretNames: string[];
  secrets: { [s: string]: IKubeItem<ISecret> };
  getSecret: (namespace: string, name: string) => void;
}

interface IError {
  resource: string;
  error: Error;
}

class SecretsTable extends React.Component<IServiceTableProps> {
  public async componentDidMount() {
    const { getSecret, secretNames, namespace } = this.props;
    secretNames.forEach(s => {
      getSecret(namespace, s);
    });
  }

  public render() {
    const { secrets } = this.props;
    const secretKeys = Object.keys(secrets);
    const secretError = this.findError();
    return (
      secretKeys.length > 0 && (
        <div>
          <h6>Secrets</h6>
          <table>
            <thead>
              <tr className="flex">
                <th className="col-2">NAME</th>
                <th className="col-2">TYPE</th>
                <th className="col-7">DATA</th>
              </tr>
            </thead>
            <tbody>
              {secretKeys.map(
                k => secrets[k].item && <SecretItem key={k} secret={secrets[k].item as ISecret} />,
              )}
            </tbody>
          </table>
          {secretError && (
            <ErrorSelector
              error={secretError.error}
              action="get"
              resource={`Secret ${secretError.resource}`}
              namespace={this.props.namespace}
            />
          )}
        </div>
      )
    );
  }

  private findError = (): IError | null => {
    let error = null;
    _.each(this.props.secrets, (i, k) => {
      if (i.error) {
        error = { resource: k, error: i.error };
      }
    });
    return error;
  };
}

export default SecretsTable;
