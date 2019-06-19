import * as React from "react";
import AceEditor from "react-ace";

interface IAppValuesProps {
  values: string;
}

const AppValues: React.SFC<IAppValuesProps> = props => {
  return (
    <React.Fragment>
      <h6>Installation Values</h6>
      <AceEditor
        mode="yaml"
        theme="xcode"
        name="values"
        width="100%"
        setOptions={{ showPrintMargin: false }}
        editorProps={{ $blockScrolling: Infinity }}
        value={props.values}
        readOnly={true}
      />
    </React.Fragment>
  );
};

export default AppValues;
