// Copyright 2019-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { MonacoDiffEditor } from "react-monaco-editor";
import { useSelector } from "react-redux";
import { SupportedThemes } from "shared/Config";
import { IStoreState } from "shared/types";
import "./Differential.css";

export interface IDifferentialProps {
  oldValues: string;
  newValues: string;
  emptyDiffElement: JSX.Element;
}

function Differential(props: IDifferentialProps) {
  const { oldValues, newValues, emptyDiffElement } = props;
  const {
    config: { theme },
  } = useSelector((state: IStoreState) => state);

  return (
    <div className="diff deployment-form-tabs-data">
      {oldValues === newValues ? (
        emptyDiffElement
      ) : (
        <MonacoDiffEditor
          value={newValues}
          original={oldValues}
          className="editor"
          height="90vh"
          language="yaml"
          theme={theme === SupportedThemes.dark ? "vs-dark" : "light"}
          options={{
            automaticLayout: true,
            readOnly: true,
          }}
        />
      )}
    </div>
  );
}

export default Differential;
