import * as React from "react";
import * as ReactMarkdown from "react-markdown";

import "./ChartReadme.css";

interface IChartReadmeProps {
  markdown?: string;
}

class ChartReadme extends React.Component<IChartReadmeProps> {
  public render() {
    let { markdown } = this.props;
    if (markdown === "") {
      markdown = "No README for this chart";
    }
    return (
      <div className="ChartReadme">
        {markdown ? <ReactMarkdown source={markdown} /> : "Loading"}
      </div>
    );
  }
}

export default ChartReadme;
