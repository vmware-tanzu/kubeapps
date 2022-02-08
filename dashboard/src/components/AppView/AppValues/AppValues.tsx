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
import "./AppValues.css";

interface IAppValuesProps {
  values: string;
}

function AppValues(props: IAppValuesProps) {
  const {
    config: { theme },
  } = useSelector((state: IStoreState) => state);

  let values = <p>The current application was installed without specifying any values</p>;
  if (props.values !== "") {
    values = (
      <AceEditor
        mode="yaml"
        theme={theme === "dark" ? "solarized_dark" : "xcode"}
        name="values"
        className="installation-values"
        width="100%"
        maxLines={40}
        setOptions={{ showPrintMargin: false }}
        editorProps={{ $blockScrolling: Infinity }}
        value={props.values}
        readOnly={true}
      />
    );
  }
  return (
    <section aria-labelledby="installation-values">
      <h3 className="section-title" id="installation-values">
        Installation Values
      </h3>
      {values}
    </section>
  );
}

export default AppValues;
