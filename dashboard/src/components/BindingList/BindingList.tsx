import * as React from "react";

import { IServiceBindingWithSecret } from "../../shared/ServiceBinding";
import BindingListEntry from "./BindingListEntry";

import "./BindingList.css";

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
            <tr>
              <th>Binding</th>
              <th>Status</th>
              <th>Message</th>
              <th>Secret</th>
              <th />
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
                <td> No bindings found</td>
              </tr>
            )}
          </tbody>
        </table>
      </div>
    );
  }
}

export default BindingList;
