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
          fontSize="15px"
        />
        {this.props.children}
      </div>
    );
  }
}

export default AdvancedDeploymentForm;
