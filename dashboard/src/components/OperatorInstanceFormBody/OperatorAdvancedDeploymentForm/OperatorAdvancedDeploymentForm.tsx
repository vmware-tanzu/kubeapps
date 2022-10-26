// Copyright 2019-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { MonacoDiffEditor } from "react-monaco-editor";
import { useSelector } from "react-redux";
import { SupportedThemes } from "shared/Config";
import { IStoreState } from "shared/types";

export interface IOperatorAdvancedDeploymentFormProps {
  appValues?: string;
  oldAppValues?: string;
  handleValuesChange: (value: string) => void;
  children?: JSX.Element;
}

function OperatorAdvancedDeploymentForm(props: IOperatorAdvancedDeploymentFormProps) {
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
    <div className="deployment-form-tabs-data operator-editor">
      <MonacoDiffEditor
        value={props.appValues}
        original={props.oldAppValues}
        height="90vh"
        language="yaml"
        onChange={onChange}
        theme={theme === SupportedThemes.dark ? "vs-dark" : "light"}
        options={{
          renderSideBySide: false,
          automaticLayout: true,
        }}
      />
      {props.children}
    </div>
  );
}

export default OperatorAdvancedDeploymentForm;
