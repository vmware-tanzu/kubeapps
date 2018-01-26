import * as React from "react";
import * as Modal from "react-modal";
import { RouterAction } from "react-router-redux";
import { IChartVersion } from "../../shared/types";

interface IChartDeployButtonProps {
  version: IChartVersion;
  deployChart: (chartVersion: IChartVersion, releaseName: string, namespace: string) => Promise<{}>;
  push: (location: string) => RouterAction;
}

interface IChartDeployButtonState {
  isDeploying: boolean;
  modalIsOpen: boolean;
  // deployment options
  releaseName: string;
  namespace: string;
  error: string | null;
}

class ChartDeployButton extends React.Component<IChartDeployButtonProps, IChartDeployButtonState> {
  public state: IChartDeployButtonState = {
    error: null,
    isDeploying: false,
    modalIsOpen: false,
    namespace: "default",
    releaseName: "",
  };

  public render() {
    return (
      <div className="ChartDeployButton">
        {this.state.isDeploying && <div>Deploying...</div>}
        <button
          className="button button-primary"
          onClick={this.openModel}
          disabled={this.state.isDeploying}
        >
          Deploy using Helm
        </button>
        <Modal
          isOpen={this.state.modalIsOpen}
          onRequestClose={this.closeModal}
          contentLabel="Modal"
        >
          {this.state.error && (
            <div className="container padding-v-bigger bg-action">{this.state.error}</div>
          )}
          <form onSubmit={this.handleDeploy}>
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
              <label htmlFor="namespace">Namespace</label>
              <input
                name="namespace"
                onChange={this.handleNamespaceChange}
                value={this.state.namespace}
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

  public handleDeploy = (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    const { version, deployChart, push } = this.props;
    this.setState({ isDeploying: true });
    const { releaseName, namespace } = this.state;
    deployChart(version, releaseName, namespace)
      .then(() => push(`/apps/${releaseName}`))
      .catch(err => this.setState({ isDeploying: false, error: err.toString() }));
  };

  public handleReleaseNameChange = (e: React.FormEvent<HTMLInputElement>) => {
    this.setState({ releaseName: e.currentTarget.value });
  };
  public handleNamespaceChange = (e: React.FormEvent<HTMLInputElement>) => {
    this.setState({ namespace: e.currentTarget.value });
  };
}

export default ChartDeployButton;
