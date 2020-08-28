import * as React from "react";
import AceEditor from "react-ace";

import "ace-builds/src-noconflict/mode-yaml";
import "ace-builds/src-noconflict/theme-xcode";

export interface IAdvancedDeploymentForm {
  appValues?: string;
  handleValuesChange: (value: string) => void;
  children?: JSX.Element;
}

function AdvancedDeploymentForm(props: IAdvancedDeploymentForm) {
  let timeout: NodeJS.Timeout;
  const onChange = (value: string) => {
    // Gather changes before submitting
    clearTimeout(timeout);
    timeout = setTimeout(() => props.handleValuesChange(value), 500);
  };
  return (
    <div className="deployment-form-tabs-data">
      <AceEditor
        mode="yaml"
        theme="xcode"
        width="100%"
        onChange={onChange}
        setOptions={{ showPrintMargin: false }}
        editorProps={{ $blockScrolling: Infinity }}
        value={props.appValues}
        className="editor"
        fontSize="15px"
      />
      {props.children}
    </div>
  );
}

export default AdvancedDeploymentForm;
