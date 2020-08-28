import * as jsdiff from "diff";
import { Diff2Html } from "diff2html";
import * as React from "react";

import "diff2html/dist/diff2html.css";

export interface IDifferentialProps {
  title: string;
  oldValues: string;
  newValues: string;
  emptyDiffText: string;
}

class Differential extends React.Component<IDifferentialProps> {
  public render = () => {
    const { oldValues, newValues, title, emptyDiffText } = this.props;
    const sdiff = jsdiff.createPatch(title, oldValues, newValues);
    const outputHtml = Diff2Html.getPrettyHtml(sdiff, {
      inputFormat: "diff",
      showFiles: false,
      matching: "lines",
      maxLineSizeInBlockForComparison: 20,
    });
    return (
      <div className="diff deployment-form-tabs-data">
        {oldValues === newValues ? (
          <span>{emptyDiffText}</span>
        ) : (
          <div dangerouslySetInnerHTML={{ __html: outputHtml }} />
        )}
      </div>
    );
  };
}

export default Differential;
