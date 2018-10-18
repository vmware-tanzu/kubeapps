import * as React from "react";

import { IServiceBindingWithSecret } from "../../shared/ServiceBinding";
import BindingListEntry from "./BindingListEntry";

interface IBindingList {
  bindingsWithSecrets: IServiceBindingWithSecret[];
  removeBinding: (name: string, namespace: string) => Promise<boolean>;
}

class BindingList extends React.Component<IBindingList> {
  public render() {
    const { removeBinding, bindingsWithSecrets } = this.props;
    return (
      <div className="BindingList">
        <table>
          <thead>
            <tr className="row">
              <th className="col-2">Binding</th>
              <th className="col-2">Status</th>
              <th className="col-2">Message</th>
              <th className="col-4">Details</th>
              <th className="col-2" />
            </tr>
          </thead>
          <tbody>
            {bindingsWithSecrets.length > 0 ? (
              bindingsWithSecrets.map(b => [
                <BindingListEntry
                  key={b.binding.metadata.uid}
                  bindingWithSecret={b}
                  removeBinding={removeBinding}
                />,
              ])
            ) : (
              <tr>
                <td colSpan={3}>No bindings found</td>
              </tr>
            )}
          </tbody>
        </table>
      </div>
    );
  }
}

export default BindingList;
