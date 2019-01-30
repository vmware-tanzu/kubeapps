import * as _ from "lodash";
import * as React from "react";

import ResourceRef from "shared/ResourceRef";
import SecretItem from "../../../containers/SecretItemContainer";

interface IServiceTableProps {
  secretRefs: ResourceRef[];
}

class SecretsTable extends React.Component<IServiceTableProps> {
  public render() {
    return (
      <React.Fragment>
        <h6>Secrets</h6>
        {this.secretSection()}
      </React.Fragment>
    );
  }

  private secretSection() {
    const { secretRefs } = this.props;
    let secretSection = <p>The current application does not contain any Secret objects.</p>;
    if (secretRefs.length > 0) {
      secretSection = (
        <React.Fragment>
          <table>
            <thead>
              <tr className="flex">
                <th className="col-3">NAME</th>
                <th className="col-2">TYPE</th>
                <th className="col-7">DATA</th>
              </tr>
            </thead>
            <tbody>
              {secretRefs.map(s => (
                <SecretItem key={s.getResourceURL()} secretRef={s} />
              ))}
            </tbody>
          </table>
        </React.Fragment>
      );
    }
    return secretSection;
  }
}

export default SecretsTable;
