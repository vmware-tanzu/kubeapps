import * as React from "react";
import AceEditor from "react-ace";
import * as Modal from "react-modal";
import { RouterAction } from "react-router-redux";

import { IClusterServiceClass } from "../../shared/ClusterServiceClass";
import { IServicePlan } from "../../shared/ServiceCatalog";
import { ForbiddenError, IRBACRole, NotFoundError } from "../../shared/types";
import { NotFoundErrorAlert, PermissionsErrorAlert, UnexpectedErrorAlert } from "../ErrorAlert";

import "brace/mode/json";
import "brace/theme/xcode";

interface IProvisionButtonProps {
  namespace: string;
  error: Error;
  plans: IServicePlan[];
  classes: IClusterServiceClass[];
  selectedClass?: IClusterServiceClass;
  selectedPlan?: IServicePlan;
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
  // deployment options
  releaseName: string;
  selectedPlan: IServicePlan | undefined;
  selectedClass: IClusterServiceClass | undefined;
  parameters: string;
}

const RequiredRBACRoles: IRBACRole[] = [
  {
    apiGroup: "servicecatalog.k8s.io",
    resource: "serviceinstances",
    verbs: ["create"],
  },
];

class ProvisionButton extends React.Component<IProvisionButtonProps, IProvisionButtonState> {
  public state: IProvisionButtonState = {
    isProvisioning: false,
    modalIsOpen: false,
    parameters: JSON.stringify(
      {
        firewallRules: [
          {
            endIPAddress: "255.255.255.255",
            name: "AllowAll",
            startIPAddress: "0.0.0.0",
          },
        ],
        location: "eastus",
        resourceGroup: "default",
        sslEnforcement: "disabled",
      },
      undefined,
      2,
    ),
    releaseName: "",
    selectedClass: this.props.selectedClass,
    selectedPlan: this.props.selectedPlan,
  };

  public render() {
    const { plans, classes } = this.props;
    const { selectedClass, selectedPlan } = this.state;
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
          <form onSubmit={this.handleProvision}>
            <div>
              <label htmlFor="releaseName">Name</label>
              <input
                id="releaseName"
                onChange={this.handleReleaseNameChange}
                value={this.state.releaseName}
                required={true}
              />
            </div>
            <div>
              <label htmlFor="classes">Classes</label>
              <select onChange={this.onClassChange}>
                {classes.map(c => (
                  <option
                    key={c.spec.externalName}
                    selected={c.metadata.name === (selectedClass && selectedClass.metadata.name)}
                    value={c.spec.externalName}
                  >
                    {c.spec.externalName}
                  </option>
                ))}
              </select>
            </div>
            <div>
              <label htmlFor="plans">Plans</label>
              <select onChange={this.onPlanChange}>
                {plans
                  .filter(
                    plan =>
                      plan.spec.clusterServiceClassRef.name ===
                      (selectedClass && selectedClass.metadata.name),
                  )
                  .map(p => (
                    <option
                      key={p.spec.externalName}
                      value={p.spec.externalName}
                      selected={p.metadata.name === (selectedPlan && selectedPlan.metadata.name)}
                    >
                      {p.spec.externalName}
                    </option>
                  ))}
              </select>
            </div>
            <div style={{ marginBottom: "1em" }}>
              <label htmlFor="values">Parameters (JSON)</label>
              <AceEditor
                mode="json"
                theme="xcode"
                name="values"
                width="100%"
                height="200px"
                onChange={this.handleParametersChange}
                setOptions={{ showPrintMargin: false }}
                value={this.state.parameters}
              />
            </div>
            <div>
              <button className="button button-primary" type="submit">
                Submit
              </button>
              <button className="button" onClick={this.closeModal}>
                Cancel
              </button>
            </div>
          </form>
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

  public handleProvision = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    const { namespace, provision, push } = this.props;
    this.setState({ isProvisioning: true });
    const { releaseName, selectedClass, selectedPlan, parameters } = this.state;

    const parametersObject = JSON.parse(parameters);
    if (selectedClass && selectedPlan) {
      const provisioned = await provision(
        releaseName,
        namespace,
        selectedClass.spec.externalName,
        selectedPlan.spec.externalName,
        parametersObject,
      );
      if (provisioned) {
        push(
          `/services/brokers/${
            selectedClass.spec.clusterServiceBrokerName
          }/instances/ns/${namespace}/${releaseName}`,
        );
      } else {
        this.setState({ isProvisioning: false });
      }
    }
  };

  public handleReleaseNameChange = (e: React.FormEvent<HTMLInputElement>) => {
    this.setState({ releaseName: e.currentTarget.value });
  };
  public handleParametersChange = (parameter: string) => {
    this.setState({ parameters: parameter });
  };

  public onClassChange = (e: React.ChangeEvent<HTMLSelectElement>) =>
    this.setState({
      selectedClass:
        this.props.classes.find(svcClass => svcClass.spec.externalName === e.target.value) ||
        undefined,
    });
  public onPlanChange = (e: React.ChangeEvent<HTMLSelectElement>) =>
    this.setState({
      selectedPlan:
        this.props.plans.find(plan => plan.spec.externalName === e.target.value) || undefined,
    });

  private renderError() {
    const { error, namespace } = this.props;
    const { releaseName } = this.state;
    switch (error && error.constructor) {
      case ForbiddenError:
        return (
          <PermissionsErrorAlert
            namespace={namespace}
            roles={RequiredRBACRoles}
            action={`provision Service Instance "${releaseName}"`}
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
