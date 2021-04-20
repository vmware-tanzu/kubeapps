import TableRenderer from "components/ChartView/TableRenderer";
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
        plugins={[remarkGfm]}
        components={{
          h1: HeadingRenderer,
          h2: HeadingRenderer,
          h3: HeadingRenderer,
          h4: HeadingRenderer,
          h5: HeadingRenderer,
          h6: HeadingRenderer,
          a: LinkRenderer,
          table: TableRenderer,
        }}
        skipHtml={true}
      >
        {props.description}
      </ReactMarkdown>
    </div>
  );
}
