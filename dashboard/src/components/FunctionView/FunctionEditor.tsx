import * as React from "react";
import AceEditor from "react-ace";

import "brace/mode/golang";
import "brace/mode/javascript";
import "brace/mode/php";
import "brace/mode/python";
import "brace/mode/ruby";
import "brace/theme/xcode";

interface IFunctionEditorProps {
  runtime: string;
  value: string;
  onChange: (value: string) => void;
}

class FunctionEditor extends React.Component<IFunctionEditorProps> {
  public render() {
    const { value, onChange } = this.props;
    return (
      <div className="FunctionEditor">
        <AceEditor
          mode={this.runtimeToMode()}
          theme="xcode"
          name="values"
          width="100%"
          onChange={onChange}
          setOptions={{ showPrintMargin: false }}
          value={value}
        />
      </div>
    );
  }

  private runtimeToMode() {
    const { runtime } = this.props;
    if (runtime.match(/go/)) {
      return "golang";
    } else if (runtime.match(/node/)) {
      return "javascript";
    } else if (runtime.match(/ruby/)) {
      return "ruby";
    } else if (runtime.match(/php/)) {
      return "php";
    } else if (runtime.match(/python/)) {
      return "python";
    }
    return "";
  }
}

export default FunctionEditor;
