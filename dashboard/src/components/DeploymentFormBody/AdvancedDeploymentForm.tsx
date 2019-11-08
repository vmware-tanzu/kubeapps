import * as React from "react";
import AceEditor from "react-ace";

import "brace/mode/yaml";
import "brace/theme/xcode";

export interface IAdvancedDeploymentForm {
  appValues?: string;
  handleValuesChange: (value: string) => void;
}

class AdvancedDeploymentForm extends React.Component<IAdvancedDeploymentForm> {
  public render() {
    return (
      <div className="margin-t-normal">
        <label htmlFor="values">Values (YAML)</label>
        <AceEditor
          mode="yaml"
          theme="xcode"
          name="values"
          width="100%"
          onChange={this.props.handleValuesChange}
          setOptions={{ showPrintMargin: false }}
          editorProps={{ $blockScrolling: Infinity }}
          value={this.props.appValues}
          className="editor"
        />
      </div>
    );
  }
}

export default AdvancedDeploymentForm;
