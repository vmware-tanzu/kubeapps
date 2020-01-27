import * as React from "react";
import * as Select from "react-select";
import * as ReactTooltip from "react-tooltip";

import { AlertCircle } from "react-feather";
import { INamespaceState } from "../../reducers/namespace";
import { definedNamespaces } from "../../shared/Namespace";

import "./NamespaceSelector.css";

interface INamespaceSelectorProps {
  namespace: INamespaceState;
  defaultNamespace: string;
  onChange: (ns: string) => any;
  fetchNamespaces: () => void;
}

class NamespaceSelector extends React.Component<INamespaceSelectorProps> {
  public componentDidMount() {
    this.props.fetchNamespaces();
  }

  public render() {
    const {
      namespace: { current, namespaces, error },
      defaultNamespace,
    } = this.props;
    const options =
      namespaces.length > 0
        ? namespaces.map(n => ({ value: n, label: n }))
        : [{ value: defaultNamespace, label: defaultNamespace }];
    const allOption = { value: definedNamespaces.all, label: "All Namespaces" };
    options.unshift(allOption);
    const selected = current || defaultNamespace;
    const value =
      selected === definedNamespaces.all ? allOption : { value: selected, label: selected };
    return (
      <div className="NamespaceSelector margin-r-normal">
        <label className="NamespaceSelector__label type-tiny">NAMESPACE</label>
        {error && error.action === "create" && this.renderError(error.errorMsg)}
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
    return `Create namespace "${text}"`;
  };

  private renderError = (errMsg: string) => {
    return (
      <>
        <a data-tip={true} data-for="ns-error">
          <AlertCircle className="NamespaceSelectorError" color="white" fill="red" />
        </a>
        <ReactTooltip id="ns-error" className="extraClass" effect="solid">
          {errMsg}
        </ReactTooltip>
      </>
    );
  };
}

export default NamespaceSelector;
