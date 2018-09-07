import * as React from "react";
import { Redirect } from "react-router";
import { definedNamespaces } from "../../shared/Namespace";
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
    let { namespace } = this.props;
    const repoName = version.relationships.chart.data.repo.name;
    const chartName = version.relationships.chart.data.name;
    const versionStr = version.attributes.version;

    // If our current namespace is not set a.k.a '_all' we actually set the default one
    if (namespace === definedNamespaces.all) {
      namespace = definedNamespaces.default;
    }

    return (
      <div className="ChartDeployButton">
        <button className="button button-primary" onClick={this.handleClick}>
          Deploy using Helm
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
