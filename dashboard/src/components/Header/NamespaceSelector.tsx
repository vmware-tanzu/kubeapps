import * as React from "react";
import * as Select from "react-select";

import { INamespaceState } from "../../reducers/namespace";
import { definedNamespaces } from "../../shared/Namespace";

import "./NamespaceSelector.css";

interface INamespaceSelectorProps {
  namespace: INamespaceState;
  onChange: (ns: string) => any;
  fetchNamespaces: () => void;
}

class NamespaceSelector extends React.Component<INamespaceSelectorProps> {
  public componentDidMount() {
    this.props.fetchNamespaces();
  }

  public render() {
    const {
      namespace: { current, namespaces },
    } = this.props;
    const options = namespaces.map(n => ({ value: n, label: n }));
    const allOption = { value: definedNamespaces.all, label: "All Namespaces" };
    options.unshift(allOption);
    const value =
      current === definedNamespaces.all ? allOption : { value: current, label: current };
    return (
      <div className="NamespaceSelector margin-r-normal">
        <label className="NamespaceSelector__label type-tiny">NAMESPACE</label>
        <Select.Creatable
          className="NamespaceSelector__select type-small"
          value={value}
          options={options}
          multi={false}
          onChange={this.handleNamespaceChange}
          promptTextCreator={this.promptTextCreator}
          clearable={false}
        />
      </div>
    );
  }

  private handleNamespaceChange = (value: any) => {
    if (value) {
      this.props.onChange(value.value);
    }
  };

  private promptTextCreator = (text: string) => {
    return `Use namespace "${text}"`;
  };
}

export default NamespaceSelector;
