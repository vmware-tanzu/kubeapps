import * as _ from "lodash";
import * as React from "react";

import { ErrorSelector } from "../../../components/ErrorAlert";
import LoadingWrapper from "../../../components/LoadingWrapper";
import { IKubeItem, ISecret } from "../../../shared/types";
import isSomeResourceLoading from "../helpers";
import SecretItem from "./SecretItem";

interface IServiceTableProps {
  namespace: string;
  secretNames: string[];
  secrets: Array<IKubeItem<ISecret>>;
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
    return (
      <React.Fragment>
        <h6>Secrets</h6>
        <LoadingWrapper loaded={!isSomeResourceLoading(secrets)} size="small">
          {this.secretSection()}
        </LoadingWrapper>
      </React.Fragment>
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

  private secretSection() {
    const { secrets } = this.props;
    const secretError = this.findError();
    let secretSection = <p>The current application does not contain any secret.</p>;
    if (secrets.length > 0) {
      secretSection = (
        <React.Fragment>
          <table>
            <thead>
              <tr className="flex">
                <th className="col-2">NAME</th>
                <th className="col-2">TYPE</th>
                <th className="col-7">DATA</th>
              </tr>
            </thead>
            <tbody>
              {secrets.map(
                s =>
                  s.item && <SecretItem key={`secrets/${s.item.metadata.name}`} secret={s.item} />,
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
        </React.Fragment>
      );
    }
    return secretSection;
  }
}

export default SecretsTable;
