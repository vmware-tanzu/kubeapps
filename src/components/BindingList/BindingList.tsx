import * as React from "react";
import { IServiceBinding } from "../../shared/ServiceBinding";
import { BindingEntry } from "./BindingListEntry";

interface IBindingList {
  bindings: IServiceBinding[];
  addBinding: (bindingName: string, instanceName: string, namespace: string) => Promise<any>;
  getCatalog: () => Promise<any>;
}

export class BindingList extends React.Component<IBindingList> {
  public render() {
    const { bindings, getCatalog, addBinding } = this.props;
    return (
      <div className="BindingEntryList">
        <table>
          <thead>
            <tr>
              <th>Binding</th>
              <th>Status</th>
              <th />
            </tr>
          </thead>
          <tbody>
            {bindings.map(binding => [
              <BindingEntry
                key={binding.metadata.uid}
                binding={binding}
                addBinding={addBinding}
                getCatalog={getCatalog}
              />,
            ])}
          </tbody>
        </table>
      </div>
    );
  }
}
