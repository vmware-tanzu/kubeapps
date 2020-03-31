import * as React from "react";
import { FileText } from "react-feather";
import * as ReactMarkdown from "react-markdown";

import LoadingWrapper from "../LoadingWrapper";
import HeadingRenderer from "./HeadingRenderer";

import "./ChartReadme.css";
import LinkRenderer from "./LinkRenderer";

interface IChartReadmeProps {
  getChartReadme: (namespace: string, version: string) => void;
  hasError: boolean;
  readme?: string;
  version: string;
  chartNamespace: string;
}

class ChartReadme extends React.Component<IChartReadmeProps> {
  public componentDidMount() {
    const { getChartReadme, chartNamespace, version } = this.props;
    getChartReadme(chartNamespace, version);
  }

  public componentDidUpdate(prevProps: IChartReadmeProps) {
    const { getChartReadme, chartNamespace, version } = this.props;
    if (version !== prevProps.version) {
      getChartReadme(chartNamespace, version);
    }
  }

  public render() {
    const { hasError, readme } = this.props;
    if (hasError) {
      return this.renderError();
    }
    return (
      <LoadingWrapper loaded={!!readme}>
        {readme && (
          <div className="ChartReadme">
            <ReactMarkdown
              source={readme}
              renderers={{
                heading: HeadingRenderer,
                link: LinkRenderer,
              }}
              skipHtml={true}
            />
          </div>
        )}
      </LoadingWrapper>
    );
  }

  public renderError() {
    return (
      <div className="ChartReadme ChartReadme--error flex align-center text-c">
        <FileText size={64} />
        <h4>No README found</h4>
      </div>
    );
  }
}

export default ChartReadme;
