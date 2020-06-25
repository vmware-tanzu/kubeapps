import * as React from "react";
import ReactMarkdown from "react-markdown";

import HeadingRenderer from "../ChartView/HeadingRenderer";
import LinkRenderer from "../ChartView/LinkRenderer";

interface IOperatorDescriptionProps {
  description: string;
}

class OperatorDescription extends React.Component<IOperatorDescriptionProps> {
  public render() {
    return (
      <div className="ChartReadme">
        <ReactMarkdown
          source={this.props.description}
          renderers={{
            heading: HeadingRenderer,
            link: LinkRenderer,
          }}
          skipHtml={true}
        />
      </div>
    );
  }
}

export default OperatorDescription;
