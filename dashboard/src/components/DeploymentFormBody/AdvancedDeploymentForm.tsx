import * as React from "react";
import AceEditor from "react-ace";

import "ace-builds/src-noconflict/mode-yaml";
import "ace-builds/src-noconflict/theme-xcode";

export interface IAdvancedDeploymentForm {
  appValues?: string;
  handleValuesChange: (value: string) => void;
}

class AdvancedDeploymentForm extends React.Component<IAdvancedDeploymentForm> {
  public render() {
    return (
      <div className="margin-t-normal">
        <AceEditor
          mode="yaml"
          theme="xcode"
          width="100%"
          onChange={this.props.handleValuesChange}
          setOptions={{ showPrintMargin: false }}
          editorProps={{ $blockScrolling: Infinity }}
          value={this.props.appValues}
          className="editor"
        />
        <p>
          <b>Note:</b> Only comments from the original chart values will be preserved.
        </p>
      </div>
    );
  }
}

export default AdvancedDeploymentForm;
