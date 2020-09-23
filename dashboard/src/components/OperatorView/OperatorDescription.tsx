import * as React from "react";
import ReactMarkdown from "react-markdown";

import HeadingRenderer from "../ChartView/HeadingRenderer";
import LinkRenderer from "../ChartView/LinkRenderer";

interface IOperatorDescriptionProps {
  description: string;
}

export default function OperatorDescription(props: IOperatorDescriptionProps) {
  return (
    <div className="application-readme">
      <ReactMarkdown
        source={props.description}
        renderers={{
          heading: HeadingRenderer,
          link: LinkRenderer,
        }}
        skipHtml={true}
      />
    </div>
  );
}
