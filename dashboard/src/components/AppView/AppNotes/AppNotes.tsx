// Copyright 2021-2023 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import HeadingRenderer from "components/MarkdownRenderer/HeadingRenderer";
import LinkRenderer from "components/MarkdownRenderer/LinkRenderer";
import TableRenderer from "components/MarkdownRenderer/TableRenderer";
import ReactMarkdown from "react-markdown";
import remarkBreaks from "remark-breaks";
import remarkGfm from "remark-gfm";

export interface IAppNotesProps {
  title?: string;
  notes?: string | null;
}

function AppNotes(props: IAppNotesProps) {
  const { title, notes } = props;
  return notes ? (
    <>
      <h3 className="section-title">{title ? title : "Installation Notes"}</h3>
      <div className="application-notes">
        <ReactMarkdown
          remarkPlugins={[remarkGfm, remarkBreaks]}
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
          {notes}
        </ReactMarkdown>
      </div>
    </>
  ) : (
    <div />
  );
}

export default AppNotes;
