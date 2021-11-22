import TableRenderer from "components/PackageHeader/TableRenderer";
import ReactMarkdown from "react-markdown";
import remarkGfm from "remark-gfm";
import HeadingRenderer from "../PackageHeader/HeadingRenderer";
import LinkRenderer from "../PackageHeader/LinkRenderer";

interface IOperatorDescriptionProps {
  description: string;
}

export default function OperatorDescription(props: IOperatorDescriptionProps) {
  return (
    <div className="application-readme">
      <ReactMarkdown
        remarkPlugins={[remarkGfm]}
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
