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
    const repoNamespace = version.relationships.chart.data.repo.namespace;
    const chartName = version.relationships.chart.data.name;
    const versionStr = version.attributes.version;
    const newSegment = repoNamespace === namespace ? "new" : "new-from-global";

    return (
      <div className="ChartDeployButton text-r">
        <button className="button button-primary button-accent" onClick={this.handleClick}>
          Deploy
        </button>
        {this.state.clicked && (
          <Redirect
            push={true}
            to={`/ns/${namespace}/apps/${newSegment}/${repoName}/${chartName}/versions/${versionStr}`}
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
