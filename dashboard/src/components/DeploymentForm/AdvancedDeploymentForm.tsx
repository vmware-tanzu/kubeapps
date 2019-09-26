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
      <div style={{ marginBottom: "1em" }}>
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
        />
      </div>
    );
  }
}

export default AdvancedDeploymentForm;
