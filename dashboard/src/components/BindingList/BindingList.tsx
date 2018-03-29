import * as React from "react";
import { IServiceBinding } from "../../shared/ServiceBinding";
import { BindingEntry } from "./BindingListEntry";

interface IBindingList {
  bindings: IServiceBinding[];
  getCatalog: () => Promise<any>;
}

export class BindingList extends React.Component<IBindingList> {
  public render() {
    const { bindings, getCatalog } = this.props;
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
            {bindings.length > 0 ? (
              bindings.map(binding => [
                <BindingEntry
                  key={binding.metadata.uid}
                  binding={binding}
                  getCatalog={getCatalog}
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
