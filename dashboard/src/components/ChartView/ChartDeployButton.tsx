import * as React from "react";
import { Redirect } from "react-router";
import { IChartVersion } from "../../shared/types";

interface IChartDeployButtonProps {
  version: IChartVersion;
}

interface IChartDeployButtonState {
  clicked: boolean;
}

class ChartDeployButton extends React.Component<IChartDeployButtonProps, IChartDeployButtonState> {
  public state: IChartDeployButtonState = {
    clicked: false,
  };

  public render() {
    const { version } = this.props;
    const repoName = version.relationships.chart.data.repo.name;
    const chartName = version.relationships.chart.data.name;
    const versionStr = version.attributes.version;

    return (
      <div className="ChartDeployButton">
        <button className="button button-primary" onClick={this.handleClick}>
          Deploy using Helm
        </button>
        {this.state.clicked && (
          <Redirect to={`/apps/new/${repoName}/${chartName}/versions/${versionStr}`} />
        )}
      </div>
    );
  }

  private handleClick = () => {
    this.setState({ clicked: true });
  };
}

export default ChartDeployButton;
