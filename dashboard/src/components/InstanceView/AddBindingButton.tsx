import * as React from "react";
import * as Modal from "react-modal";

import { JSONSchema6 } from "json-schema";
import { ISubmitEvent } from "react-jsonschema-form";
import { ForbiddenError, IRBACRole, NotFoundError } from "../../shared/types";
import { NotFoundErrorAlert, PermissionsErrorAlert, UnexpectedErrorAlert } from "../ErrorAlert";
import SchemaForm from "../SchemaForm";

interface IAddBindingButtonProps {
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
}

const RequiredRBACRoles: IRBACRole[] = [
  {
    apiGroup: "servicecatalog.k8s.io",
    resource: "servicebindings",
    verbs: ["create"],
  },
];

export class AddBindingButton extends React.Component<
  IAddBindingButtonProps,
  IAddBindingButtonState
> {
  public state = {
    bindingName: `${this.props.instanceRefName}-binding`,
    displayNameForm: true,
    modalIsOpen: false,
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
        <button className="button button-primary" onClick={this.openModal}>
          Add Binding
        </button>
        <Modal isOpen={modalIsOpen} onRequestClose={this.closeModal}>
          {this.props.error && <div className="margin-b-big">{this.renderError()}</div>}
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

  private renderError() {
    const { error, namespace } = this.props;
    const { bindingName } = this.state;
    switch (error && error.constructor) {
      case ForbiddenError:
        return (
          <PermissionsErrorAlert
            namespace={namespace}
            roles={RequiredRBACRoles}
            action={`create Service Binding "${bindingName}"`}
          />
        );
      case NotFoundError:
        return <NotFoundErrorAlert resource={`Namespace "${namespace}"`} />;
      default:
        return <UnexpectedErrorAlert />;
    }
  }

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
