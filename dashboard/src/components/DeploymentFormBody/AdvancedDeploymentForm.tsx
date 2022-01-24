// Copyright 2019-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

// Import ace first
import AceEditor from "react-ace";
import "ace-builds/src-noconflict/ext-searchbox";
import "ace-builds/src-noconflict/mode-yaml";
import "ace-builds/src-noconflict/theme-solarized_dark";
import "ace-builds/src-noconflict/theme-xcode";
import { useSelector } from "react-redux";
import { IStoreState } from "shared/types";

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
  const {
    config: { theme },
  } = useSelector((state: IStoreState) => state);

  return (
    <div className="deployment-form-tabs-data">
      <AceEditor
        mode="yaml"
        theme={theme === "dark" ? "solarized_dark" : "xcode"}
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
