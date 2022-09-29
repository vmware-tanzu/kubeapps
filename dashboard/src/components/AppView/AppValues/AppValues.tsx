// Copyright 2019-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import MonacoEditor from "react-monaco-editor";
import { useSelector } from "react-redux";
import { SupportedThemes } from "shared/Config";
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
      <MonacoEditor
        language="yaml"
        theme={theme === SupportedThemes.dark ? "vs-dark" : "light"}
        className="installation-values"
        height="50vh"
        value={props.values}
        options={{
          automaticLayout: true,
          readOnly: true,
        }}
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
