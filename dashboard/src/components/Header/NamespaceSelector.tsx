import * as React from "react";
import { AlertCircle } from "react-feather";
import * as Select from "react-select";
import * as ReactTooltip from "react-tooltip";

import { INamespaceState } from "../../reducers/namespace";
import { definedNamespaces } from "../../shared/Namespace";
import { ForbiddenError, NotFoundError } from "../../shared/types";

import "./NamespaceSelector.css";
import NewNamespace from "./NewNamespace";

interface INamespaceSelectorProps {
  namespace: INamespaceState;
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
    return this.props.namespace.current || this.props.defaultNamespace;
  }

  public componentDidMount() {
    this.props.fetchNamespaces();
    this.props.getNamespace(this.selected);
  }

  public render() {
    const {
      namespace: { namespaces, error },
      defaultNamespace,
    } = this.props;
    const options =
      namespaces.length > 0
        ? namespaces.map(n => ({ value: n, label: n }))
        : [{ value: defaultNamespace, label: defaultNamespace }];
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
        {error && this.renderError(error.error)}
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

  private errorText = (err: Error) => {
    switch (err.constructor) {
      case ForbiddenError:
        return `You don't have sufficient permissions to use the namespace ${this.selected}`;
      case NotFoundError:
        return `Namespace ${this.selected} not found. Create it before using it.`;
      default:
        return err.message;
    }
  };

  private renderError = (err: Error) => {
    return (
      <>
        <a data-tip={true} data-for="ns-error">
          <AlertCircle className="NamespaceSelectorError" color="white" fill="red" />
        </a>
        <ReactTooltip
          className="NamespaceTooltipError"
          delayHide={1000}
          id="ns-error"
          effect="solid"
          place="bottom"
        >
          {this.errorText(err)}
        </ReactTooltip>
      </>
    );
  };
}

export default NamespaceSelector;
