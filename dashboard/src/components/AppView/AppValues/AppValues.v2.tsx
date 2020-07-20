import * as React from "react";
import AceEditor from "react-ace";

import "./AppValues.css";

interface IAppValuesProps {
  values: string;
}

const AppValues: React.SFC<IAppValuesProps> = props => {
  if (props.values === "") {
    return (
      <React.Fragment>
        <h6>Installation Values</h6>
        <p>The current application was installed without specifying any values</p>
      </React.Fragment>
    );
  }
  return (
    <section aria-labelledby="installation-values">
      <h5 className="section-title" id="installation-values">
        Installation Values
      </h5>
      <AceEditor
        mode="yaml"
        theme="xcode"
        name="values"
        className="installation-values"
        width="100%"
        maxLines={40}
        setOptions={{ showPrintMargin: false }}
        editorProps={{ $blockScrolling: Infinity }}
        value={props.values}
        readOnly={true}
      />
    </section>
  );
};

export default AppValues;
