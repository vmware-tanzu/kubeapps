import ReactMarkdown from "react-markdown";
import remarkGfm from "remark-gfm";

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
        plugins={[remarkGfm]}
        renderers={{
          heading: HeadingRenderer,
          link: LinkRenderer,
        }}
        skipHtml={true}
      />
    </div>
  );
}
