import * as React from "react";
import * as Modal from "react-modal";

import { JSONSchema6 } from "json-schema";
import { ISubmitEvent } from "react-jsonschema-form";
import { IRBACRole } from "../../shared/types";
import { ErrorSelector } from "../ErrorAlert";
import SchemaForm from "../SchemaForm";

interface IAddBindingButtonProps {
  disabled: boolean;
  error?: Error;
  instanceRefName: string;
  namespace: string;
  addBinding: (
    bindingName: string,
    instanceName: string,
    namespace: string,
    parameters: {},
  ) => Promise<boolean>;
  onAddBinding: () => void;
  bindingSchema?: JSONSchema6;
}

interface IAddBindingButtonState {
  modalIsOpen: boolean;
  // deployment options
  bindingName: string;
  displayNameForm: boolean;
  // Name of the binding that was submitted for creation
  // This is different than bindingName since it is also used in the error banner
  // and we do not want to use bindingName since it is controller by the form field.
  latestSubmittedReleaseName: string;
}

const RequiredRBACRoles: IRBACRole[] = [
  {
    apiGroup: "servicecatalog.k8s.io",
    resource: "servicebindings",
    verbs: ["create"],
  },
];

class AddBindingButton extends React.Component<IAddBindingButtonProps, IAddBindingButtonState> {
  public state = {
    bindingName: `${this.props.instanceRefName}-binding`,
    displayNameForm: true,
    modalIsOpen: false,
    latestSubmittedReleaseName: `${this.props.instanceRefName}-binding`,
  };

  public render() {
    let { bindingSchema } = this.props;
    const { modalIsOpen, displayNameForm } = this.state;
    if (!bindingSchema || !(bindingSchema.$ref || bindingSchema.properties)) {
      bindingSchema = {
        properties: {
          kubeappsRawParameters: {
            title: "Parameters",
            type: "object",
          },
        },
        type: "object",
      };
    }
    return (
      <div className="AddBindingButton">
        <button
          className="button button-primary"
          onClick={this.openModal}
          disabled={this.props.disabled}
        >
          Add Binding
        </button>
        <Modal isOpen={modalIsOpen} onRequestClose={this.closeModal}>
          {this.props.error && (
            <ErrorSelector
              error={this.props.error}
              resource={`Service Binding "${this.state.latestSubmittedReleaseName}"`}
              action="create"
              namespace={this.props.namespace}
              defaultRequiredRBACRoles={{ create: RequiredRBACRoles }}
            />
          )}
          <div className="bind-form">
            <h1>Add Binding</h1>
            {displayNameForm ? (
              <SchemaForm schema={this.nameSchema()} onSubmit={this.handleNameChange}>
                <div>
                  <button className="button button-primary" type="submit">
                    Continue
                  </button>
                  <button className="button" onClick={this.closeModal}>
                    Cancel
                  </button>
                </div>
              </SchemaForm>
            ) : (
              <SchemaForm schema={bindingSchema} onSubmit={this.handleBind}>
                <div>
                  <button className="button button-primary" type="submit">
                    Submit
                  </button>
                  <button className="button" onClick={this.handleBackButton}>
                    Back
                  </button>
                </div>
              </SchemaForm>
            )}
          </div>
        </Modal>
      </div>
    );
  }

  private closeModal = () => this.setState({ modalIsOpen: false });
  private openModal = () => this.setState({ modalIsOpen: true });
  private handleNameChange = ({ formData }: ISubmitEvent<{ Name: string }>) => {
    this.setState({ bindingName: formData.Name, displayNameForm: false });
  };
  private handleBackButton = (e: React.MouseEvent<HTMLButtonElement>) => {
    e.preventDefault();
    this.setState({ displayNameForm: true });
  };

  private handleBind = async ({ formData }: ISubmitEvent<{ kubeappsRawParameters: {} }>) => {
    const { kubeappsRawParameters, ...rest } = formData;
    this.setState({ latestSubmittedReleaseName: this.state.bindingName });
    const added = await this.props.addBinding(
      this.state.bindingName,
      this.props.instanceRefName,
      this.props.namespace,
      kubeappsRawParameters || rest,
    );
    if (added) {
      this.closeModal();
      this.props.onAddBinding();
    }
  };

  private nameSchema(): JSONSchema6 {
    return {
      properties: {
        Name: {
          default: this.state.bindingName,
          description: "Name for ServiceBinding",
          type: "string",
        },
      },
      required: ["Name"],
      type: "object",
    };
  }
}

export default AddBindingButton;
