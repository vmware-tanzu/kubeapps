// Copyright 2019-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import MonacoEditor from "react-monaco-editor";
import { useSelector } from "react-redux";
import { SupportedThemes } from "shared/Config";
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
      <MonacoEditor
        language="yaml"
        theme={theme === SupportedThemes.dark ? "vs-dark" : "light"}
        height="90vh"
        onChange={onChange}
        value={props.appValues}
        className="editor"
        options={{
          automaticLayout: true,
        }}
      />
      {props.children}
    </div>
  );
}

export default AdvancedDeploymentForm;
