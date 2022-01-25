// Copyright 2021-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { CdsIcon } from "@cds/react/icon";
import Alert from "components/js/Alert";
import ReactMarkdown from "react-markdown";
import remarkGfm from "remark-gfm";
import LoadingWrapper from "../LoadingWrapper/LoadingWrapper";
import HeadingRenderer from "../MarkdownRenderer/HeadingRenderer";
import LinkRenderer from "../MarkdownRenderer/LinkRenderer";
import TableRenderer from "../MarkdownRenderer/TableRenderer";

interface IPackageReadmeProps {
  error?: string;
  readme?: string;
}

function PackageReadme({ error, readme }: IPackageReadmeProps) {
  if (error) {
    if (error.toLocaleLowerCase().includes("not found")) {
      return (
        <div className="section-not-found">
          <div>
            <CdsIcon shape="file" size="64" />
            <h4>No README found</h4>
          </div>
        </div>
      );
    }
    return <Alert theme="danger">Unable to fetch package README: {error}</Alert>;
  }
  return (
    <LoadingWrapper
      className="margin-t-xxl"
      loadingText="Fetching application README..."
      loaded={!!readme}
    >
      {readme && (
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
            {readme}
          </ReactMarkdown>
        </div>
      )}
    </LoadingWrapper>
  );
}

export default PackageReadme;
