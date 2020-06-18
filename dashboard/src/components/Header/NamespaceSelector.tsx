import * as React from "react";
import * as Select from "react-select";

import { IClusterState } from "../../reducers/namespace";
import { definedNamespaces } from "../../shared/Namespace";

import "./NamespaceSelector.css";
import NewNamespace from "./NewNamespace";

export interface INamespaceSelectorProps {
  cluster: IClusterState;
  defaultNamespace: string;
  onChange: (ns: string) => any;
  fetchNamespaces: () => void;
  createNamespace: (ns: string) => Promise<boolean>;
  getNamespace: (ns: string) => void;
}

interface INamespaceSelectorState {
  modalIsOpen: boolean;
  newNamespace: string;
}

class NamespaceSelector extends React.Component<INamespaceSelectorProps, INamespaceSelectorState> {
  public state: INamespaceSelectorState = {
    modalIsOpen: false,
    newNamespace: "",
  };

  get selected() {
    return this.props.cluster.currentNamespace || this.props.defaultNamespace;
  }

  public componentDidMount() {
    this.props.fetchNamespaces();
    if (this.selected !== definedNamespaces.all) {
      this.props.getNamespace(this.selected);
    }
  }

  public render() {
    const {
      cluster: { namespaces, error },
    } = this.props;
    const options = namespaces.length > 0 ? namespaces.map(n => ({ value: n, label: n })) : [];
    const allOption = { value: definedNamespaces.all, label: "All Namespaces" };
    options.unshift(allOption);
    const newOption = { value: "_new", label: "Create New" };
    options.push(newOption);
    const value =
      this.selected === definedNamespaces.all
        ? allOption
        : { value: this.selected, label: this.selected };
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
        <NewNamespace
          modalIsOpen={this.state.modalIsOpen}
          namespace={this.state.newNamespace}
          onConfirm={this.onConfirmNewNS}
          onChange={this.handleChangeNewNS}
          closeModal={this.closeModal}
          error={error && error.action === "create" ? error.error : undefined}
        />
      </div>
    );
  }

  private handleNamespaceChange = (value: any) => {
    if (value.value === "_new") {
      this.setState({ modalIsOpen: true });
      return;
    }
    if (value) {
      this.props.onChange(value.value);
    }
  };

  private promptTextCreator = (text: string) => {
    return `Use namespace "${text}"`;
  };

  private onConfirmNewNS = async () => {
    const success = await this.props.createNamespace(this.state.newNamespace);
    if (success) {
      this.handleNamespaceChange({ value: this.state.newNamespace });
      this.closeModal();
    }
  };

  private handleChangeNewNS = (e: React.FormEvent<HTMLInputElement>) => {
    this.setState({ newNamespace: e.currentTarget.value });
  };

  private closeModal = async () => {
    this.setState({
      modalIsOpen: false,
    });
  };
}

export default NamespaceSelector;
