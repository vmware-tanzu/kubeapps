import * as React from "react";
import { Redirect } from "react-router";
import { IChartVersion } from "../../shared/types";

interface IChartDeployButtonProps {
  version: IChartVersion;
  namespace: string;
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
    const { namespace } = this.props;
    const repoName = version.relationships.chart.data.repo.name;
    const chartName = version.relationships.chart.data.name;
    const versionStr = version.attributes.version;

    return (
      <div className="ChartDeployButton text-r">
        <button className="button button-primary button-accent" onClick={this.handleClick}>
          Deploy
        </button>
        {this.state.clicked && (
          <Redirect
            push={true}
            to={`/apps/ns/${namespace}/new/${repoName}/${chartName}/versions/${versionStr}`}
          />
        )}
      </div>
    );
  }

  private handleClick = () => {
    this.setState({ clicked: true });
  };
}

export default ChartDeployButton;
