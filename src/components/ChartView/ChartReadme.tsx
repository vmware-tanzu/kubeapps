import * as React from "react";
import * as ReactMarkdown from "react-markdown";

import "./ChartReadme.css";

interface IChartReadmeProps {
  markdown?: string;
}

class ChartReadme extends React.Component<IChartReadmeProps> {
  public render() {
    const { markdown } = this.props;
    return (
      <div className="ChartReadme">
        {markdown ? <ReactMarkdown source={markdown} /> : "Loading"}
      </div>
    );
  }
}

export default ChartReadme;
