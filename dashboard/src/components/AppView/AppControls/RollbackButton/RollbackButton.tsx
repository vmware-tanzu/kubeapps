import * as React from "react";
import * as Modal from "react-modal";
import { IAppRepository, IChartVersion, IRelease } from "shared/types";
import SelectRepoForm from "../../../../components/SelectRepoForm";
import RollbackDialog from "./RollbackDialog";

interface IRollbackButtonProps {
  app: IRelease;
  rollbackApp: (
    chartVersion: IChartVersion,
    releaseName: string,
    revision: number,
    namespace: string,
    values?: string,
  ) => Promise<boolean>;
  chart?: IChartVersion;
  getChartVersion: (id: string, version: string) => Promise<void>;
  loading: boolean;
  repos: IAppRepository[];
  repo: IAppRepository;
  kubeappsNamespace: string;
  fetchRepositories: () => void;
  checkChart: (repo: string, chartName: string) => Promise<boolean>;
  error?: Error;
}

interface IRollbackButtonState {
  modalIsOpen: boolean;
  loading: boolean;
  chartName: string;
  chartVersion: string;
}

class RollbackButton extends React.Component<IRollbackButtonProps> {
  public static getDerivedStateFromProps(props: IRollbackButtonProps) {
    // Store the chart name and version in the state for convenience
    if (props.app) {
      if (
        props.app.chart &&
        props.app.chart.metadata &&
        props.app.chart.metadata.name &&
        props.app.chart.metadata.version
      ) {
        return {
          chartName: props.app.chart.metadata.name,
          chartVersion: props.app.chart.metadata.version,
        };
      } else {
        // This should not be reached, unexpected error
        throw new Error("The current app is missing its chart information");
      }
    }
    return null;
  }

  public state: IRollbackButtonState = {
    modalIsOpen: false,
    loading: false,
    chartName: "",
    chartVersion: "",
  };

  public render() {
    return (
      <React.Fragment>
        <Modal
          style={{
            content: {
              bottom: "auto",
              left: "50%",
              marginRight: "-50%",
              right: "auto",
              top: "50%",
              transform: "translate(-50%, -50%)",
            },
          }}
          isOpen={this.state.modalIsOpen}
          onRequestClose={this.closeModal}
          contentLabel="Modal"
        >
          {/* If we were not able to resolve the chart, ask for the repository */}
          {this.props.chart ? (
            <RollbackDialog
              onConfirm={this.handleRollback}
              loading={this.state.loading}
              closeModal={this.closeModal}
              revision={this.props.app.version}
            />
          ) : (
            <SelectRepoForm
              repos={this.props.repos}
              repo={this.props.repo}
              kubeappsNamespace={this.props.kubeappsNamespace}
              checkChart={this.getChart}
              error={this.props.error}
              chartName={this.state.chartName}
            />
          )}
        </Modal>
        <button className="button" onClick={this.openModal}>
          Rollback
        </button>
      </React.Fragment>
    );
  }

  public openModal = () => {
    const { repos, fetchRepositories, chart, app, getChartVersion } = this.props;

    if (!chart && app.updateInfo) {
      // If there is updateInfo we can retrieve the chart
      const chartID = `${app.updateInfo.repository.name}/${this.state.chartName}`;
      getChartVersion(chartID, this.state.chartVersion);
    } else {
      // In other case we need to ask for the repository so we fecth the available ones
      if (repos.length === 0) {
        fetchRepositories();
      }
    }
    this.setState({
      modalIsOpen: true,
    });
  };

  public closeModal = async () => {
    this.setState({
      modalIsOpen: false,
    });
  };

  private handleRollback = (revision: number) => {
    return async () => {
      this.setState({ loading: true });
      const success = await this.props.rollbackApp(
        this.props.chart!, // Chart should be defined to reach this point
        this.props.app.name,
        revision,
        this.props.app.namespace,
        (this.props.app.config && this.props.app.config.raw) || "",
      );
      // If there is an error it's catched at AppView level
      if (success) {
        this.setState({ loading: false });
        this.closeModal();
      }
    };
  };

  private getChart = async (repo: string, chartName: string) => {
    const exists = await this.props.checkChart(repo, chartName);
    if (exists) {
      const chartID = `${repo}/${chartName}`;
      this.props.getChartVersion(chartID, this.state.chartVersion);
    }
  };
}

export default RollbackButton;
