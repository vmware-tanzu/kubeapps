import * as React from "react";
import * as Modal from "react-modal";
import { RouterAction } from "react-router-redux";

import { IClusterServiceClass } from "../../shared/ClusterServiceClass";
import { IServicePlan } from "../../shared/ServiceCatalog";
import { ForbiddenError, IRBACRole, NotFoundError } from "../../shared/types";
import { NotFoundErrorAlert, PermissionsErrorAlert, UnexpectedErrorAlert } from "../ErrorAlert";
import SchemaForm from "../SchemaForm";

import { JSONSchema6 } from "json-schema";
import { ISubmitEvent } from "react-jsonschema-form";

interface IProvisionButtonProps {
  namespace: string;
  error: Error;
  selectedClass?: IClusterServiceClass;
  selectedPlan: IServicePlan;
  provision: (
    releaseName: string,
    namespace: string,
    className: string,
    planName: string,
    parameters: {},
  ) => Promise<boolean>;
  push: (location: string) => RouterAction;
}

interface IProvisionButtonState {
  isProvisioning: boolean;
  modalIsOpen: boolean;
}

const RequiredRBACRoles: IRBACRole[] = [
  {
    apiGroup: "servicecatalog.k8s.io",
    resource: "serviceinstances",
    verbs: ["create"],
  },
];

const NameProperty: JSONSchema6 = {
  description: "Name for ServiceInstance",
  type: "string",
};

class ProvisionButton extends React.Component<IProvisionButtonProps, IProvisionButtonState> {
  public state: IProvisionButtonState = {
    isProvisioning: false,
    modalIsOpen: false,
  };

  public render() {
    const { selectedPlan } = this.props;
    let schema = selectedPlan.spec.instanceCreateParameterSchema;
    if (schema) {
      schema.properties = {
        name: NameProperty,
        ...(schema.properties || {}),
      };
      schema.required = [...(schema.required || []), "name"];
    } else {
      // If the Service Broker does not define a schema, default to a raw
      // parameters JSON object
      schema = {
        properties: {
          kubeappsRawParameters: {
            type: "object",
          },
          name: NameProperty,
        },
        type: "object",
      };
    }

    return (
      <div className="ProvisionButton">
        {this.state.isProvisioning && <div>Provisioning...</div>}
        <button
          className="button button-primary button-small"
          onClick={this.openModel}
          disabled={this.state.isProvisioning}
        >
          Provision
        </button>
        <Modal
          isOpen={this.state.modalIsOpen}
          onRequestClose={this.closeModal}
          contentLabel="Modal"
        >
          {this.props.error && <div className="margin-b-big">{this.renderError()}</div>}
          <SchemaForm schema={schema} onSubmit={this.handleProvision}>
            <div>
              <button className="button button-primary" type="submit">
                Submit
              </button>
              <button className="button" onClick={this.closeModal}>
                Cancel
              </button>
            </div>
          </SchemaForm>
        </Modal>
      </div>
    );
  }

  public openModel = () => {
    this.setState({
      modalIsOpen: true,
    });
  };

  public closeModal = () => {
    this.setState({
      modalIsOpen: false,
    });
  };

  public handleProvision = async ({
    formData,
  }: ISubmitEvent<{ name: string; kubeappsRawParameters: {} }>) => {
    const { namespace, provision, push, selectedClass, selectedPlan } = this.props;
    this.setState({ isProvisioning: true });

    const { name, kubeappsRawParameters, ...rest } = formData;
    if (selectedClass && selectedPlan) {
      const provisioned = await provision(
        name,
        namespace,
        selectedClass.spec.externalName,
        selectedPlan.spec.externalName,
        kubeappsRawParameters || rest,
      );
      if (provisioned) {
        push(
          `/services/brokers/${
            selectedClass.spec.clusterServiceBrokerName
          }/instances/ns/${namespace}/${name}`,
        );
      } else {
        this.setState({ isProvisioning: false });
      }
    }
  };

  private renderError() {
    const { error, namespace } = this.props;
    switch (error && error.constructor) {
      case ForbiddenError:
        return (
          <PermissionsErrorAlert
            namespace={namespace}
            roles={RequiredRBACRoles}
            action={`provision Service Instance`}
          />
        );
      case NotFoundError:
        return <NotFoundErrorAlert resource={`Namespace "${namespace}"`} />;
      default:
        return <UnexpectedErrorAlert />;
    }
  }
}

export default ProvisionButton;
