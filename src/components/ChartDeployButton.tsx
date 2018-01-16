import * as React from 'react';
import { Chart } from '../store/types';
import { RouterAction } from 'react-router-redux';
import * as Modal from 'react-modal';

interface Props {
  chart: Chart;
  deployChart: (chart: Chart, releaseName: string) => Promise<{}>;
  push: (location: string) => RouterAction;
}

interface State {
  isDeploying: boolean;
  modalIsOpen: boolean;
  // deployment options
  releaseName: string;
  namespace: string;
}

class ChartDeployButton extends React.Component<Props> {
  state: State = {
    isDeploying: false,
    modalIsOpen: false,
    releaseName: '',
    namespace: 'default'
  };

  render() {
    return (
      <div className="ChartDeployButton">
        {this.state.isDeploying &&
          <div>Deploying...</div>
        }
        <button
          className="button button-primary"
          onClick={this.openModel}
          disabled={this.state.isDeploying}
        >
          Deploy
        </button>
        <Modal
          isOpen={this.state.modalIsOpen}
          onRequestClose={this.closeModal}
          contentLabel="Modal"
        >
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
              <button className="button button-primary" type="submit">Submit</button>
              <button className="button" onClick={this.closeModal}>Cancel</button>
            </div>
          </form>
        </Modal>
      </div>
    );
  }

  openModel = () => {
    this.setState({
      modalIsOpen: true
    });
  }

  closeModal = () => {
    this.setState({
      modalIsOpen: false
    });
  }

  handleDeploy = (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    const { chart, deployChart, push } = this.props;
    this.setState({
      isDeploying: true,
    });
    deployChart(chart, this.state.releaseName)
      .then(() => push(`/apps/${this.state.releaseName}`))
      .catch(err => {
      this.setState({
        isDeploying: false,
      });
    });
  }

  handleReleaseNameChange = (e: React.FormEvent<HTMLInputElement>) => {
    this.setState({ releaseName: e.currentTarget.value });
  }
  handleNamespaceChange = (e: React.FormEvent<HTMLInputElement>) => {
    this.setState({ namespace: e.currentTarget.value });
  }
}

export default ChartDeployButton;
