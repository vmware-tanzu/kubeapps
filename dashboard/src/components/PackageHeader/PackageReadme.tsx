// Copyright 2021-2023 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { CdsIcon } from "@cds/react/icon";
import AlertGroup from "components/AlertGroup";
import LoadingWrapper from "components/LoadingWrapper";
import ReactMarkdown from "react-markdown";
import remarkGfm from "remark-gfm";
import HeadingRenderer from "../MarkdownRenderer/HeadingRenderer";
import LinkRenderer from "../MarkdownRenderer/LinkRenderer";
import TableRenderer from "../MarkdownRenderer/TableRenderer";

export interface IPackageReadmeProps {
  error?: string;
  readme?: string;
  isFetching?: boolean;
}

function PackageReadme({ error, readme, isFetching }: IPackageReadmeProps) {
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
    return <AlertGroup status="danger">Unable to fetch the package's README: {error}.</AlertGroup>;
  }
  return (
    <LoadingWrapper
      className="margin-t-xxl"
      loadingText="Fetching application README..."
      loaded={!isFetching}
    >
      <div className="application-readme">
        {readme ? (
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
        ) : (
          <p> This package does not contain a README file.</p>
        )}
      </div>
    </LoadingWrapper>
  );
}

export default PackageReadme;
